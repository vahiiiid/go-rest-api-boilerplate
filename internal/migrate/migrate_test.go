package migrate

import (
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunMigrationsAndRollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	// Run migrations
	RunMigrations(db)
	for _, model := range Models {
		typeName := getTypeName(model)
		if !db.Migrator().HasTable(model) {
			t.Errorf("expected table for %s to exist after migration", typeName)
		}
	}

	// Rollback migrations
	RollbackMigrations(db)
	for _, model := range Models {
		typeName := getTypeName(model)
		if db.Migrator().HasTable(model) {
			t.Errorf("expected table for %s to be dropped after rollback", typeName)
		}
	}
}

func getTypeName(i interface{}) string {
	return fmt.Sprintf("%T", i)
}
