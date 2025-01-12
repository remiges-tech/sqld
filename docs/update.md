# Type-Safe SQL Updates

The `sqld` package provides a type-safe approach to SQL updates through the `ExecuteUpdate` function and associated types. This document explains the type safety features and advantages over traditional ORMs like GORM.

## Basic Usage

```go
// Define your model
type Employee struct {
    ID        int64   `json:"id" db:"id"`
    FirstName string  `json:"first_name" db:"first_name"`
    Salary    float64 `json:"salary" db:"salary"`
}

// Create an update request
req := sqld.UpdateRequest{
    Set: map[string]interface{}{
        "salary": 85000.0,
        "first_name": "John",
    },
    Where: []sqld.Condition{
        {Field: "id", Operator: sqld.OpEqual, Value: 1},
    },
}

// Execute the update
rowsAffected, err := sqld.ExecuteUpdate[Employee](ctx, db, req)
```

## Type Safety Features

### 1. Compile-time Type Checking with Generics

The `ExecuteUpdate` function uses Go generics to ensure type safety at compile time:

```go
func ExecuteUpdate[T Model](ctx context.Context, db interface{}, req UpdateRequest) (int64, error)
```

- Requires a concrete type `T` that implements the `Model` interface
- Validates the model structure at compile time
- Prevents using invalid model types

Compared to GORM:
```go
// GORM - No compile-time type checking
db.Model(&User{}).Updates(map[string]interface{}{...}) // Any type allowed
```

### 2. Field Name Validation

All field names are validated against the model's metadata:

```go
// Our approach - Strict field validation
req := UpdateRequest{
    Set: map[string]interface{}{
        "non_existent_field": "value", // Error: invalid field
    },
}

// GORM - Silently ignores invalid fields
db.Model(&User{}).Updates(map[string]interface{}{
    "non_existent_field": "value", // No error, ignored
})
```

### 3. Type Compatibility Checking

Every value is checked for type compatibility with its corresponding field:

```go
// Our approach - Strict type checking
req := UpdateRequest{
    Set: map[string]interface{}{
        "salary": "not_a_number", // Error: type mismatch
    },
}

// GORM - Runtime errors or implicit conversion
db.Model(&User{}).Updates(map[string]interface{}{
    "age": "not_a_number", // Runtime error or implicit conversion
})
```

### 4. Required WHERE Clause

Protection against accidental mass updates:

```go
// Our approach - WHERE clause required
req := UpdateRequest{
    Set: map[string]interface{}{
        "salary": 85000.0,
    },
    // Error: WHERE clause required
}

// GORM - No WHERE clause required
db.Model(&User{}).Updates(map[string]interface{}{
    "status": "active", // Updates ALL records!
})
```

### 5. JSON to Database Field Mapping

Automatic handling of JSON and database field names:

```go
type Employee struct {
    FirstName string `json:"first_name" db:"first_name"`
}

// Our approach - Use JSON field names
req.Set["first_name"] = "John" // Automatically maps to correct DB field

// GORM - Must use struct field names
db.Model(&User{}).Updates(map[string]interface{}{
    "FirstName": "John", // Must match struct field name
})
```

### 6. Type-Safe Operators

Strongly typed operators prevent SQL injection and invalid operations:

```go
// Our approach - Type-safe operators
type Operator string
const (
    OpEqual     Operator = "="
    OpNotEqual  Operator = "!="
    OpIn        Operator = "IN"
    OpIsNull    Operator = "IS NULL"
)

// GORM - String-based operators
db.Where("age > ?", 20) // Raw string operators
```

### 7. Collection Type Safety

Type checking for collection operations (IN, NOT IN):

```go
// Our approach - Type-checked collections
req.Where = []Condition{
    {Field: "id", Operator: OpIn, Value: []int64{1, 2, 3}}, // Type checked
}

// GORM - No collection type checking
db.Where("id IN ?", []interface{}{1, "2", true}) // Mixed types allowed
```

### 8. Null Value Handling

Explicit and type-safe null handling:

```go
// Our approach - Explicit NULL operators
req.Where = []Condition{
    {Field: "manager_id", Operator: OpIsNull}, // Explicit NULL check
}

// GORM - Inconsistent null handling
db.Where("manager_id = ?", nil)      // Generates IS NULL
db.Where("manager_id IS NULL")       // Raw SQL required
```

## Benefits

1. **Compile-time Safety**
   - Type errors caught during compilation
   - Invalid model types prevented
   - Required fields enforced

2. **Runtime Safety**
   - Field name validation
   - Type compatibility checking
   - Protection against mass updates
   - Collection type validation

3. **Developer Experience**
   - Clear error messages
   - Consistent API
   - IDE autocompletion support
   - Self-documenting code

4. **Data Integrity**
   - Prevents accidental updates
   - Ensures type consistency
   - Validates all operations

## Example HTTP Handler

```go
func (s *Server) UpdateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
    var req sqld.UpdateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    rowsAffected, err := sqld.ExecuteUpdate[Employee](r.Context(), s.db, req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "rows_affected": rowsAffected,
        "message": "Update successful",
    })
}
```

## Transactional Updates Example

Here's an example of updating multiple tables within a transaction. This example demonstrates promoting an employee to a manager position while also updating their account balance:

```go
// Request structure for the promotion operation
type PromotionRequest struct {
    EmployeeID   int64   `json:"employee_id"`
    NewPosition  string  `json:"new_position"`
    SalaryRaise  float64 `json:"salary_raise"`
    BonusAmount  float64 `json:"bonus_amount"`
}

func (s *Server) PromoteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse request
    var promReq PromotionRequest
    if err := json.NewDecoder(r.Body).Decode(&promReq); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Start transaction
    tx, err := s.db.Begin(r.Context())
    if err != nil {
        http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback() // Rollback if not committed

    // 1. Update employee position and salary
    empUpdate := sqld.UpdateRequest{
        Set: map[string]interface{}{
            "position": promReq.NewPosition,
            "salary": promReq.SalaryRaise,
        },
        Where: []sqld.Condition{
            {Field: "id", Operator: sqld.OpEqual, Value: promReq.EmployeeID},
            {Field: "is_active", Operator: sqld.OpEqual, Value: true},
        },
    }

    empRows, err := sqld.ExecuteUpdate[Employee](r.Context(), tx, empUpdate)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to update employee: %v", err), http.StatusInternalServerError)
        return
    }
    if empRows == 0 {
        http.Error(w, "Employee not found or not active", http.StatusNotFound)
        return
    }

    // 2. Update account balance with bonus
    accUpdate := sqld.UpdateRequest{
        Set: map[string]interface{}{
            "balance": squirrel.Expr("balance + ?", promReq.BonusAmount),
        },
        Where: []sqld.Condition{
            {Field: "owner_id", Operator: sqld.OpEqual, Value: promReq.EmployeeID},
            {Field: "status", Operator: sqld.OpEqual, Value: "active"},
        },
    }

    accRows, err := sqld.ExecuteUpdate[Account](r.Context(), tx, accUpdate)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to update account: %v", err), http.StatusInternalServerError)
        return
    }
    if accRows == 0 {
        http.Error(w, "No active account found for employee", http.StatusNotFound)
        return
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
        return
    }

    // Return success response
    response := map[string]interface{}{
        "message": "Promotion successful",
        "employee_updated": empRows > 0,
        "account_updated": accRows > 0,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

Example request:
```bash
curl -X POST http://localhost:8080/employees/promote \
-H "Content-Type: application/json" \
-d '{
    "employee_id": 1,
    "new_position": "Senior Manager",
    "salary_raise": 120000.00,
    "bonus_amount": 10000.00
}'
```

This example demonstrates:
1. Transaction handling with proper rollback
2. Multiple table updates in a single atomic operation
3. Conditional updates with validation
4. Using SQL expressions for numeric operations
5. Proper error handling and status codes
6. Atomic commit or rollback

## Comparison with Raw SQL

```go
// Raw SQL - No type safety
db.Exec("UPDATE employees SET salary = ? WHERE id = ?", salary, id)

// Our approach - Full type safety
sqld.ExecuteUpdate[Employee](ctx, db, UpdateRequest{
    Set: map[string]interface{}{"salary": salary},
    Where: []Condition{{Field: "id", Operator: OpEqual, Value: id}},
})
```

The type-safe update system provides a robust foundation for database operations while preventing common errors and maintaining clean, maintainable code.
