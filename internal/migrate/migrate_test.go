package migrate

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMigrate struct {
	upFunc      func() error
	stepsFunc   func(n int) error
	migrateFunc func(version uint) error
	versionFunc func() (uint, bool, error)
	forceFunc   func(version int) error
	dropFunc    func() error
	closeFunc   func() (error, error)
}

func (m *mockMigrate) Up() error {
	if m.upFunc != nil {
		return m.upFunc()
	}
	return nil
}

func (m *mockMigrate) Steps(n int) error {
	if m.stepsFunc != nil {
		return m.stepsFunc(n)
	}
	return nil
}

func (m *mockMigrate) Migrate(version uint) error {
	if m.migrateFunc != nil {
		return m.migrateFunc(version)
	}
	return nil
}

func (m *mockMigrate) Version() (uint, bool, error) {
	if m.versionFunc != nil {
		return m.versionFunc()
	}
	return 0, false, nil
}

func (m *mockMigrate) Force(version int) error {
	if m.forceFunc != nil {
		return m.forceFunc(version)
	}
	return nil
}

func (m *mockMigrate) Drop() error {
	if m.dropFunc != nil {
		return m.dropFunc()
	}
	return nil
}

func (m *mockMigrate) Close() (error, error) {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil, nil
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

func TestNew_InvalidDatabase(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: "./testdata",
		Timeout:       30 * time.Second,
		LockTimeout:   10 * time.Second,
	}

	_, err = New(db, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create")
}

func TestNew_InvalidMigrationsDir(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MigrationsDir: "/nonexistent/path/that/does/not/exist",
		Timeout:       30 * time.Second,
		LockTimeout:   10 * time.Second,
	}

	_, err = New(db, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create")
}

func TestMigrator_Up_Success(t *testing.T) {
	mock := &mockMigrate{
		upFunc: func() error {
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Up(ctx)
	assert.NoError(t, err)
}

func TestMigrator_Up_NoChange(t *testing.T) {
	mock := &mockMigrate{
		upFunc: func() error {
			return migrate.ErrNoChange
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Up(ctx)
	assert.NoError(t, err)
}

func TestMigrator_Up_Error(t *testing.T) {
	mock := &mockMigrate{
		upFunc: func() error {
			return errors.New("migration failed")
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Up(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration failed")
}

func TestMigrator_Up_ContextTimeout(t *testing.T) {
	mock := &mockMigrate{
		upFunc: func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 10 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := migrator.Up(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestMigrator_Down_Success(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			assert.Equal(t, -1, n)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Down(ctx, 1)
	assert.NoError(t, err)
}

func TestMigrator_Down_MultipleSteps(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			assert.Equal(t, -3, n)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Down(ctx, 3)
	assert.NoError(t, err)
}

func TestMigrator_Down_NegativeSteps(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			assert.Equal(t, -1, n)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Down(ctx, -1)
	assert.NoError(t, err)
}

func TestMigrator_Down_ZeroSteps(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			assert.Equal(t, -1, n)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Down(ctx, 0)
	assert.NoError(t, err)
}

func TestMigrator_Down_Error(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			return errors.New("rollback failed")
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Down(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
}

func TestMigrator_Down_ContextTimeout(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			time.Sleep(200 * time.Millisecond)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 10 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := migrator.Down(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestMigrator_Steps_Forward(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			assert.Equal(t, 2, n)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Steps(ctx, 2)
	assert.NoError(t, err)
}

func TestMigrator_Steps_Backward(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			assert.Equal(t, -2, n)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Steps(ctx, -2)
	assert.NoError(t, err)
}

func TestMigrator_Steps_Error(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			return errors.New("steps failed")
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Steps(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "steps failed")
}

func TestMigrator_Steps_ContextTimeout(t *testing.T) {
	mock := &mockMigrate{
		stepsFunc: func(n int) error {
			time.Sleep(200 * time.Millisecond)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 10 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := migrator.Steps(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestMigrator_Goto_Success(t *testing.T) {
	mock := &mockMigrate{
		migrateFunc: func(version uint) error {
			assert.Equal(t, uint(5), version)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Goto(ctx, 5)
	assert.NoError(t, err)
}

func TestMigrator_Goto_NoChange(t *testing.T) {
	mock := &mockMigrate{
		migrateFunc: func(version uint) error {
			return migrate.ErrNoChange
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Goto(ctx, 5)
	assert.NoError(t, err)
}

func TestMigrator_Goto_Error(t *testing.T) {
	mock := &mockMigrate{
		migrateFunc: func(version uint) error {
			return errors.New("goto failed")
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := migrator.Goto(ctx, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goto failed")
	assert.Contains(t, err.Error(), "version 5")
}

func TestMigrator_Goto_ContextTimeout(t *testing.T) {
	mock := &mockMigrate{
		migrateFunc: func(version uint) error {
			time.Sleep(200 * time.Millisecond)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
		config: Config{
			Timeout: 10 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := migrator.Goto(ctx, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestMigrator_Version_Success(t *testing.T) {
	mock := &mockMigrate{
		versionFunc: func() (uint, bool, error) {
			return 3, false, nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	version, dirty, err := migrator.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint(3), version)
	assert.False(t, dirty)
}

func TestMigrator_Version_Dirty(t *testing.T) {
	mock := &mockMigrate{
		versionFunc: func() (uint, bool, error) {
			return 3, true, nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	version, dirty, err := migrator.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint(3), version)
	assert.True(t, dirty)
}

func TestMigrator_Version_NilVersion(t *testing.T) {
	mock := &mockMigrate{
		versionFunc: func() (uint, bool, error) {
			return 0, false, migrate.ErrNilVersion
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	version, dirty, err := migrator.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint(0), version)
	assert.False(t, dirty)
}

func TestMigrator_Version_Error(t *testing.T) {
	mock := &mockMigrate{
		versionFunc: func() (uint, bool, error) {
			return 0, false, errors.New("version error")
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	_, _, err := migrator.Version()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get migration version")
}

func TestMigrator_Force_Success(t *testing.T) {
	mock := &mockMigrate{
		forceFunc: func(version int) error {
			assert.Equal(t, 3, version)
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	err := migrator.Force(3)
	assert.NoError(t, err)
}

func TestMigrator_Force_Error(t *testing.T) {
	mock := &mockMigrate{
		forceFunc: func(version int) error {
			return errors.New("force failed")
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	err := migrator.Force(3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to force version")
}

func TestMigrator_Drop_Success(t *testing.T) {
	mock := &mockMigrate{
		dropFunc: func() error {
			return nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	err := migrator.Drop()
	assert.NoError(t, err)
}

func TestMigrator_Drop_Error(t *testing.T) {
	mock := &mockMigrate{
		dropFunc: func() error {
			return errors.New("drop failed")
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	err := migrator.Drop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to drop tables")
}

func TestMigrator_Close_Success(t *testing.T) {
	mock := &mockMigrate{
		closeFunc: func() (error, error) {
			return nil, nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	err := migrator.Close()
	assert.NoError(t, err)
}

func TestMigrator_Close_SourceError(t *testing.T) {
	mock := &mockMigrate{
		closeFunc: func() (error, error) {
			return errors.New("source close failed"), nil
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	err := migrator.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close source")
}

func TestMigrator_Close_DatabaseError(t *testing.T) {
	mock := &mockMigrate{
		closeFunc: func() (error, error) {
			return nil, errors.New("db close failed")
		},
	}

	migrator := &Migrator{
		migrate: mock,
	}

	err := migrator.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close database")
}
