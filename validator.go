package sqld

import (
	"fmt"
	"reflect"
	"time"
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

// areTypesCompatible checks if two types are compatible for validation
func areTypesCompatible(fieldType, valueType reflect.Type) bool {
	// If they're literally the same, all good
	if fieldType == valueType {
		return true
	}

	// Handle string types
	if fieldType.Kind() == reflect.String && valueType.Kind() == reflect.String {
		return true
	}

	// Handle numeric types
	if isNumericType(fieldType) && isNumericType(valueType) {
		return true
	}

	// Handle time types
	if isTimeType(fieldType) && isTimeType(valueType) {
		return true
	}

	// Handle bool types
	if fieldType.Kind() == reflect.Bool && valueType.Kind() == reflect.Bool {
		return true
	}

	// Handle slices (for IN/NOT IN operators)
	if valueType.Kind() == reflect.Slice {
		// Get the element type of the slice
		elemType := valueType.Elem()
		
		// If the element type is interface{}, we need to check the actual values
		if elemType.Kind() == reflect.Interface {
			return true // We'll check actual values during validation
		}

		// Otherwise check if the element type is compatible
		return areTypesCompatible(fieldType, elemType)
	}

	return false
}

// isNumericType checks if a type is numeric (int*, uint*, float*)
func isNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// isTimeType checks if a type represents time
func isTimeType(t reflect.Type) bool {
	return t == reflect.TypeOf(time.Time{}) ||
		(t.Kind() == reflect.String && (t.Name() == "Time" || t.Name() == "Date"))
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

			// Special case for IN/NOT IN which expect slices
			if cond.Operator == OpIn || cond.Operator == OpNotIn {
				if valueType.Kind() != reflect.Slice {
					return fmt.Errorf("value for IN/NOT IN must be a slice")
				}

				// For IN/NOT IN with []interface{}, check each element's actual type
				if valueType.Elem().Kind() == reflect.Interface {
					sliceValue := reflect.ValueOf(cond.Value)
					for i := 0; i < sliceValue.Len(); i++ {
						elemValue := sliceValue.Index(i).Interface()
						elemType := reflect.TypeOf(elemValue)
						if !areTypesCompatible(field.NormalizedType, elemType) {
							return fmt.Errorf(
								"invalid type for field %s at index %d: expected %v, got %v",
								cond.Field, i, field.NormalizedType, elemType)
						}
					}
				} else {
					// For typed slices, check the element type
					if !areTypesCompatible(field.NormalizedType, valueType.Elem()) {
						return fmt.Errorf("invalid type for field %s: expected %v, got %v",
							cond.Field, field.NormalizedType, valueType.Elem())
					}
				}
			} else if !areTypesCompatible(field.NormalizedType, valueType) {
				return fmt.Errorf("invalid type for field %s: expected %v, got %v",
					cond.Field, field.NormalizedType, valueType)
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
