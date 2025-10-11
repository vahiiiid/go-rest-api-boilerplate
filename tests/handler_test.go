package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/server"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Use the test config helper to get a valid configuration
	testCfg := config.NewTestConfig()

	// Setup in-memory SQLite database for testing
	database, err := db.NewSQLiteDB(":memory:")
	assert.NoError(t, err)

	// Run migrations
	err = database.AutoMigrate(&user.User{})
	assert.NoError(t, err)

	// Initialize services with the test config
	authService := auth.NewService(&testCfg.JWT)
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

	// Setup router with all dependencies and the test config
	router := server.SetupRouter(userHandler, authService, testCfg)

	return router
}

func setupRateLimitTestRouter(t *testing.T) *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Use the test config helper to get a valid base configuration
	testCfg := config.NewTestConfig()
	// Override rate limit settings specifically for this test
	testCfg.Ratelimit.Enabled = true
	testCfg.Ratelimit.Requests = 10
	testCfg.Ratelimit.Window = time.Minute

	// Create in-memory SQLite database for testing
	database, err := db.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(&user.User{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services with the test config
	authService := auth.NewService(&testCfg.JWT)
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

	// Setup router
	return server.SetupRouter(userHandler, authService, testCfg)
}

func TestRegisterHandler(t *testing.T) {
	router := setupTestRouter(t)

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful registration",
			payload: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if token, ok := body["token"].(string); !ok || token == "" {
					t.Error("Expected token in response")
				}
				if userData, ok := body["user"].(map[string]interface{}); !ok {
					t.Error("Expected user object in response")
				} else {
					if email, ok := userData["email"].(string); !ok || email != "john@example.com" {
						t.Errorf("Expected email 'john@example.com', got '%v'", email)
					}
				}
			},
		},
		{
			name: "duplicate email",
			payload: map[string]string{
				"name":     "Jane Doe",
				"email":    "john@example.com", // Same email as previous test
				"password": "password123",
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if errorMsg, ok := body["error"].(string); !ok || errorMsg == "" {
					t.Error("Expected error message in response")
				}
			},
		},
		{
			name: "invalid email format",
			payload: map[string]string{
				"name":     "Invalid User",
				"email":    "not-an-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if errorMsg, ok := body["error"].(string); !ok || errorMsg == "" {
					t.Error("Expected error message in response")
				}
			},
		},
		{
			name: "missing required fields",
			payload: map[string]string{
				"name": "Incomplete User",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if errorMsg, ok := body["error"].(string); !ok || errorMsg == "" {
					t.Error("Expected error message in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	router := setupTestRouter(t)

	// First, register a user
	registerPayload := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "testpassword123",
	}
	jsonPayload, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful login",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "testpassword123",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if token, ok := body["token"].(string); !ok || token == "" {
					t.Error("Expected token in response")
				}
			},
		},
		{
			name: "invalid password",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if errorMsg, ok := body["error"].(string); !ok || errorMsg == "" {
					t.Error("Expected error message in response")
				}
			},
		},
		{
			name: "non-existent user",
			payload: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if errorMsg, ok := body["error"].(string); !ok || errorMsg == "" {
					t.Error("Expected error message in response")
				}
			},
		},
		{
			name: "missing credentials",
			payload: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if errorMsg, ok := body["error"].(string); !ok || errorMsg == "" {
					t.Error("Expected error message in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}
	if status, ok := response["status"].(string); !ok || status != "ok" {
		t.Error("Expected status 'ok' in health check response")
	}
}

func TestRateLimit_BlocksThenAllows(t *testing.T) {
	// Skip this test in CI environment to avoid flaky failures
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping rate limiting test in CI environment")
	}

	r := setupRateLimitTestRouter(t)

	// Arrange: register a user (this also consumes 1 token on /auth if the limiter is on the group)
	registerBody, _ := json.Marshal(map[string]string{
		"name":     "Rate Test",
		"email":    "rate@example.com",
		"password": "secret123",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.168.1.100") // Use consistent IP for testing
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("register expected 200, got %d", rr.Code)
	}

	// Arrange: login payload
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "rate@example.com",
		"password": "secret123",
	})

	// Act: consume remaining budget (limit=10/min, 1 was used by register → 9 left)
	for i := 0; i < 9; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.100") // Use consistent IP for testing
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("login #%d expected 200, got %d", i+1, rr.Code)
		}
	}

	// Assert: next login should be blocked with 429 and include Retry-After
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.168.1.100") // Use consistent IP for testing
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after limit exhausted, got %d", rr.Code)
	}
	retryAfterStr := rr.Header().Get("Retry-After")
	if retryAfterStr == "" {
		t.Fatalf("expected Retry-After header on 429")
	}
	retryAfterSec, err := strconv.Atoi(retryAfterStr)
	if err != nil || retryAfterSec <= 0 {
		t.Fatalf("Retry-After should be positive integer seconds, got %q (err=%v)", retryAfterStr, err)
	}

	// Arrange/Act: wait for the advised cooldown, then retry once
	time.Sleep(time.Duration(retryAfterSec)*time.Second + 10*time.Millisecond)

	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.168.1.100") // Use consistent IP for testing
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Assert: request should pass after cooldown
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 after waiting Retry-After=%ds, got %d", retryAfterSec, rr.Code)
	}
}
