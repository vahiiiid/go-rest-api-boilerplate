package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	action := flag.String("action", "up", "Migration action: up, down, status")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,
		cfg.Database.Name, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Execute migration action
	switch *action {
	case "up":
		runMigrations(db)
	case "down":
		rollbackMigrations(db)
	case "status":
		showMigrationStatus(db)
	default:
		log.Fatalf("Unknown action: %s. Use 'up', 'down', or 'status'", *action)
	}
}

func runMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")

	// AutoMigrate all models here
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("✅ Migrations completed successfully")
}

func rollbackMigrations(db *gorm.DB) {
	log.Println("Rolling back database migrations...")

	// Drop tables in reverse order (consider foreign key dependencies)
	if err := db.Migrator().DropTable(&user.User{}); err != nil {
		log.Fatalf("Failed to rollback migrations: %v", err)
	}

	log.Println("✅ Migrations rolled back successfully")
}

func showMigrationStatus(db *gorm.DB) {
	log.Println("Checking migration status...")

	// Check if tables exist
	hasUsersTable := db.Migrator().HasTable(&user.User{})

	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	fmt.Printf("Users table: %s\n", tableStatus(hasUsersTable))

	// You can add more detailed status checks here
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
