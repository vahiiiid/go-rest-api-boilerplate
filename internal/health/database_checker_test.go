package health

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseChecker_Check(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	checker := NewDatabaseChecker(db)

	assert.Equal(t, "database", checker.Name())

	result := checker.Check(context.Background())

	assert.Equal(t, CheckPass, result.Status)
	assert.Contains(t, result.Message, "healthy")
	assert.NotEmpty(t, result.ResponseTime)
}
