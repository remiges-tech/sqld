package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/remiges-remiges/sqld"
	"github.com/remiges-remiges/sqld/examples/db/sqlc-gen"
)

// AdvancedQueryParams represents parameters for advanced employee queries
type AdvancedQueryParams struct {
	Department string  `db:"department" json:"department"`
	MinSalary  float64 `db:"min_salary" json:"min_salary"`
	MaxSalary  float64 `db:"max_salary" json:"max_salary"`
}

// DynamicQueryParams represents the parameters for dynamic query building
type DynamicQueryParams struct {
	Fields  []string `json:"fields"` // Fields to select
	Filters struct {
		Department *string  `json:"department,omitempty"`
		MinSalary  *float64 `json:"min_salary,omitempty"`
		MaxSalary  *float64 `json:"max_salary,omitempty"`
	} `json:"filters"`
}

// PaginatedDynamicQueryParams extends DynamicQueryParams with pagination and ordering
type PaginatedDynamicQueryParams struct {
	Fields  []string `json:"fields"` // Fields to select
	Filters struct {
		Department *string  `json:"department,omitempty"`
		MinSalary  *float64 `json:"min_salary,omitempty"`
		MaxSalary  *float64 `json:"max_salary,omitempty"`
	} `json:"filters"`
	Pagination struct {
		Limit  *int `json:"limit,omitempty"`  // Number of records to return
		Offset *int `json:"offset,omitempty"` // Number of records to skip
	} `json:"pagination"`
	OrderBy []struct {
		Field string `json:"field"` // Field to order by
		Desc  bool   `json:"desc"`  // If true, order descending
	} `json:"order_by,omitempty"`
}

// AdvancedQueryHandler demonstrates using multiple WHERE clauses
func (s *Server) AdvancedQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params AdvancedQueryParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, first_name, last_name, department, salary
		FROM employees
		WHERE department = {{department}}
		AND salary >= {{min_salary}}
		AND salary <= {{max_salary}}
		ORDER BY salary DESC
	`

	paramMap := map[string]interface{}{
		"department": params.Department,
		"min_salary": params.MinSalary,
		"max_salary": params.MaxSalary,
	}

	req := sqld.ExecuteRawRequest{
		Query:  query,
		Params: paramMap,
	}

	results, err := sqld.ExecuteRaw[AdvancedQueryParams, EmployeeRow](
		r.Context(),
		s.db,
		req,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Data []map[string]interface{} `json:"data"`
	}{
		Data: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdvancedSQLCHandler demonstrates using multiple WHERE clauses with SQLC types
func (s *Server) AdvancedSQLCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use the simple params structure for request parsing
	var requestParams AdvancedQueryParams
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to SQLC types
	params := sqlc.GetEmployeesAdvancedParams{
		Department: pgtype.Text{String: requestParams.Department, Valid: true},
		MinSalary:  pgtype.Numeric{Int: big.NewInt(int64(requestParams.MinSalary)), Valid: true},
		MaxSalary:  pgtype.Numeric{Int: big.NewInt(int64(requestParams.MaxSalary)), Valid: true},
	}

	query := `
		SELECT id, first_name, last_name, department, salary
		FROM employees
		WHERE department = {{department}}
		AND salary >= {{min_salary}}
		AND salary <= {{max_salary}}
		ORDER BY salary DESC
	`

	paramMap := map[string]interface{}{
		"department": params.Department,
		"min_salary": params.MinSalary,
		"max_salary": params.MaxSalary,
	}

	req := sqld.ExecuteRawRequest{
		Query:  query,
		Params: paramMap,
	}

	results, err := sqld.ExecuteRaw[sqlc.GetEmployeesAdvancedParams, sqlc.GetEmployeesAdvancedRow](
		r.Context(),
		s.db,
		req,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Data []map[string]interface{} `json:"data"`
	}{
		Data: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdvancedSQLCHandlerSimple demonstrates using multiple WHERE clauses with SQLC types
// but returns fewer fields in the result
func (s *Server) AdvancedSQLCHandlerSimple(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use the simple params structure for request parsing
	var requestParams AdvancedQueryParams
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to SQLC types
	params := sqlc.GetEmployeesAdvancedParams{
		Department: pgtype.Text{String: requestParams.Department, Valid: true},
		MinSalary:  pgtype.Numeric{Int: big.NewInt(int64(requestParams.MinSalary)), Valid: true},
		MaxSalary:  pgtype.Numeric{Int: big.NewInt(int64(requestParams.MaxSalary)), Valid: true},
	}

	query := `
		SELECT first_name, department 
		FROM employees
		WHERE department = {{department}}
		AND salary >= {{min_salary}}
		AND salary <= {{max_salary}}
		ORDER BY salary DESC
	`

	paramMap := map[string]interface{}{
		"department": params.Department,
		"min_salary": params.MinSalary,
		"max_salary": params.MaxSalary,
	}

	req := sqld.ExecuteRawRequest{
		Query:        query,
		Params:       paramMap,
		SelectFields: []string{"first_name", "department"}, // Specify fields to be returned
	}

	results, err := sqld.ExecuteRaw[sqlc.GetEmployeesAdvancedParams, sqlc.GetEmployeesAdvancedRow](
		r.Context(),
		s.db,
		req,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Data []map[string]interface{} `json:"data"`
	}{
		Data: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdvancedSQLCHandlerJoins demonstrates a complex query with joins and aliases
func (s *Server) AdvancedSQLCHandlerJoins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use the simple params structure for request parsing
	var requestParams AdvancedQueryParams
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to SQLC types
	params := sqlc.GetEmployeesWithAccountsParams{
		Department: pgtype.Text{String: requestParams.Department, Valid: true},
		MinSalary:  pgtype.Numeric{Int: big.NewInt(int64(requestParams.MinSalary)), Valid: true},
		MaxSalary:  pgtype.Numeric{Int: big.NewInt(int64(requestParams.MaxSalary)), Valid: true},
	}

	query := `
		SELECT 
			e.first_name as employee_name,
			e.department as dept,
			COALESCE(COUNT(a.id), 0) as account_count,
			COALESCE(SUM(a.balance), 0) as total_balance
		FROM employees e
		LEFT JOIN accounts a ON a.owner_id = e.id
		WHERE e.department = {{department}}
		AND e.salary >= {{min_salary}}
		AND e.salary <= {{max_salary}}
		GROUP BY e.first_name, e.department
		ORDER BY total_balance DESC
	`

	paramMap := map[string]interface{}{
		"department": params.Department,
		"min_salary": params.MinSalary,
		"max_salary": params.MaxSalary,
	}

	req := sqld.ExecuteRawRequest{
		Query:        query,
		Params:       paramMap,
		SelectFields: []string{"employee_name", "dept", "account_count", "total_balance"},
	}

	results, err := sqld.ExecuteRaw[sqlc.GetEmployeesWithAccountsParams, sqlc.GetEmployeesWithAccountsRow](
		r.Context(),
		s.db,
		req,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Data []map[string]interface{} `json:"data"`
	}{
		Data: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdvancedSQLCHandlerDynamic demonstrates dynamic query building with field selection
func (s *Server) AdvancedSQLCHandlerDynamic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var requestParams DynamicQueryParams
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and prepare fields
	availableFields := map[string]string{
		"employee_name": "e.first_name",
		"dept":          "e.department",
		"account_count": "COALESCE(COUNT(a.id), 0)",
		"total_balance": "COALESCE(SUM(a.balance), 0)",
	}

	// If no fields specified, use all available fields
	if len(requestParams.Fields) == 0 {
		for field := range availableFields {
			requestParams.Fields = append(requestParams.Fields, field)
		}
	}

	// Validate requested fields
	var selectFields []string
	for _, field := range requestParams.Fields {
		if expr, ok := availableFields[field]; ok {
			selectFields = append(selectFields, fmt.Sprintf("%s as %s", expr, field))
		} else {
			http.Error(w, fmt.Sprintf("Invalid field: %s", field), http.StatusBadRequest)
			return
		}
	}

	// Build WHERE clause and params
	var whereConditions []string
	paramMap := make(map[string]interface{})

	if requestParams.Filters.Department != nil {
		whereConditions = append(whereConditions, "e.department = {{department}}")
		paramMap["department"] = pgtype.Text{String: *requestParams.Filters.Department, Valid: true}
	}
	if requestParams.Filters.MinSalary != nil {
		whereConditions = append(whereConditions, "e.salary >= {{min_salary}}")
		paramMap["min_salary"] = pgtype.Numeric{Int: big.NewInt(int64(*requestParams.Filters.MinSalary)), Valid: true}
	}
	if requestParams.Filters.MaxSalary != nil {
		whereConditions = append(whereConditions, "e.salary <= {{max_salary}}")
		paramMap["max_salary"] = pgtype.Numeric{Int: big.NewInt(int64(*requestParams.Filters.MaxSalary)), Valid: true}
	}

	// Build the complete query
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Only include ORDER BY total_balance if it's in the selected fields
	orderByClause := ""
	for _, field := range requestParams.Fields {
		if field == "total_balance" {
			orderByClause = "ORDER BY total_balance DESC"
			break
		}
	}

	query := fmt.Sprintf(`
		SELECT 
			%s
		FROM employees e
		LEFT JOIN accounts a ON a.owner_id = e.id
		%s
		GROUP BY e.first_name, e.department
		%s
	`, strings.Join(selectFields, ", "),
		whereClause,
		orderByClause)

	req := sqld.ExecuteRawRequest{
		Query:        query,
		Params:       paramMap,
		SelectFields: requestParams.Fields,
	}

	results, err := sqld.ExecuteRaw[sqlc.GetEmployeesWithAccountsParams, sqlc.GetEmployeesWithAccountsRow](
		r.Context(),
		s.db,
		req,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Data []map[string]interface{} `json:"data"`
	}{
		Data: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AdvancedSQLCHandlerDynamicPaginated demonstrates dynamic query building with pagination
func (s *Server) AdvancedSQLCHandlerDynamicPaginated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var requestParams PaginatedDynamicQueryParams
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and prepare fields
	availableFields := map[string]string{
		"employee_name": "e.first_name",
		"dept":          "e.department",
		"account_count": "COALESCE(COUNT(a.id), 0)",
		"total_balance": "COALESCE(SUM(a.balance), 0)",
	}

	// If no fields specified, use all available fields
	if len(requestParams.Fields) == 0 {
		for field := range availableFields {
			requestParams.Fields = append(requestParams.Fields, field)
		}
	}

	// Validate requested fields
	var selectFields []string
	validFields := make(map[string]bool)
	for _, field := range requestParams.Fields {
		if expr, ok := availableFields[field]; ok {
			selectFields = append(selectFields, fmt.Sprintf("%s as %s", expr, field))
			validFields[field] = true
		} else {
			http.Error(w, fmt.Sprintf("Invalid field: %s", field), http.StatusBadRequest)
			return
		}
	}

	// Build WHERE clause and params
	var whereConditions []string
	paramMap := make(map[string]interface{})

	if requestParams.Filters.Department != nil {
		whereConditions = append(whereConditions, "e.department = {{department}}")
		paramMap["department"] = pgtype.Text{String: *requestParams.Filters.Department, Valid: true}
	}
	if requestParams.Filters.MinSalary != nil {
		whereConditions = append(whereConditions, "e.salary >= {{min_salary}}")
		paramMap["min_salary"] = pgtype.Numeric{Int: big.NewInt(int64(*requestParams.Filters.MinSalary)), Valid: true}
	}
	if requestParams.Filters.MaxSalary != nil {
		whereConditions = append(whereConditions, "e.salary <= {{max_salary}}")
		paramMap["max_salary"] = pgtype.Numeric{Int: big.NewInt(int64(*requestParams.Filters.MaxSalary)), Valid: true}
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY clause
	var orderClauses []string
	if len(requestParams.OrderBy) > 0 {
		for _, order := range requestParams.OrderBy {
			if !validFields[order.Field] {
				http.Error(w, fmt.Sprintf("Invalid order field: %s", order.Field), http.StatusBadRequest)
				return
			}
			direction := "ASC"
			if order.Desc {
				direction = "DESC"
			}
			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", order.Field, direction))
		}
	}
	orderByClause := ""
	if len(orderClauses) > 0 {
		orderByClause = "ORDER BY " + strings.Join(orderClauses, ", ")
	}

	// Build LIMIT and OFFSET
	limitOffsetClause := ""
	if requestParams.Pagination.Limit != nil {
		limitOffsetClause = fmt.Sprintf("LIMIT %d", *requestParams.Pagination.Limit)
		if requestParams.Pagination.Offset != nil {
			limitOffsetClause += fmt.Sprintf(" OFFSET %d", *requestParams.Pagination.Offset)
		}
	}

	// Build the complete query
	query := fmt.Sprintf(`
		SELECT 
			%s
		FROM employees e
		LEFT JOIN accounts a ON a.owner_id = e.id
		%s
		GROUP BY e.first_name, e.department
		%s
		%s
	`, strings.Join(selectFields, ", "),
		whereClause,
		orderByClause,
		limitOffsetClause)

	req := sqld.ExecuteRawRequest{
		Query:        query,
		Params:       paramMap,
		SelectFields: requestParams.Fields,
	}

	results, err := sqld.ExecuteRaw[sqlc.GetEmployeesWithAccountsParams, sqlc.GetEmployeesWithAccountsRow](
		r.Context(),
		s.db,
		req,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Data []map[string]interface{} `json:"data"`
	}{
		Data: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
