package sqld

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ValidatorTestModel struct {
	ID        int     `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	Age       int     `json:"age" db:"age"`
	Email     string  `json:"email" db:"email"`
	Active    bool    `json:"active" db:"active"`
	Salary    float64 `json:"salary" db:"salary"`
	Nullable  *string `json:"nullable" db:"nullable"`
}

func (ValidatorTestModel) TableName() string {
	return "test_models"
}

func TestValidateQueryRequest(t *testing.T) {
	// Register test model
	if err := Register[ValidatorTestModel](); err != nil {
		t.Fatalf("Failed to register test model: %v", err)
	}

	tests := []struct {
		name    string
		request QueryRequest
		wantErr bool
	}{
		{
			name: "valid basic request",
			request: QueryRequest{
				Select: []string{"name", "age", "email"},
			},
			wantErr: false,
		},
		{
			name: "valid request with where clause",
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
			wantErr: false,
		},
		{
			name: "valid request with multiple conditions",
			request: QueryRequest{
				Select: []string{"name", "email"},
				Where: []Condition{
					{
						Field:    "age",
						Operator: OpGreaterThanOrEqual,
						Value:    21,
					},
					{
						Field:    "active",
						Operator: OpEqual,
						Value:    true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with order by",
			request: QueryRequest{
				Select: []string{"name", "age"},
				OrderBy: []OrderByClause{
					{Field: "age", Desc: true},
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with pagination",
			request: QueryRequest{
				Select: []string{"name", "age"},
				Limit:  intPtr(10),
				Offset: intPtr(0),
			},
			wantErr: false,
		},
		{
			name: "invalid - empty select",
			request: QueryRequest{
				Select: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid - non-existent field in select",
			request: QueryRequest{
				Select: []string{"invalid_field"},
			},
			wantErr: true,
		},
		{
			name: "invalid - non-existent field in where",
			request: QueryRequest{
				Select: []string{"name"},
				Where: []Condition{
					{
						Field:    "invalid_field",
						Operator: OpEqual,
						Value:    "value",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - wrong value type in where",
			request: QueryRequest{
				Select: []string{"name"},
				Where: []Condition{
					{
						Field:    "age",
						Operator: OpEqual,
						Value:    "not_a_number",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - non-existent field in order by",
			request: QueryRequest{
				Select: []string{"name"},
				OrderBy: []OrderByClause{
					{Field: "invalid_field", Desc: true},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - negative limit",
			request: QueryRequest{
				Select: []string{"name"},
				Limit:  intPtr(-1),
			},
			wantErr: true,
		},
		{
			name: "invalid - negative offset",
			request: QueryRequest{
				Select: []string{"name"},
				Offset: intPtr(-1),
			},
			wantErr: true,
		},
		{
			name: "valid - null check operators",
			request: QueryRequest{
				Select: []string{"name", "nullable"},
				Where: []Condition{
					{
						Field:    "nullable",
						Operator: OpIsNull,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - IN operator with slice",
			request: QueryRequest{
				Select: []string{"name"},
				Where: []Condition{
					{
						Field:    "age",
						Operator: OpIn,
						Value:    []int{18, 21, 25},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid - IN operator with non-slice value",
			request: QueryRequest{
				Select: []string{"name"},
				Where: []Condition{
					{
						Field:    "age",
						Operator: OpIn,
						Value:    18,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid - pattern matching operators",
			request: QueryRequest{
				Select: []string{"name"},
				Where: []Condition{
					{
						Field:    "name",
						Operator: OpLike,
						Value:    "%John%",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var model ValidatorTestModel
			metadata, err := getModelMetadata(model)
			assert.NoError(t, err)

			validator := BasicValidator{}
			err = validator.ValidateQuery(tt.request, metadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
