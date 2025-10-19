package migrate

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

var Models = []interface{}{
	&user.User{},
}

func RunMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(Models...); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✅ Migrations completed successfully")
}

func RollbackMigrations(db *gorm.DB) {
	log.Println("Rolling back database migrations...")
	for i := len(Models) - 1; i >= 0; i-- {
		if err := db.Migrator().DropTable(Models[i]); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
	}
	log.Println("✅ Migrations rolled back successfully")
}

func ShowMigrationStatus(db *gorm.DB) {
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
