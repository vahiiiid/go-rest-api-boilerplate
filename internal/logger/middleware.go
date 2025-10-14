package logger

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDKey = "request_id"

// Middleware logs HTTP requests with structured fields
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Skip logging for health checks (optional)
		if c.Request.URL.Path == "/health" {
			return
		}

		// Log after request completes
		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("size", c.Writer.Size()),
		}

		// Add authenticated user if available
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// Log level based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			Error("Request completed with error", fields...)
		case status >= 400:
			Warn("Request completed with client error", fields...)
		default:
			Info("Request completed", fields...)
		}
	}
}

// GetRequestID retrieves request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}
