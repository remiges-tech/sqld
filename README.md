# SQLD For Dynamic SQL

`sqld` is a package that provides type-safe dynamic SQL query building and execution. It offers two distinct subsystems:

1. **Structured Query System**: Type-safe query building using squirrel with model metadata validation
2. **Raw Query System**: Execute any SELECT statement with named parameters and type validation

## Design Philosophy

This library was built for a specific use case: **frontend list views with filters**.

A typical frontend displays data in a table with filter controls. Users select filters, the frontend calls a list API, and the backend fetches matching rows. These calls need dynamic field selection, filter conditions, sorting, and pagination.

The **Structured Query System** handles this pattern with built-in pagination and type safety. It assumes:

- **Single table queries** - List views typically show data from one entity. For data spread across tables, create a database view or use the raw query system.
- **AND conditions only** - UI filters combine with AND. Users expect "department=Engineering AND status=active".
- **PostgreSQL** - Parameter binding uses `$1, $2` syntax. Array operators use PostgreSQL syntax.

The **Raw Query System** handles everything else - JOINs, GROUP BY, subqueries, window functions. You generate the SQL dynamically using a query builder like squirrel or your own code. sqld validates the SQL using a PostgreSQL parser to catch syntax errors before runtime, and handles parameter binding safely.

## Key Features

### Structured Query System
- Type-safe query building using squirrel
- Runtime validation against model metadata
- Field validation and mapping between JSON and database fields
- Built-in pagination (page-based or limit/offset)
- Multiple field ordering with direction control
- Support for both sql.DB and pgx.Conn

### Safe Raw Query System
- Named parameters using {{param_name}} syntax
- Runtime type validation against parameter struct
- SQL syntax validation using PostgreSQL parser
- SQL injection prevention through parameter validation
- Optional field selection in results

## Usage

### Structured Query System
```go
// Define and register your model
type Employee struct {
    ID         int64     `json:"id" db:"id"`
    FirstName  string    `json:"first_name" db:"first_name"`
    Department string    `json:"department" db:"department"`
    Salary     float64   `json:"salary" db:"salary"`
    IsActive   bool      `json:"is_active" db:"is_active"`
}

// Required: Implement Model interface
func (Employee) TableName() string {
    return "employees"
}

// Register model for metadata validation
if err := sqld.Register(Employee{}); err != nil {
    log.Fatal(err)
}

// Execute a query with type-safe building
resp, err := sqld.Execute[Employee](ctx, db, sqld.QueryRequest{
    Select: []string{"id", "first_name", "department", "salary"},
    Where: map[string]interface{}{
        "department": "Engineering",
        "is_active": true,
    },
    OrderBy: []sqld.OrderByClause{
        {Field: "salary", Desc: true},
    },
    Pagination: &sqld.PaginationRequest{
        Page: 1,      // Page numbers start at 1
        PageSize: 10, // Automatically capped at MaxPageSize (100)
    },
})

// Response includes data and pagination metadata
fmt.Printf("Total Items: %d\n", resp.Pagination.TotalItems)
for _, employee := range resp.Data {
    // Access fields using type-safe struct
}
```

### Safe Raw Query System
```go
// Define your parameter struct with both db and json tags
type QueryParams struct {
    Department string  `db:"department" json:"department"`
    MinSalary  float64 `db:"min_salary" json:"min_salary"`
}

// Define your result struct with db tags
type EmployeeResult struct {
    ID        int64   `db:"id"`
    FirstName string  `db:"first_name"`
    Salary    float64 `db:"salary"`
}

// Execute a raw query with validation
results, err := sqld.ExecuteRaw[QueryParams, EmployeeResult](
    ctx, 
    db,
    sqld.ExecuteRawRequest{
        Query: `
            SELECT id, first_name, salary
            FROM employees
            WHERE department = {{department}}
            AND salary >= {{min_salary}}
            ORDER BY salary DESC
        `,
        Params: map[string]interface{}{
            "department": "Engineering",
            "min_salary": 50000,
        },
        SelectFields: []string{"first_name", "salary"}, // Optional: filters which fields from the result struct are included in the output
    },
)
```

## Architecture

The package is built around these core components:

1. **Model Metadata System**
   - Runtime model registration and validation
   - Field mapping between database, Go types, and JSON
   - Type validation for query parameters

2. **Query Builder**
   - Type-safe query building using squirrel
   - Parameter binding and validation
   - Support for pagination and ordering

3. **Executor**
   - Support for both sql.DB and pgx.Conn
   - Result mapping with scany
   - Pagination metadata handling

## Documentation
For more detailed documentation and examples:
- Check the `examples/` directory for working examples
- See `doc.go` for package documentation
- Try the example server in `examples/main.go`
