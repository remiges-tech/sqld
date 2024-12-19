package sqlc

// TableName methods for parameter types
func (GetEmployeesAdvancedParams) TableName() string {
	return "employees_advanced_params"
}

func (GetEmployeesWithAccountsParams) TableName() string {
	return "employees_with_accounts_params"
}

func (UCCListParams) TableName() string {
	return "ucc_list_params"
}

// TableName methods for result types
func (GetEmployeesAdvancedRow) TableName() string {
	return "employees"
}

func (GetEmployeesWithAccountsRow) TableName() string {
	return "employees"
}

func (UCCListRow) TableName() string {
	return "ucc"
}
