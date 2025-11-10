package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	response HealthResponse
}

func (m *mockService) GetHealth(ctx context.Context) HealthResponse {
	return m.response
}

func (m *mockService) GetLiveness(ctx context.Context) HealthResponse {
	return m.response
}

func (m *mockService) GetReadiness(ctx context.Context) HealthResponse {
	return m.response
}

func TestHandler_Health(t *testing.T) {
	tests := []struct {
		name           string
		response       HealthResponse
		expectedStatus int
	}{
		{
			name: "healthy response",
			response: HealthResponse{
				Status:      StatusHealthy,
				Version:     "1.0.0",
				Timestamp:   time.Now(),
				Uptime:      "1h 30m",
				Environment: "test",
				Checks:      make(map[string]CheckResult),
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{response: tt.response}
			handler := NewHandler(mockSvc)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/health", handler.Health)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), "status")
		})
	}
}

func TestHandler_Live(t *testing.T) {
	mockSvc := &mockService{
		response: HealthResponse{
			Status:      StatusHealthy,
			Version:     "1.0.0",
			Timestamp:   time.Now(),
			Uptime:      "1h 30m",
			Environment: "test",
			Checks:      make(map[string]CheckResult),
		},
	}
	handler := NewHandler(mockSvc)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health/live", handler.Live)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
}

func TestHandler_Ready(t *testing.T) {
	tests := []struct {
		name           string
		response       HealthResponse
		expectedStatus int
	}{
		{
			name: "ready - all checks pass",
			response: HealthResponse{
				Status:      StatusHealthy,
				Version:     "1.0.0",
				Timestamp:   time.Now(),
				Uptime:      "1h 30m",
				Environment: "test",
				Checks: map[string]CheckResult{
					"database": {Status: CheckPass, Message: "OK"},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not ready - unhealthy",
			response: HealthResponse{
				Status:      StatusUnhealthy,
				Version:     "1.0.0",
				Timestamp:   time.Now(),
				Uptime:      "1h 30m",
				Environment: "test",
				Checks: map[string]CheckResult{
					"database": {Status: CheckFail, Message: "Connection failed"},
				},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "ready - degraded but still serving",
			response: HealthResponse{
				Status:      StatusDegraded,
				Version:     "1.0.0",
				Timestamp:   time.Now(),
				Uptime:      "1h 30m",
				Environment: "test",
				Checks: map[string]CheckResult{
					"database": {Status: CheckWarn, Message: "Slow response"},
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{response: tt.response}
			handler := NewHandler(mockSvc)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/health/ready", handler.Ready)

			req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), "status")
		})
	}
}
