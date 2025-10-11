package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"helixtrack.ru/core/internal/config"
)

var (
	globalLogger *zap.Logger
	sugared      *zap.SugaredLogger
)

// Initialize initializes the global logger with the given configuration
func Initialize(cfg config.LogConfig) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(cfg.LogPath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Determine log level
	level := zapcore.InfoLevel
	if cfg.Level != "" {
		if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
			return fmt.Errorf("invalid log level '%s': %w", cfg.Level, err)
		}
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Determine log file path
	logFilePath := filepath.Join(cfg.LogPath, "htCore.log")
	if cfg.LogfileBaseName != "" {
		logFilePath = filepath.Join(cfg.LogPath, cfg.LogfileBaseName+".log")
	}

	// Configure log rotation with lumberjack
	fileWriter := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    int(cfg.LogSizeLimit / 1024 / 1024), // Convert bytes to MB
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}

	// Create core that writes to both file and console
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		level,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Combine cores
	core := zapcore.NewTee(fileCore, consoleCore)

	// Create logger
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugared = globalLogger.Sugar()

	return nil
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to a default logger if not initialized
		globalLogger, _ = zap.NewProduction()
		sugared = globalLogger.Sugar()
	}
	return globalLogger
}

// GetSugared returns the global sugared logger instance
func GetSugared() *zap.SugaredLogger {
	if sugared == nil {
		Get() // Initialize if not already done
	}
	return sugared
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		err := globalLogger.Sync()
		// Ignore "sync /dev/stdout: invalid argument" and "sync /dev/stderr: invalid argument" errors
		// These are expected on Linux and other Unix-like systems
		if err != nil && (strings.Contains(err.Error(), "sync /dev/stdout") ||
			strings.Contains(err.Error(), "sync /dev/stderr") ||
			errors.Is(err, os.ErrInvalid)) {
			return nil
		}
		return err
	}
	return nil
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// Debugf logs a debug message with fmt.Sprintf-style formatting
func Debugf(template string, args ...interface{}) {
	GetSugared().Debugf(template, args...)
}

// Infof logs an info message with fmt.Sprintf-style formatting
func Infof(template string, args ...interface{}) {
	GetSugared().Infof(template, args...)
}

// Warnf logs a warning message with fmt.Sprintf-style formatting
func Warnf(template string, args ...interface{}) {
	GetSugared().Warnf(template, args...)
}

// Errorf logs an error message with fmt.Sprintf-style formatting
func Errorf(template string, args ...interface{}) {
	GetSugared().Errorf(template, args...)
}

// Fatalf logs a fatal message with fmt.Sprintf-style formatting and exits
func Fatalf(template string, args ...interface{}) {
	GetSugared().Fatalf(template, args...)
}
