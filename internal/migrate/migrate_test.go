package migrate

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunMigrationsAndRollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	for _, model := range Models {
		typeName := getTypeName(model)
		if !db.Migrator().HasTable(model) {
			t.Errorf("expected table for %s to exist after migration", typeName)
		}
	}

	if err := RollbackMigrations(db); err != nil {
		t.Fatalf("RollbackMigrations failed: %v", err)
	}
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

func TestShowMigrationStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ShowMigrationStatus(db)

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()
	if !strings.Contains(output, "Users table: ✓ EXISTS") {
		t.Errorf("expected status output to mention table exists, got: %s", output)
	}
}

func TestTableStatus(t *testing.T) {
	if tableStatus(true) != "✓ EXISTS" {
		t.Error("tableStatus(true) should return '✓ EXISTS'")
	}
	if tableStatus(false) != "✗ NOT FOUND" {
		t.Error("tableStatus(false) should return '✗ NOT FOUND'")
	}
}

func TestShowMigrationStatus_CountError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql DB: %v", err)
	}
	if cerr := sqlDB.Close(); cerr != nil {
		t.Fatalf("failed to close sql DB: %v", cerr)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ShowMigrationStatus(db)

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()
	if strings.Contains(output, "Users table: ✗ NOT FOUND") {
		// Table is missing, which is expected after DB close
		return
	}
	if !strings.Contains(output, "Users count: ERROR") {
		t.Errorf("expected error output for user count, got: %s", output)
	}
}
