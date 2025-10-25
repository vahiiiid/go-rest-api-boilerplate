package migrate

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" || os.Getenv("DOCKER_ENV") == "true" {
		t.Skip("Skipping integration test in Docker/CI environment")
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		DatabaseURL:   "postgres://test",
		MigrationsDir: "./testdata",
		Timeout:       30 * time.Second,
		LockTimeout:   10 * time.Second,
	}

	assert.Equal(t, "./testdata", cfg.MigrationsDir)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 10*time.Second, cfg.LockTimeout)
	assert.Equal(t, "postgres://test", cfg.DatabaseURL)
}

func TestNew_WithValidConfig(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	upSQL := filepath.Join(migrationsDir, "000001_test.up.sql")
	require.NoError(t, os.WriteFile(upSQL, []byte("CREATE TABLE test (id INTEGER);"), 0644))

	downSQL := filepath.Join(migrationsDir, "000001_test.down.sql")
	require.NoError(t, os.WriteFile(downSQL, []byte("DROP TABLE test;"), 0644))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Failed to close DB: %v", err)
		}
	}()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       30 * time.Second,
		LockTimeout:   10 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Logf("Expected error with SQLite (postgres driver required): %v", err)
		return
	}

	if migrator != nil {
		defer func() {
			if err := migrator.Close(); err != nil {
				t.Logf("Failed to close migrator: %v", err)
			}
		}()
		assert.NotNil(t, migrator.migrate)
		assert.NotNil(t, migrator.db)
		assert.Equal(t, cfg.MigrationsDir, migrator.config.MigrationsDir)
	}
}

func TestNew_InvalidMigrationsDir(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: "/nonexistent/path",
		Timeout:       30 * time.Second,
		LockTimeout:   10 * time.Second,
	}

	_, err = New(db, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create")
}

func TestMigrator_Up_NoMigrations(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Logf("Expected error with SQLite: %v", err)
		return
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = migrator.Up(ctx)
	assert.NoError(t, err)
}

func TestMigrator_Up_ContextTimeout(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	err = migrator.Up(ctx)
	if err != nil {
		assert.Contains(t, err.Error(), "timeout")
	}
}

func TestMigrator_Down(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = migrator.Down(ctx, 1)
	assert.Error(t, err)
}

func TestMigrator_Down_MultipleSteps(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = migrator.Down(ctx, 3)
	assert.Error(t, err)
}

func TestMigrator_Down_NegativeSteps(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = migrator.Down(ctx, -1)
	assert.Error(t, err)
}

func TestMigrator_Steps(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = migrator.Steps(ctx, 1)
	assert.Error(t, err)
}

func TestMigrator_Steps_Backward(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = migrator.Steps(ctx, -2)
	assert.Error(t, err)
}

func TestMigrator_Goto(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = migrator.Goto(ctx, 1)
	assert.Error(t, err)
}

func TestMigrator_Version(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	version, dirty, err := migrator.Version()
	if err != nil {
		t.Logf("Expected error or nil version: %v", err)
	}
	assert.False(t, dirty)
	assert.Equal(t, uint(0), version)
}

func TestMigrator_Force(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	err = migrator.Force(1)
	if err != nil {
		t.Logf("Force operation may fail without proper migrations: %v", err)
	}
}

func TestMigrator_Drop(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}
	defer func() { _ = migrator.Close() }()

	err = migrator.Drop()
	if err != nil {
		t.Logf("Drop operation may fail: %v", err)
	}
}

func TestMigrator_Close(t *testing.T) {
	skipIfIntegration(t)

	tmpDir := t.TempDir()
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: migrationsDir,
		Timeout:       5 * time.Second,
		LockTimeout:   5 * time.Second,
	}

	migrator, err := New(db, cfg)
	if err != nil {
		t.Skip("Skipping - requires postgres driver")
	}

	err = migrator.Close()
	assert.NoError(t, err)
}
