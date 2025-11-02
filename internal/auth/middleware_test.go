package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of Service interface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateToken(userID uint, email string, name string) (string, error) {
	args := m.Called(userID, email, name)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateTokenPair(ctx context.Context, userID uint, email string, name string) (*TokenPair, error) {
	args := m.Called(ctx, userID, email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
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

func setupTestRouter(authService Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	protected := r.Group("/api")
	protected.Use(AuthMiddleware(authService))
	protected.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	return r
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful authentication",
			authHeader: "Bearer valid-token",
			setupMock: func(m *MockAuthService) {
				claims := &Claims{
					UserID: 123,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				m.On("ValidateToken", "valid-token").Return(claims, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"authorization header required"}`,
		},
		{
			name:           "invalid authorization header format - no Bearer",
			authHeader:     "invalid-token",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid authorization header format"}`,
		},
		{
			name:           "invalid authorization header format - wrong scheme",
			authHeader:     "Basic dGVzdDp0ZXN0",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid authorization header format"}`,
		},
		{
			name:           "invalid authorization header format - no token",
			authHeader:     "Bearer",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid authorization header format"}`,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			setupMock: func(m *MockAuthService) {
				m.On("ValidateToken", "invalid-token").Return(nil, ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid or expired token"}`,
		},
		{
			name:       "expired token",
			authHeader: "Bearer expired-token",
			setupMock: func(m *MockAuthService) {
				m.On("ValidateToken", "expired-token").Return(nil, ErrExpiredToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid or expired token"}`,
		},
		{
			name:       "service error",
			authHeader: "Bearer error-token",
			setupMock: func(m *MockAuthService) {
				m.On("ValidateToken", "error-token").Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid or expired token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{}
			tt.setupMock(mockService)

			router := setupTestRouter(mockService)

			req, _ := http.NewRequest("GET", "/api/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set(AuthorizationHeader, tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_ContextSetting(t *testing.T) {
	mockService := &MockAuthService{}
	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
		Name:   "Test User",
	}
	mockService.On("ValidateToken", "valid-token").Return(claims, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(AuthMiddleware(mockService))
	r.GET("/test", func(c *gin.Context) {
		user, exists := c.Get(KeyUser)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found in context"})
			return
		}

		userClaims, ok := user.(*Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type in context"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id": userClaims.UserID,
			"email":   userClaims.Email,
			"name":    userClaims.Name,
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(AuthorizationHeader, "Bearer valid-token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedBody := `{"user_id":123,"email":"test@example.com","name":"Test User"}`
	assert.JSONEq(t, expectedBody, w.Body.String())

	mockService.AssertExpectations(t)
}

func TestGetUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("user ID exists in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UserIDKey, uint(123))

		userID, exists := GetUserIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, uint(123), userID)
	})

	t.Run("user ID does not exist in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userID, exists := GetUserIDFromContext(c)
		assert.False(t, exists)
		assert.Equal(t, uint(0), userID)
	})

	t.Run("user ID has wrong type in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UserIDKey, "not-a-uint")

		userID, exists := GetUserIDFromContext(c)
		assert.False(t, exists)
		assert.Equal(t, uint(0), userID)
	})
}
