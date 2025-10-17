package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"helixtrack.ru/chat/internal/models"
)

var (
	// Log is the global logger instance
	Log *zap.Logger
)

// InitLogger initializes the logger with the given configuration
func InitLogger(config *models.LoggerConfig) error {
	// Ensure log directory exists
	if err := os.MkdirAll(config.LogPath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file path
	logFile := filepath.Join(config.LogPath, config.LogfileBaseName+".log")

	// Setup file rotation
	fileWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    config.LogSizeLimit / (1024 * 1024), // Convert bytes to MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create file core
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		level,
	)

	// Create console core for development
	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Combine cores
	core := zapcore.NewTee(fileCore, consoleCore)

	// Create logger
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	Log.Info("Logger initialized",
		zap.String("log_file", logFile),
		zap.String("level", config.Level),
	)

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	if Log != nil {
		return Log.With(fields...)
	}
	return zap.NewNop()
}
