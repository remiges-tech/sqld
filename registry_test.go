package sqld

import (
	"database/sql"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModel is a simple model for testing
type TestModel struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CustomInt CustomInt `json:"custom_int" db:"custom_int"`
}

func (TestModel) TableName() string {
	return "test_models"
}

// CustomInt is a custom type for testing scanner registration
type CustomInt int

// CustomScanner is a scanner for CustomInt
type CustomScanner struct {
	value CustomInt
	valid bool
}

func (s *CustomScanner) Scan(src interface{}) error {
	if src == nil {
		s.valid = false
		return nil
	}
	if v, ok := src.(int64); ok {
		s.value = CustomInt(v)
		s.valid = true
		return nil
	}
	return nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.models)
	assert.NotNil(t, registry.scanners)
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	model := TestModel{}

	// Test registration
	err := registry.Register(model)
	assert.NoError(t, err)

	// Verify metadata was stored correctly
	metadata, err := registry.GetModelMetadata(model)
	assert.NoError(t, err)
	assert.Equal(t, "test_models", metadata.TableName)

	// Check field mappings
	expectedFields := map[string]Field{
		"id": {
			Name:     "id",
			JSONName: "id",
			Type:     reflect.TypeOf(int64(0)),
		},
		"name": {
			Name:     "name",
			JSONName: "name",
			Type:     reflect.TypeOf(""),
		},
		"is_active": {
			Name:     "is_active",
			JSONName: "is_active",
			Type:     reflect.TypeOf(false),
		},
		"custom_int": {
			Name:     "custom_int",
			JSONName: "custom_int",
			Type:     reflect.TypeOf(CustomInt(0)),
		},
	}

	assert.Equal(t, len(expectedFields), len(metadata.Fields))
	for name, expectedField := range expectedFields {
		actualField, ok := metadata.Fields[name]
		assert.True(t, ok)
		assert.Equal(t, expectedField.Name, actualField.Name)
		assert.Equal(t, expectedField.JSONName, actualField.JSONName)
		assert.Equal(t, expectedField.Type, actualField.Type)
	}
}

func TestRegistry_RegisterScanner(t *testing.T) {
	registry := NewRegistry()
	customIntType := reflect.TypeOf(CustomInt(0))

	// Test scanner registration
	registry.RegisterScanner(customIntType, func() sql.Scanner {
		return &CustomScanner{}
	})

	// Verify scanner was stored
	factory, ok := registry.GetScanner(customIntType)
	assert.True(t, ok)
	assert.NotNil(t, factory)

	// Test scanner creation
	scanner := factory()
	assert.NotNil(t, scanner)
	assert.IsType(t, &CustomScanner{}, scanner)
}

func TestRegistry_GetModelMetadata_NotFound(t *testing.T) {
	registry := NewRegistry()
	model := TestModel{}

	// Try to get metadata for unregistered model
	metadata, err := registry.GetModelMetadata(model)
	assert.Error(t, err)
	assert.Equal(t, ModelMetadata{}, metadata)
	assert.Contains(t, err.Error(), "not registered")
}

func TestRegistry_GetScanner_NotFound(t *testing.T) {
	registry := NewRegistry()
	customIntType := reflect.TypeOf(CustomInt(0))

	// Try to get unregistered scanner
	factory, ok := registry.GetScanner(customIntType)
	assert.False(t, ok)
	assert.Nil(t, factory)
}

func TestDefaultRegistry(t *testing.T) {
	// Test using default registry functions
	model := TestModel{}
	customIntType := reflect.TypeOf(CustomInt(0))

	// Test Register
	err := Register(model)
	assert.NoError(t, err)

	// Test RegisterScanner
	RegisterScanner(customIntType, func() sql.Scanner {
		return &CustomScanner{}
	})

	// Test getModelMetadata
	metadata, err := getModelMetadata(model)
	assert.NoError(t, err)
	assert.Equal(t, "test_models", metadata.TableName)

	// Verify scanner was registered in default registry
	factory, ok := defaultRegistry.GetScanner(customIntType)
	assert.True(t, ok)
	assert.NotNil(t, factory)
}

func TestRegistry_Concurrency(t *testing.T) {
	registry := NewRegistry()
	model := TestModel{}
	customIntType := reflect.TypeOf(CustomInt(0))

	// Test concurrent registrations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			err := registry.Register(model)
			assert.NoError(t, err)
			registry.RegisterScanner(customIntType, func() sql.Scanner {
				return &CustomScanner{}
			})
			_, err = registry.GetModelMetadata(model)
			assert.NoError(t, err)
			_, _ = registry.GetScanner(customIntType)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

type TestModel2 struct {
	ID        int    `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

func (t TestModel2) TableName() string {
	return "test_models"
}

func TestRegistry_Register2(t *testing.T) {
	registry := NewRegistry()
	model := TestModel2{}

	err := registry.Register(model)
	assert.NoError(t, err)

	metadata, err := registry.GetModelMetadata(model)
	assert.NoError(t, err)
	assert.Equal(t, "test_models", metadata.TableName)

	// Check fields
	assert.Len(t, metadata.Fields, 3)

	// Check ID field
	idField, ok := metadata.Fields["id"]
	assert.True(t, ok)
	assert.Equal(t, "id", idField.Name)
	assert.Equal(t, "id", idField.JSONName)
	assert.Equal(t, reflect.TypeOf(0), idField.Type)

	// Check Name field
	nameField, ok := metadata.Fields["name"]
	assert.True(t, ok)
	assert.Equal(t, "name", nameField.Name)
	assert.Equal(t, "name", nameField.JSONName)
	assert.Equal(t, reflect.TypeOf(""), nameField.Type)
}

func TestRegistry_GetModelMetadata_Unregistered2(t *testing.T) {
	registry := NewRegistry()
	model := TestModel2{}

	_, err := registry.GetModelMetadata(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model TestModel2 not registered")
}

type CustomScanner2 struct {
	sql.NullString
}

func TestRegistry_RegisterScanner2(t *testing.T) {
	registry := NewRegistry()
	scannerType := reflect.TypeOf("")
	factory := func() sql.Scanner { return &CustomScanner2{} }

	registry.RegisterScanner(scannerType, factory)

	// Verify scanner is registered
	gotFactory, ok := registry.GetScanner(scannerType)
	assert.True(t, ok)
	assert.NotNil(t, gotFactory)

	// Verify scanner factory works
	scanner := gotFactory()
	assert.IsType(t, &CustomScanner2{}, scanner)
}

func TestRegistry_GetScanner_Unregistered2(t *testing.T) {
	registry := NewRegistry()
	scannerType := reflect.TypeOf("")

	factory, ok := registry.GetScanner(scannerType)
	assert.False(t, ok)
	assert.Nil(t, factory)
}

func TestRegistry_Concurrency2(t *testing.T) {
	registry := NewRegistry()
	model := TestModel2{}
	scannerType := reflect.TypeOf("")
	factory := func() sql.Scanner { return &CustomScanner2{} }

	var wg sync.WaitGroup
	workers := 10

	// Test concurrent model registration
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			err := registry.Register(model)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	// Test concurrent scanner registration
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			registry.RegisterScanner(scannerType, factory)
		}()
	}
	wg.Wait()

	// Test concurrent metadata retrieval
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			metadata, err := registry.GetModelMetadata(model)
			assert.NoError(t, err)
			assert.Equal(t, "test_models", metadata.TableName)
		}()
	}
	wg.Wait()

	// Test concurrent scanner retrieval
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			factory, ok := registry.GetScanner(scannerType)
			assert.True(t, ok)
			assert.NotNil(t, factory)
		}()
	}
	wg.Wait()
}
