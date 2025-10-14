package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// createTempConfigFile creates a temporary YAML config file for testing.
func createTempConfigFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	return path
}

func TestLoadConfig_Comprehensive(t *testing.T) {
	// Reset viper before each test to ensure a clean state
	viper.Reset()

	t.Run("loads from default config file", func(t *testing.T) {
		viper.Reset()
		// Clear environment variables that might interfere
		t.Setenv("APP_NAME", "")
		t.Setenv("DATABASE_HOST", "")
		t.Setenv("JWT_SECRET", "")

		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
app:
  name: "Test API"
database:
  host: "testhost"
jwt:
  secret: "test-secret-for-validation"
`)
		cfg, err := LoadConfig(path) // Pass the explicit path
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "Test API", cfg.App.Name)
		assert.Equal(t, "testhost", cfg.Database.Host)
		assert.Equal(t, "test-secret-for-validation", cfg.JWT.Secret)
	})

	t.Run("environment variables override file values", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
database:
  host: "filehost"
  port: 5432
jwt:
  secret: "file-secret-for-validation"
`)
		// Set env vars that should override the file
		t.Setenv("DATABASE_HOST", "envhost")
		t.Setenv("JWT_SECRET", "env-secret-for-validation")

		cfg, err := LoadConfig(path) // Pass the explicit path
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "envhost", cfg.Database.Host)                // Assert override
		assert.Equal(t, 5432, cfg.Database.Port)                     // Assert value from file is still present
		assert.Equal(t, "env-secret-for-validation", cfg.JWT.Secret) // Assert override
	})

	t.Run("uses config file defaults when no env var is set", func(t *testing.T) {
		viper.Reset()
		// Clear environment variables that might interfere
		t.Setenv("JWT_SECRET", "")
		t.Setenv("APP_ENVIRONMENT", "")
		t.Setenv("DATABASE_HOST", "")
		t.Setenv("DATABASE_PASSWORD", "")

		// Create a complete config file with all required fields
		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
app:
  name: "GRAB API (development)"
  environment: "development"
  debug: true
database:
  host: "testhost"
  port: 5432
  user: "testuser"
  password: "testpass"
  name: "testdb"
  sslmode: "disable"
jwt:
  secret: "file-secret-for-validation"
  ttlhours: 24
server:
  port: "8080"
  readtimeout: 10
  writetimeout: 10
logging:
  level: "info"
ratelimit:
  enabled: false
  requests: 100
  window: "1m"
`)

		cfg, err := LoadConfig(path)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		// These values should come from config file defaults
		assert.Equal(t, 10, cfg.Server.ReadTimeout)
		assert.Equal(t, "development", cfg.App.Environment)
		assert.Equal(t, "GRAB API (development)", cfg.App.Name)
		assert.Equal(t, "file-secret-for-validation", cfg.JWT.Secret)
	})

	t.Run("fails validation if required JWT_SECRET is missing", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(t.TempDir()) // No config file

		// Ensure JWT_SECRET is not set in environment
		t.Setenv("JWT_SECRET", "")

		_, err := LoadConfig("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret is required")
	})

	t.Run("fails validation if DB_PASSWORD is missing in production", func(t *testing.T) {
		viper.Reset()
		// Create a minimal config with production environment but no database password
		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
app:
  environment: "production"
database:
  host: "testhost"
  port: 5432
  user: "testuser"
  password: ""  # Empty password should fail validation in production
  name: "testdb"
  sslmode: "require"
jwt:
  secret: "this-is-a-very-strong-production-secret-for-testing"
  ttlhours: 24
`)
		t.Setenv("APP_ENVIRONMENT", "production")
		t.Setenv("DATABASE_PASSWORD", "") // Explicitly empty

		_, err := LoadConfig(path)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database.password is required in production")
	})

	t.Run("fails validation for short JWT secret in production", func(t *testing.T) {
		viper.Reset()
		// Clear environment variables that might interfere
		t.Setenv("JWT_SECRET", "")
		t.Setenv("APP_ENVIRONMENT", "")
		t.Setenv("DATABASE_PASSWORD", "")

		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
app:
  environment: "production"
database:
  host: "testhost"
  port: 5432
  user: "testuser"
  password: "prod-password"
  name: "testdb"
  sslmode: "require"
jwt:
  secret: "short"
  ttlhours: 24
`)
		_, err := LoadConfig(path)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "JWT secret must be at least 32 characters long in production")
		}
	})

	t.Run("loads environment-specific config file when no path is given", func(t *testing.T) {
		viper.Reset()
		// Clear environment variables that might interfere
		t.Setenv("APP_NAME", "")
		t.Setenv("DATABASE_SSLMODE", "")
		t.Setenv("DATABASE_PASSWORD", "")
		t.Setenv("JWT_SECRET", "")

		tempDir := t.TempDir()
		configsDir := filepath.Join(tempDir, "configs")
		err := os.Mkdir(configsDir, 0755)
		assert.NoError(t, err)

		// Create a default and a production config file inside the temp configs dir
		createTempConfigFile(t, configsDir, "config.yaml", `
app:
  name: "Default API"
database:
  host: "testhost"
jwt:
  secret: "default-secret"
`)
		createTempConfigFile(t, configsDir, "config.production.yaml", `
app:
  name: "Production API"
  environment: "production"
database:
  host: "testhost"
  port: 5432
  user: "testuser"
  password: "prod-password"
  name: "testdb"
  sslmode: "require"
jwt:
  secret: "this-is-a-very-strong-production-secret-for-testing-purposes-only"
  ttlhours: 24
`)
		// Temporarily change working directory so LoadConfig can find the "configs" folder
		oldWd, err := os.Getwd()
		assert.NoError(t, err)
		err = os.Chdir(tempDir)
		assert.NoError(t, err)
		defer func() {
			err := os.Chdir(oldWd)
			if err != nil {
				t.Logf("Failed to restore working directory: %v", err)
			}
		}()

		t.Setenv("APP_ENVIRONMENT", "production")

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		if cfg != nil {
			// Assert it loaded the production file, not the default one
			assert.Equal(t, "Production API", cfg.App.Name)
		}
	})
}

func TestLoggingConfig_GetLogLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo}, // Should default to info
		{"", slog.LevelInfo},        // Should default to info
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			cfg := &LoggingConfig{Level: tt.level}
			result := cfg.GetLogLevel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSkipPaths(t *testing.T) {
	tests := []struct {
		env      string
		expected []string
	}{
		{"production", []string{"/health", "/metrics", "/debug", "/pprof"}},
		{"development", []string{"/health"}},
		{"test", []string{"/health"}},
		{"staging", []string{"/health"}}, // default case
		{"", []string{"/health"}},        // default case
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			result := GetSkipPaths(tt.env)
			assert.Equal(t, tt.expected, result)
		})
	}
}
