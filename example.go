package sqld

import (
	"encoding/json"
)

// Example of how to use the package
func Example() {
	// Sample JSON query from user
	queryJSON := `{
		"select": ["id", "name"],
		"from": "users",
		"where": {
			"status": "active"
		}
	}`

	// Parse the JSON request
	var req QueryRequest
	if err := json.Unmarshal([]byte(queryJSON), &req); err != nil {
		// handle error
	}

	// Execute query and get results
	// resp, err := Execute(ctx, db, req)
	// if err != nil {
	// handle error
	// }

	// Convert response to JSON
	// jsonResponse, err := json.Marshal(resp)
	// if err != nil {
	// handle error
	// }

	// jsonResponse would be a string like this:
	/*
		{
			"data": [
				{
					"id": 1,
					"name": "John Doe"
				},
				{
					"id": 2,
					"name": "Jane Smith"
				}
			]
		}
	*/
}
