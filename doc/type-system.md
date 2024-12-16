# Type System in SQLD

SQLD provides a type system that ensures safety for dynamic queries. This document explains how SQLD's type system works and how to use it.

## Core Concepts

### 1. Query Types

SQLD supports two types of queries:

#### Raw Queries with ExecuteRaw
The Raw Query System allows you to write SQL queries with named parameters while ensuring type safety through runtime validation:

```go
// Raw SQL with named parameters using {{param_name}} syntax
query := `
    SELECT id, name, description
    FROM users
    WHERE status = {{status}}
    AND created_at > {{since_date}}
`
```

#### Structured Queries with Execute
The Structured Query System uses a JSON-based format to build queries:

```go
type QueryRequest struct {
    Select     []string                // Fields to select
    Where      map[string]interface{}  // WHERE conditions
    OrderBy    []OrderByClause         // ORDER BY clauses
    Pagination *PaginationRequest      // Optional pagination
}
```

### 2. Type Parameters

Both query types use Go generics for type safety:

```go
// For raw queries
func ExecuteRaw[P any, R any](
    ctx context.Context,
    db interface{},
    req ExecuteRawRequest,
) ([]map[string]interface{}, error)

// For structured queries
func Execute[T Model](
    ctx context.Context,
    db interface{},
    req QueryRequest,
) (QueryResponse[T], error)
```

Where:
- `P`: Parameter struct type with both `db` and `json` tags (required)
- `R`: Result struct type with `db` tags for column mapping
- `T`: Model type implementing the `Model` interface:
  ```go
  type Model interface {
      TableName() string
  }
  ```

### 3. Response Types

#### Raw Query Response
ExecuteRaw returns a slice of maps:
```go
[]map[string]interface{}
```
Each map contains field names as keys (from `db` tags) and their values.

#### Structured Query Response
Execute returns a QueryResponse:
```go
type QueryResponse[T Model] struct {
    Data       []QueryResult       `json:"data"`
    Pagination *PaginationResponse `json:"pagination,omitempty"`
    Error      string             `json:"error,omitempty"`
}

type PaginationResponse struct {
    Page       int `json:"page"`        // Current page (1-based)
    PageSize   int `json:"page_size"`   // Items per page
    TotalItems int `json:"total_items"` // Total number of items
    TotalPages int `json:"total_pages"` // Total number of pages
}
```

## Type Safety Features

### 1. Parameter Validation

SQLD performs comprehensive validation:

```go
// Parameter struct must have both db and json tags
type QueryParams struct {
    Status    string    `db:"status" json:"status"`
    SinceDate time.Time `db:"since_date" json:"since_date"`
}

// Validation ensures:
// - All {{param}} placeholders have matching struct fields
// - Parameter types match struct field types
// - No extra parameters are provided
// - SQL injection prevention through parameter substitution
```

### 2. Result Mapping

Results are mapped using struct tags:

```go
type UserProfile struct {
    ID          int64       `db:"id" json:"id"`
    Name        string      `db:"name" json:"name"`
    Description pgtype.Text `db:"description" json:"description"`
    CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}

// Mapping features:
// - Automatic mapping using db tags
// - Proper handling of null values with pgx types
// - JSON serialization support
// - Type conversion where appropriate
```

### 3. Using sqlc for Type Generation

We recommend using sqlc to generate type-safe structs:

```sql
-- Write a query to generate parameter and result structs
-- name: GetUserProfiles :many
SELECT 
    u.id,           -- int64
    u.name,         -- string
    p.description,  -- pgtype.Text
    p.created_at    -- time.Time
FROM users u
JOIN profiles p ON p.user_id = u.id
WHERE u.id = $1
  AND p.type = $2;
```

sqlc generates:
```go
type GetUserProfilesParams struct {
    ID   int64  `db:"id" json:"id"`
    Type string `db:"type" json:"type"`
}

type GetUserProfilesRow struct {
    ID          int64       `db:"id" json:"id"`
    Name        string      `db:"name" json:"name"`
    Description pgtype.Text `db:"description" json:"description"`
    CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}
```

Then use these types with ExecuteRaw:
```go
results, err := sqld.ExecuteRaw[sqlc.GetUserProfilesParams, sqlc.GetUserProfilesRow](
    ctx,
    db,
    sqld.ExecuteRawRequest{
        Query: query,
        Params: map[string]interface{}{
            "id": 123,
            "type": "personal",
        },
        SelectFields: []string{"name", "description"}, // Optional: filters output fields
    },
)
```

### 4. Return Type Handling
   - ExecuteRaw returns `[]map[string]interface{}`
   - Each map contains the selected fields as keys
   - Use SelectFields to control which fields appear in results

### 5. Pagination
   - Page numbers start at 1 (not 0)
   - PageSize is automatically capped at MaxPageSize (100)
   - Default PageSize is 10 if not specified
   - Response includes total items and total pages
   - Example:
     ```go
     req := sqld.ExecuteRawRequest{
         Query: query,
         Params: params,
         Pagination: &sqld.PaginationRequest{
             Page: 1,      // First page
             PageSize: 10, // Results per page
         },
     }
     ```

### 6. Error Handling
   - Type errors if P or R are not structs
   - Missing or extra parameter errors
   - Missing json tag errors during parameter struct field processing
   - Parameter type mismatch errors
   - SQL syntax errors after parameter substitution
   - Non-SELECT statement errors
   - Database query execution errors
   - Row scanning errors
