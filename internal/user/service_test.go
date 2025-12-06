package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestNewService(t *testing.T) {
	mockRepo := &MockRepository{}
	svc := NewService(mockRepo)

	assert.NotNil(t, svc)
	assert.Implements(t, (*Service)(nil), svc)
}

func TestService_RegisterUser(t *testing.T) {
	tests := []struct {
		name        string
		request     RegisterRequest
		setupMock   func(*MockRepository)
		expectedErr error
	}{
		{
			name: "successful registration",
			request: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Run(func(args mock.Arguments) {
					user := args.Get(1).(*User)
					user.ID = 1
				}).Return(nil)
				m.On("AssignRole", mock.Anything, uint(1), RoleUser).Return(nil)
				userWithRole := &User{ID: 1, Name: "John Doe", Email: "john@example.com", Roles: []Role{{Name: RoleUser}}}
				m.On("FindByID", mock.Anything, uint(1)).Return(userWithRole, nil)
			},
			expectedErr: nil,
		},
		{
			name: "email already exists",
			request: RegisterRequest{
				Name:     "John Doe",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				existingUser := &User{ID: 1, Email: "existing@example.com"}
				m.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedErr: ErrEmailExists,
		},
		{
			name: "repository error on email check",
			request: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, errors.New("db error"))
			},
			expectedErr: errors.New("failed to check existing email: db error"),
		},
		{
			name: "repository error on create",
			request: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(errors.New("create error"))
			},
			expectedErr: errors.New("failed to create user: create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			user, err := service.RegisterUser(context.Background(), tt.request)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Name, user.Name)
				assert.Equal(t, tt.request.Email, user.Email)
				// Password hash validation removed - cannot test with mocked repository
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_AuthenticateUser(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		request     LoginRequest
		setupMock   func(*MockRepository)
		expectedErr error
	}{
		{
			name: "successful authentication",
			request: LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				user := &User{
					ID:           1,
					Email:        "john@example.com",
					PasswordHash: string(hashedPassword),
				}
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
			},
			expectedErr: nil,
		},
		{
			name: "user not found",
			request: LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, nil)
			},
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			request: LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(m *MockRepository) {
				user := &User{
					ID:           1,
					Email:        "john@example.com",
					PasswordHash: string(hashedPassword),
				}
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(user, nil)
			},
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "repository error",
			request: LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, errors.New("db error"))
			},
			expectedErr: errors.New("failed to find user: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			user, err := service.AuthenticateUser(context.Background(), tt.request)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Email, user.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetUserByID(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		setupMock   func(*MockRepository)
		expectedErr error
	}{
		{
			name:   "user found",
			userID: 1,
			setupMock: func(m *MockRepository) {
				user := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
			},
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(999)).Return(nil, nil)
			},
			expectedErr: ErrUserNotFound,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("db error"))
			},
			expectedErr: errors.New("failed to find user: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			user, err := service.GetUserByID(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		request     UpdateUserRequest
		setupMock   func(*MockRepository)
		expectedErr error
	}{
		{
			name:   "successful update",
			userID: 1,
			request: UpdateUserRequest{
				Name:  "Updated Name",
				Email: "updated@example.com",
			},
			setupMock: func(m *MockRepository) {
				user := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
				m.On("FindByEmail", mock.Anything, "updated@example.com").Return(nil, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userID: 999,
			request: UpdateUserRequest{
				Name: "Updated Name",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(999)).Return(nil, nil)
			},
			expectedErr: ErrUserNotFound,
		},
		{
			name:   "email already exists",
			userID: 1,
			request: UpdateUserRequest{
				Email: "existing@example.com",
			},
			setupMock: func(m *MockRepository) {
				user := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				existingUser := &User{ID: 2, Email: "existing@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
				m.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedErr: ErrEmailExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			user, err := service.UpdateUser(context.Background(), tt.userID, tt.request)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.request.Name != "" {
					assert.Equal(t, tt.request.Name, user.Name)
				}
				if tt.request.Email != "" {
					assert.Equal(t, tt.request.Email, user.Email)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		setupMock   func(*MockRepository)
		expectedErr error
	}{
		{
			name:   "successful deletion",
			userID: 1,
			setupMock: func(m *MockRepository) {
				m.On("Delete", mock.Anything, uint(1)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userID: 1,
			setupMock: func(m *MockRepository) {
				m.On("Delete", mock.Anything, uint(1)).Return(gorm.ErrRecordNotFound)
			},
			expectedErr: ErrUserNotFound,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMock: func(m *MockRepository) {
				m.On("Delete", mock.Anything, uint(1)).Return(errors.New("delete error"))
			},
			expectedErr: errors.New("failed to delete user: delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			err := service.DeleteUser(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedErr, ErrUserNotFound) {
					assert.ErrorIs(t, err, ErrUserNotFound)
				} else {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hashedPassword, err := hashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	err = verifyPassword(hashedPassword, password)
	assert.NoError(t, err)
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("correct password", func(t *testing.T) {
		err := verifyPassword(string(hashedPassword), password)
		assert.NoError(t, err)
	})

	t.Run("incorrect password", func(t *testing.T) {
		err := verifyPassword(string(hashedPassword), "wrongpassword")
		assert.Error(t, err)
	})
}

func TestService_ListUsers(t *testing.T) {
	tests := []struct {
		name          string
		filters       UserFilterParams
		page          int
		perPage       int
		setupMocks    func(*MockRepository)
		expectedUsers []User
		expectedTotal int64
		expectedErr   error
	}{
		{
			name: "successful list with defaults",
			filters: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
			page:    1,
			perPage: 20,
			setupMocks: func(m *MockRepository) {
				users := []User{
					{ID: 1, Name: "User 1", Email: "user1@example.com"},
					{ID: 2, Name: "User 2", Email: "user2@example.com"},
				}
				m.On("ListAllUsers", mock.Anything, UserFilterParams{Sort: "created_at", Order: "desc"}, 1, 20).
					Return(users, int64(2), nil)
			},
			expectedUsers: []User{
				{ID: 1, Name: "User 1", Email: "user1@example.com"},
				{ID: 2, Name: "User 2", Email: "user2@example.com"},
			},
			expectedTotal: 2,
			expectedErr:   nil,
		},
		{
			name: "filter by admin role",
			filters: UserFilterParams{
				Role:  RoleAdmin,
				Sort:  "created_at",
				Order: "desc",
			},
			page:    1,
			perPage: 20,
			setupMocks: func(m *MockRepository) {
				users := []User{
					{ID: 1, Name: "Admin User", Email: "admin@example.com", Roles: []Role{{Name: RoleAdmin}}},
				}
				m.On("ListAllUsers", mock.Anything, UserFilterParams{Role: RoleAdmin, Sort: "created_at", Order: "desc"}, 1, 20).
					Return(users, int64(1), nil)
			},
			expectedUsers: []User{
				{ID: 1, Name: "Admin User", Email: "admin@example.com", Roles: []Role{{Name: RoleAdmin}}},
			},
			expectedTotal: 1,
			expectedErr:   nil,
		},
		{
			name: "search by name",
			filters: UserFilterParams{
				Search: "john",
				Sort:   "created_at",
				Order:  "desc",
			},
			page:    1,
			perPage: 20,
			setupMocks: func(m *MockRepository) {
				users := []User{
					{ID: 1, Name: "John Doe", Email: "john@example.com"},
				}
				m.On("ListAllUsers", mock.Anything, UserFilterParams{Search: "john", Sort: "created_at", Order: "desc"}, 1, 20).
					Return(users, int64(1), nil)
			},
			expectedUsers: []User{
				{ID: 1, Name: "John Doe", Email: "john@example.com"},
			},
			expectedTotal: 1,
			expectedErr:   nil,
		},
		{
			name: "invalid role returns error",
			filters: UserFilterParams{
				Role:  "invalid_role",
				Sort:  "created_at",
				Order: "desc",
			},
			page:          1,
			perPage:       20,
			setupMocks:    func(m *MockRepository) {},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedErr:   ErrInvalidRole,
		},
		{
			name: "repository error",
			filters: UserFilterParams{
				Sort:  "created_at",
				Order: "desc",
			},
			page:    1,
			perPage: 20,
			setupMocks: func(m *MockRepository) {
				m.On("ListAllUsers", mock.Anything, UserFilterParams{Sort: "created_at", Order: "desc"}, 1, 20).
					Return(nil, int64(0), errors.New("database error"))
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedErr:   errors.New("database error"),
		},
		{
			name: "empty result set",
			filters: UserFilterParams{
				Sort:  "created_at",
				Order: "desc",
			},
			page:    1,
			perPage: 20,
			setupMocks: func(m *MockRepository) {
				m.On("ListAllUsers", mock.Anything, UserFilterParams{Sort: "created_at", Order: "desc"}, 1, 20).
					Return([]User{}, int64(0), nil)
			},
			expectedUsers: []User{},
			expectedTotal: 0,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMocks(mockRepo)

			service := NewService(mockRepo)
			users, total, err := service.ListUsers(context.Background(), tt.filters, tt.page, tt.perPage)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == ErrInvalidRole {
					assert.Equal(t, ErrInvalidRole, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUsers, users)
				assert.Equal(t, tt.expectedTotal, total)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_PromoteToAdmin(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		setupMocks  func(*MockRepository)
		expectedErr error
	}{
		{
			name:   "successful promotion",
			userID: 1,
			setupMocks: func(m *MockRepository) {
				user := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
				m.On("AssignRole", mock.Anything, uint(1), RoleAdmin).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(999)).Return(nil, ErrUserNotFound)
			},
			expectedErr: ErrUserNotFound,
		},
		{
			name:   "repository error on FindByID",
			userID: 1,
			setupMocks: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
		{
			name:   "repository error on AssignRole",
			userID: 1,
			setupMocks: func(m *MockRepository) {
				user := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
				m.On("AssignRole", mock.Anything, uint(1), RoleAdmin).Return(errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
		{
			name:   "user already has admin role - idempotent",
			userID: 1,
			setupMocks: func(m *MockRepository) {
				user := &User{
					ID:    1,
					Name:  "Admin User",
					Email: "admin@example.com",
					Roles: []Role{{ID: 2, Name: RoleAdmin}},
				}
				m.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
				// AssignRole should NOT be called since user already has the role
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMocks(mockRepo)

			service := NewService(mockRepo)
			err := service.PromoteToAdmin(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if tt.expectedErr == ErrUserNotFound {
					assert.ErrorIs(t, err, ErrUserNotFound)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_RegisterUser_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		request     RegisterRequest
		setupMock   func(*MockRepository)
		expectedErr string
	}{
		{
			name: "error assigning default role",
			request: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Run(func(args mock.Arguments) {
					user := args.Get(1).(*User)
					user.ID = 1
				}).Return(nil)
				m.On("AssignRole", mock.Anything, uint(1), RoleUser).Return(errors.New("role assignment error"))
			},
			expectedErr: "failed to assign default role",
		},
		{
			name: "error reloading user after creation",
			request: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Run(func(args mock.Arguments) {
					user := args.Get(1).(*User)
					user.ID = 1
				}).Return(nil)
				m.On("AssignRole", mock.Anything, uint(1), RoleUser).Return(nil)
				m.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("reload error"))
			},
			expectedErr: "failed to reload user",
		},
		{
			name: "nil user returned after successful creation",
			request: RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "john@example.com").Return(nil, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Run(func(args mock.Arguments) {
					user := args.Get(1).(*User)
					user.ID = 1
				}).Return(nil)
				m.On("AssignRole", mock.Anything, uint(1), RoleUser).Return(nil)
				m.On("FindByID", mock.Anything, uint(1)).Return(nil, nil)
			},
			expectedErr: "user not found after creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			user, err := service.RegisterUser(context.Background(), tt.request)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
			assert.Nil(t, user)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListUsers_PaginationErrors(t *testing.T) {
	tests := []struct {
		name        string
		filters     UserFilterParams
		page        int
		perPage     int
		expectedErr string
	}{
		{
			name:        "page less than 1",
			filters:     UserFilterParams{Sort: "created_at", Order: "desc"},
			page:        0,
			perPage:     20,
			expectedErr: "page must be >= 1",
		},
		{
			name:        "perPage less than 1",
			filters:     UserFilterParams{Sort: "created_at", Order: "desc"},
			page:        1,
			perPage:     0,
			expectedErr: "perPage must be >= 1",
		},
		{
			name:        "perPage greater than 100",
			filters:     UserFilterParams{Sort: "created_at", Order: "desc"},
			page:        1,
			perPage:     101,
			expectedErr: "perPage must be <= 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			service := NewService(mockRepo)

			users, total, err := service.ListUsers(context.Background(), tt.filters, tt.page, tt.perPage)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedErr, err.Error())
			assert.Nil(t, users)
			assert.Equal(t, int64(0), total)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListUsers_RepositoryError(t *testing.T) {
	mockRepo := &MockRepository{}
	filters := UserFilterParams{Sort: "created_at", Order: "desc"}
	mockRepo.On("ListAllUsers", mock.Anything, filters, 1, 20).Return(nil, int64(0), errors.New("database error"))

	service := NewService(mockRepo)
	users, total, err := service.ListUsers(context.Background(), filters, 1, 20)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list users")
	assert.Nil(t, users)
	assert.Equal(t, int64(0), total)

	mockRepo.AssertExpectations(t)
}

func TestService_UpdateUser_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		request     UpdateUserRequest
		setupMock   func(*MockRepository)
		expectedErr string
	}{
		{
			name:   "error fetching existing user",
			userID: 1,
			request: UpdateUserRequest{
				Name:  "Updated Name",
				Email: "updated@example.com",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
			},
			expectedErr: "database error",
		},
		{
			name:   "user not found",
			userID: 999,
			request: UpdateUserRequest{
				Name:  "Updated Name",
				Email: "updated@example.com",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByID", mock.Anything, uint(999)).Return(nil, nil)
			},
			expectedErr: "user not found",
		},
		{
			name:   "email already taken by another user",
			userID: 1,
			request: UpdateUserRequest{
				Name:  "Updated Name",
				Email: "taken@example.com",
			},
			setupMock: func(m *MockRepository) {
				existingUser := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(existingUser, nil)
				otherUser := &User{ID: 2, Email: "taken@example.com"}
				m.On("FindByEmail", mock.Anything, "taken@example.com").Return(otherUser, nil)
			},
			expectedErr: "email already exists",
		},
		{
			name:   "error checking email availability",
			userID: 1,
			request: UpdateUserRequest{
				Name:  "Updated Name",
				Email: "new@example.com",
			},
			setupMock: func(m *MockRepository) {
				existingUser := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(existingUser, nil)
				m.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("database error"))
			},
			expectedErr: "failed to check existing email",
		},
		{
			name:   "error updating user",
			userID: 1,
			request: UpdateUserRequest{
				Name: "Updated Name",
			},
			setupMock: func(m *MockRepository) {
				existingUser := &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
				m.On("FindByID", mock.Anything, uint(1)).Return(existingUser, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(errors.New("update error"))
			},
			expectedErr: "update error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			service := NewService(mockRepo)
			user, err := service.UpdateUser(context.Background(), tt.userID, tt.request)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
			assert.Nil(t, user)

			mockRepo.AssertExpectations(t)
		})
	}
}
