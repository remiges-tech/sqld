package sqld

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// TODO: Add input validation for maximum number of selected columns
// TODO: Add SQL injection protection checks for WHERE values
// TODO: Add validation for LIMIT/OFFSET values
// TODO: Add query timeout configuration
// TODO: Add metrics/logging for query performance monitoring

// buildQuery creates a type-safe query for the given model.
// To achieve safety, it does the following:
// - Validates the select fields against the model metadata
// - Converts JSON field names to actual field names for SELECT
// - Converts JSON field names to actual field names for WHERE
// - Validates operator compatibility with field types
// - Other validations -- TODO
func buildQuery[T Model](req QueryRequest) (squirrel.SelectBuilder, error) {
	var model T
	metadata, err := getModelMetadata(model)
	if err != nil {
		return squirrel.SelectBuilder{}, fmt.Errorf("failed to get model metadata: %w", err)
	}

	// Call the validator before building the query
	validator := BasicValidator{}
	if err := validator.ValidateQuery(req, metadata); err != nil {
		return squirrel.SelectBuilder{}, fmt.Errorf("validation failed: %w", err)
	}

	// Use Postgres placeholder format ($1, $2, etc)
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Convert JSON field names to actual field names for SELECT
	selectFields := make([]string, len(req.Select))
	for i, jsonName := range req.Select {
		field := metadata.Fields[jsonName] // Safe to use directly as validation passed
		selectFields[i] = field.Name
	}

	// Build query with converted field names
	query := builder.Select(selectFields...).
		From(model.TableName())

	// Build WHERE conditions
	if len(req.Where) > 0 {
		for _, cond := range req.Where {
			field := metadata.Fields[cond.Field] // Safe to use directly as validation passed

			switch cond.Operator {
			case OpEqual:
				query = query.Where(squirrel.Eq{field.Name: cond.Value})
			case OpNotEqual:
				query = query.Where(squirrel.NotEq{field.Name: cond.Value})
			case OpGreaterThan:
				query = query.Where(squirrel.Gt{field.Name: cond.Value})
			case OpLessThan:
				query = query.Where(squirrel.Lt{field.Name: cond.Value})
			case OpGreaterThanOrEqual:
				query = query.Where(squirrel.GtOrEq{field.Name: cond.Value})
			case OpLessThanOrEqual:
				query = query.Where(squirrel.LtOrEq{field.Name: cond.Value})
			case OpLike, OpILike:
				op := string(cond.Operator)
				query = query.Where(squirrel.Expr(field.Name+" "+op+" ?", cond.Value))
			case OpIn:
				query = query.Where(squirrel.Eq{field.Name: cond.Value})
			case OpNotIn:
				query = query.Where(squirrel.NotEq{field.Name: cond.Value})
			case OpIsNull:
				query = query.Where(squirrel.Eq{field.Name: nil})
			case OpIsNotNull:
				query = query.Where(squirrel.NotEq{field.Name: nil})
			default:
				return squirrel.SelectBuilder{}, fmt.Errorf("unsupported operator: %s", cond.Operator)
			}
		}
	}

	// Handle ORDER BY clauses
	if len(req.OrderBy) > 0 {
		for _, orderBy := range req.OrderBy {
			field, ok := metadata.Fields[orderBy.Field]
			if !ok {
				return squirrel.SelectBuilder{}, fmt.Errorf("invalid field in order by clause: %s", orderBy.Field)
			}
			if orderBy.Desc {
				query = query.OrderBy(field.Name + " DESC")
			} else {
				query = query.OrderBy(field.Name + " ASC")
			}
		}
	}

	// Handle LIMIT and OFFSET
	if req.Limit != nil {
		if *req.Limit < 0 {
			return squirrel.SelectBuilder{}, fmt.Errorf("limit must be non-negative")
		}
		query = query.Limit(uint64(*req.Limit))
	}

	if req.Offset != nil {
		if *req.Offset < 0 {
			return squirrel.SelectBuilder{}, fmt.Errorf("offset must be non-negative")
		}
		query = query.Offset(uint64(*req.Offset))
	}

	// TODO: Add support for GROUP BY

	return query, nil
}
