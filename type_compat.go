package sqld

import (
	"reflect"
	"time"
)

// AreTypesCompatible checks if valueType is suitable for fieldType.
// Returns true if they are compatible, false otherwise.
func AreTypesCompatible(fieldType, valueType reflect.Type) bool {
	// nil check or interface check
	if valueType == nil || fieldType == nil {
		return false
	}

	// If the field type is an interface{} with no methods, it's basically "any"
	if fieldType.Kind() == reflect.Interface && fieldType.NumMethod() == 0 {
		return true
	}

	// Unwrap pointers
	for fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}

	// Exact match
	if fieldType == valueType {
		return true
	}

	// Check numeric
	if IsNumericType(fieldType) && IsNumericType(valueType) {
		return true
	}

	// Check strings
	if fieldType.Kind() == reflect.String && valueType.Kind() == reflect.String {
		return true
	}

	// Check time.Time
	if IsTimeType(fieldType) && IsTimeType(valueType) {
		return true
	}

	// Check bool
	if fieldType.Kind() == reflect.Bool && valueType.Kind() == reflect.Bool {
		return true
	}

	return false
}

// IsNumericType returns true if the reflect.Type is one of the integer or float kinds.
func IsNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// IsTimeType returns true if the reflect.Type is time.Time or a known date/time type.
func IsTimeType(t reflect.Type) bool {
	return t == reflect.TypeOf(time.Time{})
}
