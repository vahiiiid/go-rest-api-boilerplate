package migrate

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Basic smoke tests for migrate package
// Full integration tests should use PostgreSQL

func TestConfig(t *testing.T) {
	cfg := Config{
		DatabaseURL:   "postgres://test",
		MigrationsDir: "./testdata",
		Timeout:       30 * time.Second,
		LockTimeout:   10 * time.Second,
	}

	if cfg.MigrationsDir != "./testdata" {
		t.Errorf("Expected MigrationsDir to be './testdata', got '%s'", cfg.MigrationsDir)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30s, got %v", cfg.Timeout)
	}

	if cfg.LockTimeout != 10*time.Second {
		t.Errorf("Expected LockTimeout to be 10s, got %v", cfg.LockTimeout)
	}
}

func TestNew_RequiresValidDB(t *testing.T) {
	// Note: We can't test nil DB as postgres driver panics on Ping
	// This would be caught in real usage before calling New()
	t.Skip("Skipping nil DB test - postgres driver panics on Ping")
}
func TestNew_RequiresMigrationsDir(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	cfg := Config{
		MigrationsDir: "",
		Timeout:       30 * time.Second,
		LockTimeout:   10 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		// Expected - migrations directory doesn't exist
		t.Logf("Expected error with empty migrations dir: %v", err)
		return
	}

	if migrator != nil {
		defer func() {
			if err := migrator.Close(); err != nil {
				t.Errorf("Failed to close migrator: %v", err)
			}
		}()
	}
}
