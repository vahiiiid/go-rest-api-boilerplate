package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository defines user repository interface
type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	ListAllUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error)
	AssignRole(ctx context.Context, userID uint, roleName string) error
	RemoveRole(ctx context.Context, userID uint, roleName string) error
	FindRoleByName(ctx context.Context, name string) (*Role, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
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
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByEmail finds a user by email
func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *repository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).Preload("Roles").First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update updates a user in the database
func (r *repository) Update(ctx context.Context, user *User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete soft deletes a user from the database
func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListAllUsers retrieves paginated list of users with filters
func (r *repository) ListAllUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error) {
	var users []User
	var total int64

	query := r.db.WithContext(ctx).Model(&User{}).Preload("Roles")

	if filters.Role != "" {
		query = query.Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("roles.name = ?", filters.Role)
	}

	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("users.name LIKE ? OR users.email LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	orderClause := filters.Sort + " " + filters.Order

	if err := query.Order(orderClause).Limit(perPage).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// AssignRole assigns a role to a user
func (r *repository) AssignRole(ctx context.Context, userID uint, roleName string) error {
	role, err := r.FindRoleByName(ctx, roleName)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	// Check if association already exists
	var count int64
	r.db.WithContext(ctx).Table("user_roles").
		Where("user_id = ? AND role_id = ?", userID, role.ID).
		Count(&count)

	if count > 0 {
		return nil // Already assigned
	}

	// Use raw SQL that works with both PostgreSQL and SQLite
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO user_roles (user_id, role_id, assigned_at) VALUES (?, ?, ?)",
		userID, role.ID, time.Now(),
	).Error
}

// RemoveRole removes a role from a user
func (r *repository) RemoveRole(ctx context.Context, userID uint, roleName string) error {
	role, err := r.FindRoleByName(ctx, roleName)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	return r.db.WithContext(ctx).Exec(
		"DELETE FROM user_roles WHERE user_id = ? AND role_id = ?",
		userID, role.ID,
	).Error
}

// FindRoleByName finds a role by name
func (r *repository) FindRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &role, nil
}

// GetUserRoles retrieves all roles for a user
func (r *repository) GetUserRoles(ctx context.Context, userID uint) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
