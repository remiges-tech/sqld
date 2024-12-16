// Package sqld provides dynamic SQL query generation and execution from JSON requests.
// It offers two distinct subsystems for handling different query needs:
//
// 1. Structured Query System: A high-level, type-safe abstraction for building queries using Go generics
// 2. Safe Raw Query System: A flexible system for writing raw SQL with safety guarantees
//
// Key Features:
// - Type-safe query building using Go generics
// - Dynamic field selection with validation
// - Built-in pagination with metadata
// - Named parameter support with type checking
// - SQL injection protection
// - SQLC integration
//
// Architecture:
//
// The package follows a clean architecture with these components:
//  1. Query Builder: Constructs SQL queries with runtime validation
//  2. Type System: Manages metadata and type validation using generics
//  3. Validator: Ensures field and parameter correctness
//  4. Executor: Safely executes queries and maps results
//
// Example usage of the Structured Query System:
//
//	resp, err := sqld.Execute[Employee](ctx, db, sqld.QueryRequest{
//	    Select: []string{"id", "name", "email"},
//	    Where: map[string]interface{}{
//	        "is_active": true,
//	    },
//	    Pagination: &sqld.PaginationRequest{
//	        Page: 1,
//	        PageSize: 10,
//	    },
//	})
//
// Raw Query Example:
//
//	results, err := sqld.ExecuteRaw[QueryParams, Employee](ctx, db, sqld.ExecuteRawRequest{
//	    Query: `SELECT id, name, salary
//	            FROM employees
//	            WHERE department = {{department}}
//	            AND salary >= {{min_salary}}`,
//	    Params: map[string]interface{}{
//	        "department": "Engineering",
//	        "min_salary": 50000,
//	    },
//	})
//
// For more examples and detailed documentation, visit the examples directory.
package sqld
