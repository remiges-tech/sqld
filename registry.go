package sqld

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
)

// Registry is a type-safe registry for model metadata and scanners
type Registry struct {
	models     map[reflect.Type]ModelMetadata
	scanners   map[reflect.Type]func() sql.Scanner
	converters map[reflect.Type]TypeConverter
	mu         sync.RWMutex
}

// NewRegistry returns a new instance of the registry
func NewRegistry() *Registry {
	return &Registry{
		models:     make(map[reflect.Type]ModelMetadata),
		scanners:   make(map[reflect.Type]func() sql.Scanner),
		converters: make(map[reflect.Type]TypeConverter),
	}
}

// defaultRegistry is the default global registry instance
var defaultRegistry = NewRegistry()

// Register adds a model's metadata to the registry
func Register[T Model]() error {
	var model T
	return defaultRegistry.Register(model)
}

// RegisterScanner registers a function that creates scanners for a specific type
func RegisterScanner(t reflect.Type, scannerFactory func() sql.Scanner) {
	defaultRegistry.RegisterScanner(t, scannerFactory)
}

// RegisterConverter registers a function that converts values for a specific type
func RegisterConverter(t reflect.Type, converter TypeConverter) {
	defaultRegistry.RegisterConverter(t, converter)
}

// getModelMetadata retrieves metadata for a model type
func getModelMetadata(model Model) (ModelMetadata, error) {
	return defaultRegistry.GetModelMetadata(model)
}

// Register adds a model's metadata to the registry
func (r *Registry) Register(model Model) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := reflect.TypeOf(model)
	// Check if model is already registered
	if _, exists := r.models[t]; exists {
		return fmt.Errorf("model %s already registered", t.Name())
	}

	metadata := ModelMetadata{
		TableName: model.TableName(),
		Fields:    make(map[string]Field),
	}

	// Reflect over the struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get database column name from db tag
		dbName := field.Tag.Get("db")
		if dbName == "" {
			return fmt.Errorf("field %q missing required db tag", field.Name)
		}

		// Get JSON name from json tag
		jsonName := field.Tag.Get("json")
		if jsonName == "" {
			return fmt.Errorf("field %q missing required json tag", field.Name)
		}

		// For SQLC-generated types, we need to handle the field type differently
		fieldType := field.Type
		if field.Type.PkgPath() == "github.com/jackc/pgx/v5/pgtype" {
			// Register a type converter for this type if we don't have one yet
			if _, ok := r.converters[fieldType]; !ok {
				log.Printf("Registering auto-converter for type %v", fieldType)
				r.converters[fieldType] = func(v interface{}) (interface{}, error) {
					// Handle basic Go types to pgtype conversions
					switch ft := field.Type.Name(); ft {
					case "Bool":
						if b, ok := v.(bool); ok {
							return &pgtype.Bool{Bool: b, Valid: true}, nil
						}
					case "Int8":
						if i, ok := v.(int64); ok {
							return &pgtype.Int8{Int64: i, Valid: true}, nil
						}
					case "Text":
						if s, ok := v.(string); ok {
							return &pgtype.Text{String: s, Valid: true}, nil
						}
					case "Numeric":
						if _, ok := v.(float64); ok {
							return &pgtype.Numeric{Valid: true}, nil // TODO: proper numeric conversion
						}
					}
					return nil, fmt.Errorf("unsupported conversion to %v: %T", fieldType, v)
				}
			}
		}

		metadata.Fields[jsonName] = Field{
			Name:     dbName,    // Store DB column name
			JSONName: jsonName,  // Store JSON field name
			Type:     fieldType, // Store the field type
		}
	}

	r.models[t] = metadata
	return nil
}

// RegisterScanner registers a function that creates scanners for a specific type
func (r *Registry) RegisterScanner(t reflect.Type, scannerFactory func() sql.Scanner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scanners[t] = scannerFactory
}

// RegisterConverter registers a function that converts values for a specific type
func (r *Registry) RegisterConverter(t reflect.Type, converter TypeConverter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.Printf("Registering converter for type %v", t)
	r.converters[t] = converter
}

// GetModelMetadata retrieves metadata for a model type
func (r *Registry) GetModelMetadata(model Model) (ModelMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t := reflect.TypeOf(model)
	metadata, ok := r.models[reflect.TypeOf(model)]
	if !ok {
		return ModelMetadata{}, fmt.Errorf("model %s not registered", t.Name())
	}
	return metadata, nil
}

// GetScanner returns a scanner factory for the given type, if registered
func (r *Registry) GetScanner(t reflect.Type) (func() sql.Scanner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.scanners[t]
	return factory, ok
}

// GetConverter returns a converter for the given type, if registered
func (r *Registry) GetConverter(t reflect.Type) (TypeConverter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	converter, ok := r.converters[t]
	return converter, ok
}

// TypeConverter converts a value from one type to another
type TypeConverter func(interface{}) (interface{}, error)
