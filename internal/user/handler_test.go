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

func (m *MockAuthService) RevokeUserRefreshToken(ctx context.Context, userID uint, refreshToken string) error {
	args := m.Called(ctx, userID, refreshToken)
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
				assert.Equal(t, true, response["success"])
				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok, "data should be a map")
				assert.Contains(t, data, "access_token")
				assert.Contains(t, data, "refresh_token")
				assert.Contains(t, data, "user")
			},
		},
		{
			name:        "invalid JSON format",
			requestBody: `{"name": "John", "email": invalid-json`,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "VALIDATION_ERROR", errorInfo["code"])
			},
		},
		{
			name: "missing required fields",
			requestBody: RegisterRequest{
				Name: "John Doe",
			},
			setupMocks: func(ms *MockService, mas *MockAuthService) {
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "VALIDATION_ERROR", errorInfo["code"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Email already exists", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "database connection error", errorInfo["details"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "token generation failed", errorInfo["details"])
			},
		},
		{
			name:        "empty request body",
			requestBody: `{}`,
			setupMocks: func(ms *MockService, mas *MockAuthService) {
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "VALIDATION_ERROR", errorInfo["code"])
				assert.Equal(t, "Validation failed", errorInfo["message"])
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
				assert.Equal(t, true, response["success"])
				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok, "data should be a map")
				assert.Equal(t, float64(1), data["id"])
				assert.Equal(t, "John Doe", data["name"])
				assert.Equal(t, "john@example.com", data["email"])
			},
		},
		{
			name:   "invalid user ID format",
			userID: "invalid",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "VALIDATION_ERROR", errorInfo["code"])
				assert.Equal(t, "Invalid user ID", errorInfo["message"])
			},
		},
		{
			name:   "unauthenticated user - no context",
			userID: "1",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
			},
			setupContext: func(c *gin.Context) {
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Forbidden user ID", errorInfo["message"])
			},
		},
		{
			name:   "forbidden access - different user",
			userID: "2",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Forbidden user ID", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "User not found", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "database connection error", errorInfo["details"])
			},
		},
		{
			name:   "zero user ID",
			userID: "0",
			setupMocks: func(ms *MockService, mas *MockAuthService) {
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Forbidden user ID", errorInfo["message"])
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
				assert.Equal(t, true, response["success"])
				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok, "data should be a map")
				assert.Equal(t, "mock-access-token", data["access_token"])
				assert.Equal(t, "mock-refresh-token", data["refresh_token"])

				user := data["user"].(map[string]interface{})
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Invalid email or password", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "failed to authenticate user", errorInfo["details"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "failed to generate token", errorInfo["details"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "VALIDATION_ERROR", errorInfo["code"])
				assert.Equal(t, "Invalid request data format", errorInfo["message"])
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
				if str, ok := tt.requestBody.(string); ok {
					requestBody = []byte(str)
				} else {
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
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, true, response["success"])
				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok, "data should be a map")
				assert.Equal(t, float64(1), data["id"])
				assert.Equal(t, "John Updated", data["name"])
				assert.Equal(t, "john.updated@example.com", data["email"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Invalid user ID", errorInfo["message"])
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
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Forbidden user ID", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "User not found", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Email already exists", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "failed to update user", errorInfo["details"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "VALIDATION_ERROR", errorInfo["code"])
				assert.Equal(t, "Invalid request data format", errorInfo["message"])
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
				if str, ok := tt.requestBody.(string); ok {
					requestBody = []byte(str)
				} else {
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
			expectedStatus: http.StatusOK, // Note: Gin test recorder returns 200 for c.Status(204) without response body
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Invalid user ID", errorInfo["message"])
			},
		},
		{
			name:       "forbidden access",
			userID:     "2",
			setupMocks: func(ms *MockService, mas *MockAuthService) {},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "Forbidden user ID", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "User not found", errorInfo["message"])
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
				assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "failed to delete user", errorInfo["details"])
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

func TestHandler_GetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         uint
		setupMocks     func(*MockService)
		expectedStatus int
	}{
		{
			name:   "successful get current user",
			userID: 1,
			setupMocks: func(ms *MockService) {
				ms.On("GetUserByID", mock.Anything, uint(1)).Return(&User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not authenticated",
			userID: 0,
			setupMocks: func(ms *MockService) {
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func(ms *MockService) {
				ms.On("GetUserByID", mock.Anything, uint(999)).Return(nil, ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "service error",
			userID: 1,
			setupMocks: func(ms *MockService) {
				ms.On("GetUserByID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			mockAuthService := new(MockAuthService)
			handler := NewHandler(mockService, mockAuthService)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
			c.Request = req

			if tt.userID > 0 {
				claims := &auth.Claims{
					UserID: tt.userID,
					Email:  "test@example.com",
				}
				c.Set(auth.KeyUser, claims)
			}

			handler.GetMe(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_ListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func(*MockService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful list with defaults",
			queryParams: "",
			setupMocks: func(ms *MockService) {
				users := []User{
					{ID: 1, Name: "User 1", Email: "user1@example.com"},
					{ID: 2, Name: "User 2", Email: "user2@example.com"},
				}
				ms.On("ListUsers", mock.Anything, mock.MatchedBy(func(f UserFilterParams) bool {
					return f.Sort == "created_at" && f.Order == "desc"
				}), 1, 20).Return(users, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.Equal(t, float64(2), data["total"])
			},
		},
		{
			name:        "list with role filter",
			queryParams: "?role=admin&page=1&per_page=10",
			setupMocks: func(ms *MockService) {
				users := []User{
					{ID: 1, Name: "Admin User", Email: "admin@example.com"},
				}
				ms.On("ListUsers", mock.Anything, mock.MatchedBy(func(f UserFilterParams) bool {
					return f.Role == "admin"
				}), 1, 10).Return(users, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, float64(1), data["total"])
			},
		},
		{
			name:        "list with search",
			queryParams: "?search=john",
			setupMocks: func(ms *MockService) {
				users := []User{
					{ID: 1, Name: "John Doe", Email: "john@example.com"},
				}
				ms.On("ListUsers", mock.Anything, mock.MatchedBy(func(f UserFilterParams) bool {
					return f.Search == "john"
				}), 1, 20).Return(users, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
			},
		},
		{
			name:        "empty result set",
			queryParams: "",
			setupMocks: func(ms *MockService) {
				ms.On("ListUsers", mock.Anything, mock.Anything, 1, 20).Return([]User{}, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, float64(0), data["total"])
			},
		},
		{
			name:        "service error",
			queryParams: "",
			setupMocks: func(ms *MockService) {
				ms.On("ListUsers", mock.Anything, mock.Anything, 1, 20).Return(nil, int64(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
			},
		},
		{
			name:        "invalid role filter",
			queryParams: "",
			setupMocks: func(ms *MockService) {
				ms.On("ListUsers", mock.Anything, mock.Anything, 1, 20).Return(nil, int64(0), ErrInvalidRole)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			mockAuthService := new(MockAuthService)
			handler := NewHandler(mockService, mockAuthService)

			tt.setupMocks(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users"+tt.queryParams, nil)
			c.Request = req

			handler.ListUsers(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			mockService.AssertExpectations(t)
		})
	}
}
