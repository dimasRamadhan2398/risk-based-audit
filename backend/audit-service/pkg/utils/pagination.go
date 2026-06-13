package utils

import (
	"strconv"
)

type Pagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Order    string `json:"order"` // asc, desc
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

// NewPagination creates a new pagination with defaults
func NewPagination(page, pageSize int, sort, order string) *Pagination {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}
}

// GetOffset calculates the offset for database query
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size
func (p *Pagination) GetLimit() int {
	return p.PageSize
}

// BuildPaginationResponse builds pagination response metadata
func BuildPaginationResponse(page, pageSize int, totalCount int64) *PaginationResponse {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

// ParsePagination parses pagination from query string
func ParsePagination(pageStr, pageSizeStr, sort, order string) (*Pagination, error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	return NewPagination(page, pageSize, sort, order), nil
}
