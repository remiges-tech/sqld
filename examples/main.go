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

// AccountQueryParams represents parameters for account queries
type AccountQueryParams struct {
	MinBalance float64 `db:"min_balance" json:"min_balance"`
}

type Server struct {
	db *pgx.Conn
}

func NewServer(db *pgx.Conn) *Server {
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

	resp, err := sqld.Execute[Employee](r.Context(), s.db, req)
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

// CustomFilterHandler demonstrates using custom WHERE conditions
func (s *Server) CustomFilterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req sqld.QueryRequest
	req.Select = []string{"id", "account_number", "balance", "status"}
	req.Where = map[string]interface{}{
		"status":  "active",
		"balance": 1000.00,
	}

	resp, err := sqld.Execute[Account](r.Context(), s.db, req)
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

func init() {
	sqld.RegisterScanner(reflect.TypeOf((*EmployeeID)(nil)).Elem(), func() sql.Scanner {
		return &EmployeeIDScanner{}
	})

	if err := sqld.Register(sqlc.Employee{}); err != nil {
		log.Fatalf("failed to register sqlc.Employee model: %v", err)
	}
}

func main() {
	ctx := context.Background()

	config, err := pgx.ParseConfig("postgres://alyatest:alyatest@localhost:5432/alyatest?sslmode=disable")
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := sqld.Register(Employee{}); err != nil {
		log.Fatal(err)
	}
	if err := sqld.Register(Account{}); err != nil {
		log.Fatal(err)
	}
	if err := sqld.Register(sqlc.Employee{}); err != nil {
		log.Fatal(err)
	}

	server := NewServer(conn)

	// Basic CRUD endpoints
	http.HandleFunc("/api/dynamic", server.DynamicQueryHandler)       // Dynamic field selection
	http.HandleFunc("/api/paginated", server.PaginatedQueryHandler)   // Pagination example
	http.HandleFunc("/api/custom-filter", server.CustomFilterHandler) // Custom filtering
	http.HandleFunc("/api/sqlc", server.SQLCQueryHandler)             // SQLC integration

	// Advanced query handlers
	http.HandleFunc("/api/employees/advanced", server.AdvancedQueryHandler)                  // Multiple WHERE clauses example
	http.HandleFunc("/api/employees/advanced-sqlc", server.AdvancedSQLCHandler)              // Multiple WHERE clauses with SQLC types
	http.HandleFunc("/api/employees/advanced-sqlc-simple", server.AdvancedSQLCHandlerSimple) // Simple fields example
	http.HandleFunc("/api/employees/advanced-sqlc-joins", server.AdvancedSQLCHandlerJoins)   // Complex join query example
	http.HandleFunc("/api/employees/advanced-sqlc-dynamic", server.AdvancedSQLCHandlerDynamic)
	http.HandleFunc("/api/employees/advanced-sqlc-dynamic-paginated", server.AdvancedSQLCHandlerDynamicPaginated)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
