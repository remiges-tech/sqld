package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/remiges-tech/sqld"
	"github.com/remiges-tech/sqld/examples/db/sqlc-gen"
)

// DepartmentTransferRequest represents the request for transferring an employee to a new department
type DepartmentTransferRequest struct {
	EmployeeID       int64   `json:"employee_id"`
	NewDepartment    string  `json:"new_department"`
	SalaryAdjustment float64 `json:"salary_adjustment"`
	TransferBonus    float64 `json:"transfer_bonus"`
}

// TransferEmployeeHandler demonstrates mixing SQLC and SQLD operations in a transaction
func (s *Server) TransferEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Parse request
	var req DepartmentTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// 1. Get employee details using SQLC
	emp, err := sqlc.New(tx).GetEmployee(ctx, req.EmployeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get employee: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert pgtype.Numeric to float64 for calculation
	salaryValue, err := emp.Salary.Float64Value()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get current salary value: %v", err), http.StatusInternalServerError)
		return
	}
	currentSalary := salaryValue.Float64

	// 2. Update employee department first using SQLD
	empDeptUpdate := sqld.UpdateRequest{
		Set: map[string]interface{}{
			"department": req.NewDepartment,
		},
		Where: []sqld.Condition{
			{Field: "id", Operator: sqld.OpEqual, Value: req.EmployeeID},
			{Field: "is_active", Operator: sqld.OpEqual, Value: true},
		},
	}

	empRows, err := sqld.ExecuteUpdate[Employee](ctx, tx, empDeptUpdate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee department: %v", err), http.StatusInternalServerError)
		return
	}
	if empRows == 0 {
		http.Error(w, "Employee not found or not active", http.StatusNotFound)
		return
	}

	// 3. Update employee salary using SQLD
	newSalary := currentSalary + req.SalaryAdjustment
	empSalaryUpdate := sqld.UpdateRequest{
		Set: map[string]interface{}{
			"salary": newSalary,
		},
		Where: []sqld.Condition{
			{Field: "id", Operator: sqld.OpEqual, Value: req.EmployeeID},
			{Field: "is_active", Operator: sqld.OpEqual, Value: true},
		},
	}

	empRows, err = sqld.ExecuteUpdate[Employee](ctx, tx, empSalaryUpdate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee salary: %v", err), http.StatusInternalServerError)
		return
	}
	if empRows == 0 {
		http.Error(w, "Employee not found or not active", http.StatusNotFound)
		return
	}

	// 4. Get employee's accounts using SQLC
	ownerID := pgtype.Int8{Int64: req.EmployeeID, Valid: true}
	accounts, err := sqlc.New(tx).GetEmployeeAccounts(ctx, ownerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get accounts: %v", err), http.StatusInternalServerError)
		return
	}

	// 5. Update account balance with transfer bonus using SQLD
	if req.TransferBonus > 0 && len(accounts) > 0 {
		// Update all active accounts for the employee
		for _, account := range accounts {
			balanceValue, err := account.Balance.Float64Value()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get current balance value: %v", err), http.StatusInternalServerError)
				return
			}
			currentBalance := balanceValue.Float64
			newBalance := currentBalance + req.TransferBonus

			accUpdate := sqld.UpdateRequest{
				Set: map[string]interface{}{
					"balance": newBalance,
				},
				Where: []sqld.Condition{
					{Field: "id", Operator: sqld.OpEqual, Value: account.ID},
					{Field: "status", Operator: sqld.OpEqual, Value: "active"},
				},
			}

			accRows, err := sqld.ExecuteUpdate[Account](ctx, tx, accUpdate)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to update account: %v", err), http.StatusInternalServerError)
				return
			}
			if accRows == 0 {
				// Log warning but don't fail - account might have been deactivated
				log.Printf("Warning: Failed to update account %d for employee %d", account.ID, req.EmployeeID)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message":             "Transfer successful",
		"new_department":      req.NewDepartment,
		"new_salary":          newSalary,
		"transfer_bonus":      req.TransferBonus,
		"previous_department": emp.Department,
		"previous_salary":     currentSalary,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
