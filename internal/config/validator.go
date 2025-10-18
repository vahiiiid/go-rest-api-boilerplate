package config

import "fmt"

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

		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("JWT secret must be at least 32 characters long in production (current length: %d)", len(c.JWT.Secret))
		}

		if c.Database.SSLMode == "disable" {
			return fmt.Errorf("database SSL mode cannot be 'disable' in production")
		}
	}

	return nil
}
