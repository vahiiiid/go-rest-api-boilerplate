package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockChecker struct {
	name   string
	result CheckResult
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Check(ctx context.Context) CheckResult {
	return m.result
}

func TestService_GetHealth(t *testing.T) {
	svc := NewService([]Checker{}, "1.0.0", "test")

	response := svc.GetHealth(context.Background())

	assert.Equal(t, StatusHealthy, response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	assert.Equal(t, "test", response.Environment)
	assert.NotZero(t, response.Timestamp)
}

func TestService_GetLiveness(t *testing.T) {
	svc := NewService([]Checker{}, "1.0.0", "test")

	response := svc.GetLiveness(context.Background())

	assert.Equal(t, StatusHealthy, response.Status)
	assert.Equal(t, "1.0.0", response.Version)
}

func TestService_GetReadiness(t *testing.T) {
	tests := []struct {
		name           string
		checkers       []Checker
		expectedStatus HealthStatus
	}{
		{
			name: "all checks pass",
			checkers: []Checker{
				&mockChecker{name: "db", result: CheckResult{Status: CheckPass, Message: "OK"}},
			},
			expectedStatus: StatusHealthy,
		},
		{
			name: "one check fails",
			checkers: []Checker{
				&mockChecker{name: "db", result: CheckResult{Status: CheckFail, Message: "Failed"}},
			},
			expectedStatus: StatusUnhealthy,
		},
		{
			name: "one check warns",
			checkers: []Checker{
				&mockChecker{name: "db", result: CheckResult{Status: CheckWarn, Message: "Slow"}},
			},
			expectedStatus: StatusDegraded,
		},
		{
			name: "mixed - fail takes precedence",
			checkers: []Checker{
				&mockChecker{name: "db", result: CheckResult{Status: CheckPass, Message: "OK"}},
				&mockChecker{name: "cache", result: CheckResult{Status: CheckFail, Message: "Failed"}},
			},
			expectedStatus: StatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.checkers, "1.0.0", "test")

			response := svc.GetReadiness(context.Background())

			assert.Equal(t, tt.expectedStatus, response.Status)
			assert.Len(t, response.Checks, len(tt.checkers))
		})
	}
}

func TestService_FormatUptime(t *testing.T) {
	svc := &service{
		startTime: time.Now().Add(-1 * time.Hour).Add(-30 * time.Minute),
	}

	uptime := svc.formatUptime()

	assert.Contains(t, uptime, "1h")
	assert.Contains(t, uptime, "30m")
}
