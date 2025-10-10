package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(userHandler *user.Handler, authService auth.Service, cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Global middleware
	// Middleware - use configurable logging with environment-based skip paths
	skipPaths := config.GetSkipPaths(cfg.App.Environment)
	loggerConfig := middleware.NewLoggerConfig(
		cfg.Logging.GetLogLevel(),
		skipPaths,
	)
	router.Use(middleware.Logger(loggerConfig))
	router.Use(gin.Recovery())

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

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
