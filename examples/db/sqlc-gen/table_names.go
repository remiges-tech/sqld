package sqlc

// TableName implements the Model interface for GetEmployeesAdvancedRow
func (GetEmployeesAdvancedRow) TableName() string {
	return "employees"
}

// TableName implements the Model interface for GetEmployeesWithAccountsRow
func (GetEmployeesWithAccountsRow) TableName() string {
	return "employees"
}
