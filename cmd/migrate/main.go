package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/migrate"
)

func main() {
	action := flag.String("action", "up", "Migration action: up, down, status")
	flag.Parse()

	cfg, err := config.LoadConfig("")
	if err != nil {
		slog.Error("Failed to load configuration", "err", err)
		os.Exit(1)
	}

	database, err := db.NewPostgresDBFromDatabaseConfig(cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database",
			"name", cfg.Database.Name,
			"host", cfg.Database.Host,
			"port", cfg.Database.Port,
			"err", err)
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

	switch *action {
	case "up":
		if err := migrate.RunMigrations(database); err != nil {
			slog.Error("Migration error", "err", err)
			os.Exit(1)
		}
	case "down":
		if err := migrate.RollbackMigrations(database); err != nil {
			slog.Error("Migration error", "err", err)
			os.Exit(1)
		}
	case "status":
		migrate.ShowMigrationStatus(database)
	default:
		slog.Error("Unknown action", "action", *action)
		os.Exit(1)
	}
}
