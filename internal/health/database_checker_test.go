package health

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseChecker_Name(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	checker := NewDatabaseChecker(db)
	assert.Equal(t, "database", checker.Name())
}

func TestDatabaseChecker_Check_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	checker := NewDatabaseChecker(db)
	result := checker.Check(context.Background())

	assert.Equal(t, CheckPass, result.Status)
	assert.Contains(t, result.Message, "healthy")
	assert.NotEmpty(t, result.ResponseTime)
}

func TestDatabaseChecker_Check_MultipleRuns(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	checker := NewDatabaseChecker(db)

	for i := 0; i < 5; i++ {
		result := checker.Check(context.Background())
		assert.Equal(t, CheckPass, result.Status)
		assert.NotEmpty(t, result.ResponseTime)
	}
}
