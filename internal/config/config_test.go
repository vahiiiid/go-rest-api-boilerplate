package config

import (
	"log/slog"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config from default path
	cfg, err := LoadConfig("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify basic structure
	if cfg.Server.Port == "" {
		t.Error("Server port should not be empty")
	}

	if cfg.Database.Host == "" {
		t.Error("Database host should not be empty")
	}

	if cfg.JWT.Secret == "" {
		t.Error("JWT secret should not be empty")
	}

	if cfg.Logging.Level == "" {
		t.Error("Logging level should not be empty")
	}
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
			if result != tt.expected {
				t.Errorf("GetLogLevel() = %v, want %v", result, tt.expected)
			}
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
			if len(result) != len(tt.expected) {
				t.Errorf("GetSkipPaths() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, path := range result {
				if path != tt.expected[i] {
					t.Errorf("GetSkipPaths()[%d] = %v, want %v", i, path, tt.expected[i])
				}
			}
		})
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("PORT", "9090"); err != nil {
		t.Fatalf("Failed to set PORT env var: %v", err)
	}
	if err := os.Setenv("DB_HOST", "test-host"); err != nil {
		t.Fatalf("Failed to set DB_HOST env var: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("PORT"); err != nil {
			t.Errorf("Failed to unset PORT env var: %v", err)
		}
		if err := os.Unsetenv("DB_HOST"); err != nil {
			t.Errorf("Failed to unset DB_HOST env var: %v", err)
		}
		if err := os.Unsetenv("LOG_LEVEL"); err != nil {
			t.Errorf("Failed to unset LOG_LEVEL env var: %v", err)
		}
	}()

	cfg := &Config{
		Server:   ServerConfig{Port: "8080"},
		Database: DatabaseConfig{Host: "localhost"},
		Logging:  LoggingConfig{Level: "info"},
	}

	cfg.loadFromEnv()

	if cfg.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.Server.Port)
	}

	if cfg.Database.Host != "test-host" {
		t.Errorf("Expected host test-host, got %s", cfg.Database.Host)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.Logging.Level)
	}
}
