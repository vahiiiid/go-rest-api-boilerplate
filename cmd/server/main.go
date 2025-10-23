package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	maxHeaderBytes := cfg.Server.MaxHeaderBytes
	if maxHeaderBytes == 0 {
		maxHeaderBytes = 1 << 20
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}

	go func() {
		logger.Info("Server starting", "address", srv.Addr)
		logger.Info("Swagger UI available", "url", fmt.Sprintf("http://localhost:%s/swagger/index.html", port))
		logger.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/health", port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("Received shutdown signal", "signal", sig)
	logger.Info("Shutting down server gracefully...")

	sqlDB, err := database.DB()
	if err == nil {
		logger.Info("Closing database connections...")
		if err := sqlDB.Close(); err != nil {
			logger.Error("Error closing database", "error", err)
		}
	}

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeout) * time.Second
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited gracefully")
	return nil
}
