package sqld

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

func buildWhereClause(fieldName string, cond Condition) (squirrel.Sqlizer, error) {
	switch cond.Operator {
	case OpEqual:
		return squirrel.Eq{fieldName: cond.Value}, nil
	case OpNotEqual:
		return squirrel.NotEq{fieldName: cond.Value}, nil
	case OpGreaterThan:
		return squirrel.Gt{fieldName: cond.Value}, nil
	case OpLessThan:
		return squirrel.Lt{fieldName: cond.Value}, nil
	case OpGreaterThanOrEqual:
		return squirrel.GtOrEq{fieldName: cond.Value}, nil
	case OpLessThanOrEqual:
		return squirrel.LtOrEq{fieldName: cond.Value}, nil
	case OpLike, OpILike:
		return squirrel.Expr(fieldName+" "+string(cond.Operator)+" ?", cond.Value), nil
	case OpIn:
		return squirrel.Eq{fieldName: cond.Value}, nil
	case OpNotIn:
		return squirrel.NotEq{fieldName: cond.Value}, nil
	case OpIsNull:
		return squirrel.Eq{fieldName: nil}, nil
	case OpIsNotNull:
		return squirrel.NotEq{fieldName: nil}, nil
	case OpAny:
		return squirrel.Expr("? = ANY("+fieldName+")", cond.Value), nil
	default:
		return nil, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}
}

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
// - Other validations -- TODO
func buildQuery[T Model](req QueryRequest) (squirrel.SelectBuilder, error) {
	var model T
	metadata, err := getModelMetadata(model)
	if err != nil {
		return squirrel.SelectBuilder{}, fmt.Errorf("failed to get model metadata: %w", err)
	}

	// Validate select fields
	if len(req.Select) == 0 {
		return squirrel.SelectBuilder{}, fmt.Errorf("select fields cannot be empty")
	}

	// Use Postgres placeholder format ($1, $2, etc)
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Handle special "ALL" value in Select
	var selectFields []string
	if len(req.Select) == 1 && req.Select[0] == SelectAll {
		// When "ALL" is specified, include all fields from the model
		selectFields = make([]string, 0, len(metadata.Fields))
		for _, field := range metadata.Fields {
			selectFields = append(selectFields, field.Name)
		}
	} else {
		// Convert JSON field names to actual field names for SELECT
		selectFields = make([]string, len(req.Select))
		for i, jsonName := range req.Select {
			field, ok := metadata.Fields[jsonName]
			if !ok {
				return squirrel.SelectBuilder{}, fmt.Errorf("invalid field in select: %s", jsonName)
			}
			selectFields[i] = field.Name
		}
	}

	// Build query with converted field names
	query := builder.Select(selectFields...).
		From(model.TableName())

	// Build WHERE conditions
	if len(req.Where) > 0 {
		for _, cond := range req.Where {
			field, ok := metadata.Fields[cond.Field]
			if !ok {
				return squirrel.SelectBuilder{}, fmt.Errorf("invalid field in where clause: %s", cond.Field)
			}

			whereClause, err := buildWhereClause(field.Name, cond)
			if err != nil {
				return squirrel.SelectBuilder{}, err
			}
			query = query.Where(whereClause)
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
