package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	Logging   LoggingConfig   `yaml:"logging"`
	Ratelimit RateLimitConfig `yaml:"ratelimit"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string `yaml:"port"`
	Env  string `yaml:"env"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret   string `yaml:"secret"`
	TTLHours int    `yaml:"ttl_hours"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// RateLimitConfig holds rate-limit configuration
type RateLimitConfig struct {
	Enabled  bool          `yaml:"rate_limit_enabled"`
	Requests int           `yaml:"rate_limit_requests"`
	Window   time.Duration `yaml:"rate_limit_window"`
}

// LoadConfig loads configuration from YAML file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Default config path if not provided
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// Read YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if present
	config.loadFromEnv()

	return &config, nil
}

// loadFromEnv overrides configuration values with environment variables
func (c *Config) loadFromEnv() {
	// Server config
	if port := os.Getenv("PORT"); port != "" {
		c.Server.Port = port
	}
	if env := os.Getenv("ENV"); env != "" {
		c.Server.Env = env
	}

	// Database config
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		c.Database.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		c.Database.Name = name
	}

	// JWT config
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.JWT.Secret = secret
	}
	if ttlHours := os.Getenv("JWT_TTL_HOURS"); ttlHours != "" {
		// Parse TTL hours from environment variable
		// Note: This would need proper parsing in a real implementation
		// For now, we'll keep the YAML value if env var is not set
		_ = ttlHours // Acknowledge the variable to avoid unused variable warning
	}

	// Logging config
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}

	// Ratelimit config
	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			c.Ratelimit.Enabled = enabledBool
		}
	}
	if requests := os.Getenv("RATE_LIMIT_REQUESTS"); requests != "" {
		if requestsNum, err := strconv.Atoi(requests); err == nil && requestsNum > 0 {
			c.Ratelimit.Requests = requestsNum
		}
	}
	if window := os.Getenv("RATE_LIMIT_WINDOW"); window != "" {
		if windowDur, err := time.ParseDuration(window); err == nil && windowDur > 0 {
			c.Ratelimit.Window = windowDur
		}
	}
}

// GetLogLevel converts string log level to slog.Level
func (l *LoggingConfig) GetLogLevel() slog.Level {
	switch l.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to info level
	}
}

// GetSkipPaths returns the appropriate skip paths based on environment
func GetSkipPaths(env string) []string {
	switch env {
	case "production":
		return []string{"/health", "/metrics", "/debug", "/pprof"}
	case "development":
		return []string{"/health"}
	case "test":
		return []string{"/health"}
	default:
		return []string{"/health"}
	}
}

// GetConfigPath returns the default config path
func GetConfigPath() string {
	// Try to find config.yaml in common locations
	paths := []string{
		"configs/config.yaml",
		"./configs/config.yaml",
		"../configs/config.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Return default path if none found
	return "configs/config.yaml"
}
