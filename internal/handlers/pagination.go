package handlers

import "github.com/gofiber/fiber/v3"

// PaginationParams holds pagination query parameters
type PaginationParams struct {
	Page   int `json:"page"` // 1-indexed page number (default: 1)
	Per    int `json:"per"`  // Items per page (default: 10, max: 100)
	Offset int `json:"-"`    // Calculated offset for SQL (not exposed in JSON)
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Per        int   `json:"per"`
	Total      int64 `json:"total"`       // Total items across all pages
	TotalPages int   `json:"total_pages"` // Calculated total pages
	HasMore    bool  `json:"has_more"`    // Whether more pages exist
}

// PaginatedResponse wraps any list response with pagination metadata
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// ParsePaginationParams extracts and validates pagination from request
func ParsePaginationParams(c fiber.Ctx) PaginationParams {
	page := max(fiber.Query[int](c, "page", 1), 1)
	per := min(max(fiber.Query[int](c, "per", 10), 1), 100)
	offset := (page - 1) * per

	return PaginationParams{
		Page:   page,
		Per:    per,
		Offset: offset,
	}
}

// BuildPaginationMeta creates pagination metadata from query results
func BuildPaginationMeta(params PaginationParams, total int64) PaginationMeta {
	var totalPages int
	if total > 0 && params.Per > 0 {
		totalPages = int((total + int64(params.Per) - 1) / int64(params.Per))
	}
	hasMore := params.Page < totalPages

	return PaginationMeta{
		Page:       params.Page,
		Per:        params.Per,
		Total:      total,
		TotalPages: totalPages,
		HasMore:    hasMore,
	}
}

// NewPaginatedResponse wraps data with pagination metadata
func NewPaginatedResponse(data interface{}, params PaginationParams, total int64) PaginatedResponse {
	return PaginatedResponse{
		Data:       data,
		Pagination: BuildPaginationMeta(params, total),
	}
}
