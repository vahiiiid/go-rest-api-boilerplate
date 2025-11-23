package contextutil

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected *auth.Claims
	}{
		{
			name: "successful get user",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: &auth.Claims{
				UserID: 1,
				Email:  "test@example.com",
				Name:   "Test User",
			},
		},
		{
			name:     "user not found in context",
			setup:    func(c *gin.Context) {}, // Don't set anything
			expected: nil,
		},
		{
			name: "invalid type in context",
			setup: func(c *gin.Context) {
				c.Set(auth.KeyUser, "invalid-type")
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := GetUser(c)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.UserID, result.UserID)
				assert.Equal(t, tt.expected.Email, result.Email)
				assert.Equal(t, tt.expected.Name, result.Name)
			}
		})
	}
}

func TestMustGetUser(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*gin.Context)
		expectError bool
		expected    *auth.Claims
	}{
		{
			name: "successful get user",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			expectError: false,
			expected: &auth.Claims{
				UserID: 1,
				Email:  "test@example.com",
				Name:   "Test User",
			},
		},
		{
			name:        "user not found",
			setup:       func(c *gin.Context) {}, // Don't set anything
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result, err := MustGetUser(c)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, "user not found in context", err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.UserID, result.UserID)
				assert.Equal(t, tt.expected.Email, result.Email)
				assert.Equal(t, tt.expected.Name, result.Name)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected uint
	}{
		{
			name: "successful get user ID",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 42,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: 42,
		},
		{
			name:     "user not found",
			setup:    func(c *gin.Context) {}, // Don't set anything
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := GetUserID(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMustGetUserID(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*gin.Context)
		expectError bool
		expected    uint
	}{
		{
			name: "successful get user ID",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 42,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			expectError: false,
			expected:    42,
		},
		{
			name:        "user not found",
			setup:       func(c *gin.Context) {}, // Don't set anything
			expectError: true,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result, err := MustGetUserID(c)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, uint(0), result)
				assert.Equal(t, "user ID not found in context", err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetEmail(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected string
	}{
		{
			name: "successful get email",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: "test@example.com",
		},
		{
			name:     "user not found",
			setup:    func(c *gin.Context) {}, // Don't set anything
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := GetEmail(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected bool
	}{
		{
			name: "authenticated user",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: true,
		},
		{
			name:     "unauthenticated user",
			setup:    func(c *gin.Context) {}, // Don't set anything
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := IsAuthenticated(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCanAccessUser(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*gin.Context)
		targetUserID uint
		expected     bool
	}{
		{
			name: "can access own user",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			targetUserID: 1,
			expected:     true,
		},
		{
			name: "cannot access other user",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			targetUserID: 2,
			expected:     false,
		},
		{
			name:         "unauthenticated user",
			setup:        func(c *gin.Context) {}, // Don't set anything
			targetUserID: 1,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := CanAccessUser(c, tt.targetUserID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUserName(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected string
	}{
		{
			name: "successful get user name",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "test@example.com",
					Name:   "Test User",
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: "Test User",
		},
		{
			name:     "user not found",
			setup:    func(c *gin.Context) {}, // Don't set anything
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := GetUserName(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		role     string
		expected bool
	}{
		{
			name: "user has admin role",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "admin@example.com",
					Name:   "Admin User",
					Roles:  []string{"admin", "user"},
				}
				c.Set(auth.KeyUser, claims)
			},
			role:     "admin",
			expected: true,
		},
		{
			name: "user does not have role",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "user@example.com",
					Name:   "Regular User",
					Roles:  []string{"user"},
				}
				c.Set(auth.KeyUser, claims)
			},
			role:     "admin",
			expected: false,
		},
		{
			name:     "unauthenticated user",
			setup:    func(c *gin.Context) {},
			role:     "admin",
			expected: false,
		},
		{
			name: "user with no roles",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "user@example.com",
					Name:   "User",
					Roles:  []string{},
				}
				c.Set(auth.KeyUser, claims)
			},
			role:     "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := HasRole(c, tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRoles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected []string
	}{
		{
			name: "user with multiple roles",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "admin@example.com",
					Roles:  []string{"admin", "user", "editor"},
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: []string{"admin", "user", "editor"},
		},
		{
			name: "user with single role",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "user@example.com",
					Roles:  []string{"user"},
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: []string{"user"},
		},
		{
			name:     "unauthenticated user",
			setup:    func(c *gin.Context) {},
			expected: []string{},
		},
		{
			name: "user with no roles",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "user@example.com",
					Roles:  []string{},
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := GetRoles(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		expected bool
	}{
		{
			name: "user is admin",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "admin@example.com",
					Roles:  []string{"admin", "user"},
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: true,
		},
		{
			name: "user is not admin",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "user@example.com",
					Roles:  []string{"user", "editor"},
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: false,
		},
		{
			name:     "unauthenticated user",
			setup:    func(c *gin.Context) {},
			expected: false,
		},
		{
			name: "user with no roles",
			setup: func(c *gin.Context) {
				claims := &auth.Claims{
					UserID: 1,
					Email:  "user@example.com",
					Roles:  []string{},
				}
				c.Set(auth.KeyUser, claims)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setup(c)

			result := IsAdmin(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}
