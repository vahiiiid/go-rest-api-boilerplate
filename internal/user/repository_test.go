package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	return db
}

func TestNewRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	assert.NotNil(t, repo)
	assert.IsType(t, &repository{}, repo)
}

func TestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	user := &User{
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
	}

	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create first user
	user1 := &User{
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(context.Background(), user1)
	assert.NoError(t, err)

	// Try to create second user with same email
	user2 := &User{
		Name:         "Jane Doe",
		Email:        "john@example.com", // Same email
		PasswordHash: "another_password",
	}
	err = repo.Create(context.Background(), user2)
	assert.Error(t, err)
}

func TestRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create a test user
	originalUser := &User{
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(context.Background(), originalUser)
	require.NoError(t, err)

	t.Run("user found", func(t *testing.T) {
		user, err := repo.FindByEmail(context.Background(), "john@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		user, err := repo.FindByEmail(context.Background(), "notfound@example.com")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create a test user
	originalUser := &User{
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(context.Background(), originalUser)
	require.NoError(t, err)

	t.Run("user found", func(t *testing.T) {
		user, err := repo.FindByID(context.Background(), originalUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, originalUser.ID, user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		user, err := repo.FindByID(context.Background(), 999999)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create a test user
	user := &User{
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Update user
	user.Name = "Updated Name"
	user.Email = "updated@example.com"

	err = repo.Update(context.Background(), user)
	assert.NoError(t, err)

	// Verify update
	updatedUser, err := repo.FindByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
	assert.Equal(t, "updated@example.com", updatedUser.Email)
}

func TestRepository_Update_NonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	user := &User{
		ID:           999999, // Non-existent ID
		Name:         "Ghost User",
		Email:        "ghost@example.com",
		PasswordHash: "password",
	}

	err := repo.Update(context.Background(), user)
	// GORM doesn't return error for updating non-existent records
	// Instead, it returns no error but affects 0 rows
	assert.NoError(t, err)
}

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create a test user
	user := &User{
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
	}
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Delete user
	err = repo.Delete(context.Background(), user.ID)
	assert.NoError(t, err)

	// Verify deletion
	deletedUser, err := repo.FindByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Nil(t, deletedUser)
}

func TestRepository_Delete_NonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Try to delete non-existent user
	err := repo.Delete(context.Background(), 999999)
	// Repository returns error when no rows are affected (record not found)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}
