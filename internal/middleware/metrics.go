package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpRequestsTotal counts total HTTP requests by method, path, and status
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestDuration tracks HTTP request latency in seconds
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// httpRequestsInProgress tracks the number of HTTP requests currently being processed
	httpRequestsInProgress = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_progress",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// httpRequestSize tracks HTTP request size in bytes
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// httpResponseSize tracks HTTP response size in bytes
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)
)

// MetricsConfig defines the configuration for the metrics middleware
type MetricsConfig struct {
	// SkipPaths is a list of paths that should not be tracked
	SkipPaths []string
}

// DefaultMetricsConfig returns a default configuration for the metrics middleware
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		SkipPaths: []string{"/health", "/metrics"},
	}
}

// NewMetricsConfig creates a metrics configuration with custom skip paths
func NewMetricsConfig(skipPaths []string) *MetricsConfig {
	return &MetricsConfig{
		SkipPaths: skipPaths,
	}
}

// Metrics returns a Gin middleware for Prometheus metrics collection
func Metrics(config *MetricsConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	// Build a map for fast path lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip metrics collection for specified paths
		if skipPaths[path] {
			c.Next()
			return
		}

		// Normalize path to avoid high cardinality (e.g., /users/123 -> /users/:id)
		normalizedPath := normalizePath(c)

		// Start timer
		start := time.Now()

		// Increment in-progress requests
		httpRequestsInProgress.Inc()

		// Track request size
		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(
				c.Request.Method,
				normalizedPath,
			).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Decrement in-progress requests
		httpRequestsInProgress.Dec()

		// Calculate request duration
		duration := time.Since(start).Seconds()

		// Get response status
		statusCode := c.Writer.Status()
		statusStr := strconv.Itoa(statusCode)

		// Record metrics
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			normalizedPath,
			statusStr,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			normalizedPath,
		).Observe(duration)

		// Track response size
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(
				c.Request.Method,
				normalizedPath,
			).Observe(float64(responseSize))
		}
	}
}

// normalizePath converts actual paths to route patterns to avoid high cardinality
// Example: /api/v1/users/123 -> /api/v1/users/:id
func normalizePath(c *gin.Context) string {
	// Get the matched route pattern if available
	route := c.FullPath()
	if route != "" {
		return route
	}

	// Fallback: use the actual path but sanitize it
	path := c.Request.URL.Path

	// Remove trailing slash for consistency
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return path
}
