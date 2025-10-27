package errors

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// APIError represents a structured API error with code, message, details and HTTP status.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Status  int    `json:"-"`
}

// RateLimitError extends APIError with retry-after information for rate limiting.
type RateLimitError struct {
	APIError
	RetryAfter int `json:"retry_after"`
}

func (e *APIError) Error() string {
	return e.Message
}

// NotFound creates a 404 Not Found error.
func NotFound(message string) *APIError {
	return &APIError{
		Code:    CodeNotFound,
		Message: message,
		Status:  http.StatusNotFound,
	}
}

// BadRequest creates a 400 Bad Request error for validation failures.
func BadRequest(message string) *APIError {
	return &APIError{
		Code:    CodeValidation,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

// Conflict creates a 409 Conflict error for duplicate resources.
func Conflict(message string) *APIError {
	return &APIError{
		Code:    CodeConflict,
		Message: message,
		Status:  http.StatusConflict,
	}
}

// Forbidden creates a 403 Forbidden error for authorization failures.
func Forbidden(message string) *APIError {
	return &APIError{
		Code:    CodeForbidden,
		Message: message,
		Status:  http.StatusForbidden,
	}
}

// Unauthorized creates a 401 Unauthorized error for authentication failures.
func Unauthorized(message string) *APIError {
	return &APIError{
		Code:    CodeUnauthorized,
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}

// InternalServerError creates a 500 Internal Server Error with details from the original error.
func InternalServerError(err error) *APIError {
	return &APIError{
		Code:    CodeInternal,
		Message: "Internal server error",
		Details: err.Error(),
		Status:  http.StatusInternalServerError,
	}
}

// TooManyRequests creates a 429 Too Many Requests error with retry-after seconds.
func TooManyRequests(ra int) *RateLimitError {
	return &RateLimitError{
		APIError: APIError{
			Code:    CodeTooManyRequests,
			Message: "Rate limit exceeded",
			Details: fmt.Sprintf("Too many requests. Please try again in %s seconds.", strconv.Itoa(ra)),
			Status:  http.StatusTooManyRequests,
		},
		RetryAfter: ra,
	}
}

// ValidationError creates a validation error with field-level details.
func ValidationError(details interface{}) *APIError {
	return &APIError{
		Code:    CodeValidation,
		Message: "Validation failed",
		Details: details,
		Status:  http.StatusBadRequest,
	}
}

// FromGinValidation converts Gin/validator errors to structured APIError with field-level details.
func FromGinValidation(err error) *APIError {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		details := make(map[string]string)

		for _, fieldErr := range validationErrs {
			details[fieldErr.Field()] = formatValidationError(fieldErr)
		}

		return ValidationError(details)
	}

	return &APIError{
		Code:    CodeValidation,
		Message: "Invalid request data format",
		Details: err.Error(),
		Status:  http.StatusBadRequest,
	}
}

// formatValidationError converts validator field errors to human-readable messages.
// Handles common validation tags: required, email, min, max.
func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "min":
		return fe.Field() + " is too short (minimum " + fe.Param() + ")"
	case "max":
		return fe.Field() + " is too long (maximum " + fe.Param() + ")"
	default:
		return fe.Field() + " failed validation on tag " + fe.Tag()
	}
}
