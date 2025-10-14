package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/vahiiiid/go-rest-api-boilerplate/api/docs" // swagger docs
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/logger"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/server"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
	"go.uber.org/zap"
)

// @title Go REST API Boilerplate
// @version 1.0
// @description A production-ready REST API boilerplate in Go with JWT authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize logger
	env := os.Getenv("ENV")
	version := os.Getenv("VERSION")
	if env == "" {
		env = "development"
	}

	if err := logger.Init(env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync() // Flush buffered logs on shutdown

	logger.Info("Starting Go REST API Boilerplate",
		zap.String("environment", env),
		zap.String("version", version),
	)

	// Load configuration (viper-based)
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Fatal("Failed to load configuration: ", zap.Error(err))
	}

	// Connect to database using typed config
	database, err := db.NewPostgresDBFromDatabaseConfig(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database: ", zap.Error(err))
	}

	// Run migrations
	// logger.Info("Database connection established")
	logger.Info("Connected to database",
		zap.String("host", dbConfig.Host),
		zap.String("port", dbConfig.Port),
	)
	if err := database.AutoMigrate(&user.User{}); err != nil {
		logger.Fatal("Failed to run migrations: ", zap.Error(err))
	}
	logger.Info("Migrations completed successfully")

	// Initialize services
	authService := auth.NewService(&cfg.JWT)
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

	// Setup router with configuration
	router := server.SetupRouter(userHandler, authService, cfg)

	// Get port from config
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)

	logger.Info("Server starting", zap.String("server", fmt.Sprintf("http://localhost:%s", port)))

	logger.Info("Swagger UI available at", zap.String("swagger", fmt.Sprintf("http://localhost:%s/swagger/index.html", port)))

	logger.Info("Health check available at", zap.String("health", fmt.Sprintf("http://localhost:%s/health", port)))

	if err := router.Run(addr); err != nil {
		logger.Fatal("Failed to start server: ", zap.Error(err))
	}
}
