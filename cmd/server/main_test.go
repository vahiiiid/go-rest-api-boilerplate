package main

import (
	"os"
	"testing"
)

func TestRun_ConfigLoadError(t *testing.T) {
	if err := os.Setenv("APP_ENVIRONMENT", "nonexistent"); err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("APP_ENVIRONMENT"); err != nil {
			t.Errorf("failed to unset environment variable: %v", err)
		}
	}()

	err := run()
	if err == nil {
		t.Error("expected error when config validation fails, got nil")
	}
}

func TestRun_WithTestConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Setenv("JWT_SECRET", "test-secret-key-for-testing-minimum-32-chars")
	t.Setenv("DATABASE_HOST", "invalid-host-to-trigger-error")
	t.Setenv("DATABASE_PORT", "5432")

	err := run()
	if err == nil {
		t.Error("expected error when database connection fails, got nil")
	}
}
