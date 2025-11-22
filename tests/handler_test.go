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

// createTestSchema creates the SQLite test schema using GORM AutoMigrate for consistency
func createTestSchema(t *testing.T, database *gorm.DB) {
	t.Helper()

	err := database.AutoMigrate(&user.User{}, &user.Role{}, &auth.RefreshToken{})
	assert.NoError(t, err)

	// Drop the auto-created user_roles table (created by GORM for many2many)
	// and recreate it with our custom schema including assigned_at column
	database.Exec("DROP TABLE IF EXISTS user_roles")
	
	// Manually create the user_roles junction table with assigned_at column
	err = database.Exec(`
		CREATE TABLE user_roles (
			user_id INTEGER NOT NULL,
			role_id INTEGER NOT NULL,
			assigned_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, role_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
		)
	`).Error
	assert.NoError(t, err)

	// Seed role data - use FirstOrCreate to avoid duplicate errors
	roles := []user.Role{
		{ID: 1, Name: "user", Description: "Standard user role with basic permissions"},
		{ID: 2, Name: "admin", Description: "Administrator role with full system access"},
	}
	for _, role := range roles {
		var existingRole user.Role
		result := database.Where("name = ?", role.Name).FirstOrCreate(&existingRole, &role)
		if result.Error != nil {
			t.Fatalf("Failed to create role %s: %v", role.Name, result.Error)
		}
	}
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

	router := server.SetupRouter(userHandler, authService, testCfg, database)

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

	return server.SetupRouter(userHandler, authService, testCfg, database)
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
				if success, ok := body["success"].(bool); !ok || !success {
					t.Error("Expected success to be true in response")
				}
				data, ok := body["data"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected data object in response")
				}
				if accessToken, ok := data["access_token"].(string); !ok || accessToken == "" {
					t.Error("Expected access_token in response data")
				}
				if refreshToken, ok := data["refresh_token"].(string); !ok || refreshToken == "" {
					t.Error("Expected refresh_token in response data")
				}
				if userData, ok := data["user"].(map[string]interface{}); !ok {
					t.Error("Expected user object in response data")
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
				"email":    "john@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if success, ok := body["success"].(bool); !ok || success {
					t.Error("Expected success to be false for error response")
				}
				errorInfo, ok := body["error"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected error object in response")
				}
				if errorMsg, ok := errorInfo["message"].(string); !ok || errorMsg == "" {
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
				if success, ok := body["success"].(bool); !ok || success {
					t.Error("Expected success to be false for error response")
				}
				errorInfo, ok := body["error"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected error object in response")
				}
				if errorMsg, ok := errorInfo["message"].(string); !ok || errorMsg == "" {
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
				if success, ok := body["success"].(bool); !ok || success {
					t.Error("Expected success to be false for error response")
				}
				errorInfo, ok := body["error"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected error object in response")
				}
				if errorMsg, ok := errorInfo["message"].(string); !ok || errorMsg == "" {
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
				t.Logf("Response body: %s", w.Body.String())
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
				if success, ok := body["success"].(bool); !ok || !success {
					t.Error("Expected success to be true in response")
				}
				data, ok := body["data"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected data object in response")
				}
				if accessToken, ok := data["access_token"].(string); !ok || accessToken == "" {
					t.Error("Expected access_token in response data")
				}
				if refreshToken, ok := data["refresh_token"].(string); !ok || refreshToken == "" {
					t.Error("Expected refresh_token in response data")
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
				if success, ok := body["success"].(bool); !ok || success {
					t.Error("Expected success to be false for error response")
				}
				errorInfo, ok := body["error"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected error object in response")
				}
				if errorMsg, ok := errorInfo["message"].(string); !ok || errorMsg == "" {
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
				if success, ok := body["success"].(bool); !ok || success {
					t.Error("Expected success to be false for error response")
				}
				errorInfo, ok := body["error"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected error object in response")
				}
				if errorMsg, ok := errorInfo["message"].(string); !ok || errorMsg == "" {
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
				if success, ok := body["success"].(bool); !ok || success {
					t.Error("Expected success to be false for error response")
				}
				errorInfo, ok := body["error"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected error object in response")
				}
				if errorMsg, ok := errorInfo["message"].(string); !ok || errorMsg == "" {
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
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy' in health check response, got %v", status)
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
		t.Logf("Response body: %s", rr.Body.String())
		t.Fatalf("register expected 200, got %d", rr.Code)
	}

	var registerResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("Failed to unmarshal register response: %v", err)
	}
	dataResp, ok := registerResp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data object in register response")
	}
	userResp, ok := dataResp["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user object in register response data")
	}
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
