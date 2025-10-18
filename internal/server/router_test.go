package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	// Create a simple mock user handler
	mockUserHandler := &user.Handler{}

	// Create a simple mock auth service
	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	mockAuthService := auth.NewService(cfg)

	// Create a test config
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
		},
		Ratelimit: config.RateLimitConfig{
			Enabled:  true,
			Requests: 100,
			Window:   time.Minute,
		},
	}

	// Setup router
	router := SetupRouter(mockUserHandler, mockAuthService, testConfig)

	// Test that router is not nil
	assert.NotNil(t, router)

	// Test health endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "ok")
}
