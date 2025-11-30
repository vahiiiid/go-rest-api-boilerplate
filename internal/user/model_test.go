package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_TableName(t *testing.T) {
	user := User{}
	tableName := user.TableName()

	assert.Equal(t, "users", tableName)
}

func TestToUserResponse_WithDates(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := ToUserResponse(user)

	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "John Doe", response.Name)
	assert.Equal(t, "john@example.com", response.Email)
	assert.NotEmpty(t, response.CreatedAt)
	assert.NotEmpty(t, response.UpdatedAt)
}

func TestUser_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		roles    []Role
		roleName string
		expected bool
	}{
		{
			name: "user has role",
			roles: []Role{
				{Name: "admin"},
				{Name: "user"},
			},
			roleName: "admin",
			expected: true,
		},
		{
			name: "user does not have role",
			roles: []Role{
				{Name: "user"},
			},
			roleName: "admin",
			expected: false,
		},
		{
			name:     "user has no roles",
			roles:    []Role{},
			roleName: "admin",
			expected: false,
		},
		{
			name:     "nil roles",
			roles:    nil,
			roleName: "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:    1,
				Roles: tt.roles,
			}
			result := user.HasRole(tt.roleName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		roles    []Role
		expected bool
	}{
		{
			name: "user is admin",
			roles: []Role{
				{Name: "admin"},
			},
			expected: true,
		},
		{
			name: "user is not admin",
			roles: []Role{
				{Name: "user"},
			},
			expected: false,
		},
		{
			name:     "user has no roles",
			roles:    []Role{},
			expected: false,
		},
		{
			name: "user has multiple roles including admin",
			roles: []Role{
				{Name: "user"},
				{Name: "admin"},
				{Name: "editor"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:    1,
				Roles: tt.roles,
			}
			result := user.IsAdmin()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_GetRoleNames(t *testing.T) {
	tests := []struct {
		name     string
		roles    []Role
		expected []string
	}{
		{
			name: "single role",
			roles: []Role{
				{Name: "admin"},
			},
			expected: []string{"admin"},
		},
		{
			name: "multiple roles",
			roles: []Role{
				{Name: "user"},
				{Name: "admin"},
				{Name: "editor"},
			},
			expected: []string{"user", "admin", "editor"},
		},
		{
			name:     "no roles",
			roles:    []Role{},
			expected: []string{},
		},
		{
			name:     "nil roles",
			roles:    nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:    1,
				Roles: tt.roles,
			}
			result := user.GetRoleNames()
			assert.Equal(t, tt.expected, result)
		})
	}
}
