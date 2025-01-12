package sqld

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"database/sql"
)

// UpdateRequest represents the structure for building dynamic UPDATE queries.
// It provides type-safe query building with runtime validation against model metadata.
type UpdateRequest struct {
	// Set specifies which fields to update with their new values.
	// Field names must match the JSON tags in your model struct.
	// Each field name is validated against the model's metadata.
	Set map[string]interface{} `json:"set"`

	// Where specifies filter conditions using operators. Each condition consists of
	// a field name (matching JSON field names), an operator, and a value.
	// Required - to prevent accidental updates of all rows.
	Where []Condition `json:"where"`
}

// buildUpdateQuery creates a type-safe UPDATE query for the given model.
// To achieve safety, it does the following:
// - Validates the update fields against the model metadata
// - Converts JSON field names to actual field names for SET and WHERE clauses
// - Validates field types match the model's field types
func buildUpdateQuery[T Model](req UpdateRequest) (squirrel.UpdateBuilder, error) {
	var model T
	metadata, err := getModelMetadata(model)
	if err != nil {
		return squirrel.UpdateBuilder{}, fmt.Errorf("failed to get model metadata: %w", err)
	}

	// Start building the query
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := builder.Update(model.TableName())

	// Process SET clause
	if len(req.Set) == 0 {
		return squirrel.UpdateBuilder{}, fmt.Errorf("update request must include at least one field to update")
	}

	for jsonName, value := range req.Set {
		field, ok := metadata.Fields[jsonName]
		if !ok {
			return squirrel.UpdateBuilder{}, fmt.Errorf("invalid field in update set: %s", jsonName)
		}

		// Type validation
		valueType := reflect.TypeOf(value)
		if value != nil && !AreTypesCompatible(field.Type, valueType) {
			return squirrel.UpdateBuilder{}, fmt.Errorf("type mismatch for field %s: expected %v, got %v", jsonName, field.Type, valueType)
		}

		query = query.Set(field.Name, value)
	}

	// Process WHERE conditions
	if len(req.Where) == 0 {
		return squirrel.UpdateBuilder{}, fmt.Errorf("update request must include where conditions")
	}

	for _, cond := range req.Where {
		field, ok := metadata.Fields[cond.Field]
		if !ok {
			return squirrel.UpdateBuilder{}, fmt.Errorf("invalid field in update where: %s", cond.Field)
		}

		// Type validation for the condition value
		if cond.Value != nil {
			valueType := reflect.TypeOf(cond.Value)
			if !AreTypesCompatible(field.Type, valueType) {
				return squirrel.UpdateBuilder{}, fmt.Errorf("type mismatch for where condition %s: expected %v, got %v", cond.Field, field.Type, valueType)
			}
		}

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
		case OpIn:
			query = query.Where(squirrel.Eq{field.Name: cond.Value})
		case OpNotIn:
			query = query.Where(squirrel.NotEq{field.Name: cond.Value})
		case OpIsNull:
			query = query.Where(squirrel.Eq{field.Name: nil})
		case OpIsNotNull:
			query = query.Where(squirrel.NotEq{field.Name: nil})
		default:
			return squirrel.UpdateBuilder{}, fmt.Errorf("unsupported operator: %s", cond.Operator)
		}
	}

	return query, nil
}

// ExecuteUpdate executes an UPDATE query with the given request and returns the number of rows affected.
// It provides parameter validation and safe query execution.
func ExecuteUpdate[T Model](ctx context.Context, db interface{}, req UpdateRequest) (int64, error) {
	// Build the update query
	builder, err := buildUpdateQuery[T](req)
	if err != nil {
		return 0, fmt.Errorf("failed to build update query: %w", err)
	}

	// Generate SQL
	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to generate update sql: %w", err)
	}

	// Execute based on database type
	switch dbType := db.(type) {
	case *sql.DB:
		res, execErr := dbType.ExecContext(ctx, sqlStr, args...)
		if execErr != nil {
			return 0, fmt.Errorf("failed to execute update: %w", execErr)
		}
		return res.RowsAffected()

	case *pgxpool.Pool:
		ct, execErr := dbType.Exec(ctx, sqlStr, args...)
		if execErr != nil {
			return 0, fmt.Errorf("failed to execute update: %w", execErr)
		}
		return ct.RowsAffected(), nil

	default:
		return 0, fmt.Errorf("unsupported database type: %T", dbType)
	}
}
