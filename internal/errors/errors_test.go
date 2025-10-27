package errors

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:    "TEST_ERROR",
		Message: "test error message",
		Status:  http.StatusBadRequest,
	}

	assert.Equal(t, "test error message", err.Error())
}

func TestNotFound(t *testing.T) {
	err := NotFound("Resource not found")

	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, "Resource not found", err.Message)
	assert.Equal(t, http.StatusNotFound, err.Status)
	assert.Nil(t, err.Details)
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("Invalid input")

	assert.Equal(t, CodeValidation, err.Code)
	assert.Equal(t, "Invalid input", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Nil(t, err.Details)
}

func TestConflict(t *testing.T) {
	err := Conflict("Resource already exists")

	assert.Equal(t, CodeConflict, err.Code)
	assert.Equal(t, "Resource already exists", err.Message)
	assert.Equal(t, http.StatusConflict, err.Status)
	assert.Nil(t, err.Details)
}

func TestForbidden(t *testing.T) {
	err := Forbidden("Access denied")

	assert.Equal(t, CodeForbidden, err.Code)
	assert.Equal(t, "Access denied", err.Message)
	assert.Equal(t, http.StatusForbidden, err.Status)
	assert.Nil(t, err.Details)
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("Authentication required")

	assert.Equal(t, CodeUnauthorized, err.Code)
	assert.Equal(t, "Authentication required", err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.Status)
	assert.Nil(t, err.Details)
}

func TestInternalServerError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := InternalServerError(originalErr)

	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, "Internal server error", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.Status)
	assert.Equal(t, "database connection failed", err.Details)
}

func TestTooManyRequests(t *testing.T) {
	retryAfter := 60
	err := TooManyRequests(retryAfter)

	assert.Equal(t, CodeTooManyRequests, err.Code)
	assert.Equal(t, "Rate limit exceeded", err.Message)
	assert.Equal(t, http.StatusTooManyRequests, err.Status)
	assert.Equal(t, retryAfter, err.RetryAfter)
	assert.Contains(t, err.Details, "60 seconds")
}

func TestValidationError(t *testing.T) {
	details := map[string]string{
		"email":    "Invalid email format",
		"password": "Password too short",
	}
	err := ValidationError(details)

	assert.Equal(t, CodeValidation, err.Code)
	assert.Equal(t, "Validation failed", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, details, err.Details)
}

func TestFormatValidationError(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		field    string
		param    string
		expected string
	}{
		{
			name:     "required field",
			tag:      "required",
			field:    "Email",
			param:    "",
			expected: "Email is required",
		},
		{
			name:     "email validation",
			tag:      "email",
			field:    "Email",
			param:    "",
			expected: "Email must be a valid email address",
		},
		{
			name:     "min length validation",
			tag:      "min",
			field:    "Password",
			param:    "6",
			expected: "Password is too short (minimum 6)",
		},
		{
			name:     "max length validation",
			tag:      "max",
			field:    "Name",
			param:    "100",
			expected: "Name is too long (maximum 100)",
		},
		{
			name:     "unknown validation tag",
			tag:      "custom",
			field:    "Field",
			param:    "",
			expected: "Field failed validation on tag custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &mockFieldError{
				tag:   tt.tag,
				field: tt.field,
				param: tt.param,
			}
			result := formatValidationError(fe)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type mockFieldError struct {
	tag   string
	field string
	param string
}

func (m *mockFieldError) Tag() string                    { return m.tag }
func (m *mockFieldError) ActualTag() string              { return m.tag }
func (m *mockFieldError) Namespace() string              { return "" }
func (m *mockFieldError) StructNamespace() string        { return "" }
func (m *mockFieldError) Field() string                  { return m.field }
func (m *mockFieldError) StructField() string            { return m.field }
func (m *mockFieldError) Value() interface{}             { return nil }
func (m *mockFieldError) Param() string                  { return m.param }
func (m *mockFieldError) Kind() reflect.Kind             { return reflect.String }
func (m *mockFieldError) Type() reflect.Type             { return nil }
func (m *mockFieldError) Translate(ut.Translator) string { return "" }
func (m *mockFieldError) Error() string                  { return "" }

func TestFromGinValidation_WithValidationErrors(t *testing.T) {
	validate := validator.New()

	type TestStruct struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,min=6"`
		Name     string `validate:"required,max=100"`
	}

	testData := TestStruct{
		Email:    "invalid-email",
		Password: "123",
		Name:     "",
	}

	err := validate.Struct(testData)
	assert.Error(t, err)

	apiErr := FromGinValidation(err)

	assert.Equal(t, CodeValidation, apiErr.Code)
	assert.Equal(t, "Validation failed", apiErr.Message)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.NotNil(t, apiErr.Details)

	details, ok := apiErr.Details.(map[string]string)
	assert.True(t, ok)
	assert.Contains(t, details, "Email")
	assert.Contains(t, details, "Password")
	assert.Contains(t, details, "Name")
}

func TestFromGinValidation_WithNonValidationError(t *testing.T) {
	err := errors.New("some random error")
	apiErr := FromGinValidation(err)

	assert.Equal(t, CodeValidation, apiErr.Code)
	assert.Equal(t, "Invalid request data format", apiErr.Message)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.Equal(t, "some random error", apiErr.Details)
}

func TestRateLimitError_Structure(t *testing.T) {
	err := TooManyRequests(30)

	assert.IsType(t, &RateLimitError{}, err)
	assert.Equal(t, 30, err.RetryAfter)
	assert.Equal(t, CodeTooManyRequests, err.Code)
	assert.Equal(t, http.StatusTooManyRequests, err.Status)
}

func TestAPIError_WithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field":  "email",
		"reason": "already exists",
	}

	err := &APIError{
		Code:    CodeConflict,
		Message: "Conflict occurred",
		Details: details,
		Status:  http.StatusConflict,
	}

	assert.Equal(t, details, err.Details)
	assert.Equal(t, "Conflict occurred", err.Error())
}

func TestAPIError_WithoutDetails(t *testing.T) {
	err := NotFound("User not found")
	assert.Nil(t, err.Details)
}
