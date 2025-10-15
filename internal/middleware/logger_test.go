package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/logger"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// TestLogger tests the logger middleware basic functionality
func TestLogger(t *testing.T) {
	// Create a buffer to capture logs
	var buf bytes.Buffer

	// Initialize logger with the buffer
	if err := logger.InitWithWriter("test", &buf); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Setup router
	router := gin.New()

	router.Use(logger.Middleware())
	router.Use(gin.Recovery())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "Request") {
		t.Error("Expected log to contain 'Request'")
	}
	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain request method 'GET'")
	}
	if !strings.Contains(logOutput, "/test") {
		t.Error("Expected log to contain request path '/test'")
	}
	if !strings.Contains(logOutput, "INFO") {
		t.Error("Expected log to contain request log 'INFO'")
	}
}

// TestLoggerSkipPaths tests that specified paths are skipped
func TestLoggerSkipPaths(t *testing.T) {
	// Create a buffer to capture logs
	var buf bytes.Buffer

	// Initialize logger with the buffer
	if err := logger.InitWithWriter("test", &buf); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Setup router
	router := gin.New()

	router.Use(logger.Middleware())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request to /health
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response is OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify log is empty (path was skipped)
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output for skipped path, got: %s", logOutput)
	}
}

// TestLoggerRequestID tests request ID generation and propagation
func TestLoggerRequestID(t *testing.T) {
	// Create a buffer to capture logs
	var buf bytes.Buffer

	// Initialize logger with the buffer
	if err := logger.InitWithWriter("test", &buf); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Setup router
	router := gin.New()

	router.Use(logger.Middleware())
	router.Use(gin.Recovery())

	router.GET("/test", func(c *gin.Context) {
		// Verify request ID is set in context
		requestID, exists := c.Get("request_id")
		if !exists {
			t.Error("Expected request_id to be set in context")
		}
		if requestID == "" {
			t.Error("Expected request_id to be non-empty")
		}
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify X-Request-ID header is set in response
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set in response")
	}

	// Verify request ID is in log
	logOutput := buf.String()
	if !strings.Contains(logOutput, requestID) {
		t.Error("Expected log to contain request ID")
	}
}

// TestLoggerWithProvidedRequestID tests that provided request ID is used
func TestLoggerWithProvidedRequestID(t *testing.T) {
	// Create a buffer to capture logs
	var buf bytes.Buffer

	// Initialize logger with the buffer
	if err := logger.InitWithWriter("test", &buf); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Setup router
	router := gin.New()

	router.Use(logger.Middleware())
	router.Use(gin.Recovery())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request with X-Request-ID header
	providedID := "test-request-id-123"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", providedID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify provided request ID is used
	logOutput := buf.String()
	if !strings.Contains(logOutput, providedID) {
		t.Errorf("Expected log to contain provided request ID: %s", providedID)
	}
}

// // TestLoggerStatusCodes tests logging of different status codes
func TestLoggerStatusCodes(t *testing.T) {
	testCases := []struct {
		name          string
		statusCode    int
		expectedLevel string
	}{
		{"Success", http.StatusOK, "info"},
		{"Client Error", http.StatusBadRequest, "warn"},
		{"Not Found", http.StatusNotFound, "warn"},
		{"Server Error", http.StatusInternalServerError, "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a buffer to capture logs
			var buf bytes.Buffer

			// Initialize logger with the buffer
			// Using production log in this test because we are parsing json
			if err := logger.InitWithWriter("production", &buf); err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			// Setup router
			router := gin.New()

			router.Use(logger.Middleware())
			router.Use(gin.Recovery())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(tc.statusCode, gin.H{"status": "test"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify status code
			if w.Code != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, w.Code)
			}

			// Parse log output
			logOutput := buf.String()
			var logData map[string]interface{}
			if err := json.Unmarshal([]byte(logOutput), &logData); err != nil {
				t.Fatalf("Failed to parse log JSON: %v", err)
			}

			// Verify log level
			if level, ok := logData["level"].(string); ok {
				if level != tc.expectedLevel {
					t.Errorf("Expected log level %s, got %s", tc.expectedLevel, level)
				}
			} else {
				t.Error("Expected 'level' field in log output")
			}

			// Verify status in log
			if status, ok := logData["status"].(float64); ok {
				if int(status) != tc.statusCode {
					t.Errorf("Expected status %d in log, got %d", tc.statusCode, int(status))
				}
			} else {
				t.Error("Expected 'status' field in log output")
			}
		})
	}
}

// TestLoggerQueryParameters tests logging with query parameters
func TestLoggerQueryParameters(t *testing.T) {
	// Create a buffer to capture logs
	var buf bytes.Buffer

	// Initialize logger with the buffer
	if err := logger.InitWithWriter("test", &buf); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Setup router
	router := gin.New()

	router.Use(logger.Middleware())
	router.Use(gin.Recovery())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test?foo=bar&baz=qux", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "foo=bar") {
		t.Error("Expected log to contain query parameters")
	}
}
