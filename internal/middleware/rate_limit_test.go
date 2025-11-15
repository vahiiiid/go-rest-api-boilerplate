package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// MockStorage is a mock implementation of Storage interface for testing
type MockStorage struct {
	store map[string]*rate.Limiter
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		store: make(map[string]*rate.Limiter),
	}
}

func (m *MockStorage) Add(key string, limiter *rate.Limiter) bool {
	if _, exists := m.store[key]; exists {
		return false
	}
	m.store[key] = limiter
	return true
}

func (m *MockStorage) Get(key string) (*rate.Limiter, bool) {
	limiter, exists := m.store[key]
	return limiter, exists
}

// TestNewRateLimitMiddleware tests the NewRateLimitMiddleware function
func TestNewRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		window       time.Duration
		requests     int
		keyFunc      func(*gin.Context) string
		store        Storage
		testRequests int
		expectBlocks int
		description  string
	}{
		{
			name:     "basic rate limiting with IP-based key",
			window:   time.Second,
			requests: 2,
			keyFunc: func(c *gin.Context) string {
				return c.ClientIP()
			},
			store:        NewMockStorage(),
			testRequests: 5,
			expectBlocks: 3, // First 2 pass, next 3 blocked
			description:  "2 requests per second should allow first 2, block remaining",
		},
		{
			name:     "rate limiting with custom key function",
			window:   2 * time.Second,
			requests: 3,
			keyFunc: func(c *gin.Context) string {
				return c.GetHeader("X-User-ID")
			},
			store:        NewMockStorage(),
			testRequests: 4,
			expectBlocks: 1, // First 3 pass, last 1 blocked
			description:  "3 requests per 2 seconds with user header key",
		},
		{
			name:     "rate limiting with nil store uses default",
			window:   time.Second,
			requests: 1,
			keyFunc: func(c *gin.Context) string {
				return "test-key"
			},
			store:        nil, // Should use default store
			testRequests: 3,
			expectBlocks: 2, // First 1 passes, next 2 blocked
			description:  "nil store should use default LRU store",
		},
		{
			name:     "high rate limit allows many requests",
			window:   time.Second,
			requests: 100,
			keyFunc: func(c *gin.Context) string {
				return c.ClientIP()
			},
			store:        NewMockStorage(),
			testRequests: 10,
			expectBlocks: 0, // All should pass with high limit
			description:  "100 requests per second should allow all 10 test requests",
		},
		{
			name:     "very strict rate limiting",
			window:   10 * time.Second,
			requests: 1,
			keyFunc: func(c *gin.Context) string {
				return "single-key"
			},
			store:        NewMockStorage(),
			testRequests: 2,
			expectBlocks: 1, // Only first request passes
			description:  "1 request per 10 seconds should be very restrictive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRateLimitMiddleware(tt.window, tt.requests, tt.keyFunc, tt.store)
			assert.NotNil(t, middleware, "Middleware should not be nil")

			router := gin.New()
			router.Use(apiErrors.ErrorHandler())
			router.Use(middleware)
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			successCount := 0
			blockedCount := 0

			// Make test requests
			for i := 0; i < tt.testRequests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)

				// Set custom header if key function uses it
				if tt.keyFunc != nil {
					if tt.name == "rate limiting with custom key function" {
						req.Header.Set("X-User-ID", "user123")
					}
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					successCount++
				} else if w.Code == http.StatusTooManyRequests {
					blockedCount++

					// Verify rate limit headers are set
					assert.NotEmpty(t, w.Header().Get("Retry-After"), "Retry-After header should be set")
					assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"), "X-RateLimit-Limit header should be set")
					assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"), "X-RateLimit-Remaining header should be set")
					assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"), "X-RateLimit-Reset header should be set")

					var response map[string]interface{}
					assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

					assert.False(t, response["success"].(bool))
					errorObj := response["error"].(map[string]interface{})
					assert.Equal(t, "Rate limit exceeded", errorObj["message"])
					assert.Contains(t, errorObj, "retry_after")
				}
			}

			// Verify expectations
			assert.Equal(t, tt.expectBlocks, blockedCount,
				"Expected %d blocked requests, got %d. %s", tt.expectBlocks, blockedCount, tt.description)
			assert.Equal(t, tt.testRequests-tt.expectBlocks, successCount,
				"Expected %d successful requests, got %d", tt.testRequests-tt.expectBlocks, successCount)
		})
	}
}

// TestRateLimitMiddleware_DifferentKeys tests that different keys have separate limits
func TestRateLimitMiddleware_DifferentKeys(t *testing.T) {
	keyFunc := func(c *gin.Context) string {
		return c.GetHeader("X-Client-ID")
	}

	middleware := NewRateLimitMiddleware(time.Second, 1, keyFunc, NewMockStorage())

	router := gin.New()
	router.Use(apiErrors.ErrorHandler())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with client1
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Client-ID", "client1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "First request from client1 should succeed")

	// Test with client2 - should also succeed (different key)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Client-ID", "client2")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code, "First request from client2 should succeed")

	// Test with client1 again - should be blocked
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Client-ID", "client1")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code, "Second request from client1 should be blocked")
}

// TestRateLimitMiddleware_Headers tests rate limit header values
func TestRateLimitMiddleware_Headers(t *testing.T) {
	middleware := NewRateLimitMiddleware(time.Second, 5, func(c *gin.Context) string {
		return "test"
	}, NewMockStorage())

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make requests until blocked
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			// Verify header values
			retryAfter := w.Header().Get("Retry-After")
			assert.NotEmpty(t, retryAfter, "Retry-After should not be empty")

			retrySeconds, err := strconv.Atoi(retryAfter)
			assert.NoError(t, err, "Retry-After should be a valid integer")
			assert.Greater(t, retrySeconds, 0, "Retry-After should be positive")

			limit := w.Header().Get("X-RateLimit-Limit")
			assert.Equal(t, "5", limit, "X-RateLimit-Limit should be 5")

			remaining := w.Header().Get("X-RateLimit-Remaining")
			assert.Equal(t, "0", remaining, "X-RateLimit-Remaining should be 0 when blocked")

			reset := w.Header().Get("X-RateLimit-Reset")
			assert.NotEmpty(t, reset, "X-RateLimit-Reset should not be empty")

			resetTime, err := strconv.ParseInt(reset, 10, 64)
			assert.NoError(t, err, "X-RateLimit-Reset should be a valid unix timestamp")
			assert.Greater(t, resetTime, time.Now().Unix(), "Reset time should be in the future")

			break
		}
	}
}
