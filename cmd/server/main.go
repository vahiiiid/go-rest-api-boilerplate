package main

import (
	"fmt"
	"log/slog"
	"os"

	_ "github.com/vahiiiid/go-rest-api-boilerplate/api/docs" // swagger docs
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/server"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
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
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := slog.Default()
	logger.Info("Starting Go REST API Boilerplate...")

	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		return err
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Configuration validation failed", "error", err)
		return err
	}

	cfg.LogSafeConfig(logger)

	database, err := db.NewPostgresDBFromDatabaseConfig(cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return err
	}

	authService := auth.NewService(&cfg.JWT)
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, authService)

	router := server.SetupRouter(userHandler, authService, cfg)

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	logger.Info("Server starting", "address", addr)
	logger.Info("Swagger UI available", "url", fmt.Sprintf("http://localhost:%s/swagger/index.html", port))
	logger.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/health", port))

	if err := router.Run(addr); err != nil {
		logger.Error("Failed to start server", "error", err)
		return err
	}

	return nil
}
