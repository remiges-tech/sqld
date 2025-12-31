package sqld

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ArrayTestModel struct {
	ID          int64   `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	ReportingTo []int64 `json:"reporting_to" db:"reporting_to"`
}

func (ArrayTestModel) TableName() string {
	return "array_test_models"
}

func TestRegistryDetectsArrayFields(t *testing.T) {
	err := Register[ArrayTestModel]()
	require.NoError(t, err)

	var model ArrayTestModel
	metadata, err := getModelMetadata(model)
	require.NoError(t, err)

	scalarField := metadata.Fields["id"]
	assert.Nil(t, scalarField.Array)

	arrayField := metadata.Fields["reporting_to"]
	require.NotNil(t, arrayField.Array)
	assert.Equal(t, reflect.TypeOf(int64(0)), arrayField.Array.ElementType)
}

func TestValidatorAcceptsOpAnyOnArrayField(t *testing.T) {
	err := Register[ArrayTestModel]()
	require.NoError(t, err)

	var model ArrayTestModel
	metadata, err := getModelMetadata(model)
	require.NoError(t, err)

	validator := BasicValidator{}

	req := QueryRequest{
		Select: []string{"id", "name"},
		Where: []Condition{
			{
				Field:    "reporting_to",
				Operator: OpAny,
				Value:    int64(20),
			},
		},
	}

	err = validator.ValidateQuery(req, metadata)
	assert.NoError(t, err)
}

func TestBuildQueryWithOpAny(t *testing.T) {
	err := Register[ArrayTestModel]()
	require.NoError(t, err)

	req := QueryRequest{
		Select: []string{"id", "name"},
		Where: []Condition{
			{
				Field:    "reporting_to",
				Operator: OpAny,
				Value:    int64(20),
			},
		},
	}

	got, err := buildQuery[ArrayTestModel](req)
	require.NoError(t, err)

	sql, _, err := got.ToSql()
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, name FROM array_test_models WHERE $1 = ANY(reporting_to)", sql)
}

func TestValidatorAcceptsOpContainsOnArrayField(t *testing.T) {
	err := Register[ArrayTestModel]()
	require.NoError(t, err)

	var model ArrayTestModel
	metadata, err := getModelMetadata(model)
	require.NoError(t, err)

	validator := BasicValidator{}

	req := QueryRequest{
		Select: []string{"id", "name"},
		Where: []Condition{
			{
				Field:    "reporting_to",
				Operator: OpContains,
				Value:    []int64{20, 30},
			},
		},
	}

	err = validator.ValidateQuery(req, metadata)
	assert.NoError(t, err)
}

func TestValidatorRejectsOpContainsWithScalarValue(t *testing.T) {
	err := Register[ArrayTestModel]()
	require.NoError(t, err)

	var model ArrayTestModel
	metadata, err := getModelMetadata(model)
	require.NoError(t, err)

	validator := BasicValidator{}

	req := QueryRequest{
		Select: []string{"id", "name"},
		Where: []Condition{
			{
				Field:    "reporting_to",
				Operator: OpContains,
				Value:    int64(20),
			},
		},
	}

	err = validator.ValidateQuery(req, metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slice")
}

func TestBuildQueryWithOpContains(t *testing.T) {
	err := Register[ArrayTestModel]()
	require.NoError(t, err)

	req := QueryRequest{
		Select: []string{"id", "name"},
		Where: []Condition{
			{
				Field:    "reporting_to",
				Operator: OpContains,
				Value:    []int64{20, 30},
			},
		},
	}

	got, err := buildQuery[ArrayTestModel](req)
	require.NoError(t, err)

	sql, _, err := got.ToSql()
	require.NoError(t, err)
	assert.Equal(t, "SELECT id, name FROM array_test_models WHERE reporting_to @> $1", sql)
}

func TestValidatorAcceptsIsNullOnArrayField(t *testing.T) {
	err := Register[ArrayTestModel]()
	require.NoError(t, err)

	var model ArrayTestModel
	metadata, err := getModelMetadata(model)
	require.NoError(t, err)

	validator := BasicValidator{}

	tests := []struct {
		name     string
		operator Operator
	}{
		{"OpIsNull", OpIsNull},
		{"OpIsNotNull", OpIsNotNull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := QueryRequest{
				Select: []string{"id", "name"},
				Where: []Condition{
					{
						Field:    "reporting_to",
						Operator: tt.operator,
					},
				},
			}

			err = validator.ValidateQuery(req, metadata)
			assert.NoError(t, err)
		})
	}
}
