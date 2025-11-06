package health

import (
	"context"
	"fmt"
	"time"
)

type Service interface {
	GetHealth(ctx context.Context) HealthResponse
	GetLiveness(ctx context.Context) HealthResponse
	GetReadiness(ctx context.Context) HealthResponse
}

type service struct {
	checkers    []Checker
	startTime   time.Time
	version     string
	environment string
}

func NewService(checkers []Checker, version, environment string) Service {
	return &service{
		checkers:    checkers,
		startTime:   time.Now(),
		version:     version,
		environment: environment,
	}
}

func (s *service) GetHealth(ctx context.Context) HealthResponse {
	return HealthResponse{
		Status:      StatusHealthy,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Environment: s.environment,
		Checks:      make(map[string]CheckResult),
	}
}

func (s *service) GetLiveness(ctx context.Context) HealthResponse {
	return HealthResponse{
		Status:      StatusHealthy,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Environment: s.environment,
		Checks:      make(map[string]CheckResult),
	}
}

func (s *service) GetReadiness(ctx context.Context) HealthResponse {
	checks := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	for _, checker := range s.checkers {
		result := checker.Check(ctx)
		checks[checker.Name()] = result

		if result.Status == CheckFail {
			overallStatus = StatusUnhealthy
		} else if result.Status == CheckWarn && overallStatus != StatusUnhealthy {
			overallStatus = StatusDegraded
		}
	}

	return HealthResponse{
		Status:      overallStatus,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Environment: s.environment,
		Checks:      checks,
	}
}

func (s *service) formatUptime() string {
	uptime := time.Since(s.startTime)
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
