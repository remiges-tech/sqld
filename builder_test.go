package sqld

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type BuilderTestModel struct {
	ID        int     `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	Age       int     `json:"age" db:"age"`
	Email     string  `json:"email" db:"email"`
	Active    bool    `json:"active" db:"active"`
	Salary    float64 `json:"salary" db:"salary"`
	Nullable  *string `json:"nullable" db:"nullable"`
}

func (BuilderTestModel) TableName() string {
	return "test_models"
}

func TestBuildQuery(t *testing.T) {
	// Register test model
	if err := Register[BuilderTestModel](); err != nil {
		t.Fatalf("Failed to register test model: %v", err)
	}

	tests := []struct {
		name    string
		request QueryRequest
		want    string
		wantErr bool
	}{
		{
			name: "basic select",
			request: QueryRequest{
				Select: []string{"name", "age"},
			},
			want: "SELECT name, age FROM test_models",
		},
		{
			name: "with where clause",
			request: QueryRequest{
				Select: []string{"name", "age"},
				Where: []Condition{
					{
						Field:    "age",
						Operator: OpGreaterThan,
						Value:    18,
					},
				},
			},
			want: "SELECT name, age FROM test_models WHERE age > $1",
		},
		{
			name: "with multiple where conditions",
			request: QueryRequest{
				Select: []string{"name", "age", "email"},
				Where: []Condition{
					{
						Field:    "age",
						Operator: OpGreaterThanOrEqual,
						Value:    21,
					},
					{
						Field:    "email",
						Operator: OpEqual,
						Value:    "test@example.com",
					},
				},
			},
			want: "SELECT name, age, email FROM test_models WHERE age >= $1 AND email = $2",
		},
		{
			name: "with order by",
			request: QueryRequest{
				Select: []string{"name", "age"},
				OrderBy: []OrderByClause{
					{Field: "age", Desc: true},
				},
			},
			want: "SELECT name, age FROM test_models ORDER BY age DESC",
		},
		{
			name: "with pagination",
			request: QueryRequest{
				Select: []string{"name", "age"},
				Limit:  intPtr(10),
				Offset: intPtr(0),
			},
			want: "SELECT name, age FROM test_models LIMIT 10 OFFSET 0",
		},
		{
			name: "complex query",
			request: QueryRequest{
				Select: []string{"name", "age", "email"},
				Where: []Condition{
					{
						Field:    "age",
						Operator: OpGreaterThan,
						Value:    25,
					},
					{
						Field:    "email",
						Operator: OpIn,
						Value:    []string{"test1@example.com", "test2@example.com"},
					},
				},
				OrderBy: []OrderByClause{
					{Field: "name", Desc: false},
					{Field: "age", Desc: true},
				},
				Limit:  intPtr(10),
				Offset: intPtr(20),
			},
			want: "SELECT name, age, email FROM test_models WHERE age > $1 AND email IN ($2,$3) ORDER BY name ASC, age DESC LIMIT 10 OFFSET 20",
		},
		{
			name: "with like operator",
			request: QueryRequest{
				Select: []string{"name", "email"},
				Where: []Condition{
					{
						Field:    "name",
						Operator: OpLike,
						Value:    "%John%",
					},
				},
			},
			want: "SELECT name, email FROM test_models WHERE name LIKE $1",
		},
		{
			name: "with ilike operator",
			request: QueryRequest{
				Select: []string{"name", "email"},
				Where: []Condition{
					{
						Field:    "name",
						Operator: OpILike,
						Value:    "%john%",
					},
				},
			},
			want: "SELECT name, email FROM test_models WHERE name ILIKE $1",
		},
		{
			name: "with null checks",
			request: QueryRequest{
				Select: []string{"name", "email"},
				Where: []Condition{
					{
						Field:    "email",
						Operator: OpIsNull,
					},
					{
						Field:    "name",
						Operator: OpIsNotNull,
					},
				},
			},
			want: "SELECT name, email FROM test_models WHERE email IS NULL AND name IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildQuery[BuilderTestModel](tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Convert the SelectBuilder to a string
			sql, _, err := got.ToSql()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, sql)
		})
	}
}
