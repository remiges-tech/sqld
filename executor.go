package sqld

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier interface abstracts database operations
type Querier interface {
	// QueryContext is provided by sql.DB
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// PgxQuerier interface for pgx operations
type PgxQuerier interface {
	// Query is provided by pgx.Conn
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

// Execute runs the query and returns properly scanned results.
func Execute[T Model](ctx context.Context, db interface{}, req QueryRequest) (QueryResponse[T], error) {
	// Get model metadata using type parameter T
	var model T
	metadata, err := getModelMetadata(model)
	if err != nil {
		return QueryResponse[T]{}, fmt.Errorf("failed to get model metadata: %w", err)
	}

	// Call the validator before building and executing the query.
	validator := BasicValidator{}
	if err := validator.ValidateQuery(req, metadata); err != nil {
		return QueryResponse[T]{}, fmt.Errorf("failed to validate query: %w", err)
	}

	// Handle pagination if requested
	var paginationResp *PaginationResponse
	if req.Pagination != nil || req.Limit != nil || req.Offset != nil {
		if req.Pagination != nil {
			// If req.Pagination is provided, it will override any previously set limit/offset values.
			// This ensures that page-based pagination always takes precedence over direct limit/offset parameters.

			// Validate and normalize pagination parameters
			req.Pagination = ValidatePagination(req.Pagination)

			// Set limit and offset based on pagination
			limit := req.Pagination.PageSize
			offset := CalculateOffset(req.Pagination.Page, req.Pagination.PageSize)
			req.Limit = &limit
			req.Offset = &offset
		}
	}

	// Build query using the generic buildQuery
	builder, err := buildQuery[T](req)
	if err != nil {
		return QueryResponse[T]{}, fmt.Errorf("failed to build query: %w", err)
	}

	// If pagination is requested or limit/offset is set, we need to get total count
	if req.Pagination != nil || req.Limit != nil || req.Offset != nil {
		// Create a new count query builder with the same conditions
		// Use Postgres placeholder format ($1, $2, etc)
		builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		countBuilder := builder.Select("COUNT(*)").From(model.TableName())

		// Apply the same where conditions if they exist
		for _, cond := range req.Where {
			field, ok := metadata.Fields[cond.Field]
			if !ok {
				return QueryResponse[T]{}, fmt.Errorf("invalid field in where clause: %s", cond.Field)
			}

			switch cond.Operator {
			case OpEqual:
				countBuilder = countBuilder.Where(squirrel.Eq{field.Name: cond.Value})
			case OpNotEqual:
				countBuilder = countBuilder.Where(squirrel.NotEq{field.Name: cond.Value})
			case OpGreaterThan:
				countBuilder = countBuilder.Where(squirrel.Gt{field.Name: cond.Value})
			case OpLessThan:
				countBuilder = countBuilder.Where(squirrel.Lt{field.Name: cond.Value})
			case OpGreaterThanOrEqual:
				countBuilder = countBuilder.Where(squirrel.GtOrEq{field.Name: cond.Value})
			case OpLessThanOrEqual:
				countBuilder = countBuilder.Where(squirrel.LtOrEq{field.Name: cond.Value})
			case OpLike, OpILike:
				op := string(cond.Operator)
				countBuilder = countBuilder.Where(squirrel.Expr(field.Name+" "+op+" ?", cond.Value))
			case OpIn:
				countBuilder = countBuilder.Where(squirrel.Eq{field.Name: cond.Value})
			case OpNotIn:
				countBuilder = countBuilder.Where(squirrel.NotEq{field.Name: cond.Value})
			case OpIsNull:
				countBuilder = countBuilder.Where(squirrel.Eq{field.Name: nil})
			case OpIsNotNull:
				countBuilder = countBuilder.Where(squirrel.NotEq{field.Name: nil})
			default:
				return QueryResponse[T]{}, fmt.Errorf("unsupported operator: %s", cond.Operator)
			}
		}

		countQuery, countArgs, err := countBuilder.ToSql()
		if err != nil {
			return QueryResponse[T]{}, fmt.Errorf("failed to generate count sql: %w", err)
		}

		// Log the query for debugging
		log.Printf("Count Query: %s with args: %v", countQuery, countArgs)

		var totalItems int
		switch db := db.(type) {
		case *sql.DB:
			err = sqlscan.Get(ctx, db, &totalItems, countQuery, countArgs...)
		case *pgx.Conn:
			err = pgxscan.Get(ctx, db, &totalItems, countQuery, countArgs...)
		case *pgxpool.Pool:
			err = pgxscan.Get(ctx, db, &totalItems, countQuery, countArgs...)
		default:
			return QueryResponse[T]{}, fmt.Errorf("unsupported database type: %T", db)
		}

		if err != nil {
			return QueryResponse[T]{}, fmt.Errorf("failed to get total count: %w", err)
		}

		if req.Pagination != nil {
			paginationResp = CalculatePagination(totalItems, req.Pagination.PageSize, req.Pagination.Page)
		} else if req.Limit != nil {
			pageSize := *req.Limit
			currentPage := 1
			if req.Offset != nil {
				currentPage = (*req.Offset / pageSize) + 1
			}
			paginationResp = CalculatePagination(totalItems, pageSize, currentPage)
		}
	}

	// Get the query and args for the main query
	query, args, err := builder.ToSql()
	if err != nil {
		return QueryResponse[T]{}, fmt.Errorf("failed to generate sql: %w", err)
	}

	// Use appropriate scanner based on the database type
	var results []map[string]interface{}
	switch db := db.(type) {
	case *sql.DB:
		err = sqlscan.Select(ctx, db, &results, query, args...)
	case *pgx.Conn:
		err = pgxscan.Select(ctx, db, &results, query, args...)
	case *pgxpool.Pool:
		err = pgxscan.Select(ctx, db, &results, query, args...)
	default:
		return QueryResponse[T]{}, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		return QueryResponse[T]{}, fmt.Errorf("failed to execute query: %w", err)
	}

	// Convert the results to our QueryResult type
	queryResults := make([]QueryResult, len(results))
	for i, result := range results {
		queryResult := make(QueryResult)
		for _, field := range req.Select {
			fieldMeta := metadata.Fields[field]
			if val, ok := result[fieldMeta.Name]; ok { // Use database column name
				queryResult[field] = val // Use JSON name from request
			}
		}
		queryResults[i] = queryResult
	}

	return QueryResponse[T]{
		Data:       queryResults,
		Pagination: paginationResp,
	}, nil
}

// TODO: Add connection pooling configuration
// TODO: Add caching layer for frequently used queries
// TODO: Add query execution timeout handling
// TODO: Add detailed error context and error codes
