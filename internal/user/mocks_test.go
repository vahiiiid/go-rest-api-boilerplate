package user

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the user service for testing handlers
type MockService struct {
	mock.Mock
}

func (m *MockService) RegisterUser(contextutil context.Context, req RegisterRequest) (*User, error) {
	args := m.Called(contextutil, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) AuthenticateUser(contextutil context.Context, req LoginRequest) (*User, error) {
	args := m.Called(contextutil, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) GetUserByID(contextutil context.Context, id uint) (*User, error) {
	args := m.Called(contextutil, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) UpdateUser(contextutil context.Context, id uint, req UpdateUserRequest) (*User, error) {
	args := m.Called(contextutil, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) DeleteUser(contextutil context.Context, id uint) error {
	args := m.Called(contextutil, id)
	return args.Error(0)
}

// MockRepository is a mock implementation of the user repository for testing services
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(contextutil context.Context, user *User) error {
	args := m.Called(contextutil, user)
	return args.Error(0)
}

func (m *MockRepository) FindByEmail(contextutil context.Context, email string) (*User, error) {
	args := m.Called(contextutil, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) FindByID(contextutil context.Context, id uint) (*User, error) {
	args := m.Called(contextutil, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Update(contextutil context.Context, user *User) error {
	args := m.Called(contextutil, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(contextutil context.Context, id uint) error {
	args := m.Called(contextutil, id)
	return args.Error(0)
}
