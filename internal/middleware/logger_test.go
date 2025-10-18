package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// TestLogger tests the logger middleware basic functionality
func TestLogger(t *testing.T) {
	// Create a buffer to capture logs
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths: []string{},
		Logger:    logger,
	}

	// Setup router
	router := gin.New()
	router.Use(Logger(config))
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
	if !strings.Contains(logOutput, "HTTP Request") {
		t.Error("Expected log to contain 'HTTP Request'")
	}
	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain request method 'GET'")
	}
	if !strings.Contains(logOutput, "/test") {
		t.Error("Expected log to contain request path '/test'")
	}
}

// TestLoggerSkipPaths tests that specified paths are skipped
func TestLoggerSkipPaths(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths: []string{"/health"},
		Logger:    logger,
	}

	router := gin.New()
	router.Use(Logger(config))
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
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths: []string{},
		Logger:    logger,
	}

	router := gin.New()
	router.Use(Logger(config))
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
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths: []string{},
		Logger:    logger,
	}

	router := gin.New()
	router.Use(Logger(config))
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

// TestLoggerStatusCodes tests logging of different status codes
func TestLoggerStatusCodes(t *testing.T) {
	testCases := []struct {
		name          string
		statusCode    int
		expectedLevel string
	}{
		{"Success", http.StatusOK, "INFO"},
		{"Client Error", http.StatusBadRequest, "WARN"},
		{"Not Found", http.StatusNotFound, "WARN"},
		{"Server Error", http.StatusInternalServerError, "ERROR"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			config := &LoggerConfig{
				SkipPaths: []string{},
				Logger:    logger,
			}

			router := gin.New()
			router.Use(Logger(config))
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

// TestLoggerWithConfig tests LoggerWithConfig function
func TestLoggerWithConfig(t *testing.T) {
	var buf bytes.Buffer

	// Redirect log output to buffer for testing
	oldDefault := slog.Default()
	defer slog.SetDefault(oldDefault)

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	router := gin.New()
	router.Use(LoggerWithConfig([]string{"/health"}, slog.LevelDebug))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestDefaultLoggerConfig tests the default configuration
func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.Logger == nil {
		t.Error("Expected logger to be initialized")
	}

	if len(config.SkipPaths) != 1 || config.SkipPaths[0] != "/health" {
		t.Errorf("Expected default skip paths to contain '/health', got %v", config.SkipPaths)
	}
}

// TestLoggerQueryParameters tests logging with query parameters
func TestLoggerQueryParameters(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := &LoggerConfig{
		SkipPaths: []string{},
		Logger:    logger,
	}

	router := gin.New()
	router.Use(Logger(config))
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

// TestNewLoggerConfig tests the NewLoggerConfig function
func TestNewLoggerConfig(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  slog.Level
		skipPaths []string
		validate  func(*testing.T, *LoggerConfig)
	}{
		{
			name:      "info level with default skip paths",
			logLevel:  slog.LevelInfo,
			skipPaths: []string{"/health"},
			validate: func(t *testing.T, config *LoggerConfig) {
				if config == nil {
					t.Error("Expected non-nil config")
					return
				}
				if len(config.SkipPaths) != 1 {
					t.Errorf("Expected 1 skip path, got %d", len(config.SkipPaths))
				}
				if config.SkipPaths[0] != "/health" {
					t.Errorf("Expected /health skip path, got %s", config.SkipPaths[0])
				}
				if config.Logger == nil {
					t.Error("Expected non-nil logger")
				}
			},
		},
		{
			name:      "debug level with multiple skip paths",
			logLevel:  slog.LevelDebug,
			skipPaths: []string{"/health", "/metrics", "/status"},
			validate: func(t *testing.T, config *LoggerConfig) {
				if config == nil {
					t.Error("Expected non-nil config")
					return
				}
				if len(config.SkipPaths) != 3 {
					t.Errorf("Expected 3 skip paths, got %d", len(config.SkipPaths))
				}
				expectedPaths := map[string]bool{"/health": true, "/metrics": true, "/status": true}
				for _, path := range config.SkipPaths {
					if !expectedPaths[path] {
						t.Errorf("Unexpected skip path: %s", path)
					}
				}
				if config.Logger == nil {
					t.Error("Expected non-nil logger")
				}
			},
		},
		{
			name:      "error level with empty skip paths",
			logLevel:  slog.LevelError,
			skipPaths: []string{},
			validate: func(t *testing.T, config *LoggerConfig) {
				if config == nil {
					t.Error("Expected non-nil config")
					return
				}
				if len(config.SkipPaths) != 0 {
					t.Errorf("Expected 0 skip paths, got %d", len(config.SkipPaths))
				}
				if config.Logger == nil {
					t.Error("Expected non-nil logger")
				}
			},
		},
		{
			name:      "warn level with nil skip paths",
			logLevel:  slog.LevelWarn,
			skipPaths: nil,
			validate: func(t *testing.T, config *LoggerConfig) {
				if config == nil {
					t.Error("Expected non-nil config")
					return
				}
				if len(config.SkipPaths) > 0 {
					t.Errorf("Expected empty skip paths, got %v", config.SkipPaths)
				}
				if config.Logger == nil {
					t.Error("Expected non-nil logger")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewLoggerConfig(tt.logLevel, tt.skipPaths)
			tt.validate(t, config)

			// Test that the config can be used in middleware
			router := gin.New()
			router.Use(Logger(config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "test"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}
