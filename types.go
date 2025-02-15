package sqld

import (
	"reflect"
)

// Model interface that represents a database table.
// We have it so that we can ensure that any type
// used with the query builder can map to a database table.
type Model interface {
	TableName() string
}

// ModelMetadata stores information about a model that represents a database table.
// We use it where we need the list of fields of a table and their types. For example,
// validating fields names in queries, etc.
type ModelMetadata struct {
	TableName string
	Fields    map[string]Field
}

// Field represents a queryable field with its metadata.
// Field maintains the mapping between database schema, Go types and JSON field names
// that come from the user.
// It is populated when the model is registered with Register().
type Field struct {
	Name           string       // Name of the field in the database
	JSONName       string       // Name of the field in the JSON request
	GoFieldName    string       // Name of the field in the Go struct
	Type           reflect.Type // Original Go type
	NormalizedType reflect.Type // Normalized type for validation
}

// OrderByClause defines how to sort results
type OrderByClause struct {
	Field string `json:"field"` // Must match struct field tags
	Desc  bool   `json:"desc"`  // true for descending order
}

// PaginationRequest represents pagination parameters.
// If provided in QueryRequest, it takes precedence over direct Limit/Offset values.
// Page numbers start at 1 (not 0). For example, page 1 is the first page, page 2 is the second page, etc.
// PageSize is automatically capped at MaxPageSize (100).
type PaginationRequest struct {
	Page     int `json:"page"`      // Page number starting at 1 (e.g., 1 for first page, 2 for second page)
	PageSize int `json:"page_size"` // Results per page (minimum: 1, default: 10, maximum: 100)
}

// PaginationResponse contains pagination metadata
type PaginationResponse struct {
	Page       int `json:"page"`        // Current page number (1-based)
	PageSize   int `json:"page_size"`   // Items per page
	TotalItems int `json:"total_items"` // Total number of items
	TotalPages int `json:"total_pages"` // Total number of pages
}

// Operator represents a SQL comparison operator
type Operator string

const (
	OpEqual              Operator = "="
	OpNotEqual           Operator = "!="
	OpGreaterThan        Operator = ">"
	OpLessThan           Operator = "<"
	OpGreaterThanOrEqual Operator = ">="
	OpLessThanOrEqual    Operator = "<="
	OpLike              Operator = "LIKE"
	OpILike             Operator = "ILIKE"
	OpIn                Operator = "IN"
	OpNotIn             Operator = "NOT IN"
	OpIsNull            Operator = "IS NULL"
	OpIsNotNull         Operator = "IS NOT NULL"

	// SelectAll is a special value that can be used in QueryRequest.Select to select all fields
	SelectAll = "ALL"
)

// Condition represents a single WHERE condition with an operator
type Condition struct {
	Field    string      `json:"field"`    // Field name (must match JSON field name)
	Operator Operator    `json:"operator"`  // SQL operator
	Value    interface{} `json:"value"`     // Value to compare against (optional for IS NULL/IS NOT NULL)
}

// QueryRequest represents the structure for building dynamic SQL queries.
// It provides type-safe query building with runtime validation against model metadata.
type QueryRequest struct {
	// Select specifies which fields to retrieve. Field names must match the JSON tags
	// in your model struct. This field is required and cannot be empty.
	// Each field name is validated against the model's metadata.
	Select []string `json:"select"`

	// Where specifies filter conditions using operators. Each condition consists of
	// a field name (matching JSON field names), an operator, and a value.
	// Optional - if not provided, no filtering is applied.
	Where []Condition `json:"where,omitempty"`

	// OrderBy specifies sorting criteria. Each OrderByClause contains a field name
	// (must match JSON field names) and sort direction.
	// Optional - if not provided, no sorting is applied.
	// Each field name is validated against the model's metadata.
	OrderBy []OrderByClause `json:"order_by,omitempty"`

	// Pagination enables page-based result limiting. If provided, it takes precedence
	// over direct Limit/Offset values. Uses DefaultPageSize (10) if not specified,
	// and caps at MaxPageSize (100).
	// Optional - if not provided, all results are returned unless Limit is set.
	Pagination *PaginationRequest `json:"pagination,omitempty"`

	// Limit specifies maximum number of results to return.
	// Only used if Pagination is not provided.
	// Optional - nil means no limit.
	// Must be non-negative if provided.
	Limit *int `json:"limit,omitempty"`

	// Offset specifies number of results to skip.
	// Only used if Pagination is not provided.
	// Optional - nil means no offset.
	// Must be non-negative if provided.
	Offset *int `json:"offset,omitempty"`
}

// QueryResponse represents the outgoing JSON structure
type QueryResponse[T Model] struct {
	Data       []QueryResult       `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
	Error      string              `json:"error,omitempty"`
	// TODO: Add these fields for enhanced responses
	// Metadata QueryMetadata `json:"metadata,omitempty"`
}

// QueryResult represents a single row as map of field name to value
type QueryResult map[string]interface{}

// TODO: Add metadata type for enhanced responses
// type QueryMetadata struct {
//     TotalRows    int           `json:"total_rows"`
//     ExecutionTime time.Duration `json:"execution_time"`
//     Page         int           `json:"page"`
//     TotalPages   int           `json:"total_pages"`
// }
