package user

import (
	"bytes"
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

func TestHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful token refresh",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMocks: func(mas *MockAuthService) {
				tokenPair := &auth.TokenPair{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    900,
				}
				mas.On("RefreshAccessToken", mock.Anything, "valid-refresh-token").Return(tokenPair, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response auth.TokenPairResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "new-access-token", response.AccessToken)
				assert.Equal(t, "new-refresh-token", response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Equal(t, int64(900), response.ExpiresIn)
			},
		},
		{
			name:           "missing refresh token",
			requestBody:    map[string]string{},
			setupMocks:     func(mas *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "VALIDATION_ERROR", response.Code)
			},
		},
		{
			name: "invalid refresh token",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RefreshAccessToken", mock.Anything, "invalid-token").Return(nil, auth.ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "UNAUTHORIZED", response.Code)
				assert.Contains(t, response.Message, "Invalid or expired")
			},
		},
		{
			name: "expired refresh token",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "expired-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RefreshAccessToken", mock.Anything, "expired-token").Return(nil, auth.ErrExpiredToken)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "UNAUTHORIZED", response.Code)
				assert.Contains(t, response.Message, "Invalid or expired")
			},
		},
		{
			name: "token reuse detected",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "reused-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RefreshAccessToken", mock.Anything, "reused-token").Return(nil, auth.ErrTokenReuse)
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "FORBIDDEN", response.Code)
				assert.Contains(t, response.Message, "Token reuse detected")
			},
		},
		{
			name: "revoked token",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "revoked-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RefreshAccessToken", mock.Anything, "revoked-token").Return(nil, auth.ErrTokenRevoked)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "UNAUTHORIZED", response.Code)
				assert.Contains(t, response.Message, "revoked")
			},
		},
		{
			name: "internal server error",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "some-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RefreshAccessToken", mock.Anything, "some-token").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "INTERNAL_ERROR", response.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mockAuthService := new(MockAuthService)
			tt.setupMocks(mockAuthService)

			handler := &Handler{
				authService: mockAuthService,
			}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.RefreshToken(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService)
		setupContext   func(*gin.Context)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful logout",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RevokeUserRefreshToken", mock.Anything, uint(1), "valid-refresh-token").Return(nil)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Successfully logged out", response["message"])
			},
		},
		{
			name:        "missing refresh token",
			requestBody: map[string]string{},
			setupMocks:  func(mas *MockAuthService) {},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "VALIDATION_ERROR", response.Code)
			},
		},
		{
			name: "internal server error",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "some-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RevokeUserRefreshToken", mock.Anything, uint(1), "some-token").Return(errors.New("database error"))
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "INTERNAL_ERROR", response.Code)
			},
		},
		{
			name: "logout with non-existent token",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "non-existent-token",
			},
			setupMocks: func(mas *MockAuthService) {
				mas.On("RevokeUserRefreshToken", mock.Anything, uint(1), "non-existent-token").Return(nil)
			},
			setupContext: func(c *gin.Context) {
				claims := &auth.Claims{UserID: 1}
				c.Set(auth.KeyUser, claims)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Successfully logged out", response["message"])
			},
		},
		{
			name: "unauthenticated user",
			requestBody: auth.RefreshTokenRequest{
				RefreshToken: "some-token",
			},
			setupMocks: func(mas *MockAuthService) {},
			setupContext: func(c *gin.Context) {
				// No authentication context set
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response apiErrors.APIError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "UNAUTHORIZED", response.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			mockAuthService := new(MockAuthService)
			tt.setupMocks(mockAuthService)

			handler := &Handler{
				authService: mockAuthService,
			}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			tt.setupContext(c)

			handler.Logout(c)
			apiErrors.ErrorHandler()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockAuthService.AssertExpectations(t)
		})
	}
}
