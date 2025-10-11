package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/server"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
	"gorm.io/gorm"
)

// Additional imports for password reset tests

func setupTestRouter(t *testing.T) *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database for testing
	database, err := db.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(&user.User{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	authService := auth.NewService()
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

	// Create test configuration
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Env: "test",
		},
		Logging: config.LoggingConfig{
			Level: "info",
		},
	}

	// Setup router
	return server.SetupRouter(userHandler, authService, testConfig)
}

// setupTestRouterWithAuth sets up router including forgot/reset password handlers
func setupTestRouterWithAuth(t *testing.T) (*gin.Engine, *gorm.DB, auth.PasswordResetService, user.Repository, user.Service) {
	gin.SetMode(gin.TestMode)

	database, err := db.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations for users and password_reset_tokens
	if err := database.AutoMigrate(&user.User{}, &auth.PasswordResetToken{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	authService := auth.NewService()
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

	// Password reset pieces
	resetService := auth.NewPasswordResetService(database)
	userPort := user.NewUserPasswordAdapter(userService)
	authHandler := auth.NewHandler(resetService, userPort)

	testConfig := &config.Config{Server: config.ServerConfig{Env: "test"}, Logging: config.LoggingConfig{Level: "info"}}
	router := server.SetupRouterWithAuth(userHandler, authService, authHandler, testConfig)
	return router, database, resetService, userRepo, userService
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

func TestForgotPassword_CreatesTokenAndReturns200(t *testing.T) {
	router, database, _, userRepo, _ := setupTestRouterWithAuth(t)

	// Create user
	u := &user.User{Name: "Reset User", Email: "reset@example.com", PasswordHash: "hash"}
	if err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// First request
	payload := map[string]string{"email": "reset@example.com"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Check that a token exists and is unused
	var count int64
	if err := database.Table((auth.PasswordResetToken{}).TableName()).Where("user_id = ? AND used = ?", u.ID, false).Count(&count).Error; err != nil {
		t.Fatalf("failed counting tokens: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 unused token, got %d", count)
	}

	// Second request should invalidate previous and create new one (use a fresh request)
	body2, _ := json.Marshal(payload)
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on second attempt, got %d", w2.Code)
	}

	var unused int64
	if err := database.Table((auth.PasswordResetToken{}).TableName()).Where("user_id = ? AND used = ?", u.ID, false).Count(&unused).Error; err != nil {
		t.Fatalf("failed counting unused tokens: %v", err)
	}
	if unused != 1 {
		t.Fatalf("expected exactly 1 unused token after second request, got %d", unused)
	}
}

func TestResetPassword_Success(t *testing.T) {
	router, database, resetService, userRepo, userService := setupTestRouterWithAuth(t)

	// Create user with known password
	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass123"), bcrypt.DefaultCost)
	u := &user.User{Name: "Reset User", Email: "reset2@example.com", PasswordHash: string(hashed)}
	if err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create token via service to get raw token
	token, _, err := resetService.CreateToken(context.Background(), u.ID, time.Hour)
	if err != nil {
		t.Fatalf("failed to create reset token: %v", err)
	}

	// Call reset endpoint
	payload := map[string]string{"token": token, "new_password": "newpass123"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify token marked used
	var usedCount int64
	if err := database.Table((auth.PasswordResetToken{}).TableName()).Where("user_id = ? AND used = ?", u.ID, true).Count(&usedCount).Error; err != nil {
		t.Fatalf("failed counting used tokens: %v", err)
	}
	if usedCount == 0 {
		t.Fatalf("expected at least 1 used token after reset")
	}

	// Verify password updated
	refreshed, err := userService.GetUserByID(context.Background(), u.ID)
	if err != nil || refreshed == nil {
		t.Fatalf("failed to fetch user: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(refreshed.PasswordHash), []byte("newpass123")); err != nil {
		t.Fatalf("expected password to be updated")
	}
}
