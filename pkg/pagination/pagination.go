// Package pagination provides utilities for working with paginated queries.
package pagination

import (
	"math"
)

// Defaults for pagination.
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Request represents a pagination request.
type Request struct {
	Page     int32
	PageSize int32
}

// Response represents a paginated response.
type Response struct {
	Page      int32
	PageSize  int32
	Total     int64
	TotalPage int32
}

// NewRequest creates a Request with defaults applied.
func NewRequest(page, pageSize int32) Request {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		pageSize = DefaultPageSize
	}
	return Request{Page: page, PageSize: pageSize}
}

// Offset returns the SQL OFFSET value.
func (r Request) Offset() int32 {
	return (r.Page - 1) * r.PageSize
}

// Limit returns the SQL LIMIT value.
func (r Request) Limit() int32 {
	return r.PageSize
}

// NewResponse builds a Response from a Request and the total count.
func NewResponse(req Request, total int64) Response {
	totalPage := int32(math.Ceil(float64(total) / float64(req.PageSize)))
	return Response{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Total:     total,
		TotalPage: totalPage,
	}
}
