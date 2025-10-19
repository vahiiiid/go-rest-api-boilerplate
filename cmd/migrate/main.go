package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"

	"gorm.io/gorm"
)

var models = []interface{}{
	&user.User{},
}

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
		runMigrations(database)
	case "down":
		rollbackMigrations(database)
	case "status":
		showMigrationStatus(database)
	default:
		log.Fatalf("Unknown action: %s. Use 'up', 'down', or 'status'", *action)
	}
}

func runMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("✅ Migrations completed successfully")
}

func rollbackMigrations(db *gorm.DB) {
	log.Println("Rolling back database migrations...")

	for i := len(models) - 1; i >= 0; i-- {
		if err := db.Migrator().DropTable(models[i]); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
	}

	log.Println("✅ Migrations rolled back successfully")
}

func showMigrationStatus(db *gorm.DB) {
	log.Println("Checking migration status...")

	hasUsersTable := db.Migrator().HasTable(&user.User{})

	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	fmt.Printf("Users table: %s\n", tableStatus(hasUsersTable))

	if hasUsersTable {
		var count int64
		db.Model(&user.User{}).Count(&count)
		fmt.Printf("Users count: %d\n", count)
	}

	log.Println("\n✅ Status check completed")
}

func tableStatus(exists bool) string {
	if exists {
		return "✓ EXISTS"
	}
	return "✗ NOT FOUND"
}
