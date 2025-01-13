package sqld

import (
	"fmt"
	"reflect"
)

type Validator interface {
	ValidateQuery(req QueryRequest, metadata ModelMetadata) error
	ValidateUpdateRequest(req UpdateRequest, metadata ModelMetadata) error
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

// isOperatorCompatibleWithType checks if the operator is compatible with the field type
func isOperatorCompatibleWithType(op Operator, fieldType reflect.Type) bool {
	switch op {
	case OpEqual, OpNotEqual, OpIsNull, OpIsNotNull:
		// These operators work with all types
		return true
		
	case OpGreaterThan, OpLessThan, OpGreaterThanOrEqual, OpLessThanOrEqual:
		// These operators only work with ordered types (numbers, strings, dates)
		switch fieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			 reflect.Float32, reflect.Float64, reflect.String:
			return true
		default:
			// Check if it's a time.Time type
			return fieldType.String() == "time.Time"
		}
		
	case OpLike, OpILike:
		// LIKE operators only work with strings
		return fieldType.Kind() == reflect.String
		
	case OpIn, OpNotIn:
		// IN operators work with all types (slice handling is done separately)
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

		// Validate operator compatibility with field type
		if !isOperatorCompatibleWithType(cond.Operator, field.NormalizedType) {
			return fmt.Errorf("operator %s is not compatible with field type %v", cond.Operator, field.NormalizedType)
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
						if !AreTypesCompatible(field.NormalizedType, elemType) {
							return fmt.Errorf(
								"invalid type for field %s at index %d: expected %v, got %v",
								cond.Field, i, field.NormalizedType, elemType)
						}
					}
				} else {
					// For typed slices, check the element type
					if !AreTypesCompatible(field.NormalizedType, valueType.Elem()) {
						return fmt.Errorf("invalid type for field %s: expected %v, got %v",
							cond.Field, field.NormalizedType, valueType.Elem())
					}
				}
			} else if !AreTypesCompatible(field.NormalizedType, valueType) {
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

// ValidateUpdateRequest validates an update request against model metadata
func (v BasicValidator) ValidateUpdateRequest(req UpdateRequest, metadata ModelMetadata) error {
    // Validate SET clause
    if len(req.Set) == 0 {
        return fmt.Errorf("update request must include at least one field to update")
    }

    for jsonName, value := range req.Set {
        field, ok := metadata.Fields[jsonName]
        if !ok {
            return fmt.Errorf("invalid field in update set: %s", jsonName)
        }

        // Type validation for the update value
        if value != nil {
            valueType := reflect.TypeOf(value)
            if !AreTypesCompatible(field.NormalizedType, valueType) {
                return fmt.Errorf("type mismatch for field %s: expected %v, got %v", jsonName, field.NormalizedType, valueType)
            }
        }
    }

    // Validate WHERE conditions
    if len(req.Where) == 0 {
        return fmt.Errorf("update request must include where conditions")
    }

    for _, cond := range req.Where {
        field, ok := metadata.Fields[cond.Field]
        if !ok {
            return fmt.Errorf("invalid field in where clause: %s", cond.Field)
        }

        // Validate operator
        if !isValidOperator(cond.Operator) {
            return fmt.Errorf("unsupported operator: %s", cond.Operator)
        }

        // Validate operator compatibility with field type
        if !isOperatorCompatibleWithType(cond.Operator, field.NormalizedType) {
            return fmt.Errorf("operator %s is not compatible with field type %v", cond.Operator, field.NormalizedType)
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
                        if !AreTypesCompatible(field.NormalizedType, elemType) {
                            return fmt.Errorf(
                                "invalid type for field %s at index %d: expected %v, got %v",
                                cond.Field, i, field.NormalizedType, elemType)
                        }
                    }
                } else {
                    // For typed slices, check the element type
                    if !AreTypesCompatible(field.NormalizedType, valueType.Elem()) {
                        return fmt.Errorf("invalid type for field %s: expected %v, got %v",
                            cond.Field, field.NormalizedType, valueType.Elem())
                    }
                }
            } else if !AreTypesCompatible(field.NormalizedType, valueType) {
                return fmt.Errorf("invalid type for field %s: expected %v, got %v",
                    cond.Field, field.NormalizedType, valueType)
            }
        }
    }

    return nil
}
