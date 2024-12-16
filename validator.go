package sqld

import "fmt"

type Validator interface {
	ValidateQuery(req QueryRequest, metadata ModelMetadata) error
}

type BasicValidator struct{}

func (v BasicValidator) ValidateQuery(req QueryRequest, metadata ModelMetadata) error {
	if len(req.Select) == 0 {
		return fmt.Errorf("select fields cannot be empty")
	}
	for _, field := range req.Select {
		if _, ok := metadata.Fields[field]; !ok {
			return fmt.Errorf("invalid field in select: %s", field)
		}
	}
	for whereField := range req.Where {
		if _, ok := metadata.Fields[whereField]; !ok {
			return fmt.Errorf("invalid field in where clause: %s", whereField)
		}
	}
	for _, orderBy := range req.OrderBy {
		if _, ok := metadata.Fields[orderBy.Field]; !ok {
			return fmt.Errorf("invalid field in order by clause: %s", orderBy.Field)
		}
	}
	if req.Limit != nil && *req.Limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	if req.Offset != nil && *req.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	return nil
}
