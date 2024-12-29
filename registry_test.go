package sqld

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

type RegistryTestModel struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (RegistryTestModel) TableName() string {
	return "registry_test_models"
}

type RegistryTestScanner struct {
	value interface{}
}

func (s *RegistryTestScanner) Scan(value interface{}) error {
	s.value = value
	return nil
}

func (s *RegistryTestScanner) ScanRow(row pgx.Row) (*RegistryTestModel, error) {
	var model RegistryTestModel
	err := row.Scan(&model.ID, &model.Name)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *RegistryTestScanner) ScanRows(rows pgx.Rows) ([]*RegistryTestModel, error) {
	var models []*RegistryTestModel
	for rows.Next() {
		model, err := s.ScanRow(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

func TestRegisterModel(t *testing.T) {
	// Clear the registry before test
	defaultRegistry = NewRegistry()

	// Test registering a model
	err := Register[RegistryTestModel]()
	assert.NoError(t, err)

	// Test registering the same model again (should fail)
	err = Register[RegistryTestModel]()
	assert.Error(t, err)

	// Test getting model metadata
	var model RegistryTestModel
	metadata, err := getModelMetadata(model)
	assert.NoError(t, err)
	assert.Equal(t, "registry_test_models", metadata.TableName)
	assert.Contains(t, metadata.Fields, "id")
	assert.Contains(t, metadata.Fields, "name")
}

func TestRegisterScanner(t *testing.T) {
	// Clear the registry before test
	defaultRegistry = NewRegistry()

	// Test registering a scanner
	scanner := &RegistryTestScanner{}
	defaultRegistry.RegisterScanner(reflect.TypeOf(RegistryTestModel{}), func() sql.Scanner { return scanner })

	// Test registering the same scanner again (should fail)
	defaultRegistry.RegisterScanner(reflect.TypeOf(RegistryTestModel{}), func() sql.Scanner { return scanner })

	// Test getting scanner
	scannerFactory, ok := defaultRegistry.GetScanner(reflect.TypeOf(RegistryTestModel{}))
	assert.True(t, ok)
	assert.NotNil(t, scannerFactory)
}

func TestConcurrentRegistration(t *testing.T) {
	// Clear the registry before test
	defaultRegistry = NewRegistry()

	// Test concurrent model registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			err := Register[RegistryTestModel]()
			if err != nil {
				// Ignore errors as we expect some registrations to fail
				// due to concurrent registration attempts
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify that the model was registered correctly
	var model RegistryTestModel
	metadata, err := getModelMetadata(model)
	assert.NoError(t, err)
	assert.Equal(t, "registry_test_models", metadata.TableName)

	// Test concurrent scanner registration
	scanner := &RegistryTestScanner{}
	for i := 0; i < 10; i++ {
		go func() {
			defaultRegistry.RegisterScanner(reflect.TypeOf(RegistryTestModel{}), func() sql.Scanner { return scanner })
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify that the scanner was registered correctly
	scannerFactory, ok := defaultRegistry.GetScanner(reflect.TypeOf(RegistryTestModel{}))
	assert.True(t, ok)
	assert.NotNil(t, scannerFactory)
}

func TestScannerFunctionality(t *testing.T) {
	// Clear the registry before test
	defaultRegistry = NewRegistry()

	// Register the scanner
	scanner := &RegistryTestScanner{}
	defaultRegistry.RegisterScanner(reflect.TypeOf(RegistryTestModel{}), func() sql.Scanner { return scanner })

	// Get the scanner
	scannerFactory, ok := defaultRegistry.GetScanner(reflect.TypeOf(RegistryTestModel{}))
	assert.True(t, ok)
	assert.NotNil(t, scannerFactory)

	// Test scanning a single row
	row := &mockRow{
		values: []interface{}{1, "Test Model"},
	}
	model, err := scanner.ScanRow(row)
	assert.NoError(t, err)
	assert.Equal(t, 1, model.ID)
	assert.Equal(t, "Test Model", model.Name)

	// Test scanning multiple rows
	rows := &mockRows{
		data: [][]interface{}{
			{1, "Model 1"},
			{2, "Model 2"},
			{3, "Model 3"},
		},
	}
	models, err := scanner.ScanRows(rows)
	assert.NoError(t, err)
	assert.Len(t, models, 3)
	assert.Equal(t, "Model 1", models[0].Name)
	assert.Equal(t, "Model 2", models[1].Name)
	assert.Equal(t, "Model 3", models[2].Name)
}

// Mock implementations for testing

type mockRow struct {
	values []interface{}
}

func (m *mockRow) Scan(dest ...interface{}) error {
	for i, d := range dest {
		switch v := d.(type) {
		case *int:
			*v = m.values[i].(int)
		case *string:
			*v = m.values[i].(string)
		}
	}
	return nil
}

type mockRows struct {
	data     [][]interface{}
	position int
}

func (m *mockRows) Next() bool {
	m.position++
	return m.position <= len(m.data)
}

func (m *mockRows) Scan(dest ...interface{}) error {
	row := &mockRow{values: m.data[m.position-1]}
	return row.Scan(dest...)
}

func (m *mockRows) Err() error {
	return nil
}

func (m *mockRows) Close() {
	// No-op for mock
}

func (m *mockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription {
	return []pgconn.FieldDescription{}
}

func (m *mockRows) Values() ([]interface{}, error) {
	return m.data[m.position-1], nil
}

func (m *mockRows) RawValues() [][]byte {
	return nil
}

func (m *mockRows) Conn() *pgx.Conn {
	return nil
}
