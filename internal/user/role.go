package user

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Role represents a user role in the system
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Role model
func (Role) TableName() string {
	return "roles"
}
