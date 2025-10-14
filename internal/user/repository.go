package user

import (
	"context"
	"errors"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Repository defines user repository interface
type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create creates a new user in the database
func (r *repository) Create(ctx context.Context, user *User) error {
	result := r.db.WithContext(ctx).Create(user)

	// Get unique requestID
	requestID, _ := ctx.Value("request_id").(string)

	if result.Error != nil {
		logger.Error("Database error",
			zap.Error(result.Error),
			zap.String("request_id", requestID),
			zap.String("operation", "create_user"),
			zap.String("table", "users"),
		)
		return result.Error
	}

	logger.Info("User created",
		zap.String("request_id", requestID),
		zap.Uint("user_id", user.ID),
		zap.String("name", user.Name),
		zap.String("email", user.Email),
	)
	return nil
}

// FindByEmail finds a user by email
func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)

	// Get unique requestID
	requestID, _ := ctx.Value("request_id").(string)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Info("User not found",
				zap.Error(result.Error),
				zap.String("request_id", requestID),
				zap.String("email", user.Email),
				zap.String("operation", "find_user_by_email"),
				zap.String("table", "users"),
			)
			return nil, nil
		}

		logger.Error("Database error",
			zap.Error(result.Error),
			zap.String("request_id", requestID),
			zap.String("email", user.Email),
			zap.String("operation", "find_user_by_email"),
			zap.String("table", "users"),
		)
		return nil, result.Error
	}

	logger.Info("User found",
		zap.String("request_id", requestID),
		zap.Uint("user_id", user.ID),
		zap.String("name", user.Name),
		zap.String("email", user.Email),
		zap.String("operation", "find_user_by_email"),
		zap.String("table", "users"),
	)
	return &user, nil
}

// FindByID finds a user by ID
func (r *repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).First(&user, id)

	// Get unique requestID
	requestID, _ := ctx.Value("request_id").(string)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Info("User not found",
				zap.Error(result.Error),
				zap.String("request_id", requestID),
				zap.Uint("user_id", user.ID),
				zap.String("operation", "find_user_by_id"),
				zap.String("table", "users"),
			)
			return nil, nil
		}

		logger.Error("Database error",
			zap.Error(result.Error),
			zap.String("request_id", requestID),
			zap.Uint("user_id", user.ID),

			zap.String("operation", "find_user_by_id"),
			zap.String("table", "users"),
		)
		return nil, result.Error
	}

	logger.Info("User found",
		zap.String("request_id", requestID),
		zap.Uint("user_id", user.ID),
		zap.String("name", user.Name),
		zap.String("email", user.Email),
		zap.String("operation", "find_user_by_id"),
		zap.String("table", "users"),
	)
	return &user, nil
}

// Update updates a user in the database
func (r *repository) Update(ctx context.Context, user *User) error {
	result := r.db.WithContext(ctx).Save(user)

	// Get unique requestID
	requestID, _ := ctx.Value("request_id").(string)

	if result.Error != nil {
		logger.Error("Database error",
			zap.Error(result.Error),
			zap.String("request_id", requestID),
			zap.Uint("user_id", user.ID),
			zap.String("name", user.Name),
			zap.String("email", user.Email),
			zap.String("operation", "update_user"),
			zap.String("table", "users"),
		)
		return result.Error
	}

	logger.Info("User updated",
		zap.String("request_id", requestID),
		zap.Uint("user_id", user.ID),
		zap.String("name", user.Name),
		zap.String("email", user.Email),
	)
	return nil
}

// Delete soft deletes a user from the database
func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&User{}, id)

	// Get unique requestID
	requestID, _ := ctx.Value("request_id").(string)

	if result.Error != nil {
		logger.Error("Database error",
			zap.Error(result.Error),
			zap.String("request_id", requestID),
			zap.Uint("user_id", id),
			zap.String("operation", "delete_user"),
			zap.String("table", "users"),
		)
		return result.Error
	}
	if result.RowsAffected == 0 {
		logger.Info("User not found",
			zap.String("request_id", requestID),
			zap.Uint("user_id", id),
			zap.String("operation", "delete_user"),
			zap.String("table", "users"),
		)
		return gorm.ErrRecordNotFound
	}

	logger.Info("User deleted",
		zap.String("request_id", requestID),
		zap.Uint("user_id", id),
	)
	return nil
}
