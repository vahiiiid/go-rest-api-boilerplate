package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) RegisterUser(ctx context.Context, req user.RegisterRequest) (*user.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockService) AuthenticateUser(ctx context.Context, req user.LoginRequest) (*user.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockService) GetUserByID(ctx context.Context, id uint) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, id uint, req user.UpdateUserRequest) (*user.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) ListUsers(ctx context.Context, filters user.UserFilterParams, page, perPage int) ([]user.User, int64, error) {
	args := m.Called(ctx, filters, page, perPage)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]user.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) PromoteToAdmin(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

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

func TestUserIsAdmin(t *testing.T) {
	tests := []struct {
		name    string
		user    *user.User
		isAdmin bool
	}{
		{
			name: "user with admin role",
			user: &user.User{
				Roles: []user.Role{
					{ID: 1, Name: "admin"},
				},
			},
			isAdmin: true,
		},
		{
			name: "user with multiple roles including admin",
			user: &user.User{
				Roles: []user.Role{
					{ID: 2, Name: "user"},
					{ID: 1, Name: "admin"},
				},
			},
			isAdmin: true,
		},
		{
			name: "user without admin role",
			user: &user.User{
				Roles: []user.Role{
					{ID: 2, Name: "user"},
				},
			},
			isAdmin: false,
		},
		{
			name: "user with no roles",
			user: &user.User{
				Roles: []user.Role{},
			},
			isAdmin: false,
		},
		{
			name: "user with similar but not admin role",
			user: &user.User{
				Roles: []user.Role{
					{ID: 3, Name: "administrator"},
				},
			},
			isAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsAdmin()
			assert.Equal(t, tt.isAdmin, result)
		})
	}
}

func TestPasswordValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "password with multiple special chars",
			password: "P@ss!w#0rd$%",
			wantErr:  false,
		},
		{
			name:     "very long password",
			password: "ThisIsAVeryLongPassword123!WithManyCharacters",
			wantErr:  false,
		},
		{
			name:     "password with all uppercase except one",
			password: "PASSWORd123!",
			wantErr:  false,
		},
		{
			name:     "password with spaces",
			password: "Pass word123!",
			wantErr:  false,
		},
		{
			name:     "password with brackets",
			password: "Pass[word]123",
			wantErr:  false,
		},
		{
			name:     "password with braces",
			password: "Pass{word}123",
			wantErr:  false,
		},
		{
			name:     "password with slash",
			password: "Pass/word123",
			wantErr:  false,
		},
		{
			name:     "password with backslash",
			password: "Pass\\word123",
			wantErr:  false,
		},
		{
			name:     "password with pipe",
			password: "Pass|word123",
			wantErr:  false,
		},
		{
			name:     "password with comma",
			password: "Pass,word123",
			wantErr:  false,
		},
		{
			name:     "password with period",
			password: "Pass.word123",
			wantErr:  false,
		},
		{
			name:     "password with semicolon",
			password: "Pass;word123",
			wantErr:  false,
		},
		{
			name:     "password with colon",
			password: "Pass:word123",
			wantErr:  false,
		},
		{
			name:     "password with quotes",
			password: "Pass'word\"123",
			wantErr:  false,
		},
		{
			name:     "password with angle brackets",
			password: "Pass<word>123",
			wantErr:  false,
		},
		{
			name:     "password with question mark",
			password: "Pass?word123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordRequirementsCombinations(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "only lowercase and length",
			password: "password",
			wantErr:  true,
			errMsg:   "uppercase",
		},
		{
			name:     "only uppercase and length",
			password: "PASSWORD",
			wantErr:  true,
			errMsg:   "lowercase",
		},
		{
			name:     "only digits and length",
			password: "12345678",
			wantErr:  true,
			errMsg:   "uppercase",
		},
		{
			name:     "only special chars and length",
			password: "!@#$%^&*",
			wantErr:  true,
			errMsg:   "uppercase",
		},
		{
			name:     "uppercase lowercase only",
			password: "PasswordOnly",
			wantErr:  true,
			errMsg:   "digit",
		},
		{
			name:     "uppercase digit only",
			password: "PASSWORD123",
			wantErr:  true,
			errMsg:   "lowercase",
		},
		{
			name:     "lowercase digit only",
			password: "password123",
			wantErr:  true,
			errMsg:   "uppercase",
		},
		{
			name:     "uppercase lowercase digit only",
			password: "Password123",
			wantErr:  true,
			errMsg:   "special",
		},
		{
			name:     "uppercase lowercase special only",
			password: "Password!",
			wantErr:  true,
			errMsg:   "digit",
		},
		{
			name:     "uppercase digit special only",
			password: "PASSWORD123!",
			wantErr:  true,
			errMsg:   "lowercase",
		},
		{
			name:     "lowercase digit special only",
			password: "password123!",
			wantErr:  true,
			errMsg:   "uppercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailRegexEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email with hyphen in domain",
			email:    "user@my-domain.com",
			expected: true,
		},
		{
			name:     "valid email with multiple dots",
			email:    "user.name.test@example.co.uk",
			expected: true,
		},
		{
			name:     "valid email with numbers in domain",
			email:    "user@example123.com",
			expected: true,
		},
		{
			name:     "valid email with underscore",
			email:    "user_name@example.com",
			expected: true,
		},
		{
			name:     "valid email with percent",
			email:    "user%name@example.com",
			expected: true,
		},
		{
			name:     "invalid email with double @",
			email:    "user@@example.com",
			expected: false,
		},
		{
			name:     "invalid email with special chars in domain",
			email:    "user@exam!ple.com",
			expected: false,
		},
		{
			name:     "invalid email ending with dot",
			email:    "user@example.com.",
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

func TestPasswordLengthBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "7 characters with all requirements",
			password: "Pass1!a",
			wantErr:  true,
		},
		{
			name:     "8 characters with all requirements",
			password: "Pass1!ab",
			wantErr:  false,
		},
		{
			name:     "9 characters with all requirements",
			password: "Pass1!abc",
			wantErr:  false,
		},
		{
			name:     "empty string",
			password: "",
			wantErr:  true,
		},
		{
			name:     "one character",
			password: "P",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "invalid email format",
			email:   "invalid-email",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email without @",
			email:   "userexample.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "email without domain",
			email:   "user@",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid name",
			input:   "John Doe",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "single character name",
			input:   "J",
			wantErr: false,
		},
		{
			name:    "name with numbers",
			input:   "User123",
			wantErr: false,
		},
		{
			name:    "name with special characters",
			input:   "John-Doe O'Brien",
			wantErr: false,
		},
		{
			name:    "name exactly 255 characters",
			input:   string(make([]byte, 255)),
			wantErr: false,
		},
		{
			name:    "name exceeding 255 characters",
			input:   string(make([]byte, 256)),
			wantErr: true,
			errMsg:  "name is too long",
		},
		{
			name:    "name with unicode characters",
			input:   "José María",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testInput := tt.input
			if tt.name == "name exactly 255 characters" {
				testInput = ""
				for i := 0; i < 255; i++ {
					testInput += "a"
				}
			} else if tt.name == "name exceeding 255 characters" {
				testInput = ""
				for i := 0; i < 256; i++ {
					testInput += "a"
				}
			}

			err := validateName(testInput)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckPasswordsMatch(t *testing.T) {
	tests := []struct {
		name            string
		password        string
		confirmPassword string
		wantErr         bool
	}{
		{
			name:            "matching passwords",
			password:        "Password123!",
			confirmPassword: "Password123!",
			wantErr:         false,
		},
		{
			name:            "non-matching passwords",
			password:        "Password123!",
			confirmPassword: "Different123!",
			wantErr:         true,
		},
		{
			name:            "empty passwords",
			password:        "",
			confirmPassword: "",
			wantErr:         false,
		},
		{
			name:            "one empty password",
			password:        "Password123!",
			confirmPassword: "",
			wantErr:         true,
		},
		{
			name:            "case sensitive mismatch",
			password:        "Password123!",
			confirmPassword: "password123!",
			wantErr:         true,
		},
		{
			name:            "whitespace difference",
			password:        "Password123!",
			confirmPassword: "Password123! ",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkPasswordsMatch(tt.password, tt.confirmPassword)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "passwords do not match")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromoteUserToAdmin(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupMock func(*MockService)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "successful promotion",
			userID: 1,
			setupMock: func(ms *MockService) {
				existingUser := &user.User{
					ID:    1,
					Email: "user@example.com",
					Name:  "Test User",
					Roles: []user.Role{{ID: 2, Name: "user"}},
				}
				ms.On("GetUserByID", mock.Anything, uint(1)).Return(existingUser, nil)
				ms.On("PromoteToAdmin", mock.Anything, uint(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(ms *MockService) {
				ms.On("GetUserByID", mock.Anything, uint(999)).Return(nil, fmt.Errorf("user not found"))
			},
			wantErr: true,
			errMsg:  "failed to find user",
		},
		{
			name:   "user already admin",
			userID: 2,
			setupMock: func(ms *MockService) {
				adminUser := &user.User{
					ID:    2,
					Email: "admin@example.com",
					Name:  "Admin User",
					Roles: []user.Role{{ID: 1, Name: "admin"}},
				}
				ms.On("GetUserByID", mock.Anything, uint(2)).Return(adminUser, nil)
			},
			wantErr: false,
		},
		{
			name:   "promotion fails",
			userID: 3,
			setupMock: func(ms *MockService) {
				existingUser := &user.User{
					ID:    3,
					Email: "user@example.com",
					Name:  "Test User",
					Roles: []user.Role{{ID: 2, Name: "user"}},
				}
				ms.On("GetUserByID", mock.Anything, uint(3)).Return(existingUser, nil)
				ms.On("PromoteToAdmin", mock.Anything, uint(3)).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "failed to promote user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.setupMock(mockService)

			err := promoteUserToAdmin(context.Background(), mockService, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestRegisterAndPromoteUser(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		userName  string
		setupMock func(*MockService)
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "successful registration and promotion",
			email:    "newadmin@example.com",
			password: "Password123!",
			userName: "New Admin",
			setupMock: func(ms *MockService) {
				newUser := &user.User{
					ID:    1,
					Email: "newadmin@example.com",
					Name:  "New Admin",
				}
				ms.On("RegisterUser", mock.Anything, mock.MatchedBy(func(req user.RegisterRequest) bool {
					return req.Email == "newadmin@example.com" &&
						req.Password == "Password123!" &&
						req.Name == "New Admin"
				})).Return(newUser, nil)
				ms.On("PromoteToAdmin", mock.Anything, uint(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "registration fails",
			email:    "exists@example.com",
			password: "Password123!",
			userName: "Existing User",
			setupMock: func(ms *MockService) {
				ms.On("RegisterUser", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("email already exists"))
			},
			wantErr: true,
			errMsg:  "failed to create user",
		},
		{
			name:     "promotion fails after registration",
			email:    "newuser@example.com",
			password: "Password123!",
			userName: "New User",
			setupMock: func(ms *MockService) {
				newUser := &user.User{
					ID:    2,
					Email: "newuser@example.com",
					Name:  "New User",
				}
				ms.On("RegisterUser", mock.Anything, mock.Anything).Return(newUser, nil)
				ms.On("PromoteToAdmin", mock.Anything, uint(2)).Return(fmt.Errorf("role assignment failed"))
			},
			wantErr: true,
			errMsg:  "failed to promote user to admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.setupMock(mockService)

			result, err := registerAndPromoteUser(context.Background(), mockService, tt.email, tt.password, tt.userName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
				assert.Equal(t, tt.userName, result.Name)
			}

			mockService.AssertExpectations(t)
		})
	}
}
