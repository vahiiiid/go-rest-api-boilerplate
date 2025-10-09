# Middleware Package

This package contains custom middleware for the Go REST API Boilerplate.

## Logger Middleware

The logger middleware provides structured JSON logging for all HTTP requests with comprehensive tracking capabilities.

### Features

- **Structured Logging**: Uses Go's standard `log/slog` for JSON-formatted logs
- **Request Tracking**: Automatic generation and propagation of request IDs
- **Smart Log Levels**: Adjusts log level based on response status
  - `INFO`: 2xx and 3xx responses
  - `WARN`: 4xx responses (client errors)
  - `ERROR`: 5xx responses (server errors)
- **Performance Metrics**: Tracks request duration in milliseconds
- **Configurable**: Skip logging for specific paths (e.g., health checks)
- **HTTP Headers**: Automatically adds `X-Request-ID` to response headers

### Usage

#### Default Configuration

The simplest way to use the logger middleware is with default settings:

```go
import "github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"

router := gin.New()
router.Use(middleware.Logger(middleware.DefaultLoggerConfig()))
```

Default configuration:
- Skips `/health` endpoint
- Logs at `INFO` level
- JSON output to stdout

#### Custom Configuration

Create a custom configuration for more control:

```go
import (
    "log/slog"
    "os"
    "github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
)

// Create custom logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Configure middleware
config := &middleware.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
    Logger:    logger,
}

router.Use(middleware.Logger(config))
```

#### Simplified Custom Configuration

Use the helper function for quick customization:

```go
import (
    "log/slog"
    "github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
)

router.Use(middleware.LoggerWithConfig(
    []string{"/health", "/metrics"}, // paths to skip
    slog.LevelDebug,                  // log level
))
```

### Log Format

Each request generates a structured JSON log entry:

```json
{
    "time": "2025-10-06T23:45:12.123Z",
    "level": "INFO",
    "msg": "HTTP Request",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "method": "POST",
    "path": "/api/v1/users",
    "status": 200,
    "duration": 45123456,
    "duration_ms": "45.123ms",
    "client_ip": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "response_size": 256
}
```

### Logged Fields

| Field | Type | Description |
|-------|------|-------------|
| `time` | string | Timestamp of the log entry |
| `level` | string | Log level (INFO/WARN/ERROR) |
| `msg` | string | Log message ("HTTP Request") |
| `request_id` | string | Unique identifier for the request |
| `method` | string | HTTP method (GET, POST, etc.) |
| `path` | string | Request path with query parameters |
| `status` | int | HTTP response status code |
| `duration` | int64 | Request duration in nanoseconds |
| `duration_ms` | string | Human-readable duration |
| `client_ip` | string | Client IP address |
| `user_agent` | string | User agent string |
| `response_size` | int | Response body size in bytes |

### Request ID

The middleware automatically handles request IDs:

1. **Provided ID**: If the client sends `X-Request-ID` header, it's used
2. **Generated ID**: Otherwise, a UUID is generated automatically
3. **Propagation**: The ID is:
   - Stored in the Gin context as `request_id`
   - Added to response headers as `X-Request-ID`
   - Included in all log entries

Example of accessing request ID in handlers:

```go
func MyHandler(c *gin.Context) {
    requestID, exists := c.Get("request_id")
    if exists {
        // Use the request ID
        log.Printf("Processing request: %s", requestID)
    }
}
```

### Skipping Paths

To avoid cluttering logs with high-frequency endpoints:

```go
config := &middleware.LoggerConfig{
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/readiness",
    },
    Logger: logger,
}
```

### Error Logging

If errors are added to the Gin context, they're logged separately:

```go
func MyHandler(c *gin.Context) {
    err := someOperation()
    if err != nil {
        c.Error(err) // Automatically logged by middleware
        c.JSON(500, gin.H{"error": "Internal error"})
        return
    }
}
```

### Log Levels by Status Code

The middleware automatically adjusts log levels:

```go
// 200-399: INFO level
GET /api/v1/users/1 -> 200 OK -> INFO

// 400-499: WARN level
POST /api/v1/users -> 400 Bad Request -> WARN

// 500-599: ERROR level
GET /api/v1/users/999 -> 500 Internal Error -> ERROR
```

### Testing

The middleware includes comprehensive unit tests:

```bash
# Run middleware tests
go test ./internal/middleware -v

# Run with coverage
go test ./internal/middleware -cover

# Generate coverage report
go test ./internal/middleware -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Performance

The logger middleware is designed for minimal overhead:

- **Fast JSON serialization**: Uses `slog`'s optimized JSON handler
- **Efficient path matching**: Uses map lookup for skip paths
- **No blocking I/O**: Async writes to stdout
- **Memory efficient**: Reuses buffers where possible

Typical overhead: **< 100µs per request**

### Dependencies

- `log/slog` - Go's standard structured logging (Go 1.21+)
- `github.com/gin-gonic/gin` - Web framework
- `github.com/google/uuid` - UUID generation

### Configuration Examples

#### Production Configuration

```go
// Production: INFO level, skip monitoring endpoints
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

config := &middleware.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics", "/readiness"},
    Logger:    logger,
}

router.Use(middleware.Logger(config))
```

#### Development Configuration

```go
// Development: DEBUG level, skip only health
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

config := &middleware.LoggerConfig{
    SkipPaths: []string{"/health"},
    Logger:    logger,
}

router.Use(middleware.Logger(config))
```

#### Custom File Logging

```go
// Log to file instead of stdout
logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}

logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

config := &middleware.LoggerConfig{
    SkipPaths: []string{"/health"},
    Logger:    logger,
}

router.Use(middleware.Logger(config))
```

## Adding New Middleware

To add new middleware to this package:

1. Create a new file: `internal/middleware/your_middleware.go`
2. Implement the middleware function:
   ```go
   func YourMiddleware() gin.HandlerFunc {
       return func(c *gin.Context) {
           // Before request processing

           c.Next()

           // After request processing
       }
   }
   ```
3. Add tests: `internal/middleware/your_middleware_test.go`
4. Update this README with documentation
5. Register in `internal/server/router.go`

## Future Enhancements

Planned middleware additions:

- **Rate Limiter** - Request rate limiting by IP/user
- **Request ID** - Standalone request ID middleware (currently part of logger)
- **Timeout** - Request timeout handling
- **Compression** - Response compression (gzip)
- **CORS** - Enhanced CORS handling (currently using gin-contrib)
- **Metrics** - Prometheus metrics collection
- **Circuit Breaker** - Fault tolerance for external services

## Contributing

When contributing middleware:

1. Follow the existing code style
2. Include comprehensive tests (aim for 100% coverage)
3. Document all configuration options
4. Update this README with usage examples
5. Ensure thread-safety for concurrent requests

## License

This package is part of the Go REST API Boilerplate project and follows the same license (MIT).
