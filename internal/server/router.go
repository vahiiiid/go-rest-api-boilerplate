package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(userHandler *user.Handler, authService auth.Service, cfg *config.Config) *gin.Engine {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware - use configurable logging with environment-based skip paths
	skipPaths := config.GetSkipPaths(cfg.Server.Env)
	loggerConfig := middleware.NewLoggerConfig(
		cfg.Logging.GetLogLevel(),
		skipPaths,
	)
	router.Use(middleware.Logger(loggerConfig))
	router.Use(gin.Recovery())

	// Metrics middleware - track HTTP requests
	if cfg.Metrics.Enabled {
		metricsConfig := middleware.NewMetricsConfig(skipPaths)
		router.Use(middleware.Metrics(metricsConfig))
	}

	// CORS configuration - permissive for development
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Metrics endpoint - expose Prometheus metrics
	if cfg.Metrics.Enabled {
		router.GET(cfg.Metrics.Path, gin.WrapH(promhttp.Handler()))
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Global rate-limit middleware (per client IP)
	rlCfg := cfg.Ratelimit
	if rlCfg.Enabled {
		router.Use(
			middleware.NewRateLimitMiddleware(
				rlCfg.Window,
				rlCfg.Requests,
				func(c *gin.Context) string {
					ip := c.ClientIP()
					if ip == "" {
						// Fallback for test environments or when ClientIP is empty
						ip = c.GetHeader("X-Forwarded-For")
						if ip == "" {
							ip = c.GetHeader("X-Real-IP")
						}
						if ip == "" {
							ip = "unknown"
						}
					}
					return ip
				},
				nil, // default in-memory LRU
			),
		)
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", userHandler.Login)
		}

		// Protected user routes
		usersGroup := v1.Group("/users")
		usersGroup.Use(auth.AuthMiddleware(authService))
		{
			usersGroup.GET("/:id", userHandler.GetUser)
			usersGroup.PUT("/:id", userHandler.UpdateUser)
			usersGroup.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	return router
}
