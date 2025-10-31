package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/server"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// createTestSchema creates the SQLite test schema that matches PostgreSQL production schema
func createTestSchema(t *testing.T, database *gorm.DB) {
	t.Helper()
	sqlDB, err := database.DB()
	assert.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		);
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_users_deleted_at ON users(deleted_at);
	`)
	assert.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE refresh_tokens (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			token_hash TEXT NOT NULL,
			token_family TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used_at DATETIME,
			revoked_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
		CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
		CREATE INDEX idx_refresh_tokens_token_family ON refresh_tokens(token_family);
		CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
	`)
	assert.NoError(t, err)
}

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	testCfg := config.NewTestConfig()

	database, err := db.NewSQLiteDB(":memory:")
	assert.NoError(t, err)

	createTestSchema(t, database)

	authService := auth.NewServiceWithRepo(&testCfg.JWT, database)
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

	router := server.SetupRouter(userHandler, authService, testCfg)

	return router
}

func setupRateLimitTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	testCfg := config.NewTestConfig()
	testCfg.Ratelimit.Enabled = true
	testCfg.Ratelimit.Requests = 10
	testCfg.Ratelimit.Window = time.Minute

	database, err := db.NewSQLiteDB(":memory:")
	assert.NoError(t, err)

	createTestSchema(t, database)

	authService := auth.NewServiceWithRepo(&testCfg.JWT, database)
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

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
				if accessToken, ok := body["access_token"].(string); !ok || accessToken == "" {
					t.Error("Expected access_token in response")
				}
				if refreshToken, ok := body["refresh_token"].(string); !ok || refreshToken == "" {
					t.Error("Expected refresh_token in response")
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
				if errorMsg, ok := body["message"].(string); !ok || errorMsg == "" {
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
				if errorMsg, ok := body["message"].(string); !ok || errorMsg == "" {
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
				if errorMsg, ok := body["message"].(string); !ok || errorMsg == "" {
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
				if accessToken, ok := body["access_token"].(string); !ok || accessToken == "" {
					t.Error("Expected access_token in response")
				}
				if refreshToken, ok := body["refresh_token"].(string); !ok || refreshToken == "" {
					t.Error("Expected refresh_token in response")
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
				if errorMsg, ok := body["message"].(string); !ok || errorMsg == "" {
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
				if errorMsg, ok := body["message"].(string); !ok || errorMsg == "" {
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
				if errorMsg, ok := body["message"].(string); !ok || errorMsg == "" {
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
	r := setupRateLimitTestRouter(t)

	testIP := fmt.Sprintf("192.168.1.%d", time.Now().UnixNano()%255)

	registerBody, _ := json.Marshal(map[string]string{
		"name":     "Rate Test",
		"email":    fmt.Sprintf("rate%d@example.com", time.Now().UnixNano()),
		"password": "secret123",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", testIP)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("register expected 200, got %d", rr.Code)
	}

	var registerResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("Failed to unmarshal register response: %v", err)
	}
	userResp := registerResp["user"].(map[string]interface{})
	email := userResp["email"].(string)

	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": "secret123",
	})

	successCount := 0
	for i := 0; i < 15; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", testIP)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			successCount++
		} else if rr.Code == http.StatusTooManyRequests {
			retryAfterStr := rr.Header().Get("Retry-After")
			if retryAfterStr == "" {
				t.Fatalf("expected Retry-After header on 429")
			}
			retryAfterSec, err := strconv.Atoi(retryAfterStr)
			if err != nil || retryAfterSec <= 0 {
				t.Fatalf("Retry-After should be positive integer seconds, got %q (err=%v)", retryAfterStr, err)
			}
			t.Logf("Rate limit triggered after %d successful requests (including register)", successCount+1)
			return
		} else {
			t.Fatalf("login #%d expected 200 or 429, got %d", i+1, rr.Code)
		}
	}

	// If we get here, rate limiting didn't work
	t.Fatalf("expected rate limiting to trigger, but completed %d requests without 429", successCount)
}
