package migrate

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

var Models = []interface{}{
	&user.User{},
}

func RunMigrations(db *gorm.DB) error {
	slog.Info("Running database migrations...")
	if err := db.AutoMigrate(Models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("Migrations completed successfully", "status", "✅")
	return nil
}

func RollbackMigrations(db *gorm.DB) error {
	slog.Info("Rolling back database migrations...")
	for i := len(Models) - 1; i >= 0; i-- {
		if err := db.Migrator().DropTable(Models[i]); err != nil {
			return fmt.Errorf("failed to rollback migrations: %w", err)
		}
	}
	slog.Info("Migrations rolled back successfully", "status", "✅")
	return nil
}

func ShowMigrationStatus(db *gorm.DB) {
	slog.Info("Checking migration status...")
	hasUsersTable := db.Migrator().HasTable(&user.User{})
	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	fmt.Printf("Users table: %s\n", tableStatus(hasUsersTable))
	if hasUsersTable {
		var count int64
		result := db.Model(&user.User{}).Count(&count)
		if result.Error != nil {
			slog.Error("Error counting users", "err", result.Error)
			fmt.Println("Users count: ERROR")
		} else {
			fmt.Printf("Users count: %d\n", count)
		}
	}
	slog.Info("Status check completed", "status", "✅")
}

func tableStatus(exists bool) string {
	if exists {
		return "✓ EXISTS"
	}
	return "✗ NOT FOUND"
}
