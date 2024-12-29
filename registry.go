package sqld

import (
	"database/sql"
	"fmt"
	"reflect"
	"sync"
)

// Registry is a type-safe registry for model metadata and scanners
type Registry struct {
	models   map[reflect.Type]ModelMetadata
	scanners map[reflect.Type]func() sql.Scanner
	mu       sync.RWMutex
}

// NewRegistry returns a new instance of the registry
func NewRegistry() *Registry {
	return &Registry{
		models:   make(map[reflect.Type]ModelMetadata),
		scanners: make(map[reflect.Type]func() sql.Scanner),
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

		metadata.Fields[jsonName] = Field{
			Name:     dbName,   // Store DB column name
			JSONName: jsonName, // Store JSON field name
			Type:     field.Type,
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
