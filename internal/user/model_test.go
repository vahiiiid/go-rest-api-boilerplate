package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_TableName(t *testing.T) {
	user := User{}
	tableName := user.TableName()
	
	assert.Equal(t, "users", tableName)
}

func TestToUserResponse_WithDates(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	response := ToUserResponse(user)
	
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "John Doe", response.Name)
	assert.Equal(t, "john@example.com", response.Email)
	assert.NotEmpty(t, response.CreatedAt)
	assert.NotEmpty(t, response.UpdatedAt)
}