package sqld

import (
	"context"
	"fmt"
	"testing"
)

type TestParams struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (TestParams) TableName() string {
	return "test_params"
}

type TestResult struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (TestResult) TableName() string {
	return "test_results"
}

func TestExecuteRawWithRegistry(t *testing.T) {
	// Register both param and result types
	if err := Register[TestParams](); err != nil {
		t.Fatalf("Failed to register TestParams: %v", err)
	}
	if err := Register[TestResult](); err != nil {
		t.Fatalf("Failed to register TestResult: %v", err)
	}

	// Create a request
	req := ExecuteRawRequest{
		Query: "SELECT id, name FROM test WHERE id = {{id}} AND name = {{name}}",
		Params: map[string]interface{}{
			"id":   1,
			"name": "test",
		},
		SelectFields: []string{"id", "name"},
	}

	// Mock DB - we won't actually execute the query
	var mockDB *MockDB

	// Execute the query - it should use registry
	_, err := ExecuteRaw[TestParams, TestResult](context.Background(), mockDB, req)
	if err != nil {
		// We expect an error since we're using a nil DB, but it should be a DB error
		// not a metadata or validation error
		if err.Error() != "unsupported database type: *sqld.MockDB" {
			t.Errorf("Expected DB error, got: %v", err)
		}
	}
}

func TestExecuteRawRegistryForParamsAndResult(t *testing.T) {
	// Clear the registry first to ensure clean state
	defaultRegistry = NewRegistry()

	// Register both param and result types
	if err := Register[TestParams](); err != nil {
		t.Fatalf("Failed to register TestParams: %v", err)
	}
	if err := Register[TestResult](); err != nil {
		t.Fatalf("Failed to register TestResult: %v", err)
	}

	// Create multiple requests to verify metadata reuse
	requests := []ExecuteRawRequest{
		{
			Query: "SELECT id, name FROM test WHERE id = {{id}}",
			Params: map[string]interface{}{
				"id": 1,
			},
			SelectFields: []string{"id", "name"},
		},
		{
			Query: "SELECT id, name FROM test WHERE name = {{name}}",
			Params: map[string]interface{}{
				"name": "test",
			},
			SelectFields: []string{"id", "name"},
		},
	}

	// Mock DB - we won't actually execute the query
	var mockDB *MockDB

	// Execute the queries multiple times to verify metadata reuse
	for i, req := range requests {
		t.Run(fmt.Sprintf("Request_%d", i), func(t *testing.T) {
			_, err := ExecuteRaw[TestParams, TestResult](context.Background(), mockDB, req)
			if err == nil {
				t.Error("Expected error due to mock DB, got nil")
			}
			
			// We expect a specific error since we're using a nil DB
			if err.Error() != "unsupported database type: *sqld.MockDB" {
				t.Errorf("Expected DB error, got: %v", err)
			}
		})
	}

	// Verify that metadata exists in registry for both types
	if _, err := defaultRegistry.GetModelMetadata(TestParams{}); err != nil {
		t.Errorf("TestParams metadata not found in registry: %v", err)
	}
	if _, err := defaultRegistry.GetModelMetadata(TestResult{}); err != nil {
		t.Errorf("TestResult metadata not found in registry: %v", err)
	}
}

type MockDB struct{}
