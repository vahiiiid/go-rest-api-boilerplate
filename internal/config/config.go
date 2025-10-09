package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App      AppConfig      `mapstructure:"app" yaml:"app"`
	Database DatabaseConfig `mapstructure:"database" yaml:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" yaml:"jwt"`
	Server   ServerConfig   `mapstructure:"server" yaml:"server"`
	Logging  LoggingConfig  `mapstructure:"logging" yaml:"logging"`
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

// LoadConfig loads configuration using Viper. If configPath is non-empty it
// will be used as the exact config file path, otherwise Viper searches common locations.
func LoadConfig(configPath string) (*Config, error) {
	// If a specific file path is passed, use it directly
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Search default locations
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("configs")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
	}

	// Environment variable mapping
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Defaults
	setDefaults()

	// Read config file if present
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// no config file is ok; env vars and defaults will be used
	}

	// Unmarshal into struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Backwards compatibility: if Server.Env not set, prefer APP_ENVIRONMENT/ENV
	if cfg.App.Environment == "" {
		if e := os.Getenv("APP_ENVIRONMENT"); e != "" {
			cfg.App.Environment = e
		} else if e := os.Getenv("ENV"); e != "" {
			cfg.App.Environment = e
		}
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	// App
	viper.SetDefault("app.name", "GRAB API")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)

	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "grab")

	// JWT
	viper.SetDefault("jwt.ttlhours", 24)
	viper.SetDefault("jwt.secret", "")

	// Server
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.readtimeout", 10)
	viper.SetDefault("server.writetimeout", 10)

	// Logging
	viper.SetDefault("logging.level", "info")
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
