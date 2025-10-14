package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/logger"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(userHandler *user.Handler, authService auth.Service, cfg *config.Config) *gin.Engine {
	router := gin.New()

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router.Use(logger.Middleware())
	router.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

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
