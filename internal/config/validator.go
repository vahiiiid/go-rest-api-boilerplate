package config

import "fmt"

// Config represents the application configuration
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Environment string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Password string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// Validate checks required configuration values
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required (set JWT_SECRET or jwt.secret in config)")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.App.Environment == "production" {
		if c.Database.Password == "" {
			return fmt.Errorf("database.password is required in production")
		}
		if c.JWT.Secret == "" {
			return fmt.Errorf("jwt.secret is required in production")
		}
	}
	return nil
}