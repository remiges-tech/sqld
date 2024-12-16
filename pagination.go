package sqld

import "math"

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// ValidatePagination validates and normalizes pagination parameters
func ValidatePagination(req *PaginationRequest) *PaginationRequest {
	if req == nil {
		return &PaginationRequest{
			Page:     1,
			PageSize: DefaultPageSize,
		}
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = DefaultPageSize
	}
	if req.PageSize > MaxPageSize {
		req.PageSize = MaxPageSize
	}

	return req
}

// CalculateOffset converts page/pageSize into offset
func CalculateOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(totalItems, pageSize, currentPage int) *PaginationResponse {
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	return &PaginationResponse{
		Page:       currentPage,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// HasNextPage checks if there is a next page
func HasNextPage(totalItems, pageSize, currentPage int) bool {
	return CalculatePagination(totalItems, pageSize, currentPage).TotalPages > currentPage
}

// HasPreviousPage checks if there is a previous page
func HasPreviousPage(currentPage int) bool {
	return currentPage > 1
}

// GetNextPage returns the next page number
func GetNextPage(currentPage int) int {
	return currentPage + 1
}

// GetPreviousPage returns the previous page number
func GetPreviousPage(currentPage int) int {
	return currentPage - 1
}
