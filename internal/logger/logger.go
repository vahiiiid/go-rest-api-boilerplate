package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init initializes the global logger
func Init(env string) error {
	return InitWithWriter(env, nil)
}

// InitWithWriter initializes the global logger with writer (for testing)
func InitWithWriter(env string, writer io.Writer) error {
	var config zapcore.EncoderConfig
	var level zapcore.Level
	var encoder zapcore.Encoder

	if env == "production" {
		// Production: JSON format, no stacktraces for non-errors
		config = zap.NewProductionEncoderConfig()
		config.TimeKey = "timestamp"
		config.MessageKey = "msg"
		config.LevelKey = "level"
		config.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(config)
		level = zapcore.InfoLevel
	} else {
		// Development: human-readable format, colorized
		config = zap.NewDevelopmentEncoderConfig()
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewConsoleEncoder(config)
		level = zapcore.DebugLevel
	}

	// Set level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		var zapLevel zapcore.Level
		if err := zapLevel.UnmarshalText([]byte(logLevel)); err == nil {
			level = zapLevel
		}
	}

	var logger *zap.Logger

	// If a writer instance is not nil (used for testing)
	if writer != nil {
		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(writer),
			level,
		)
		logger = zap.New(core)
	} else {
		// Writing logs in "server.log" only for production

		if env == "production" {
			file, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			core := zapcore.NewTee(
				zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
				zapcore.NewCore(encoder, zapcore.AddSync(file), level),
			)
			logger = zap.New(core)
		} else {
			core := zapcore.NewCore(
				encoder,
				zapcore.AddSync(os.Stdout),
				level,
			)
			logger = zap.New(core)
		}
	}

	Log = logger
	return nil
}

// Sync flushes any buffered log entries (call on shutdown)
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// Convenience functions
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}
