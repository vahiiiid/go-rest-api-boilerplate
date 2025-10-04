# Tests

This directory contains integration and end-to-end tests for the API.

## What Goes Here

**Integration Tests**: Tests that verify the complete request/response cycle including handlers, services, and repositories.

**Unit tests** for individual packages should be placed alongside the code:
- `internal/user/service_test.go` - for user service tests
- `internal/auth/middleware_test.go` - for auth middleware tests

## Running Tests

```bash
# Run all tests
make test

# Or using go test directly
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run tests in this directory only
go test ./tests/...
```

## Writing a New Test

### 1. Create a test file
```bash
# Name it *_test.go
touch tests/my_feature_test.go
```

### 2. Basic structure
```go
package tests

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
)

func TestMyFeature(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    
    // Create test database (in-memory SQLite)
    db := setupTestDB(t)
    
    // Create router
    router := server.SetupRouter(db)
    
    // Make request
    req := httptest.NewRequest("GET", "/api/v1/endpoint", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

### 3. Use existing helpers
See `handler_test.go` for helper functions:
- `setupTestDB(t)` - Creates in-memory SQLite database
- `createTestUser(t, db)` - Creates a test user
- `getAuthToken(t, db)` - Gets JWT token for testing

## Test Database

Tests use **SQLite in-memory** database, not PostgreSQL. This makes tests:
- ✅ Fast
- ✅ Isolated
- ✅ No external dependencies

## Best Practices

1. **Clean up**: Each test should be independent
2. **Use subtests**: Group related tests with `t.Run()`
3. **Test errors**: Don't just test happy paths
4. **Mock external services**: Don't make real API calls
5. **Use table-driven tests**: For testing multiple scenarios

## Example: Table-Driven Test

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"invalid email", "not-an-email", true},
        {"empty email", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEmail(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## CI Integration

Tests run automatically on:
- Every push to `main` or `develop`
- Every pull request

See `.github/workflows/ci.yml` for CI configuration.

---

**Need help?** Check existing tests in `handler_test.go` for examples.
