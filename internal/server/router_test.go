package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	mockUserHandler := &user.Handler{}

	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	mockAuthService := auth.NewService(cfg)

	testConfig := &config.Config{
		App: config.AppConfig{
			Version:     "1.0.0",
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port: "8080",
		},
		Ratelimit: config.RateLimitConfig{
			Enabled:  true,
			Requests: 100,
			Window:   time.Minute,
		},
		Health: config.HealthConfig{
			Timeout:              5,
			DatabaseCheckEnabled: true,
		},
	}

	router := SetupRouter(mockUserHandler, mockAuthService, testConfig, db)

	assert.NotNil(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "healthy")
}
