# Query Building in SQLD

SQLD provides two query systems:
1. **Raw Query System** (`ExecuteRaw`): For executing custom SQL with type-safe parameters
2. **Structured Query System** (`Execute`): For building queries using a structured format

## Raw Query System

The Raw Query System allows you to write SQL queries with named parameters while ensuring type safety through runtime validation.

### Using sqlc with ExecuteRaw

While ExecuteRaw provides flexibility for dynamic queries, we strongly recommend using [sqlc](https://sqlc.dev/) to generate type-safe structs. 

Instead of manually defining parameter and result structs, use sqlc purely for generating the type definitions. Write a simple sqlc query that includes all the fields you need - you won't actually execute this query using sqlc, but will use its generated structs with ExecuteRaw.

### 1. sqlc Generated Structs

**sqlc Query Definition (For Struct Generation Only)**:
```sql
-- name: GetUserProfiles :many
-- This query is only for generating structs, we won't execute it via sqlc
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

> **Important**: We just need this query to generate the required model struct. We are not actually going to use sqlc generated code to execute the query.

**sqlc Generated Parameter Struct:**
```go
type GetUserProfilesParams struct {
    ID   int64  `db:"id" json:"id"`
    Type string `db:"type" json:"type"`
}
```

**sqlc Generated Row Struct:**
```go
type GetUserProfilesRow struct {
    ID          int64       `db:"id" json:"id"`
    Name        string      `db:"name" json:"name"`
    Description pgtype.Text `db:"description" json:"description"`
    CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}
```

These structs are type-safe and include proper `db` and `json` tags. They ensure that when we later execute a custom query using `ExecuteRaw`, the parameters and results align correctly with the database schema.

### 2. Dynamic Query Using ExecuteRaw

Now we can write a dynamic query that we will actually execute using `ExecuteRaw`. We'll use the sqlc-generated structs as type parameters, giving us type safety while maintaining query flexibility:

```go
query := `
    SELECT u.id, u.name, p.description, p.created_at
    FROM users u
    JOIN profiles p ON p.user_id = u.id
    WHERE u.id = {{id}}
      AND p.type = {{type}}
    ORDER BY p.created_at DESC
`

params := GetUserProfilesParams{
    ID:   12345,
    Type: "personal",
}

result, err := db.ExecuteRaw[sqlc.GetUserProfilesParams, sqlc.GetUserProfilesRow](ctx, query, params, &results)
```

The key benefit here is that we get sqlc's type safety while maintaining complete freedom in writing our dynamic queries:
- Use `GetUserProfilesParams` for type-safe parameters
- Use `GetUserProfilesRow` for properly typed results
- Write any dynamic query while ensuring schema compatibility through the generated types

### Parameter Format
- Use `{{param_name}}` for parameters (not `:param_name`)
- Parameters are automatically converted to PostgreSQL positional parameters ($1, $2, etc.)

Example:
```sql
SELECT id, name, salary 
FROM employees 
WHERE department = {{department}} 
AND salary >= {{min_salary}}
```

### Type Safety
SQLD ensures type safety through several validations:

1. **Parameter Validation**:
   - All `{{param}}` placeholders must have matching values in the params map
   - No extra parameters allowed in the params map
   - Parameter types must match the struct field types exactly
   - All fields with `db` tags must have corresponding `json` tags

2. **SQL Validation**:
   - Must be a SELECT statement
   - SQL syntax validated using PostgreSQL parser
   - Parameter substitution checked for correctness

3. **Result Validation**:
   - Result struct must be a valid struct type
   - All struct fields must have `db` tags
   - Column names must match the `db` tags

### Usage Example

```go
type QueryParams struct {
    Department string  `db:"department" json:"department"`
    MinSalary  float64 `db:"min_salary" json:"min_salary"`
}

type Result struct {
    ID        int64   `db:"id"`
    FirstName string  `db:"first_name"`
    Salary    float64 `db:"salary"`
}

results, err := sqld.ExecuteRaw[QueryParams, Result](
    ctx, 
    db,  // can be *sql.DB or *pgx.Conn
    sqld.ExecuteRawRequest{
        Query: `
            SELECT id, first_name, salary
            FROM employees
            WHERE department = {{department}}
            AND salary >= {{min_salary}}
        `,
        Params: map[string]interface{}{
            "department": "Engineering",
            "min_salary": 50000,
        },
        SelectFields: []string{"first_name", "salary"}, // Optional: filters output fields
    },
)

## Structured Query System

The Structured Query System uses [Squirrel](https://github.com/Masterminds/squirrel) internally to build safe SQL queries.

### Components

1. **SELECT**:
   ```go
   Select: []string{"id", "first_name", "salary"}
   ```
   - Fields must exist in your model
   - Cannot be empty

2. **WHERE**:
   ```go
   Where: map[string]interface{}{
       "department": "Engineering",
       "is_active": true,
   }
   ```
   - Field names must exist in your model
   - Values are automatically parameterized
   - Only supports equality conditions

3. **ORDER BY**:
   ```go
   OrderBy: []OrderBy{
       {Field: "salary", Desc: true}
   }
   ```
   - Field must exist in your model
   - Supports both ASC and DESC

4. **Pagination**:
   ```go
   Pagination: &PaginationRequest{
       Page: 1,
       PageSize: 10,
   }
   ```
   - Both values must be non-negative
   - Page numbers start at 1
   - Automatically calculates LIMIT and OFFSET

### Example

```go
resp, err := sqld.Execute[Employee](ctx, db, sqld.QueryRequest{
    Select: []string{"id", "first_name", "salary"},
    Where: map[string]interface{}{
        "department": "Engineering",
        "is_active": true,
    },
    OrderBy: []OrderBy{{Field: "salary", Desc: true}},
    Pagination: &PaginationRequest{
        Page: 1,
        PageSize: 10,
    },
})

## Current Limitations

1. Raw Query System:
   - Only supports SELECT statements
   - Parameters must use {{param_name}} format
   - No support for IN clauses with arrays
   - No support for subqueries

2. Structured Query System:
   - No support for JOIN operations
   - No support for GROUP BY clauses
   - WHERE clause limited to equality conditions
   - No support for complex conditions (OR, IN, etc.)
