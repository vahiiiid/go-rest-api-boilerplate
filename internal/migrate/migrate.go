package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	DatabaseURL   string
	MigrationsDir string
	Timeout       time.Duration
	LockTimeout   time.Duration
}

type migrateInterface interface {
	Up() error
	Steps(n int) error
	Migrate(version uint) error
	Version() (uint, bool, error)
	Force(version int) error
	Drop() error
	Close() (error, error)
}

type Migrator struct {
	migrate migrateInterface
	db      *sql.DB
	config  Config
}

func New(db *sql.DB, cfg Config) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable:       "schema_migrations",
		MultiStatementEnabled: true,
		StatementTimeout:      cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", cfg.MigrationsDir),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate: m,
		db:      db,
		config:  cfg,
	}, nil
}

func (m *Migrator) Up(ctx context.Context) error {
	slog.Info("Running migrations...")

	done := make(chan error, 1)
	go func() {
		done <- m.migrate.Up()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("migration timeout exceeded (%v). Use --timeout flag for longer migrations (e.g., --timeout=30m)", m.config.Timeout)
	case err := <-done:
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migration failed: %w", err)
		}
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No pending migrations")
			return nil
		}
		slog.Info("Migrations completed successfully", "status", "✅")
		return nil
	}
}

func (m *Migrator) Down(ctx context.Context, steps int) error {
	if steps < 1 {
		steps = 1
	}

	slog.Info("Rolling back migrations...", "steps", steps)

	done := make(chan error, 1)
	go func() {
		done <- m.migrate.Steps(-steps)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("migration timeout exceeded (%v). Use --timeout flag for longer migrations", m.config.Timeout)
	case err := <-done:
		if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}
		slog.Info("Rollback completed successfully", "status", "✅")
		return nil
	}
}

func (m *Migrator) Steps(ctx context.Context, n int) error {
	action := "forward"
	if n < 0 {
		action = "backward"
	}

	slog.Info("Executing migration steps...", "steps", n, "direction", action)

	done := make(chan error, 1)
	go func() {
		done <- m.migrate.Steps(n)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("migration timeout exceeded (%v). Use --timeout flag for longer migrations", m.config.Timeout)
	case err := <-done:
		if err != nil {
			return fmt.Errorf("migration steps failed: %w", err)
		}
		slog.Info("Migration steps completed successfully", "status", "✅")
		return nil
	}
}

func (m *Migrator) Goto(ctx context.Context, version uint) error {
	slog.Info("Migrating to version...", "version", version)

	done := make(chan error, 1)
	go func() {
		done <- m.migrate.Migrate(version)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("migration timeout exceeded (%v). Use --timeout flag for longer migrations", m.config.Timeout)
	case err := <-done:
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migration to version %d failed: %w", version, err)
		}
		slog.Info("Migration to version completed", "version", version, "status", "✅")
		return nil
	}
}

func (m *Migrator) Version() (uint, bool, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}

	return version, dirty, nil
}

func (m *Migrator) Force(version int) error {
	slog.Warn("Forcing migration version", "version", version)

	if err := m.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force version: %w", err)
	}

	slog.Info("Migration version forced", "version", version, "status", "✅")
	return nil
}

func (m *Migrator) Drop() error {
	slog.Warn("Dropping all tables...")

	if err := m.migrate.Drop(); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	slog.Info("All tables dropped", "status", "✅")
	return nil
}

func (m *Migrator) Close() error {
	srcErr, dbErr := m.migrate.Close()
	if srcErr != nil {
		return fmt.Errorf("failed to close source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}
