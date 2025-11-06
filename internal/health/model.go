package health

import "time"

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type CheckStatus string

const (
	CheckPass CheckStatus = "pass"
	CheckWarn CheckStatus = "warn"
	CheckFail CheckStatus = "fail"
)

type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Checks      map[string]CheckResult `json:"checks"`
	Environment string                 `json:"environment"`
}

type CheckResult struct {
	Status       CheckStatus `json:"status"`
	Message      string      `json:"message,omitempty"`
	ResponseTime string      `json:"response_time,omitempty"`
	Details      interface{} `json:"details,omitempty"`
}
