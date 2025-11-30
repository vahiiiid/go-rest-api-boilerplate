package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type txKey struct{}

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
	Transaction(ctx context.Context, fn func(context.Context) error) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// getDB returns the DB from context if in transaction, otherwise returns the repository's DB
func (r *repository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// Create creates a new user in the database
func (r *repository) Create(ctx context.Context, user *User) error {
	result := r.getDB(ctx).WithContext(ctx).Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByEmail finds a user by email
func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	result := r.getDB(ctx).WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&user)
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
	result := r.getDB(ctx).WithContext(ctx).Preload("Roles").First(&user, id)
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
	// WHY: Save() syncs associations, potentially clearing roles
	result := r.getDB(ctx).WithContext(ctx).Select("name", "email", "password_hash", "updated_at").Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete soft deletes a user from the database
func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.getDB(ctx).WithContext(ctx).Delete(&User{}, id)
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

	query := r.getDB(ctx).WithContext(ctx).Model(&User{}).Preload("Roles")

	if filters.Role != "" {
		query = query.Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("roles.name = ?", filters.Role)
	}

	if filters.Search != "" {
		// WHY: Escape SQL LIKE wildcards to prevent incorrect matches
		escapedSearch := strings.ReplaceAll(filters.Search, "%", "\\%")
		escapedSearch = strings.ReplaceAll(escapedSearch, "_", "\\_")
		searchPattern := "%" + escapedSearch + "%"
		query = query.Where("users.name LIKE ? OR users.email LIKE ?", searchPattern, searchPattern)
	}

	// WHY: Count distinct user IDs when using JOINs to avoid inflated totals
	if err := query.Distinct("users.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage

	// Defense-in-depth: Validate sort parameters at repository layer
	validSorts := map[string]bool{
		"name": true, "email": true, "created_at": true, "updated_at": true,
	}
	if !validSorts[filters.Sort] {
		return nil, 0, errors.New("invalid sort field")
	}
	if filters.Order != "asc" && filters.Order != "desc" {
		return nil, 0, errors.New("invalid sort order")
	}

	// Use type-safe GORM clause to prevent SQL injection
	orderColumn := clause.OrderByColumn{
		Column: clause.Column{Table: "users", Name: filters.Sort},
		Desc:   filters.Order == "desc",
	}

	// WHY: Use Distinct with explicit columns to avoid duplicate users with JOINs
	if err := query.Distinct("users.*").Order(orderColumn).Limit(perPage).Offset(offset).Find(&users).Error; err != nil {
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

	// Use database-level conflict handling for race-safe, idempotent role assignment
	// Works with both PostgreSQL and SQLite
	return r.getDB(ctx).WithContext(ctx).Exec(`
		INSERT INTO user_roles (user_id, role_id, assigned_at)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, role.ID, time.Now()).Error
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

	return r.getDB(ctx).WithContext(ctx).Exec(
		"DELETE FROM user_roles WHERE user_id = ? AND role_id = ?",
		userID, role.ID,
	).Error
}

// FindRoleByName finds a role by name
func (r *repository) FindRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	result := r.getDB(ctx).WithContext(ctx).Where("name = ?", name).First(&role)
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
	err := r.getDB(ctx).WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// Transaction executes a function within a database transaction
func (r *repository) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Inject transaction into context
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
