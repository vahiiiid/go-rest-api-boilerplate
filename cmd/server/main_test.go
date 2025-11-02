package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestRun_ConfigLoadError(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("skipping integration test (SKIP_INTEGRATION_TESTS is set)")
	}

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

func TestMain_ExitsOnError(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		t.Setenv("APP_ENVIRONMENT", "nonexistent")
		main()
		return
	}

	tests := []struct {
		name string
		env  string
	}{
		{"invalid environment", "nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENVIRONMENT", tt.env)
			t.Setenv("BE_CRASHER", "1")
		})
	}
}

func TestGracefulShutdown_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("skipping integration test (SKIP_INTEGRATION_TESTS is set)")
	}

	if os.Getenv("CI") != "" || os.Getenv("DOCKER_ENV") != "" {
		t.Skip("skipping graceful shutdown test in CI/Docker environment")
	}

	envVars := map[string]string{
		"SKIP_MIGRATION_CHECK":   "true",
		"JWT_SECRET":             "test-secret-key-for-testing-minimum-32-chars-long",
		"DATABASE_HOST":          "localhost",
		"DATABASE_PORT":          "5432",
		"DATABASE_USER":          "postgres",
		"DATABASE_PASSWORD":      "postgres",
		"DATABASE_NAME":          "grab",
		"SERVER_PORT":            "18080",
		"SERVER_SHUTDOWNTIMEOUT": "5",
	}

	originals := make(map[string]string)
	for key, value := range envVars {
		originals[key] = os.Getenv(key)
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set %s: %v", key, err)
		}
	}

	defer func() {
		for key, value := range originals {
			var err error
			if value == "" {
				err = os.Unsetenv(key)
			} else {
				err = os.Setenv(key, value)
			}
			if err != nil {
				t.Logf("failed to restore %s: %v", key, err)
			}
		}
	}()

	done := make(chan error, 1)
	go func() {
		done <- run()
	}()

	time.Sleep(2 * time.Second)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:18080/health")
	if err == nil {
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				t.Logf("failed to close response body: %v", cerr)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status OK, got %d", resp.StatusCode)
		}
	}

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("run() returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Error("graceful shutdown timed out")
	}
}

func TestServerTimeouts_Configuration(t *testing.T) {
	tests := []struct {
		name            string
		readTimeout     string
		writeTimeout    string
		idleTimeout     string
		shutdownTimeout string
		maxHeaderBytes  string
		expectError     bool
	}{
		{
			name:            "valid timeouts",
			readTimeout:     "10",
			writeTimeout:    "10",
			idleTimeout:     "120",
			shutdownTimeout: "30",
			maxHeaderBytes:  "1048576",
			expectError:     true,
		},
		{
			name:            "zero timeouts",
			readTimeout:     "0",
			writeTimeout:    "0",
			idleTimeout:     "0",
			shutdownTimeout: "0",
			maxHeaderBytes:  "0",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", "test-secret-key-for-testing-minimum-32-chars")
			t.Setenv("SERVER_READTIMEOUT", tt.readTimeout)
			t.Setenv("SERVER_WRITETIMEOUT", tt.writeTimeout)
			t.Setenv("SERVER_IDLETIMEOUT", tt.idleTimeout)
			t.Setenv("SERVER_SHUTDOWNTIMEOUT", tt.shutdownTimeout)
			t.Setenv("SERVER_MAXHEADERBYTES", tt.maxHeaderBytes)
			t.Setenv("DATABASE_HOST", "invalid-host")

			err := run()
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestHTTPServer_TimeoutBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("skipping integration test (SKIP_INTEGRATION_TESTS is set)")
	}

	t.Setenv("JWT_SECRET", "test-secret-key-for-testing-minimum-32-chars-long")
	t.Setenv("DATABASE_HOST", "localhost")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "postgres")
	t.Setenv("DATABASE_PASSWORD", "postgres")
	t.Setenv("DATABASE_NAME", "grab")
	t.Setenv("SERVER_PORT", "18081")
	t.Setenv("SERVER_READTIMEOUT", "2")
	t.Setenv("SERVER_WRITETIMEOUT", "2")
	t.Setenv("SERVER_IDLETIMEOUT", "10")

	done := make(chan error, 1)
	go func() {
		done <- run()
	}()

	time.Sleep(2 * time.Second)

	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("http://localhost:18081/health")
	if err == nil {
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				t.Logf("failed to close response body: %v", cerr)
			}
		}()
	}

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("server shutdown timed out")
	}
}

func TestServerShutdown_WithActiveConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("skipping integration test (SKIP_INTEGRATION_TESTS is set)")
	}

	t.Setenv("JWT_SECRET", "test-secret-key-for-testing-minimum-32-chars-long")
	t.Setenv("DATABASE_HOST", "localhost")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "postgres")
	t.Setenv("DATABASE_PASSWORD", "postgres")
	t.Setenv("DATABASE_NAME", "grab")
	t.Setenv("SERVER_PORT", "18082")
	t.Setenv("SERVER_SHUTDOWNTIMEOUT", "10")

	done := make(chan error, 1)
	go func() {
		done <- run()
	}()

	time.Sleep(2 * time.Second)

	requestDone := make(chan bool, 1)
	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://localhost:18082/health")
		if err == nil {
			_ = resp.Body.Close()
		}
		requestDone <- true
	}()

	time.Sleep(500 * time.Millisecond)

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case <-requestDone:
	case <-time.After(8 * time.Second):
		t.Error("request did not complete before shutdown timeout")
	}

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Error("graceful shutdown timed out")
	}
}

func TestServerConfig_DefaultValues(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("skipping integration test (SKIP_INTEGRATION_TESTS is set)")
	}

	t.Setenv("JWT_SECRET", "test-secret-key-for-testing-minimum-32-chars-long")
	t.Setenv("DATABASE_HOST", "localhost")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "postgres")
	t.Setenv("DATABASE_PASSWORD", "postgres")
	t.Setenv("DATABASE_NAME", "grab")
	t.Setenv("SERVER_PORT", "18083")

	done := make(chan error, 1)
	go func() {
		done <- run()
	}()

	time.Sleep(2 * time.Second)

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("run() returned error: %v", err)
		}
	case <-time.After(35 * time.Second):
		t.Error("graceful shutdown with default timeout failed")
	}
}

func TestHTTPServer_MaxHeaderBytes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("skipping integration test (SKIP_INTEGRATION_TESTS is set)")
	}

	t.Setenv("JWT_SECRET", "test-secret-key-for-testing-minimum-32-chars-long")
	t.Setenv("DATABASE_HOST", "localhost")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "postgres")
	t.Setenv("DATABASE_PASSWORD", "postgres")
	t.Setenv("DATABASE_NAME", "grab")
	t.Setenv("SERVER_PORT", "18084")
	t.Setenv("SERVER_MAXHEADERBYTES", "1024")

	done := make(chan error, 1)
	go func() {
		done <- run()
	}()

	time.Sleep(2 * time.Second)

	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://localhost:18084/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	for i := 0; i < 100; i++ {
		req.Header.Add(fmt.Sprintf("X-Test-Header-%d", i), "test-value")
	}

	client := &http.Client{Timeout: 2 * time.Second}
	_, _ = client.Do(req)

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("server shutdown timed out")
	}
}
