package sqld

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/jackc/pgx/v5"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type fieldInfo struct {
	jsonKey   string
	goType    reflect.Type
	fieldName string
}

// BuildMetadataMap uses reflection on the model struct to map db tags to fieldInfo.
// It extracts the 'db' and 'json' tags from the struct fields and creates a map
// where the key is the 'db' tag and the value is a fieldInfo struct containing the
// 'json' tag and the Go type of the field. This map is used later in the ExecuteRaw
// function to map database column names to JSON keys in the result.
func BuildMetadataMap[T any]() (map[string]fieldInfo, error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct")
	}

	metaMap := make(map[string]fieldInfo)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		jsonTag := field.Tag.Get("json")
		if dbTag != "" && jsonTag != "" {
			metaMap[dbTag] = fieldInfo{
				jsonKey:   jsonTag,
				goType:    field.Type,
				fieldName: field.Name,
			}
		}
	}
	return metaMap, nil
}

// isTypeCompatible checks if the runtime type of a value matches the expected type.
// It returns true if the value's type is compatible with the expected type,
// and false otherwise. It also handles the case where the expected type is an
// empty interface, in which case any type is considered compatible.
func isTypeCompatible(valType, expectedType reflect.Type) bool {
	if valType == nil || expectedType == nil {
		return false
	}

	// If the expected type is an empty interface, accept any type.
	if expectedType.Kind() == reflect.Interface && expectedType.NumMethod() == 0 {
		// This means expectedType is `interface{}`
		return true
	}

	return valType == expectedType
}

func typeNameOrNil(t reflect.Type) string {
	if t == nil {
		return "nil"
	}
	return t.String()
}

// Named parameter regex to find patterns like {{param_name}}
var namedParamRegex = regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)

// ExtractNamedPlaceholders finds all named parameters in the {{param_name}} format.
func ExtractNamedPlaceholders(query string) ([]string, error) {
	matches := namedParamRegex.FindAllStringSubmatch(query, -1)
	var params []string
	seen := make(map[string]bool)
	for _, match := range matches {
		paramName := match[1]
		if !seen[paramName] {
			seen[paramName] = true
			params = append(params, paramName)
		}
	}
	return params, nil
}

// ReplaceNamedWithDollarPlaceholders replaces {{param_name}} with $1, $2, ...
func ReplaceNamedWithDollarPlaceholders(query string, queryParams []string) (string, error) {
	for i, p := range queryParams {
		placeholder := fmt.Sprintf("{{%s}}", p)
		newPlaceholder := fmt.Sprintf("$%d", i+1)
		query = strings.ReplaceAll(query, placeholder, newPlaceholder)
	}
	return query, nil
}

// ValidateMapParamsAgainstStructNamed ensures the params map matches the expected types from P.
// It uses the isTypeCompatible function to check if the type of each parameter in the map
// matches the expected type from P. This is primarily to prevent runtime errors due to type mismatches.
func ValidateMapParamsAgainstStructNamed[P any](
	paramMap map[string]interface{},
	queryParams []string,
) ([]interface{}, error) {
	t := reflect.TypeOf((*P)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct")
	}

	typeByName := make(map[string]reflect.Type)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		jsonTag := field.Tag.Get("json")

		// Validate that all fields with db tag must have json tag
		if dbTag != "" && jsonTag == "" {
			return nil, fmt.Errorf("field %s has db tag but missing json tag", field.Name)
		}

		if dbTag != "" {
			typeByName[dbTag] = field.Type
		}
	}

	args := make([]interface{}, 0, len(queryParams))
	for _, p := range queryParams {
		expectedType, found := typeByName[p]
		if !found {
			return nil, fmt.Errorf("no type info for param %s", p)
		}

		val, present := paramMap[p]
		if !present {
			// If the parameter is optional and not present, append nil or handle as needed
			args = append(args, nil)
			continue
		}

		valType := reflect.TypeOf(val)
		if !isTypeCompatible(valType, expectedType) {
			return nil, fmt.Errorf("parameter %s type mismatch: got %s, want %s",
				p, typeNameOrNil(valType), typeNameOrNil(expectedType))
		}

		args = append(args, val)
	}

	return args, nil
}

// validateQueryParams checks if all parameters in the query have corresponding values in paramMap
func validateQueryParams(query string, paramMap map[string]interface{}) error {
	// Find all parameters in the query using regex
	re := regexp.MustCompile(`{{(\w+)}}`)
	matches := re.FindAllStringSubmatch(query, -1)

	// Create a set of required parameters from the query
	requiredParams := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			requiredParams[match[1]] = true
		}
	}

	// Check if all required parameters are in paramMap
	var missingParams []string
	for param := range requiredParams {
		if _, exists := paramMap[param]; !exists {
			missingParams = append(missingParams, param)
		}
	}

	// Check if there are any extra parameters in paramMap that aren't in the query
	var extraParams []string
	for param := range paramMap {
		if !requiredParams[param] {
			extraParams = append(extraParams, param)
		}
	}

	if len(missingParams) > 0 || len(extraParams) > 0 {
		var errMsg string
		if len(missingParams) > 0 {
			errMsg += fmt.Sprintf("missing required parameters in paramMap: %v", missingParams)
		}
		if len(extraParams) > 0 {
			if errMsg != "" {
				errMsg += "; "
			}
			errMsg += fmt.Sprintf("extra unused parameters in paramMap: %v", extraParams)
		}
		return fmt.Errorf(errMsg)
	}

	return nil
}

// validateSQLSyntax uses pg_query to validate SQL syntax and structure
func validateSQLSyntax(query string) error {
	result, err := pg_query.Parse(query)
	if err != nil {
		return fmt.Errorf("SQL syntax error: %w", err)
	}

	if len(result.Stmts) == 0 {
		return fmt.Errorf("empty SQL query")
	}

	// Get the first statement
	stmt := result.Stmts[0].Stmt

	// Check if it's a SELECT statement
	selectStmt := stmt.GetSelectStmt()
	if selectStmt == nil {
		return fmt.Errorf("only SELECT statements are allowed")
	}

	return nil
}

// ExecuteRawRequest contains all parameters needed for ExecuteRaw
type ExecuteRawRequest struct {
	Query        string                 // SQL query with {{param_name}} placeholders
	Params       map[string]interface{} // Parameter values mapped to placeholder names
	SelectFields []string               // List of fields to be returned in the result
}

// ExecuteRaw executes a dynamic SQL query with named parameters and returns the results as a slice of maps.
// It provides parameter validation and safe query execution with PostgreSQL.
//
// Generic Type Parameters:
//   - P: Parameter struct type that defines the expected query parameters.
//     Must be a struct with both `db` and `json` tags for each field used as a parameter.
//     The `db` tags must match the {{param_name}} placeholders in the query.
//   - R: Result struct type used for scanning database rows.
//     Must be a struct with `db` tags for fields that match the column names in the query.
//
// The function performs the following steps in exact order:
//
//  1. Initial Parameter Validation:
//     - Validates that P is a struct type
//     - Finds {{param}} placeholders in query using regex
//     - Validates all placeholders have values in Params map
//     - Validates no extra unused parameters in Params map
//
//  2. Parameter Processing:
//     - Extracts unique parameter names from query
//     - Builds type mapping from parameter struct fields
//     - During type mapping, validates fields have both db and json tags
//     - Validates parameter values have exactly matching types with struct fields
//     (except interface{} fields which accept any type)
//
//  3. Query Processing:
//     - Replaces {{param}} placeholders with $N positional parameters
//     - Validates modified SQL using PostgreSQL parser
//     - Verifies query is a SELECT statement
//
//  4. Result Setup:
//     - Validates that R is a struct type
//     - Builds metadata map from R struct's db tags
//
//  5. Query Execution:
//     - Executes query with positional parameters
//     - Uses scany's sqlscan/pgxscan to scan results into R structs
//
//  6. Result Processing:
//     - Converts struct fields to map entries
//     - If SelectFields is empty, includes all fields with db tags
//     - If SelectFields is provided, only includes fields whose db tags match
//
// Usage:
//
//	type QueryParams struct {
//	    Department string  `db:"department" json:"department"`
//	    MinSalary int     `db:"min_salary" json:"min_salary"`
//	}
//
//	type Employee struct {
//	    ID        int     `db:"id"`
//	    Name      string  `db:"name"`
//	    Salary    int     `db:"salary"`
//	}
//
//	req := ExecuteRawRequest{
//	    Query: `
//	        SELECT id, name, salary
//	        FROM employees
//	        WHERE department = {{department}}
//	        AND salary >= {{min_salary}}
//	    `,
//	    Params: map[string]interface{}{
//	        "department": "Engineering",
//	        "min_salary": 50000,
//	    },
//	    SelectFields: []string{"name", "salary"}, // Optional: filters output map fields by db tag names
//	}
//
//	results, err := ExecuteRaw[QueryParams, Employee](ctx, db, req)
//
// The function may return the following errors:
//   - Type errors if P or R are not structs
//   - Missing or extra parameter errors
//   - Missing json tag errors during parameter struct field processing
//   - Parameter type mismatch errors
//   - SQL syntax errors after parameter substitution
//   - Non-SELECT statement errors
//   - Database query execution errors
//   - Row scanning errors
//
// The function supports both *sql.DB and *pgx.Conn database connections through scany's
// sqlscan and pgxscan packages.
func ExecuteRaw[P any, R Model](
	ctx context.Context,
	db interface{},
	req ExecuteRawRequest,
) ([]map[string]interface{}, error) {
	// Validate that all query parameters have corresponding values
	if err := validateQueryParams(req.Query, req.Params); err != nil {
		return nil, err
	}

	// Extract named placeholders
	queryParams, err := ExtractNamedPlaceholders(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to extract named placeholders: %w", err)
	}

	// Validate and convert map params to arguments in correct order
	args, err := ValidateMapParamsAgainstStructNamed[P](req.Params, queryParams)
	if err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Replace named placeholders with $N placeholders
	finalQuery, err := ReplaceNamedWithDollarPlaceholders(req.Query, queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to replace named placeholders: %w", err)
	}

	// Validate SQL syntax
	if err := validateSQLSyntax(finalQuery); err != nil {
		return nil, err
	}

	// Get metadata from registry for result type
	var result R
	metadata, err := getModelMetadata(result)
	if err != nil {
		return nil, fmt.Errorf("failed to get model metadata: %w", err)
	}

	// Execute query and scan into slice of structs first to handle custom types
	var structResults []R
	switch db := db.(type) {
	case *sql.DB:
		if err := sqlscan.Select(ctx, db, &structResults, finalQuery, args...); err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
	case *pgx.Conn:
		if err := pgxscan.Select(ctx, db, &structResults, finalQuery, args...); err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	// Convert struct results to maps with only requested fields
	results := make([]map[string]interface{}, len(structResults))
	for i, row := range structResults {
		val := reflect.ValueOf(row)
		resultMap := make(map[string]interface{})

		// Only include fields that were specified in SelectFields
		for jsonName, field := range metadata.Fields {
			// If SelectFields is empty, include all fields
			// Otherwise, only include fields that were requested
			if len(req.SelectFields) == 0 || contains(req.SelectFields, field.Name) {
				fieldVal := val.FieldByName(field.Name)
				if fieldVal.IsValid() {
					resultMap[jsonName] = fieldVal.Interface()
				}
			}
		}
		results[i] = resultMap
	}

	return results, nil
}

// contains checks if a string is present in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
