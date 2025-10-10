package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
