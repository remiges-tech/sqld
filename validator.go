package sqld

import (
	"fmt"
	"log"
	"reflect"
)

type Validator interface {
	ValidateQuery(req QueryRequest, metadata ModelMetadata) error
}

type BasicValidator struct{}

func isValidOperator(op Operator) bool {
	switch op {
	case OpEqual, OpNotEqual, OpGreaterThan, OpLessThan,
		OpGreaterThanOrEqual, OpLessThanOrEqual, OpLike,
		OpILike, OpIn, OpNotIn, OpIsNull, OpIsNotNull:
		return true
	}
	return false
}

func (v BasicValidator) ValidateQuery(req QueryRequest, metadata ModelMetadata) error {
	// Validate select fields
	if len(req.Select) == 0 {
		return fmt.Errorf("select fields cannot be empty")
	}

	for _, field := range req.Select {
		if _, ok := metadata.Fields[field]; !ok {
			return fmt.Errorf("invalid field in select: %s", field)
		}
	}

	// Validate where conditions
	for _, cond := range req.Where {
		// Validate field exists
		field, ok := metadata.Fields[cond.Field]
		if !ok {
			return fmt.Errorf("invalid field in where clause: %s", cond.Field)
		}

		// Validate operator
		if !isValidOperator(cond.Operator) {
			return fmt.Errorf("unsupported operator: %s", cond.Operator)
		}

		// Special validation for null operators
		if cond.Operator == OpIsNull || cond.Operator == OpIsNotNull {
			if cond.Value != nil {
				return fmt.Errorf("value must be nil for IS NULL/IS NOT NULL operators")
			}
			continue
		}

		// Validate value type matches field type for non-null operators
		if cond.Value != nil {
			valueType := reflect.TypeOf(cond.Value)
			log.Printf("Validating field %s: value type %v, field type %v", cond.Field, valueType, field.Type)
			
			// Special case for IN/NOT IN which expect slices
			if cond.Operator == OpIn || cond.Operator == OpNotIn {
				if valueType.Kind() != reflect.Slice {
					return fmt.Errorf("value for IN/NOT IN must be a slice")
				}
				// Check element type matches field type
				if valueType.Elem() != field.Type {
					return fmt.Errorf("invalid type for field %s: expected %v, got %v", 
						cond.Field, field.Type, valueType.Elem())
				}
			} else {
				// Check if type matches directly or if we have a converter for it
				if valueType != field.Type {
					// Check if we have a converter registered for this type
					if _, ok := defaultRegistry.GetConverter(field.Type); !ok {
						log.Printf("No converter found for type %v", field.Type)
						return fmt.Errorf("invalid type for field %s: expected %v, got %v",
							cond.Field, field.Type, valueType)
					}
					log.Printf("Found converter for type %v", field.Type)
				}
			}
		}
	}

	// Validate order by fields
	for _, orderBy := range req.OrderBy {
		if _, ok := metadata.Fields[orderBy.Field]; !ok {
			return fmt.Errorf("invalid field in order by clause: %s", orderBy.Field)
		}
	}

	// Validate limit and offset
	if req.Limit != nil && *req.Limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	if req.Offset != nil && *req.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}
