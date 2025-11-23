package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid strong password",
			password:    "StrongPass123!",
			expectError: false,
		},
		{
			name:        "password too short",
			password:    "Pass1!",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "missing uppercase",
			password:    "password123!",
			expectError: true,
			errorMsg:    "password must contain at least one uppercase letter",
		},
		{
			name:        "missing lowercase",
			password:    "PASSWORD123!",
			expectError: true,
			errorMsg:    "password must contain at least one lowercase letter",
		},
		{
			name:        "missing digit",
			password:    "PasswordAbc!",
			expectError: true,
			errorMsg:    "password must contain at least one digit",
		},
		{
			name:        "missing special character",
			password:    "Password123",
			expectError: true,
			errorMsg:    "password must contain at least one special character",
		},
		{
			name:        "valid with all requirements",
			password:    "Admin@2024Pass",
			expectError: false,
		},
		{
			name:        "valid with different special chars",
			password:    "Test#Pass123",
			expectError: false,
		},
		{
			name:        "exactly 8 characters valid",
			password:    "Pass123!",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailRegex(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email",
			email:    "admin@example.com",
			expected: true,
		},
		{
			name:     "valid email with subdomain",
			email:    "admin@mail.example.com",
			expected: true,
		},
		{
			name:     "valid email with plus",
			email:    "admin+test@example.com",
			expected: true,
		},
		{
			name:     "valid email with dot",
			email:    "admin.user@example.com",
			expected: true,
		},
		{
			name:     "valid email with numbers",
			email:    "admin123@example.com",
			expected: true,
		},
		{
			name:     "invalid email missing @",
			email:    "adminexample.com",
			expected: false,
		},
		{
			name:     "invalid email missing domain",
			email:    "admin@",
			expected: false,
		},
		{
			name:     "invalid email missing username",
			email:    "@example.com",
			expected: false,
		},
		{
			name:     "invalid email missing TLD",
			email:    "admin@example",
			expected: false,
		},
		{
			name:     "invalid email with spaces",
			email:    "admin @example.com",
			expected: false,
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailRegex.MatchString(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}
