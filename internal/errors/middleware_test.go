package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetRequestPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		setupCtx func() *gin.Context
		expected string
	}{
		{
			name: "normal request with path",
			setupCtx: func() *gin.Context {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest("GET", "/api/users", nil)
				return c
			},
			expected: "/api/users",
		},
		{
			name: "request with query parameters",
			setupCtx: func() *gin.Context {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest("GET", "/api/users?page=1&limit=10", nil)
				return c
			},
			expected: "/api/users",
		},
		{
			name: "nil request",
			setupCtx: func() *gin.Context {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = nil
				return c
			},
			expected: "",
		},
		{
			name: "nil request URL",
			setupCtx: func() *gin.Context {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = &http.Request{URL: nil}
				return c
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupCtx()
			result := getRequestPath(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorHandler_WithAPIError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		apiError       *APIError
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "NotFound error",
			apiError:       NotFound("User not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   CodeNotFound,
		},
		{
			name:           "BadRequest error",
			apiError:       BadRequest("Invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   CodeValidation,
		},
		{
			name:           "Unauthorized error",
			apiError:       Unauthorized("Authentication required"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   CodeUnauthorized,
		},
		{
			name:           "Forbidden error",
			apiError:       Forbidden("Access denied"),
			expectedStatus: http.StatusForbidden,
			expectedCode:   CodeForbidden,
		},
		{
			name:           "Conflict error",
			apiError:       Conflict("Resource exists"),
			expectedStatus: http.StatusConflict,
			expectedCode:   CodeConflict,
		},
		{
			name:           "InternalServerError",
			apiError:       InternalServerError(errors.New("db error")),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			_ = c.Error(tt.apiError)

			ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), `"success":false`)
			assert.Contains(t, w.Body.String(), tt.expectedCode)
			assert.Contains(t, w.Body.String(), tt.apiError.Message)
		})
	}
}

func TestErrorHandler_WithUnknownError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	unknownErr := errors.New("some unexpected error")
	_ = c.Error(unknownErr)

	ErrorHandler()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"success":false`)
	assert.Contains(t, w.Body.String(), CodeInternal)
	assert.Contains(t, w.Body.String(), "Internal server error")
}

func TestErrorHandler_WithNoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	ErrorHandler()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorHandler_WithMultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	_ = c.Error(errors.New("first error"))
	_ = c.Error(NotFound("second error"))

	ErrorHandler()(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "second error")
}

func TestErrorHandler_RateLimitError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	rateLimitErr := TooManyRequests(60)
	_ = c.Error(rateLimitErr)

	ErrorHandler()(c)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	assert.False(t, response["success"].(bool))
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, CodeTooManyRequests, errorObj["code"])
	assert.Contains(t, errorObj["details"], "60")
	assert.Equal(t, float64(60), errorObj["retry_after"])
}

func TestErrorHandler_ValidationErrorWithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	details := map[string]string{
		"email":    "Invalid email format",
		"password": "Password too short",
	}
	validationErr := ValidationError(details)
	_ = c.Error(validationErr)

	ErrorHandler()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), CodeValidation)
	assert.Contains(t, w.Body.String(), "Validation failed")
	assert.Contains(t, w.Body.String(), "email")
	assert.Contains(t, w.Body.String(), "password")
}

func TestErrorHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())

	router.GET("/test-not-found", func(c *gin.Context) {
		_ = c.Error(NotFound("Resource not found"))
	})

	router.GET("/test-validation", func(c *gin.Context) {
		_ = c.Error(ValidationError(map[string]string{"field": "error"}))
	})

	router.GET("/test-internal", func(c *gin.Context) {
		_ = c.Error(InternalServerError(errors.New("db error")))
	})

	router.GET("/test-success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "not found endpoint",
			path:           "/test-not-found",
			expectedStatus: http.StatusNotFound,
			expectedCode:   CodeNotFound,
		},
		{
			name:           "validation endpoint",
			path:           "/test-validation",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   CodeValidation,
		},
		{
			name:           "internal error endpoint",
			path:           "/test-internal",
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   CodeInternal,
		},
		{
			name:           "success endpoint",
			path:           "/test-success",
			expectedStatus: http.StatusOK,
			expectedCode:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}
		})
	}
}
