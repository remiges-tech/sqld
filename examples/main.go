package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/remiges-tech/sqld"
	"github.com/remiges-tech/sqld/examples/db/sqlc-gen"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EmployeeIDScanner handles scanning for EmployeeID type
type EmployeeIDScanner struct {
	valid bool
	value EmployeeID
}

func (s *EmployeeIDScanner) Scan(src interface{}) error {
	log.Printf("EmployeeIDScanner.Scan called with: %v of type %T", src, src) // Debug print
	if src == nil {
		s.valid = false
		return nil
	}

	switch v := src.(type) {
	case int64:
		s.value = EmployeeID(v)
		s.valid = true
	case int32:
		s.value = EmployeeID(v)
		s.valid = true
	case int:
		s.value = EmployeeID(v)
		s.valid = true
	default:
		s.valid = false
		return fmt.Errorf("cannot scan type %T into EmployeeID", src)
	}
	return nil
}

func (s *EmployeeIDScanner) Value() interface{} {
	if !s.valid {
		return nil
	}
	return s.value
}

// EmployeeID is a custom type for employee IDs
type EmployeeID int64

// Value implements driver.Valuer
func (id EmployeeID) Value() (driver.Value, error) {
	return int64(id), nil
}

// Employee represents our database model matching the employees table
type Employee struct {
	ID         EmployeeID `json:"id" db:"id"`
	FirstName  string     `json:"first_name" db:"first_name"`
	LastName   string     `json:"last_name" db:"last_name"`
	Email      string     `json:"email" db:"email"`
	Phone      string     `json:"phone" db:"phone"`
	HireDate   time.Time  `json:"hire_date" db:"hire_date"`
	Salary     float64    `json:"salary" db:"salary"`
	Department string     `json:"department" db:"department"`
	Position   string     `json:"position" db:"position"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

func (Employee) TableName() string {
	return "employees"
}

// Account represents our database model matching the accounts table
type Account struct {
	ID            int64     `json:"id" db:"id"`
	AccountNumber string    `json:"account_number" db:"account_number"`
	AccountName   string    `json:"account_name" db:"account_name"`
	AccountType   string    `json:"account_type" db:"account_type"`
	Balance       float64   `json:"balance" db:"balance"`
	Currency      string    `json:"currency" db:"currency"`
	Status        string    `json:"status" db:"status"`
	OwnerID       *int64    `json:"owner_id" db:"owner_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func (Account) TableName() string {
	return "accounts"
}

// EmployeeWithAccounts represents the result of employee-account queries
type EmployeeWithAccounts struct {
	EmployeeName string  `db:"first_name" json:"employee_name"`
	Dept         string  `db:"department" json:"dept"`
	AccountCount int64   `db:"id" json:"account_count"`
	TotalBalance float64 `db:"salary" json:"total_balance"`
}

func (e EmployeeWithAccounts) TableName() string {
	return "employees"
}

// AccountQueryParams represents parameters for account queries
type AccountQueryParams struct {
	MinBalance float64 `db:"min_balance" json:"min_balance"`
}

type Server struct {
	db *pgxpool.Pool
}

func NewServer(db *pgxpool.Pool) *Server {
	return &Server{db: db}
}

// DynamicQueryHandler demonstrates dynamic field selection and filtering
func (s *Server) DynamicQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req sqld.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := sqld.Execute[EmployeeWithAccounts](r.Context(), s.db, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// PaginatedQueryHandler demonstrates pagination with dynamic queries
func (s *Server) PaginatedQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req sqld.QueryRequest
	req.Pagination = &sqld.PaginationRequest{
		Page:     1,
		PageSize: 10,
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := sqld.Execute[Employee](r.Context(), s.db, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SQLCQueryHandler demonstrates integration with SQLC-generated types
func (s *Server) SQLCQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req sqld.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := sqld.Execute[sqlc.Employee](r.Context(), s.db, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type SimpleQueryParams struct {
	Department string  `db:"department" json:"department"`
	MinSalary  float64 `db:"min_salary" json:"min_salary"`
}

type EmployeeRow struct {
	ID         int64   `db:"id" json:"id"`
	FirstName  string  `db:"first_name" json:"first_name"`
	LastName   string  `db:"last_name" json:"last_name"`
	Department string  `db:"department" json:"department"`
	Salary     float64 `db:"salary" json:"salary"`
}

func (EmployeeRow) TableName() string {
	return "employees"
}

var (
	// Example 1: Basic equality and comparison
	basicQuery = sqld.QueryRequest{
		Select: []string{"id", "first_name", "salary"},
		Where: []sqld.Condition{
			{
				Field:    "department",
				Operator: sqld.OpEqual,
				Value:    "Engineering",
			},
			{
				Field:    "salary",
				Operator: sqld.OpGreaterThan,
				Value:    75000.0,
			},
		},
	}

	// Example 2: Pattern matching and NULL checks
	patternQuery = sqld.QueryRequest{
		Select: []string{"id", "email", "phone"},
		Where: []sqld.Condition{
			{
				Field:    "email",
				Operator: sqld.OpLike,
				Value:    "%@company.com",
			},
			{
				Field:    "phone",
				Operator: sqld.OpIsNotNull,
				Value:    nil,
			},
		},
	}

	// Example 3: IN clause and range comparison
	rangeQuery = sqld.QueryRequest{
		Select: []string{"id", "department", "position"},
		Where: []sqld.Condition{
			{
				Field:    "department",
				Operator: sqld.OpIn,
				Value:    []string{"Engineering", "Sales", "Marketing"},
			},
			{
				Field:    "salary",
				Operator: sqld.OpGreaterThanOrEqual,
				Value:    50000.0,
			},
			{
				Field:    "salary",
				Operator: sqld.OpLessThanOrEqual,
				Value:    100000.0,
			},
		},
	}
)

func (s *Server) ExampleQueriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Execute each example query
	basicResp, err := sqld.Execute[Employee](r.Context(), s.db, basicQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Basic query error: %v", err), http.StatusInternalServerError)
		return
	}

	patternResp, err := sqld.Execute[Employee](r.Context(), s.db, patternQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Pattern query error: %v", err), http.StatusInternalServerError)
		return
	}

	rangeResp, err := sqld.Execute[Employee](r.Context(), s.db, rangeQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Range query error: %v", err), http.StatusInternalServerError)
		return
	}

	// Combine all results
	response := struct {
		BasicQuery   interface{} `json:"basic_query"`
		PatternQuery interface{} `json:"pattern_query"`
		RangeQuery   interface{} `json:"range_query"`
	}{
		BasicQuery:   basicResp,
		PatternQuery: patternResp,
		RangeQuery:   rangeResp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func init() {
	// Keep only the scanner registration as it's still needed
	sqld.RegisterScanner(reflect.TypeOf((*EmployeeID)(nil)).Elem(), func() sql.Scanner {
		return &EmployeeIDScanner{}
	})
}

func main() {
	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig("postgres://alyatest:alyatest@localhost:5432/alyatest?sslmode=disable")
	if err != nil {
		log.Fatal("Unable to parse pool config")
	}

	// Add AfterConnect hook to register enums
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return sqld.AutoRegisterEnums(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := sqld.Register[Employee](); err != nil {
		log.Fatal(err)
	}
	if err := sqld.Register[Account](); err != nil {
		log.Fatal(err)
	}

	server := NewServer(pool)

	// Basic CRUD endpoints
	http.HandleFunc("/api/dynamic", server.DynamicQueryHandler)       // Dynamic field selection
	http.HandleFunc("/api/paginated", server.PaginatedQueryHandler)   // Pagination example
	http.HandleFunc("/api/sqlc", server.SQLCQueryHandler)             // SQLC integration

	// Advanced query handlers
	http.HandleFunc("/api/employees/advanced", server.AdvancedQueryHandler)                  // Multiple WHERE clauses example
	http.HandleFunc("/api/employees/advanced-sqlc", server.AdvancedSQLCHandler)              // Multiple WHERE clauses with SQLC types
	http.HandleFunc("/api/employees/advanced-sqlc-simple", server.AdvancedSQLCHandlerSimple) // Simple fields example
	http.HandleFunc("/api/employees/advanced-sqlc-joins", server.AdvancedSQLCHandlerJoins)   // Complex join query example
	http.HandleFunc("/api/employees/advanced-sqlc-dynamic", server.AdvancedSQLCHandlerDynamic)
	http.HandleFunc("/api/employees/advanced-sqlc-dynamic-paginated", server.AdvancedSQLCHandlerDynamicPaginated)

	http.HandleFunc("/examples", server.ExampleQueriesHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
