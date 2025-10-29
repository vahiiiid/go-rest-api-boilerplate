package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository defines user repository interface
type Repository interface {
	Create(contextutil context.Context, user *User) error
	FindByEmail(contextutil context.Context, email string) (*User, error)
	FindByID(contextutil context.Context, id uint) (*User, error)
	Update(contextutil context.Context, user *User) error
	Delete(contextutil context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create creates a new user in the database
func (r *repository) Create(contextutil context.Context, user *User) error {
	result := r.db.WithContext(contextutil).Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByEmail finds a user by email
func (r *repository) FindByEmail(contextutil context.Context, email string) (*User, error) {
	var user User
	result := r.db.WithContext(contextutil).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *repository) FindByID(contextutil context.Context, id uint) (*User, error) {
	var user User
	result := r.db.WithContext(contextutil).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update updates a user in the database
func (r *repository) Update(contextutil context.Context, user *User) error {
	result := r.db.WithContext(contextutil).Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete soft deletes a user from the database
func (r *repository) Delete(contextutil context.Context, id uint) error {
	result := r.db.WithContext(contextutil).Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
