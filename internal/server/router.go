package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/health"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(userHandler *user.Handler, authService auth.Service, cfg *config.Config, db *gorm.DB) *gin.Engine {
	router := gin.New()

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	skipPaths := config.GetSkipPaths(cfg.App.Environment)
	loggerConfig := middleware.NewLoggerConfig(
		cfg.Logging.GetLogLevel(),
		skipPaths,
	)
	router.Use(middleware.Logger(loggerConfig))
	router.Use(errors.ErrorHandler())
	router.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

	var checkers []health.Checker
	if cfg.Health.DatabaseCheckEnabled {
		dbChecker := health.NewDatabaseChecker(db)
		checkers = append(checkers, dbChecker)
	}
	healthService := health.NewService(checkers, cfg.App.Version, cfg.App.Environment)
	healthHandler := health.NewHandler(healthService)

	router.GET("/health", healthHandler.Health)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rlCfg := cfg.Ratelimit
	if rlCfg.Enabled {
		router.Use(
			middleware.NewRateLimitMiddleware(
				rlCfg.Window,
				rlCfg.Requests,
				func(c *gin.Context) string {
					ip := c.ClientIP()
					if ip == "" {
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
				nil,
			),
		)
	}

	v1 := router.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", userHandler.Login)
			authGroup.POST("/refresh", userHandler.RefreshToken)
			authGroup.POST("/logout", auth.AuthMiddleware(authService), userHandler.Logout)
		}

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
