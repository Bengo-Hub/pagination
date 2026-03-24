// Package pagination provides shared request/response types and helpers
// for cursor-less offset pagination across all Go backend services.
package pagination

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

// ---------------------------------------------------------------------------
// Cursor-based (keyset) pagination
// ---------------------------------------------------------------------------

// CursorParams holds keyset/cursor pagination parameters.
type CursorParams struct {
	Cursor    string `json:"cursor"`    // opaque base64 cursor
	Limit     int    `json:"limit"`
	Direction string `json:"direction"` // "next" or "prev"
}

// CursorResponse holds a page of results with cursor metadata.
type CursorResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// ParseCursorParams extracts cursor pagination from an HTTP request.
func ParseCursorParams(r *http.Request) CursorParams {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > MaxLimit {
		limit = DefaultLimit
	}
	return CursorParams{
		Cursor:    r.URL.Query().Get("cursor"),
		Limit:     limit,
		Direction: r.URL.Query().Get("direction"),
	}
}

// EncodeCursor encodes an ID + timestamp into a base64 cursor string.
func EncodeCursor(id uuid.UUID, createdAt time.Time) string {
	data := fmt.Sprintf("%s|%d", id.String(), createdAt.UnixMicro())
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// DecodeCursor decodes a base64 cursor into ID + timestamp.
func DecodeCursor(cursor string) (uuid.UUID, time.Time, error) {
	data, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return uuid.Nil, time.Time{}, fmt.Errorf("invalid cursor: %w", err)
	}
	parts := strings.SplitN(string(data), "|", 2)
	if len(parts) != 2 {
		return uuid.Nil, time.Time{}, fmt.Errorf("malformed cursor")
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, time.Time{}, fmt.Errorf("invalid cursor id: %w", err)
	}
	micros, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return uuid.Nil, time.Time{}, fmt.Errorf("invalid cursor time: %w", err)
	}
	return id, time.UnixMicro(micros), nil
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
