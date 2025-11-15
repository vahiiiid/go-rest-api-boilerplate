package errors

import "time"

// Response wraps all API responses with consistent structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	Path       string      `json:"path,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	RetryAfter *int        `json:"retry_after,omitempty"`
}

// Meta contains response metadata for pagination and tracking
type Meta struct {
	RequestID  string    `json:"request_id,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	Page       int       `json:"page,omitempty"`
	PerPage    int       `json:"per_page,omitempty"`
	Total      int64     `json:"total,omitempty"`
	TotalPages int       `json:"total_pages,omitempty"`
	Links      *Links    `json:"links,omitempty"`
}

// Links provides HATEOAS navigation links
type Links struct {
	Self  string `json:"self,omitempty"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
}

// Success creates a successful response with data
func Success(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// SuccessWithMeta creates a successful response with data and metadata
func SuccessWithMeta(data interface{}, meta *Meta) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}
