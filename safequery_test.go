package sqld

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type QueryParams struct {
	ID     int64  `db:"id" db_param:"id"`
	Status string `db:"status" db_param:"status"`
}

type TestQueryResult struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CustomID is a custom type that needs special scanning
type CustomID struct {
	ID   int
	Type string
}

// Scan implements sql.Scanner for CustomID
func (c *CustomID) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	
	switch v := src.(type) {
	case int64:
		c.ID = int(v)
		c.Type = "numeric"
	case string:
		// Parse string format "id:type"
		parts := strings.Split(v, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid CustomID format: %v", v)
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid ID in CustomID: %v", err)
		}
		c.ID = id
		c.Type = parts[1]
	default:
		return fmt.Errorf("unsupported type for CustomID: %T", src)
	}
	return nil
}

// Value implements driver.Valuer for CustomID
func (c CustomID) Value() (driver.Value, error) {
	return fmt.Sprintf("%d:%s", c.ID, c.Type), nil
}

type TestCustomParams struct {
	ID CustomID `db:"id"`
}

type TestCustomResult struct {
	ID   CustomID `db:"id" json:"custom_id"`
	Name string   `db:"name" json:"name"`
}

// TestExecuteRaw tests the ExecuteRaw function, which executes a raw SQL query with named parameters.
func TestExecuteRaw(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name      string
		query     string
		params    map[string]interface{}
		mockSetup func(sqlmock.Sqlmock)
		want      []map[string]interface{}
		wantErr   bool
	}{
		{
			name: "successful query",
			// This test case checks if ExecuteRaw executes a query successfully with correct parameters and returns the expected results.
			query: "SELECT id, name, status, created_at FROM test_models WHERE id = {{id}} AND status = {{status}}",
			params: map[string]interface{}{
				"id":     int64(1),
				"status": "active",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "status", "created_at"}).
					AddRow(1, "Test Model", "active", now)
				mock.ExpectQuery("SELECT (.+) FROM test_models WHERE id = \\$1 AND status = \\$2").
					WithArgs(1, "active").
					WillReturnRows(rows)
			},
			want: []map[string]interface{}{
				{
					"id":         int64(1),
					"name":       "Test Model",
					"status":     "active",
					"created_at": now,
				},
			},
			wantErr: false,
		},
		{
			name: "parameter count mismatch",
			// This test case checks if ExecuteRaw returns an error when the number of parameters in the query does not match the number of parameters provided.
			query: "SELECT id, name FROM test_models WHERE id = {{id}} AND status = {{status}} AND other = {{other}}",
			params: map[string]interface{}{
				"id":     int64(1),
				"status": "active",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {},
			wantErr:   true,
		},
		{
			name: "query execution error",
			// This test case checks if ExecuteRaw returns an error when the underlying database query fails.
			query: "SELECT id, name FROM test_models WHERE id = {{id}} AND status = {{status}}",
			params: map[string]interface{}{
				"id":     int64(1),
				"status": "active",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM test_models WHERE id = \\$1 AND status = \\$2").
					WithArgs(1, "active").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "type mismatch",
			// This test case checks if ExecuteRaw returns an error when the type of a parameter does not match the expected type.
			query: "SELECT id, name FROM test_models WHERE id = {{id}} AND status = {{status}}",
			params: map[string]interface{}{
				"id":     "not an int",
				"status": "active",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			results, err := ExecuteRaw[QueryParams, TestQueryResult](ctx, db, tt.query, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, results)
			assert.Equal(t, tt.want, results)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

// TestExecuteRawTyped tests the generic version of ExecuteRaw
func TestExecuteRawTyped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	now := time.Now()

	query := "SELECT id, name, status, created_at FROM test_models WHERE id = {{id}} AND status = {{status}}"
	params := map[string]interface{}{
		"id":     int64(1),
		"status": "active",
	}

	// Setup mock
	rows := sqlmock.NewRows([]string{"id", "name", "status", "created_at"}).
		AddRow(1, "Test Model", "active", now)
	mock.ExpectQuery("SELECT (.+) FROM test_models WHERE id = \\$1 AND status = \\$2").
		WithArgs(1, "active").
		WillReturnRows(rows)

	// Execute query
	results, err := ExecuteRaw[QueryParams, TestQueryResult](ctx, db, query, params)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Verify result
	expected := map[string]interface{}{
		"id":         int64(1),
		"name":       "Test Model",
		"status":     "active",
		"created_at": now,
	}
	assert.Equal(t, expected, results[0])
}

func TestExecuteRawWithCustomScanner(t *testing.T) {
	// Setup test database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	// Test querying with custom type
	query := "SELECT id, name FROM test_custom WHERE id = {{id}}"
	params := map[string]interface{}{
		"id": CustomID{ID: 1, Type: "user"},
	}

	// Setup mock
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("1:user", "Alice")
	mock.ExpectQuery("SELECT id, name FROM test_custom WHERE id = \\$1").
		WithArgs("1:user").
		WillReturnRows(rows)

	// Execute query
	results, err := ExecuteRaw[TestCustomParams, TestCustomResult](
		context.Background(),
		db,
		query,
		params,
	)
	require.NoError(t, err)
	require.Len(t, results, 1)

	// Verify custom type was scanned correctly
	result := results[0]
	customID, ok := result["custom_id"].(CustomID)
	require.True(t, ok)
	require.Equal(t, 1, customID.ID)
	require.Equal(t, "user", customID.Type)
	require.Equal(t, "Alice", result["name"])

	// Verify all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestBuildMetadataMap tests the BuildMetadataMap function, which extracts metadata from a struct using reflection.
func TestBuildMetadataMap(t *testing.T) {
	tests := []struct {
		name      string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid model",
			wantCount: 4, // id, name, status, created_at
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metaMap, err := BuildMetadataMap[TestQueryResult]()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(metaMap))
		})
	}
}

// TestValidateMapParamsAgainstStructNamed tests the ValidateMapParamsAgainstStructNamed function
func TestValidateMapParamsAgainstStructNamed(t *testing.T) {
	tests := []struct {
		name        string
		paramMap    map[string]interface{}
		queryParams []string
		wantErr     bool
	}{
		{
			name: "valid params",
			paramMap: map[string]interface{}{
				"id":     int64(1),
				"status": "active",
			},
			queryParams: []string{"id", "status"},
			wantErr:    false,
		},
		{
			name: "type mismatch",
			paramMap: map[string]interface{}{
				"id":     "not an int",
				"status": "active",
			},
			queryParams: []string{"id", "status"},
			wantErr:    true,
		},
		{
			name: "missing param",
			paramMap: map[string]interface{}{
				"id": int64(1),
			},
			queryParams: []string{"id", "status"},
			wantErr:    false, // should not error as missing params are set to nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := ValidateMapParamsAgainstStructNamed[QueryParams](tt.paramMap, tt.queryParams)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.queryParams), len(args))
		})
	}
}
