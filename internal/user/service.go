package user

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailExists is returned when email already exists
	ErrEmailExists = errors.New("email already exists")
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Service defines user service interface
type Service interface {
	RegisterUser(ctx context.Context, req RegisterRequest) (*User, error)
	AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error)
	GetUserByID(ctx context.Context, id uint) (*User, error)
	UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uint) error
	UpdatePassword(ctx context.Context, id uint, newPlainPassword string) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// RegisterUser registers a new user
func (s *service) RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
	// Check if email already exists
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *service) AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := verifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser updates a user's information
func (s *service) UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// Check if new email is already taken by another user
		existingUser, err := s.repo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailExists
		}
		user.Email = req.Email
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *service) DeleteUser(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// UpdatePassword updates a user's password hash
func (s *service) UpdatePassword(ctx context.Context, id uint, newPlainPassword string) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	hashedPassword, err := hashPassword(newPlainPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hashedPassword
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}
	return nil
}

// FindByEmail exposes repository's FindByEmail for consumers that need it
func (s *service) FindByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.FindByEmail(ctx, email)
}

// UserPasswordAdapter implements auth.UserPasswordPort on top of user.Service
type UserPasswordAdapter struct {
	svc Service
}

// NewUserPasswordAdapter wraps a user.Service to satisfy the minimal port
func NewUserPasswordAdapter(svc Service) *UserPasswordAdapter {
	return &UserPasswordAdapter{svc: svc}
}

func (a *UserPasswordAdapter) FindUserIDByEmail(ctx context.Context, email string) (uint, error) {
	user, err := a.svc.FindByEmail(ctx, email)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, nil
	}
	return user.ID, nil
}

func (a *UserPasswordAdapter) UpdatePassword(ctx context.Context, id uint, newPlainPassword string) error {
	return a.svc.UpdatePassword(ctx, id, newPlainPassword)
}

// hashPassword hashes a plain text password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a hash
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
