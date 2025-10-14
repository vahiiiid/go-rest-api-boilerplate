package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App       AppConfig       `mapstructure:"app" yaml:"app"`
	Database  DatabaseConfig  `mapstructure:"database" yaml:"database"`
	JWT       JWTConfig       `mapstructure:"jwt" yaml:"jwt"`
	Server    ServerConfig    `mapstructure:"server" yaml:"server"`
	Logging   LoggingConfig   `mapstructure:"logging" yaml:"logging"`
	Ratelimit RateLimitConfig `mapstructure:"ratelimit" yaml:"ratelimit"`
}

// AppConfig holds application-related configuration.
type AppConfig struct {
	Name        string `mapstructure:"name" yaml:"name"`
	Environment string `mapstructure:"environment" yaml:"environment"`
	Debug       bool   `mapstructure:"debug" yaml:"debug"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	Name     string `mapstructure:"name" yaml:"name"`
	SSLMode  string `mapstructure:"sslmode" yaml:"sslmode"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret   string `mapstructure:"secret" yaml:"secret"`
	TTLHours int    `mapstructure:"ttlhours" yaml:"ttlhours"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string `mapstructure:"port" yaml:"port"`
	ReadTimeout  int    `mapstructure:"readtimeout" yaml:"readtimeout"`
	WriteTimeout int    `mapstructure:"writetimeout" yaml:"writetimeout"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level string `mapstructure:"level" yaml:"level"`
}

// RateLimitConfig holds rate-limit configuration
type RateLimitConfig struct {
	Enabled  bool          `mapstructure:"enabled" yaml:"enabled"`
	Requests int           `mapstructure:"requests" yaml:"requests"`
	Window   time.Duration `mapstructure:"window" yaml:"window"`
}

// LoadConfig loads configuration using Viper. If configPath is non-empty it
// will be used as the exact config file path, otherwise Viper searches common locations.
func LoadConfig(configPath string) (*Config, error) {
	// Create a new viper instance to avoid conflicts with global state
	v := viper.New()

	// Configure environment variable handling first
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// choose config.<env>.yaml if APP_ENVIRONMENT is set; otherwise use config.yaml
		env := v.GetString("APP_ENVIRONMENT")
		if env == "" {
			// fallback to "development" when APP_ENVIRONMENT is not set
			env = "development"
		}

		// Try env-specific file first
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		v.SetConfigType("yaml")
		v.AddConfigPath("configs")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	// Defaults
	setViperDefaults(v)

	// Read config file if present
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// no config file is ok; env vars and defaults will be used
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Backwards compatibility: if Server.Env not set, prefer APP_ENVIRONMENT/ENV
	if cfg.App.Environment == "" {
		// prefer viper (env vars are already bound via AutomaticEnv and key replacer)
		if e := v.GetString("app.environment"); e != "" {
			cfg.App.Environment = e
		} else if e := v.GetString("ENV"); e != "" {
			cfg.App.Environment = e
		}
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setViperDefaults(v *viper.Viper) {
	// App
	v.SetDefault("app.name", "GRAB API")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", false)

	// Database
	v.SetDefault("database.host", "db")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.name", "grab")
	// Do not set a default for password, so validation can catch it.

	// JWT
	v.SetDefault("jwt.ttlhours", 24)
	v.SetDefault("jwt.secret", "")

	// Server
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.readtimeout", 10)
	v.SetDefault("server.writetimeout", 10)

	// Logging
	v.SetDefault("logging.level", "info")

	// Ratelimit
	v.SetDefault("ratelimit.enabled", false)
	v.SetDefault("ratelimit.requests", 100)
	v.SetDefault("ratelimit.window", "1m")
}

// GetLogLevel converts string log level to slog.Level
func (l *LoggingConfig) GetLogLevel() slog.Level {
	switch strings.ToLower(l.Level) {
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

// GetConfigPath returns the default config path (kept for compatibility)
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

	return "configs/config.yaml"
}
