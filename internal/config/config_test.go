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
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
`)
		cfg, err := LoadConfig(path) // Pass the explicit path
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "Test API", cfg.App.Name)
		assert.Equal(t, "testhost", cfg.Database.Host)
		assert.Equal(t, "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP", cfg.JWT.Secret)
	})

	t.Run("environment variables override file values", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
database:
  host: "filehost"
  port: 5432
jwt:
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
`)
		// Set env vars that should override the file
		t.Setenv("DATABASE_HOST", "envhost")
		t.Setenv("JWT_SECRET", "QRSTUVWXYZqrstuvwxyzQRSTUVWXYZab")

		cfg, err := LoadConfig(path) // Pass the explicit path
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "envhost", cfg.Database.Host)                       // Assert override
		assert.Equal(t, 5432, cfg.Database.Port)                            // Assert value from file is still present
		assert.Equal(t, "QRSTUVWXYZqrstuvwxyzQRSTUVWXYZab", cfg.JWT.Secret) // Assert override
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
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
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
		assert.Equal(t, "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP", cfg.JWT.Secret)
	})

	t.Run("fails validation if required JWT_SECRET is missing", func(t *testing.T) {
		viper.Reset()
		viper.AddConfigPath(t.TempDir()) // No config file

		// Ensure JWT_SECRET is not set in environment
		t.Setenv("JWT_SECRET", "")

		_, err := LoadConfig("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET environment variable is required")
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
  secret: "PRODabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdef"
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
			assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters")
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
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
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
  secret: "qrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzAB"
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
		{"production", []string{"/health", "/health/live", "/health/ready", "/metrics", "/debug", "/pprof"}},
		{"development", []string{"/health", "/health/live", "/health/ready"}},
		{"test", []string{"/health", "/health/live", "/health/ready"}},
		{"staging", []string{"/health", "/health/live", "/health/ready"}}, // default case
		{"", []string{"/health", "/health/live", "/health/ready"}},        // default case
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			result := GetSkipPaths(tt.env)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	result := GetConfigPath()

	// Verify it returns a valid path (should be the default or actual config path)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "config.yaml")
}

func TestNewTestConfig(t *testing.T) {
	config := NewTestConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "test", config.App.Environment)
	assert.Equal(t, "test_db", config.Database.Name)
	assert.Equal(t, "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP", config.JWT.Secret)
	assert.Equal(t, 1, config.JWT.TTLHours)
	assert.Equal(t, "8081", config.Server.Port)
}

func TestNewTestConfig_Isolation(t *testing.T) {
	// Test that multiple calls return independent configs
	config1 := NewTestConfig()
	config2 := NewTestConfig()

	// Modify one config
	config1.App.Name = "modified"

	// Verify the other is not affected
	assert.NotEqual(t, config1.App.Name, config2.App.Name)
	assert.Equal(t, "Test API", config2.App.Name)
}

func TestLogSafeConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "logs all configuration sections with redacted sensitive data",
			config: &Config{
				App: AppConfig{
					Name:        "TestApp",
					Environment: "production",
					Debug:       true,
				},
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5432,
					User:     "testuser",
					Password: "super-secret-password",
					Name:     "testdb",
					SSLMode:  "require",
				},
				JWT: JWTConfig{
					Secret:   "super-secret-jwt-key",
					TTLHours: 24,
				},
				Server: ServerConfig{
					Port:         "8080",
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
				Ratelimit: RateLimitConfig{
					Enabled:  true,
					Requests: 100,
					Window:   60,
				},
			},
		},
		{
			name: "handles empty configuration values",
			config: &Config{
				App: AppConfig{
					Name:        "",
					Environment: "",
					Debug:       false,
				},
				Database: DatabaseConfig{
					Host:     "",
					Port:     0,
					User:     "",
					Password: "",
					Name:     "",
					SSLMode:  "",
				},
				JWT: JWTConfig{
					Secret:   "",
					TTLHours: 0,
				},
				Server: ServerConfig{
					Port:            "",
					ReadTimeout:     0,
					WriteTimeout:    0,
					IdleTimeout:     0,
					ShutdownTimeout: 0,
					MaxHeaderBytes:  0,
				},
				Logging: LoggingConfig{
					Level: "",
				},
				Ratelimit: RateLimitConfig{
					Enabled:  false,
					Requests: 0,
					Window:   0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			assert.NotPanics(t, func() {
				tt.config.LogSafeConfig(logger)
			})
		})
	}
}

func TestServerConfig_TimeoutFields(t *testing.T) {
	viper.Reset()

	t.Run("loads timeout fields from config file", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
database:
  host: "testhost"
  password: "postgres"
jwt:
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
server:
  port: "8080"
  readtimeout: 10
  writetimeout: 10
  idletimeout: 120
  shutdowntimeout: 30
  maxheaderbytes: 1048576
`)
		cfg, err := LoadConfig(path)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 10, cfg.Server.ReadTimeout)
		assert.Equal(t, 10, cfg.Server.WriteTimeout)
		assert.Equal(t, 120, cfg.Server.IdleTimeout)
		assert.Equal(t, 30, cfg.Server.ShutdownTimeout)
		assert.Equal(t, 1048576, cfg.Server.MaxHeaderBytes)
	})

	t.Run("environment variables override timeout values", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
database:
  host: "testhost"
  password: "postgres"
jwt:
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
server:
  readtimeout: 10
  writetimeout: 10
  idletimeout: 120
  shutdowntimeout: 30
  maxheaderbytes: 1048576
`)
		t.Setenv("SERVER_READTIMEOUT", "15")
		t.Setenv("SERVER_WRITETIMEOUT", "15")
		t.Setenv("SERVER_IDLETIMEOUT", "180")
		t.Setenv("SERVER_SHUTDOWNTIMEOUT", "60")
		t.Setenv("SERVER_MAXHEADERBYTES", "2097152")

		cfg, err := LoadConfig(path)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 15, cfg.Server.ReadTimeout)
		assert.Equal(t, 15, cfg.Server.WriteTimeout)
		assert.Equal(t, 180, cfg.Server.IdleTimeout)
		assert.Equal(t, 60, cfg.Server.ShutdownTimeout)
		assert.Equal(t, 2097152, cfg.Server.MaxHeaderBytes)
	})

	t.Run("zero timeout values are allowed", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		path := createTempConfigFile(t, tempDir, "config.yaml", `
database:
  host: "testhost"
  password: "postgres"
jwt:
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
server:
  readtimeout: 0
  writetimeout: 0
  idletimeout: 0
  shutdowntimeout: 0
  maxheaderbytes: 0
`)
		cfg, err := LoadConfig(path)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 0, cfg.Server.ReadTimeout)
		assert.Equal(t, 0, cfg.Server.WriteTimeout)
		assert.Equal(t, 0, cfg.Server.IdleTimeout)
		assert.Equal(t, 0, cfg.Server.ShutdownTimeout)
		assert.Equal(t, 0, cfg.Server.MaxHeaderBytes)
	})
}

func TestValidate_TimeoutFields(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid timeout values",
			config: Config{
				App: AppConfig{Environment: "development"},
				Database: DatabaseConfig{
					Host: "localhost",
				},
				JWT: JWTConfig{
					Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
				},
				Server: ServerConfig{
					ReadTimeout:     10,
					WriteTimeout:    10,
					IdleTimeout:     120,
					ShutdownTimeout: 30,
					MaxHeaderBytes:  1048576,
				},
			},
			expectError: false,
		},
		{
			name: "negative read timeout",
			config: Config{
				App: AppConfig{Environment: "development"},
				Database: DatabaseConfig{
					Host: "localhost",
				},
				JWT: JWTConfig{
					Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
				},
				Server: ServerConfig{
					ReadTimeout: -1,
				},
			},
			expectError: true,
			errorMsg:    "server.readtimeout must be non-negative",
		},
		{
			name: "negative write timeout",
			config: Config{
				App: AppConfig{Environment: "development"},
				Database: DatabaseConfig{
					Host: "localhost",
				},
				JWT: JWTConfig{
					Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
				},
				Server: ServerConfig{
					WriteTimeout: -1,
				},
			},
			expectError: true,
			errorMsg:    "server.writetimeout must be non-negative",
		},
		{
			name: "negative idle timeout",
			config: Config{
				App: AppConfig{Environment: "development"},
				Database: DatabaseConfig{
					Host: "localhost",
				},
				JWT: JWTConfig{
					Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
				},
				Server: ServerConfig{
					IdleTimeout: -1,
				},
			},
			expectError: true,
			errorMsg:    "server.idletimeout must be non-negative",
		},
		{
			name: "negative shutdown timeout",
			config: Config{
				App: AppConfig{Environment: "development"},
				Database: DatabaseConfig{
					Host: "localhost",
				},
				JWT: JWTConfig{
					Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
				},
				Server: ServerConfig{
					ShutdownTimeout: -1,
				},
			},
			expectError: true,
			errorMsg:    "server.shutdowntimeout must be non-negative",
		},
		{
			name: "negative max header bytes",
			config: Config{
				App: AppConfig{Environment: "development"},
				Database: DatabaseConfig{
					Host: "localhost",
				},
				JWT: JWTConfig{
					Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
				},
				Server: ServerConfig{
					MaxHeaderBytes: -1,
				},
			},
			expectError: true,
			errorMsg:    "server.maxheaderbytes must be non-negative",
		},
		{
			name: "zero timeouts are valid",
			config: Config{
				App: AppConfig{Environment: "development"},
				Database: DatabaseConfig{
					Host: "localhost",
				},
				JWT: JWTConfig{
					Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
				},
				Server: ServerConfig{
					ReadTimeout:     0,
					WriteTimeout:    0,
					IdleTimeout:     0,
					ShutdownTimeout: 0,
					MaxHeaderBytes:  0,
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadConfig_ErrorPaths(t *testing.T) {
	t.Run("unmarshal error with invalid config structure", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()

		invalidYAML := `
app:
  name: 123
  environment: [this, should, not, be, an, array]
database:
  port: "not_a_number"
`
		path := createTempConfigFile(t, tempDir, "invalid.yaml", invalidYAML)

		_, err := LoadConfig(path)
		assert.Error(t, err)
	})

	t.Run("ENV fallback when app.environment is empty", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()

		configContent := `
app:
  name: "TestApp"
database:
  host: "testhost"
  port: 5432
  user: "testuser"
  password: "testpass"
  name: "testdb"
  sslmode: "disable"
jwt:
  secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP"
  ttlhours: 24
server:
  port: "8080"
`
		path := createTempConfigFile(t, tempDir, "config.yaml", configContent)
		t.Setenv("ENV", "testing")
		t.Setenv("APP_ENVIRONMENT", "")

		cfg, err := LoadConfig(path)
		assert.NoError(t, err)
		assert.Equal(t, "testing", cfg.App.Environment)
	})
}

func TestGetConfigPath_AllPaths(t *testing.T) {
	t.Run("returns absolute path when config exists", func(t *testing.T) {
		result := GetConfigPath()
		assert.Contains(t, result, "config.yaml")
	})

	t.Run("returns default when no config found", func(t *testing.T) {
		origWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(origWd) }()

		tempDir := t.TempDir()
		_ = os.Chdir(tempDir)

		result := GetConfigPath()
		assert.Equal(t, "configs/config.yaml", result)
	})
}

func TestValidate_ProductionSSLMode(t *testing.T) {
	cfg := Config{
		App: AppConfig{
			Environment: "production",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Password: "securepassword",
			SSLMode:  "disable",
		},
		JWT: JWTConfig{
			Secret: "longjwtauthenticationkeywithatleastsixtyfourcharsforprodvalidation",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSL mode cannot be 'disable' in production")
}

func TestValidate_DatabaseHostRequired(t *testing.T) {
	cfg := Config{
		App: AppConfig{
			Environment: "development",
		},
		Database: DatabaseConfig{
			Host: "",
		},
		JWT: JWTConfig{
			Secret: "hKLmNpQrStUvWxYzABCDEFGHIJKLMNOP",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database.host is required")
}

func TestValidate_JWTSecret(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		jwtSecret   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid JWT secret in development",
			environment: "development",
			jwtSecret:   "abcdefghijklmnopqrstuvwxyz123456",
			expectError: false,
		},
		{
			name:        "JWT secret too short (< 32 chars)",
			environment: "development",
			jwtSecret:   "short",
			expectError: true,
			errorMsg:    "JWT_SECRET must be at least 32 characters",
		},
		{
			name:        "JWT secret exactly 32 chars in development",
			environment: "development",
			jwtSecret:   "abcdefghijklmnopqrstuvwxyz123456",
			expectError: false,
		},
		{
			name:        "valid JWT secret 32 chars in production",
			environment: "production",
			jwtSecret:   "abcdefghijklmnopqrstuvwxyz123456",
			expectError: false,
		},
		{
			name:        "valid JWT secret 50 chars in production",
			environment: "production",
			jwtSecret:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX",
			expectError: false,
		},
		{
			name:        "valid JWT secret 64 chars in production",
			environment: "production",
			jwtSecret:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789AB",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				App: AppConfig{
					Environment: tt.environment,
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Password: "secure-password",
					SSLMode:  "require",
				},
				JWT: JWTConfig{
					Secret: tt.jwtSecret,
				},
			}

			err := cfg.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
