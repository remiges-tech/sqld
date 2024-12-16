package sqld

import (
	"testing"
	"time"
)

// BuilderTestModel is a sample model for testing
type BuilderTestModel struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (BuilderTestModel) TableName() string {
	return "test_models"
}

func TestBuildQuery(t *testing.T) {
	// Register our test model
	if err := Register(BuilderTestModel{}); err != nil {
		t.Fatalf("Failed to register test model: %v", err)
	}

	tests := []struct {
		name    string
		req     QueryRequest
		wantSQL string
		wantErr bool
	}{
		{
			name: "basic select all queryable fields",
			req: QueryRequest{
				Select: []string{"id", "name", "age", "created_at"},
			},
			wantSQL: `SELECT id, name, age, created_at FROM test_models`,
		},
		{
			name: "select with where clause single condition",
			req: QueryRequest{
				Select: []string{"id", "name"},
				Where: map[string]interface{}{
					"age": 25,
				},
			},
			wantSQL: `SELECT id, name FROM test_models WHERE age = $1`,
		},
		{
			name: "select with where clause multiple conditions",
			req: QueryRequest{
				Select: []string{"id", "name"},
				Where: map[string]interface{}{
					"age": 25,
				},
			},
			wantSQL: `SELECT id, name FROM test_models WHERE age = $1`,
		},
		{
			name: "select with single order by ascending",
			req: QueryRequest{
				Select: []string{"id", "name"},
				OrderBy: []OrderByClause{
					{Field: "age", Desc: false},
				},
			},
			wantSQL: `SELECT id, name FROM test_models ORDER BY age ASC`,
		},
		{
			name: "select with single order by descending",
			req: QueryRequest{
				Select: []string{"id", "name"},
				OrderBy: []OrderByClause{
					{Field: "age", Desc: true},
				},
			},
			wantSQL: `SELECT id, name FROM test_models ORDER BY age DESC`,
		},
		{
			name: "select with multiple order by",
			req: QueryRequest{
				Select: []string{"id", "name"},
				OrderBy: []OrderByClause{
					{Field: "age", Desc: true},
					{Field: "name", Desc: false},
				},
			},
			wantSQL: `SELECT id, name FROM test_models ORDER BY age DESC, name ASC`,
		},
		{
			name: "select with where and order by",
			req: QueryRequest{
				Select: []string{"id", "name"},
				Where: map[string]interface{}{
					"age": 25,
				},
				OrderBy: []OrderByClause{
					{Field: "name", Desc: true},
				},
			},
			wantSQL: `SELECT id, name FROM test_models WHERE age = $1 ORDER BY name DESC`,
		},
		{
			name: "invalid field in select",
			req: QueryRequest{
				Select: []string{"invalid_field"},
			},
			wantErr: true,
		},
		{
			name: "invalid field in where",
			req: QueryRequest{
				Select: []string{"id"},
				Where: map[string]interface{}{
					"invalid_field": "value",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid field in order by",
			req: QueryRequest{
				Select: []string{"id"},
				OrderBy: []OrderByClause{
					{Field: "invalid_field", Desc: true},
				},
			},
			wantErr: true,
		},
		{
			name: "empty select fields",
			req: QueryRequest{
				Select: []string{},
			},
			wantErr: true,
		},
		{
			name:    "nil select fields",
			req:     QueryRequest{},
			wantErr: true,
		},
		{
			name: "select with limit",
			req: QueryRequest{
				Select: []string{"id", "name"},
				Limit:  intPtr(10),
			},
			wantSQL: `SELECT id, name FROM test_models LIMIT 10`,
		},
		{
			name: "select with offset",
			req: QueryRequest{
				Select: []string{"id", "name"},
				Offset: intPtr(20),
			},
			wantSQL: `SELECT id, name FROM test_models OFFSET 20`,
		},
		{
			name: "select with limit and offset",
			req: QueryRequest{
				Select: []string{"id", "name"},
				Limit:  intPtr(10),
				Offset: intPtr(20),
			},
			wantSQL: `SELECT id, name FROM test_models LIMIT 10 OFFSET 20`,
		},
		{
			name: "select with negative limit",
			req: QueryRequest{
				Select: []string{"id"},
				Limit:  intPtr(-1),
			},
			wantErr: true,
		},
		{
			name: "select with negative offset",
			req: QueryRequest{
				Select: []string{"id"},
				Offset: intPtr(-1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := buildQuery[BuilderTestModel](tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			sql, _, err := query.ToSql()
			if err != nil {
				t.Errorf("Failed to generate SQL: %v", err)
				return
			}

			if sql != tt.wantSQL {
				t.Errorf("buildQuery() generated SQL = %v, want %v", sql, tt.wantSQL)
			}
		})
	}
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}
