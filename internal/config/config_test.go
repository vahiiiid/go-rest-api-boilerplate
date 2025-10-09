package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/com/spf13/viper"
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
		tempDir := t.TempDir()
		createTempConfigFile(t, tempDir, "config.yaml", `
app:
  name: "File API"
database:
  host: "filehost"
jwt:
  secret: "file-secret"
`)
		// Point viper to our temp directory
		viper.AddConfigPath(tempDir)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "File API", cfg.App.Name)
		assert.Equal(t, "filehost", cfg.Database.Host)
		assert.Equal(t, "file-secret", cfg.JWT.Secret)
	})

	t.Run("environment variables override file values", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		createTempConfigFile(t, tempDir, "config.yaml", `
database:
  host: "filehost"
  port: 5432
jwt:
  secret: "file-secret"
`)
		viper.AddConfigPath(tempDir)

		// Set env vars that should override the file
		t.Setenv("DATABASE_HOST", "envhost")
		t.Setenv("JWT_SECRET", "env-secret")

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "envhost", cfg.Database.Host) // Assert override
		assert.Equal(t, 5432, cfg.Database.Port)      // Assert value from file is still present
		assert.Equal(t, "env-secret", cfg.JWT.Secret) // Assert override
	})

	t.Run("uses default values when no file or env var is set", func(t *testing.T) {
		viper.Reset()
		// Ensure no config file is found
		viper.AddConfigPath(t.TempDir())
		// Ensure a required value is set to pass validation
		t.Setenv("JWT_SECRET", "some-secret")

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		// This value is not in any file or env var, so it should be the default
		assert.Equal(t, 10, cfg.Server.ReadTimeout)
		assert.Equal(t, "development", cfg.App.Environment)
	})

	t.Run("fails validation if required JWT_SECRET is missing", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(t.TempDir()) // No config file

		_, err := LoadConfig("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret is required")
	})

	t.Run("fails validation if DB_PASSWORD is missing in production", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(t.TempDir()) // No config file
		t.Setenv("APP_ENVIRONMENT", "production")
		t.Setenv("JWT_SECRET", "prod-secret") // Satisfy JWT validation

		_, err := LoadConfig("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database.password is required in production")
	})

	t.Run("loads environment-specific config file", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		// Create a default and a production config file
		createTempConfigFile(t, tempDir, "config.yaml", `app: {name: "Default API"}`)
		createTempConfigFile(t, tempDir, "config.production.yaml", `
app:
  name: "Production API"
jwt:
  secret: "prod-secret"
database:
  password: "prod-password"
`)
		viper.AddConfigPath(tempDir)
		t.Setenv("APP_ENVIRONMENT", "production")

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		// Assert it loaded the production file, not the default one
		assert.Equal(t, "Production API", cfg.App.Name)
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
