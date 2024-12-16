# Getting Started with SQLD

SQLD provides two ways to execute SQL queries:
1. `ExecuteRaw`: For executing custom SQL with type-safe parameters
2. `Execute`: For executing queries using a structured JSON format (documentation coming soon)

This guide focuses on using `ExecuteRaw`.

## Using ExecuteRaw

### Step 1: Define Parameter Struct
Create a struct for your query parameters with both `db` and `json` tags:

```go
type QueryParams struct {
    Department string  `db:"department" json:"department"`
    MinSalary  float64 `db:"min_salary" json:"min_salary"`
}
```

The `db` tags must match the parameter names in your SQL query.

### Step 2: Define Result Struct
Create a struct that matches your query's result columns with `db` tags:

```go
type Result struct {
    ID        int64   `db:"id"`
    FirstName string  `db:"first_name"`
    Salary    float64 `db:"salary"`
}
```

The `db` tags must match your column names.

### Step 3: Write Your Query
Write your SQL query using {{param_name}} for parameters:

```sql
SELECT id, first_name, salary
FROM employees
WHERE department = {{department}}
AND salary >= {{min_salary}}
```

### Step 4: Execute the Query

```go
results, err := sqld.ExecuteRaw[QueryParams, Result](
    ctx, 
    db,  // can be *sql.DB or *pgx.Conn
    sqld.ExecuteRawRequest{
        Query: query,
        Params: map[string]interface{}{
            "department": "Engineering",
            "min_salary": 50000,
        },
        // Optional: control which fields appear in results
        SelectFields: []string{"first_name", "salary"},
    },
)
```

## What SQLD Validates

SQLD performs these validations:

1. Parameter Validation:
   - All {{param}} placeholders must have values
   - Parameter types must match struct field types
   - No extra parameters allowed

2. SQL Validation:
   - Must be a SELECT statement
   - SQL syntax must be valid

3. Result Validation:
   - Result struct must have db tags
   - Column names must match db tags

## Error Handling

Handle errors based on their type:

```go
results, err := sqld.ExecuteRaw[QueryParams, Result](ctx, db, req)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "parameter"):
        // Parameter validation error
    case strings.Contains(err.Error(), "syntax"):
        // SQL syntax error
    default:
        // Database error
    }
    return
}
```

## Next Steps

- See the `examples/` directory for more examples
- Read about the structured query system (coming soon)
