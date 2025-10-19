package main

import (
	"flag"
	"log"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/migrate"
)

func main() {
	action := flag.String("action", "up", "Migration action: up, down, status")
	flag.Parse()

	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	database, err := db.NewPostgresDBFromDatabaseConfig(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database %s@%s:%d: %v",
			cfg.Database.Name, cfg.Database.Host, cfg.Database.Port, err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Warning: failed to close database connection: %v", err)
		}
	}()

	switch *action {
	case "up":
		migrate.RunMigrations(database)
	case "down":
		migrate.RollbackMigrations(database)
	case "status":
		migrate.ShowMigrationStatus(database)
	default:
		log.Fatalf("Unknown action: %s. Use 'up', 'down', or 'status'", *action)
	}
}
