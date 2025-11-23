package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requiredRole     string
		userRoles        []string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:           "user has required role",
			requiredRole:   "admin",
			userRoles:      []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:             "user missing required role",
			requiredRole:     "admin",
			userRoles:        []string{"user"},
			expectedStatus:   http.StatusForbidden,
			expectedResponse: "insufficient permissions",
		},
		{
			name:             "user has no roles",
			requiredRole:     "admin",
			userRoles:        []string{},
			expectedStatus:   http.StatusForbidden,
			expectedResponse: "insufficient permissions",
		},
		{
			name:           "user has multiple roles including required",
			requiredRole:   "editor",
			userRoles:      []string{"user", "editor", "viewer"},
			expectedStatus: http.StatusOK,
		},
		{
			name:             "no authenticated user",
			requiredRole:     "admin",
			userRoles:        nil,
			expectedStatus:   http.StatusForbidden,
			expectedResponse: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(func(c *gin.Context) {
				if tt.userRoles != nil {
					claims := &auth.Claims{
						UserID: 1,
						Email:  "test@example.com",
						Roles:  tt.userRoles,
					}
					c.Set(auth.KeyUser, claims)
				}
				c.Next()
			})

			router.Use(RequireRole(tt.requiredRole))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedResponse != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				if errorMsg, ok := response["error"].(string); ok {
					assert.Contains(t, errorMsg, tt.expectedResponse)
				}
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRoles      []string
		expectedStatus int
	}{
		{
			name:           "admin user allowed",
			userRoles:      []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-admin user forbidden",
			userRoles:      []string{"user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no user forbidden",
			userRoles:      nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "admin among multiple roles",
			userRoles:      []string{"user", "admin", "editor"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(func(c *gin.Context) {
				if tt.userRoles != nil {
					claims := &auth.Claims{
						UserID: 1,
						Email:  "admin@example.com",
						Roles:  tt.userRoles,
					}
					c.Set(auth.KeyUser, claims)
				}
				c.Next()
			})

			router.Use(RequireAdmin())
			router.GET("/admin", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
			})

			c.Request = httptest.NewRequest(http.MethodGet, "/admin", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
