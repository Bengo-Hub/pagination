// Package pagination provides shared request/response types and helpers
// for cursor-less offset pagination across all Go backend services.
package pagination

import (
	"net/http"
	"strconv"
)

// Defaults and limits.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params holds parsed pagination parameters from an HTTP request.
type Params struct {
	Limit  int
	Offset int
	Page   int // 1-based; offset = (Page-1) * Limit
}

// Response wraps a paginated result set with metadata.
type Response[T any] struct {
	Data    []T  `json:"data"`
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Page    int  `json:"page"`
	HasMore bool `json:"hasMore"`
}

// NewResponse constructs a Response from a slice, total count, and params.
func NewResponse[T any](data []T, total int, p Params) Response[T] {
	if data == nil {
		data = []T{}
	}
	return Response[T]{
		Data:    data,
		Total:   total,
		Limit:   p.Limit,
		Page:    p.Page,
		HasMore: p.Offset+len(data) < total,
	}
}

// Parse extracts pagination parameters from query strings.
// Supported keys: "limit", "page", "offset".
// If both page and offset are provided, page takes precedence.
func Parse(r *http.Request) Params {
	q := r.URL.Query()

	limit := intParam(q.Get("limit"), DefaultLimit)
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	page := intParam(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Allow explicit offset to override page-based calculation.
	if q.Has("offset") && !q.Has("page") {
		offset = intParam(q.Get("offset"), 0)
		if offset < 0 {
			offset = 0
		}
		page = (offset / limit) + 1
	}

	return Params{Limit: limit, Offset: offset, Page: page}
}

func intParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
