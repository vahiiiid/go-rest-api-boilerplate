package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/migrate"
)

func main() {
	timeoutFlag := flag.String("timeout", "", "Migration timeout (e.g., 5m, 30s, 1h)")
	lockTimeoutFlag := flag.String("lock-timeout", "", "Lock acquisition timeout (e.g., 30s, 1m)")
	forceFlag := flag.Bool("force", false, "Skip confirmations for destructive operations")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	cfg, err := config.LoadConfig("")
	if err != nil {
		slog.Error("Failed to load configuration", "err", err)
		os.Exit(1)
	}

	timeout := time.Duration(cfg.Migrations.Timeout) * time.Second
	lockTimeout := time.Duration(cfg.Migrations.LockTimeout) * time.Second

	if *timeoutFlag != "" {
		if parsedTimeout, err := time.ParseDuration(*timeoutFlag); err == nil {
			timeout = parsedTimeout
		} else {
			slog.Error("Invalid timeout duration", "timeout", *timeoutFlag, "err", err)
			os.Exit(1)
		}
	}

	if *lockTimeoutFlag != "" {
		if parsedLockTimeout, err := time.ParseDuration(*lockTimeoutFlag); err == nil {
			lockTimeout = parsedLockTimeout
		} else {
			slog.Error("Invalid lock-timeout duration", "lock-timeout", *lockTimeoutFlag, "err", err)
			os.Exit(1)
		}
	}

	database, err := db.NewPostgresDBFromDatabaseConfig(cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}

	sqlDB, err := database.DB()
	if err != nil {
		slog.Error("Failed to get database instance", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Warn("Failed to close database connection", "err", err)
		}
	}()

	migrator, err := migrate.New(sqlDB, migrate.Config{
		MigrationsDir: cfg.Migrations.Directory,
		Timeout:       timeout,
		LockTimeout:   lockTimeout,
	})
	if err != nil {
		slog.Error("Failed to create migrator", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			slog.Error("Failed to close migrator", "err", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch command {
	case "up":
		handleUp(ctx, migrator, args)
	case "down":
		handleDown(ctx, migrator, args)
	case "goto":
		handleGoto(ctx, migrator, args)
	case "version":
		handleVersion(migrator)
	case "force":
		handleForce(migrator, args)
	case "drop":
		handleDrop(migrator, *forceFlag)
	case "create":
		handleCreate(cfg.Migrations.Directory, args)
	default:
		slog.Error("Unknown command", "command", command)
		printUsage()
		os.Exit(1)
	}
}

func handleUp(ctx context.Context, migrator *migrate.Migrator, args []string) {
	if len(args) > 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 {
			slog.Error("Invalid number of steps", "value", args[1])
			os.Exit(1)
		}
		if err := migrator.Steps(ctx, n); err != nil {
			slog.Error("Migration error", "err", err)
			os.Exit(1)
		}
	} else {
		if err := migrator.Up(ctx); err != nil {
			slog.Error("Migration error", "err", err)
			os.Exit(1)
		}
	}
}

func handleDown(ctx context.Context, migrator *migrate.Migrator, args []string) {
	steps := 1
	if len(args) > 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 {
			slog.Error("Invalid number of steps", "value", args[1])
			os.Exit(1)
		}
		steps = n
	}

	fmt.Printf("⚠️  WARNING: This will rollback %d migration(s)\n", steps)
	fmt.Print("Type 'yes' to confirm: ")
	var confirmation string
	if _, err := fmt.Scanln(&confirmation); err != nil {
		slog.Error("Failed to read confirmation", "err", err)
		os.Exit(1)
	}
	if confirmation != "yes" {
		slog.Info("Operation cancelled")
		return
	}

	if err := migrator.Down(ctx, steps); err != nil {
		slog.Error("Migration error", "err", err)
		os.Exit(1)
	}
}

func handleGoto(ctx context.Context, migrator *migrate.Migrator, args []string) {
	if len(args) < 2 {
		slog.Error("Version number required")
		fmt.Println("Usage: migrate goto VERSION")
		os.Exit(1)
	}

	version, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		slog.Error("Invalid version number", "value", args[1])
		os.Exit(1)
	}

	if err := migrator.Goto(ctx, uint(version)); err != nil {
		slog.Error("Migration error", "err", err)
		os.Exit(1)
	}
}

func handleVersion(migrator *migrate.Migrator) {
	version, dirty, err := migrator.Version()
	if err != nil {
		slog.Error("Failed to get version", "err", err)
		os.Exit(1)
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("=================")
	fmt.Printf("Current version: %d\n", version)
	if dirty {
		fmt.Println("Status: ⚠️  DIRTY (migration failed or interrupted)")
		fmt.Println("\nTo recover, use: migrate force VERSION")
	} else {
		fmt.Println("Status: ✅ Clean")
	}
}

func handleForce(migrator *migrate.Migrator, args []string) {
	if len(args) < 2 {
		slog.Error("Version number required")
		fmt.Println("Usage: migrate force VERSION")
		os.Exit(1)
	}

	version, err := strconv.Atoi(args[1])
	if err != nil {
		slog.Error("Invalid version number", "value", args[1])
		os.Exit(1)
	}

	fmt.Printf("⚠️  DANGER: This will force the migration version to %d\n", version)
	fmt.Println("This should only be used to recover from failed migrations")
	fmt.Print("Type 'yes' to confirm: ")
	var confirmation string
	if _, err := fmt.Scanln(&confirmation); err != nil {
		slog.Error("Failed to read confirmation", "err", err)
		os.Exit(1)
	}
	if confirmation != "yes" {
		slog.Info("Operation cancelled")
		return
	}

	if err := migrator.Force(version); err != nil {
		slog.Error("Migration error", "err", err)
		os.Exit(1)
	}
}

func handleDrop(migrator *migrate.Migrator, force bool) {
	if !force {
		fmt.Println("🚨 DANGER: This will drop all tables!")
		fmt.Print("Type 'YES' to confirm: ")
		var confirmation string
		if _, err := fmt.Scanln(&confirmation); err != nil {
			slog.Error("Failed to read confirmation", "err", err)
			os.Exit(1)
		}
		if confirmation != "YES" {
			slog.Info("Operation cancelled")
			return
		}
	}

	if err := migrator.Drop(); err != nil {
		slog.Error("Migration error", "err", err)
		os.Exit(1)
	}
}

func handleCreate(migrationsDir string, args []string) {
	if len(args) < 2 {
		slog.Error("Migration name required")
		fmt.Println("Usage: migrate create NAME")
		os.Exit(1)
	}

	name := args[1]
	timestamp := time.Now().Format("20060102150405")

	upFile := fmt.Sprintf("%s/%s_%s.up.sql", migrationsDir, timestamp, name)
	downFile := fmt.Sprintf("%s/%s_%s.down.sql", migrationsDir, timestamp, name)

	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: Add description here

BEGIN;

-- Add your migration SQL here

COMMIT;
`, name, time.Now().Format(time.RFC3339))

	downContent := fmt.Sprintf(`-- Migration: %s (rollback)
-- Created: %s

BEGIN;

-- Add your rollback SQL here

COMMIT;
`, name, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		slog.Error("Failed to create up migration", "err", err)
		os.Exit(1)
	}

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		slog.Error("Failed to create down migration", "err", err)
		os.Exit(1)
	}

	slog.Info("Migration files created", "up", upFile, "down", downFile)
}

func printUsage() {
	fmt.Println("Usage: migrate COMMAND [args] [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up [N]           Apply all pending migrations (or N migrations)")
	fmt.Println("  down [N]         Rollback last migration (or N migrations)")
	fmt.Println("  goto VERSION     Migrate to specific version")
	fmt.Println("  version          Show current migration version")
	fmt.Println("  force VERSION    Force set migration version (recovery)")
	fmt.Println("  drop             Drop all tables (requires confirmation)")
	fmt.Println("  create NAME      Create new migration files")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --timeout DURATION        Override migration timeout (e.g., 5m, 30s, 1h)")
	fmt.Println("  --lock-timeout DURATION   Override lock timeout (e.g., 30s, 1m)")
	fmt.Println("  --force                   Skip confirmations (for drop command)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  migrate up")
	fmt.Println("  migrate up 2")
	fmt.Println("  migrate down")
	fmt.Println("  migrate goto 5")
	fmt.Println("  migrate version")
	fmt.Println("  migrate create add_user_avatar")
	fmt.Println("  migrate up --timeout=30m --lock-timeout=1m")
}
