package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggerConfig defines the configuration for the logger middleware
type LoggerConfig struct {
	// SkipPaths is a list of paths that should not be logged
	SkipPaths []string
	// Logger is the slog logger instance to use
	Logger *slog.Logger
}

// DefaultLoggerConfig returns a default configuration for the logger middleware
func DefaultLoggerConfig() *LoggerConfig {
	// Create a JSON logger that writes to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &LoggerConfig{
		SkipPaths: []string{"/health"},
		Logger:    logger,
	}
}

// NewLoggerConfig creates a logger configuration from logging config
func NewLoggerConfig(logLevel slog.Level, skipPaths []string) *LoggerConfig {
	// Create a JSON logger that writes to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	return &LoggerConfig{
		SkipPaths: skipPaths,
		Logger:    logger,
	}
}

// Logger returns a Gin middleware for structured request logging
func Logger(config *LoggerConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	// Build a map for fast path lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Process request
		c.Next()

		// Skip logging for specified paths
		if skipPaths[path] {
			return
		}

		// Calculate request duration
		duration := time.Since(start)

		// Get response status
		statusCode := c.Writer.Status()

		// Add query string to path if present
		if raw != "" {
			path = path + "?" + raw
		}

		// Determine log level based on status code
		level := slog.LevelInfo
		if statusCode >= 500 {
			level = slog.LevelError
		} else if statusCode >= 400 {
			level = slog.LevelWarn
		}

		// Log structured data
		logger.Log(c.Request.Context(), level, "HTTP Request",
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.Duration("duration", duration),
			slog.String("duration_ms", formatDuration(duration)),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Int("response_size", c.Writer.Size()),
		)

		// Log error if present
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Request error",
					slog.String("request_id", requestID),
					slog.String("error", e.Error()),
				)
			}
		}
	}
}

// formatDuration formats duration to milliseconds string
func formatDuration(d time.Duration) string {
	return d.Round(time.Millisecond).String()
}

// LoggerWithConfig returns a Gin middleware for structured request logging with custom configuration
func LoggerWithConfig(skipPaths []string, logLevel slog.Level) gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	config := &LoggerConfig{
		SkipPaths: skipPaths,
		Logger:    logger,
	}

	return Logger(config)
}
