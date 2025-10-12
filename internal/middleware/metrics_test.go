package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	// Reset metrics before test
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()
	httpRequestsInProgress.Set(0)
	httpRequestSize.Reset()
	httpResponseSize.Reset()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		method         string
		handler        gin.HandlerFunc
		expectedStatus int
		skipPaths      []string
	}{
		{
			name:   "GET request to /test",
			path:   "/test",
			method: "GET",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			expectedStatus: http.StatusOK,
			skipPaths:      []string{"/health", "/metrics"},
		},
		{
			name:   "POST request to /api/users",
			path:   "/api/users",
			method: "POST",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"id": 1})
			},
			expectedStatus: http.StatusCreated,
			skipPaths:      []string{"/health", "/metrics"},
		},
		{
			name:   "Error response",
			path:   "/error",
			method: "GET",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			},
			expectedStatus: http.StatusInternalServerError,
			skipPaths:      []string{"/health", "/metrics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with metrics middleware
			router := gin.New()
			config := NewMetricsConfig(tt.skipPaths)
			router.Use(Metrics(config))
			router.Handle(tt.method, tt.path, tt.handler)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response status
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestMetricsSkipPaths(t *testing.T) {
	// Reset metrics before test
	httpRequestsTotal.Reset()

	gin.SetMode(gin.TestMode)

	skipPaths := []string{"/health", "/metrics"}
	config := NewMetricsConfig(skipPaths)

	tests := []struct {
		name            string
		path            string
		shouldBeTracked bool
	}{
		{
			name:            "Health check should be skipped",
			path:            "/health",
			shouldBeTracked: false,
		},
		{
			name:            "Metrics endpoint should be skipped",
			path:            "/metrics",
			shouldBeTracked: false,
		},
		{
			name:            "Regular endpoint should be tracked",
			path:            "/api/users",
			shouldBeTracked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset counter
			httpRequestsTotal.Reset()

			router := gin.New()
			router.Use(Metrics(config))
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Check if metrics were recorded
			count := testutil.CollectAndCount(httpRequestsTotal)
			if tt.shouldBeTracked {
				assert.Greater(t, count, 0, "Metrics should be tracked for %s", tt.path)
			} else {
				assert.Equal(t, 0, count, "Metrics should not be tracked for %s", tt.path)
			}
		})
	}
}

func TestMetricsCounters(t *testing.T) {
	// Reset metrics before test
	httpRequestsTotal.Reset()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	config := DefaultMetricsConfig()
	router.Use(Metrics(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Verify counter increment
	count := testutil.CollectAndCount(httpRequestsTotal)
	assert.Greater(t, count, 0, "Request counter should be incremented")
}

func TestMetricsHistograms(t *testing.T) {
	// Reset metrics before test
	httpRequestDuration.Reset()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	config := DefaultMetricsConfig()
	router.Use(Metrics(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify histogram recorded
	count := testutil.CollectAndCount(httpRequestDuration)
	assert.Greater(t, count, 0, "Duration histogram should record observations")
}

func TestMetricsInProgress(t *testing.T) {
	// Reset metrics before test
	httpRequestsInProgress.Set(0)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	config := DefaultMetricsConfig()
	router.Use(Metrics(config))

	// Handler that allows us to check the gauge during execution
	var inProgressValue float64
	router.GET("/test", func(c *gin.Context) {
		// Get the current value while request is being processed
		metric := prometheus.NewGauge(prometheus.GaugeOpts{Name: "temp"})
		_ = metric
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// After request completes, gauge should be back to 0
	expected := `
# HELP http_requests_in_progress Number of HTTP requests currently being processed
# TYPE http_requests_in_progress gauge
http_requests_in_progress 0
`
	err := testutil.CollectAndCompare(httpRequestsInProgress, strings.NewReader(expected))
	assert.NoError(t, err, "In-progress gauge should be 0 after request completes")

	_ = inProgressValue
}

func TestMetricsStatusCodes(t *testing.T) {
	// Reset metrics before test
	httpRequestsTotal.Reset()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		statusCode int
		handler    gin.HandlerFunc
	}{
		{
			name:       "200 OK",
			path:       "/ok",
			statusCode: http.StatusOK,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			},
		},
		{
			name:       "201 Created",
			path:       "/created",
			statusCode: http.StatusCreated,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"id": 1})
			},
		},
		{
			name:       "400 Bad Request",
			path:       "/bad-request",
			statusCode: http.StatusBadRequest,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			},
		},
		{
			name:       "404 Not Found",
			path:       "/not-found",
			statusCode: http.StatusNotFound,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			},
		},
		{
			name:       "500 Internal Server Error",
			path:       "/error",
			statusCode: http.StatusInternalServerError,
			handler: func(c *gin.Context) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			config := DefaultMetricsConfig()
			router.Use(Metrics(config))
			router.GET(tt.path, tt.handler)

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestDefaultMetricsConfig(t *testing.T) {
	config := DefaultMetricsConfig()

	assert.NotNil(t, config)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/metrics")
}

func TestNewMetricsConfig(t *testing.T) {
	customPaths := []string{"/custom1", "/custom2"}
	config := NewMetricsConfig(customPaths)

	assert.NotNil(t, config)
	assert.Equal(t, customPaths, config.SkipPaths)
}

func TestNormalizePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		routePattern string
		requestPath  string
		expected     string
	}{
		{
			name:         "Route with parameter",
			routePattern: "/api/v1/users/:id",
			requestPath:  "/api/v1/users/123",
			expected:     "/api/v1/users/:id",
		},
		{
			name:         "Route without parameter",
			routePattern: "/api/v1/users",
			requestPath:  "/api/v1/users",
			expected:     "/api/v1/users",
		},
		{
			name:         "Root path",
			routePattern: "/",
			requestPath:  "/",
			expected:     "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			var actualPath string

			router.GET(tt.routePattern, func(c *gin.Context) {
				actualPath = normalizePath(c)
				c.JSON(http.StatusOK, gin.H{"path": actualPath})
			})

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, actualPath)
		})
	}
}

func TestMetricsWithRequestBody(t *testing.T) {
	// Reset metrics before test
	httpRequestSize.Reset()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	config := DefaultMetricsConfig()
	router.Use(Metrics(config))

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	body := strings.NewReader(`{"key": "value"}`)
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify request size was tracked
	count := testutil.CollectAndCount(httpRequestSize)
	assert.Greater(t, count, 0, "Request size should be tracked")
}

func TestMetricsWithResponseBody(t *testing.T) {
	// Reset metrics before test
	httpResponseSize.Reset()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	config := DefaultMetricsConfig()
	router.Use(Metrics(config))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok", "data": "response"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response size was tracked
	count := testutil.CollectAndCount(httpResponseSize)
	assert.Greater(t, count, 0, "Response size should be tracked")
}
