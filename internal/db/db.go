package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
)

// customLogger wraps the default logger to ignore ErrRecordNotFound
type customLogger struct {
	logger.Interface
}

func (l customLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// Don't log "record not found" errors as they are expected in many cases
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	l.Interface.Trace(ctx, begin, fc, err)
}

func (l customLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	// Don't log "record not found" errors as they are expected in many cases
	if len(data) > 0 {
		if err, ok := data[0].(error); ok && errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
	}
	l.Interface.Error(ctx, msg, data...)
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: customLogger{logger.Default.LogMode(logger.Info)},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")
	return db, nil
}

// NewPostgresDBFromDatabaseConfig creates a new PostgreSQL DB connection from typed config
func NewPostgresDBFromDatabaseConfig(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: customLogger{logger.Default.LogMode(logger.Info)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm DB: %w", err)
	}

	sqlDB.SetConnMaxLifetime(time.Minute * 30)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

// NewSQLiteDB creates a new SQLite database connection (for testing)
func NewSQLiteDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite database: %w", err)
	}

	return db, nil
}

// LoadConfigFromEnv loads database configuration using Viper (env overrides + defaults)
func LoadConfigFromEnv() Config {
	return Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		Name:     viper.GetString("database.name"),
		SSLMode:  viper.GetString("database.sslmode"),
	}
}
