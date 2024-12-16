package sqld

import (
	"reflect"
	"testing"
)

func TestBasicValidator_ValidateQuery(t *testing.T) {
	// Create test metadata
	metadata := ModelMetadata{
		TableName: "test_models",
		Fields: map[string]Field{
			"id":         {Name: "id", JSONName: "id", Type: reflect.TypeOf(int64(0))},
			"name":       {Name: "name", JSONName: "name", Type: reflect.TypeOf("")},
			"is_active":  {Name: "is_active", JSONName: "is_active", Type: reflect.TypeOf(true)},
			"custom_int": {Name: "custom_int", JSONName: "custom_int", Type: reflect.TypeOf(CustomInt(0))},
		},
	}

	tests := []struct {
		name    string
		req     QueryRequest
		wantErr bool
	}{
		{
			name: "valid query",
			req: QueryRequest{
				Select: []string{"id", "name", "is_active"},
				Where: map[string]interface{}{
					"id": 1,
				},
				OrderBy: []OrderByClause{
					{Field: "custom_int", Desc: true},
				},
			},
			wantErr: false,
		},
		{
			name: "empty select",
			req: QueryRequest{
				Select: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid select field",
			req: QueryRequest{
				Select: []string{"invalid_field"},
			},
			wantErr: true,
		},
		{
			name: "invalid where field",
			req: QueryRequest{
				Select: []string{"id"},
				Where: map[string]interface{}{
					"invalid_field": "value",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid order by field",
			req: QueryRequest{
				Select: []string{"id"},
				OrderBy: []OrderByClause{
					{Field: "invalid_field"},
				},
			},
			wantErr: true,
		},
		{
			name: "negative limit",
			req: QueryRequest{
				Select: []string{"id"},
				Limit:  intPtr(-1),
			},
			wantErr: true,
		},
		{
			name: "negative offset",
			req: QueryRequest{
				Select: []string{"id"},
				Offset: intPtr(-1),
			},
			wantErr: true,
		},
	}

	validator := BasicValidator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateQuery(tt.req, metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
