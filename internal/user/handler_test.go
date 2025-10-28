package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
)

// MockAuthService is a mock implementation of the auth service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockAuthService) GenerateToken(userID uint, email string, name string) (string, error) {
	args := m.Called(userID, email, name)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateTokenPair(ctx context.Context, userID uint, email string, name string) (*auth.TokenPair, error) {
	args := m.Called(ctx, userID, email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockAuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockService, *MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				user := &User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				ms.On("RegisterUser", mock.Anything, mock.AnythingOfType("user.RegisterRequest")).Return(user, nil)
				tokenPair := &auth.TokenPair{
					AccessToken:  "mock-access-token",
					RefreshToken: "mock-refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    900,
				}
				mas.On("GenerateTokenPair", mock.Anything, uint(1), "john@example.com", "John Doe").Return(tokenPair, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "access_token")
				assert.Contains(t, response, "refresh_token")
				assert.Contains(t, response, "user")
			},
		},
		{
			name:        "invalid JSON format",
			requestBody: `{"name": "John", "email": invalid-json`,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				// No mocks needed for JSON binding error
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, "VALIDATION_ERROR", response["code"])
			},
		},
		{
			name: "missing required fields",
			requestBody: RegisterRequest{
				Name: "John Doe",
				// Missing email and password
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				// No mocks needed for validation error
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, "VALIDATION_ERROR", response["code"])
			},
		},
		{
			name: "email already exists",
			requestBody: RegisterRequest{
				Name:     "Jane Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUser", mock.Anything, mock.AnythingOfType("user.RegisterRequest")).Return(nil, ErrEmailExists)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Email already exists", response["message"])
			},
		},
		{
			name: "service database error",
			requestBody: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("RegisterUser", mock.Anything, mock.AnythingOfType("user.RegisterRequest")).Return(nil, errors.New("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "database connection error", response["details"])
			},
		},
		{
			name: "token generation failure",
			requestBody: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				user := &User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				ms.On("RegisterUser", mock.Anything, mock.AnythingOfType("user.RegisterRequest")).Return(user, nil)
				mas.On("GenerateTokenPair", mock.Anything, uint(1), "john@example.com", "John Doe").Return(nil, errors.New("token generation failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "token generation failed", response["details"])
			},
		},
		{
			name:        "empty request body",
			requestBody: `{}`,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				// No mocks needed for validation error
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "VALIDATION_ERROR", response["code"])
				assert.Equal(t, "Validation failed", response["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			mockAuthService := &MockAuthService{}
			tt.setupMocks(mockService, mockAuthService)

			handler := NewHandler(mockService, mockAuthService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request
			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			c.Request, _ = http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Register(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*MockService, *MockAuthService)
		setupContext   func(*gin.Context)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful get user",
			userID: "1",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				user := &User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				ms.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(1), response["id"])
				assert.Equal(t, "John Doe", response["name"])
				assert.Equal(t, "john@example.com", response["email"])
			},
		},
		{
			name:   "invalid user ID format",
			userID: "invalid",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				// No mocks needed for invalid ID
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, response["code"], "VALIDATION_ERROR")
				assert.Equal(t, "Invalid user ID", response["message"])
			},
		},
		{
			name:   "unauthenticated user - no context",
			userID: "1",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				// No mocks needed for auth check
			},
			setupContext: func(c *gin.Context) {
				// Don't set user context - unauthenticated
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Forbidden user ID", response["message"])
			},
		},
		{
			name:   "forbidden access - different user",
			userID: "2",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				// No mocks needed for auth check
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1} // User 1 trying to access user 2
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Forbidden user ID", response["message"])
			},
		},
		{
			name:   "user not found",
			userID: "999",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("GetUserByID", mock.Anything, uint(999)).Return(nil, ErrUserNotFound)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 999}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "User not found", response["message"])
			},
		},
		{
			name:   "database service error",
			userID: "1",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("GetUserByID", mock.Anything, uint(1)).Return(nil, errors.New("database connection error"))
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "database connection error", response["details"])
			},
		},
		{
			name:   "zero user ID",
			userID: "0",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				// No mocks needed for authorization failure
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Forbidden user ID", response["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			mockAuthService := &MockAuthService{}
			tt.setupMocks(mockService, mockAuthService)

			handler := NewHandler(mockService, mockAuthService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create a proper HTTP request
			req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			tt.setupContext(c)

			handler.GetUser(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockService, *MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				user := &User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				ms.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("user.LoginRequest")).Return(user, nil)
				tokenPair := &auth.TokenPair{
					AccessToken:  "mock-access-token",
					RefreshToken: "mock-refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    900,
				}
				mas.On("GenerateTokenPair", mock.Anything, uint(1), "john@example.com", "John Doe").Return(tokenPair, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "mock-access-token", response["access_token"])
				assert.Equal(t, "mock-refresh-token", response["refresh_token"])

				user := response["user"].(map[string]interface{})
				assert.Equal(t, float64(1), user["id"])
				assert.Equal(t, "John Doe", user["name"])
				assert.Equal(t, "john@example.com", user["email"])
			},
		},
		{
			name: "invalid credentials",
			requestBody: LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("user.LoginRequest")).Return(nil, ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid email or password", response["message"])
			},
		},
		{
			name: "service error",
			requestBody: LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("user.LoginRequest")).Return(nil, errors.New("failed to authenticate user"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "failed to authenticate user", response["details"])
			},
		},
		{
			name: "token generation error",
			requestBody: LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				user := &User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				ms.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("user.LoginRequest")).Return(user, nil)
				mas.On("GenerateTokenPair", mock.Anything, uint(1), "john@example.com", "John Doe").Return(nil, errors.New("failed to generate token"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "failed to generate token", response["details"])
			},
		},
		{
			name:           "invalid request body",
			requestBody:    `{invalid-json}`,
			setupMocks:     func(ms *MockService, mas *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "VALIDATION_ERROR", response["code"])
				assert.Equal(t, "Invalid request data format", response["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			mockAuthService := &MockAuthService{}
			tt.setupMocks(mockService, mockAuthService)

			handler := NewHandler(mockService, mockAuthService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var requestBody []byte
			if tt.requestBody != nil {
				// If requestBody is already a string, use it as raw JSON (for invalid JSON tests)
				if str, ok := tt.requestBody.(string); ok {
					requestBody = []byte(str)
				} else {
					// Otherwise, marshal it to JSON
					var err error
					requestBody, err = json.Marshal(tt.requestBody)
					assert.NoError(t, err)
				}
			}

			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.Login(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		setupMocks     func(*MockService, *MockAuthService)
		setupContext   func(*gin.Context)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful update",
			userID: "1",
			requestBody: UpdateUserRequest{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				updatedUser := &User{
					ID:    1,
					Name:  "John Updated",
					Email: "john.updated@example.com",
				}
				ms.On("UpdateUser", mock.Anything, uint(1), mock.AnythingOfType("user.UpdateUserRequest")).Return(updatedUser, nil)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, uint(1), response.ID)
				assert.Equal(t, "John Updated", response.Name)
				assert.Equal(t, "john.updated@example.com", response.Email)
			},
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			requestBody:    UpdateUserRequest{Name: "Test"},
			setupMocks:     func(ms *MockService, mas *MockAuthService) {},
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid user ID", response["message"])
			},
		},
		{
			name:   "forbidden access",
			userID: "2",
			requestBody: UpdateUserRequest{
				Name: "Unauthorized Update",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1} // Different user ID
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Forbidden user ID", response["message"])
			},
		},
		{
			name:   "user not found",
			userID: "999",
			requestBody: UpdateUserRequest{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("UpdateUser", mock.Anything, uint(999), mock.AnythingOfType("user.UpdateUserRequest")).Return(nil, ErrUserNotFound)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 999}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "User not found", response["message"])
			},
		},
		{
			name:   "email already exists",
			userID: "1",
			requestBody: UpdateUserRequest{
				Name:  "John Updated",
				Email: "existing@example.com",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("UpdateUser", mock.Anything, uint(1), mock.AnythingOfType("user.UpdateUserRequest")).Return(nil, ErrEmailExists)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Email already exists", response["message"])
			},
		},
		{
			name:   "service error",
			userID: "1",
			requestBody: UpdateUserRequest{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("UpdateUser", mock.Anything, uint(1), mock.AnythingOfType("user.UpdateUserRequest")).Return(nil, errors.New("failed to update user"))
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "failed to update user", response["details"])
			},
		},
		{
			name:        "invalid request body",
			userID:      "1",
			requestBody: `{invalid-json}`,
			setupMocks:  func(ms *MockService, mas *MockAuthService) {},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "VALIDATION_ERROR", response["code"])
				assert.Equal(t, "Invalid request data format", response["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			mockAuthService := &MockAuthService{}
			tt.setupMocks(mockService, mockAuthService)

			handler := NewHandler(mockService, mockAuthService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var requestBody []byte
			if tt.requestBody != nil {
				// If requestBody is already a string, use it as raw JSON (for invalid JSON tests)
				if str, ok := tt.requestBody.(string); ok {
					requestBody = []byte(str)
				} else {
					// Otherwise, marshal it to JSON
					var err error
					requestBody, err = json.Marshal(tt.requestBody)
					assert.NoError(t, err)
				}
			}

			req := httptest.NewRequest("PUT", "/users/"+tt.userID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			tt.setupContext(c)

			handler.UpdateUser(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*MockService, *MockAuthService)
		setupContext   func(*gin.Context)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful deletion",
			userID: "1",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("DeleteUser", mock.Anything, uint(1)).Return(nil)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusOK, // Gin test framework returns 200 by default
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Response body should be empty for successful deletion
				assert.Equal(t, "", w.Body.String())
			},
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			setupMocks:     func(ms *MockService, mas *MockAuthService) {},
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid user ID", response["message"])
			},
		},
		{
			name:       "forbidden access",
			userID:     "2",
			setupMocks: func(ms *MockService, mas *MockAuthService) {},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1} // Different user ID
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Forbidden user ID", response["message"])
			},
		},
		{
			name:   "user not found",
			userID: "1",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("DeleteUser", mock.Anything, uint(1)).Return(ErrUserNotFound)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "User not found", response["message"])
			},
		},
		{
			name:   "service error",
			userID: "1",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
				ms.On("DeleteUser", mock.Anything, uint(1)).Return(errors.New("failed to delete user"))
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "failed to delete user", response["details"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			mockAuthService := &MockAuthService{}
			tt.setupMocks(mockService, mockAuthService)

			handler := NewHandler(mockService, mockAuthService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("DELETE", "/users/"+tt.userID, nil)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			tt.setupContext(c)

			handler.DeleteUser(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}
