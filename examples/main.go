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
			{Field: "department", Operator: sqld.OpEqual, Value: "Engineering"},
			{Field: "salary", Operator: sqld.OpGreaterThan, Value: 50000},
		},
	}

	// Example 2: Multiple conditions with IN clause
	complexQuery = sqld.QueryRequest{
		Select: []string{"id", "first_name", "last_name", "department", "salary"},
		Where: []sqld.Condition{
			{Field: "department", Operator: sqld.OpIn, Value: []string{"Engineering", "Marketing"}},
			{Field: "salary", Operator: sqld.OpGreaterThanOrEqual, Value: 60000},
		},
	}

	// Example 3: Pattern matching with LIKE
	patternQuery = sqld.QueryRequest{
		Select: []string{"id", "first_name", "last_name", "email"},
		Where: []sqld.Condition{
			{Field: "email", Operator: sqld.OpLike, Value: "%@example.com"},
		},
	}

	// Example 4: Partial update example
	updateExample = sqld.UpdateRequest{
		Set: map[string]interface{}{
			"salary": 85000.0,
			"position": "Senior Software Engineer",
		},
		Where: []sqld.Condition{
			{Field: "department", Operator: sqld.OpEqual, Value: "Engineering"},
			{Field: "salary", Operator: sqld.OpLessThan, Value: 80000.0},
		},
	}
)

// UpdateEmployeeHandler demonstrates partial updates with validation
func (s *Server) UpdateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req sqld.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rowsAffected, err := sqld.ExecuteUpdate[Employee](ctx, s.db, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"rows_affected": rowsAffected,
		"message":       "Update successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExampleQueriesHandler demonstrates various query examples
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

	rangeResp, err := sqld.Execute[Employee](r.Context(), s.db, complexQuery)
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

// PromotionRequest represents the request for promoting an employee
type PromotionRequest struct {
	EmployeeID   int64   `json:"employee_id"`
	NewPosition  string  `json:"new_position"`
	SalaryRaise  float64 `json:"salary_raise"`
	BonusAmount  float64 `json:"bonus_amount"`
}

// PromoteEmployeeHandler handles employee promotions with account updates in a transaction
func (s *Server) PromoteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Parse request
	var promReq PromotionRequest
	if err := json.NewDecoder(r.Body).Decode(&promReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	// 1. Update employee position and salary
	empUpdate := sqld.UpdateRequest{
		Set: map[string]interface{}{
			"position": promReq.NewPosition,
			"salary":   promReq.SalaryRaise,
		},
		Where: []sqld.Condition{
			{Field: "id", Operator: sqld.OpEqual, Value: promReq.EmployeeID},
			{Field: "is_active", Operator: sqld.OpEqual, Value: true},
		},
	}

	empRows, err := sqld.ExecuteUpdate[Employee](ctx, tx, empUpdate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee: %v", err), http.StatusInternalServerError)
		return
	}
	if empRows == 0 {
		http.Error(w, "Employee not found or not active", http.StatusNotFound)
		return
	}

	// 2. Update account balance with bonus using raw SQL expression
	accUpdate := sqld.UpdateRequest{
		Set: map[string]interface{}{
			"balance": fmt.Sprintf("balance + %f", promReq.BonusAmount), // Using raw SQL expression
		},
		Where: []sqld.Condition{
			{Field: "owner_id", Operator: sqld.OpEqual, Value: promReq.EmployeeID},
			{Field: "status", Operator: sqld.OpEqual, Value: "active"},
		},
	}

	accRows, err := sqld.ExecuteUpdate[Account](ctx, tx, accUpdate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update account: %v", err), http.StatusInternalServerError)
		return
	}
	if accRows == 0 {
		http.Error(w, "No active account found for employee", http.StatusNotFound)
		return
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message":         "Promotion successful",
		"employee_updated": empRows > 0,
		"account_updated": accRows > 0,
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

	http.HandleFunc("/query", server.DynamicQueryHandler)
	http.HandleFunc("/paginated", server.PaginatedQueryHandler)
	http.HandleFunc("/sqlc", server.SQLCQueryHandler)
	http.HandleFunc("/examples", server.ExampleQueriesHandler)
	http.HandleFunc("/employees/update", server.UpdateEmployeeHandler)
	http.HandleFunc("/employees/promote", server.PromoteEmployeeHandler)
	http.HandleFunc("/employees/transfer", server.TransferEmployeeHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
