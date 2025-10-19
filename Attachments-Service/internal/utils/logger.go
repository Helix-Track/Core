package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new Zap logger with the specified level and format
func NewLogger(level, format string) (*zap.Logger, error) {
	// Parse log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	writer := zapcore.AddSync(os.Stdout)

	// Create core
	core := zapcore.NewCore(encoder, writer, zapLevel)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// NewDevelopmentLogger creates a logger suitable for development
func NewDevelopmentLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

// NewProductionLogger creates a logger suitable for production
func NewProductionLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

// LoggerWithFields creates a new logger with predefined fields
func LoggerWithFields(logger *zap.Logger, fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}

// LogError logs an error with context
func LogError(logger *zap.Logger, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Error(msg, fields...)
}

// LogInfo logs an info message
func LogInfo(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// LogDebug logs a debug message
func LogDebug(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// LogWarn logs a warning message
func LogWarn(logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// SugaredLogger wraps a zap.Logger for easier use
type SugaredLogger struct {
	logger *zap.SugaredLogger
}

// NewSugaredLogger creates a new sugared logger
func NewSugaredLogger(logger *zap.Logger) *SugaredLogger {
	return &SugaredLogger{
		logger: logger.Sugar(),
	}
}

// Info logs an info message
func (l *SugaredLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Infof logs a formatted info message
func (l *SugaredLogger) Infof(template string, args ...interface{}) {
	l.logger.Infof(template, args...)
}

// Error logs an error message
func (l *SugaredLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Errorf logs a formatted error message
func (l *SugaredLogger) Errorf(template string, args ...interface{}) {
	l.logger.Errorf(template, args...)
}

// Debug logs a debug message
func (l *SugaredLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Debugf logs a formatted debug message
func (l *SugaredLogger) Debugf(template string, args ...interface{}) {
	l.logger.Debugf(template, args...)
}

// Warn logs a warning message
func (l *SugaredLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Warnf logs a formatted warning message
func (l *SugaredLogger) Warnf(template string, args ...interface{}) {
	l.logger.Warnf(template, args...)
}

// Fatal logs a fatal message and exits
func (l *SugaredLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *SugaredLogger) Fatalf(template string, args ...interface{}) {
	l.logger.Fatalf(template, args...)
}

// Sync flushes any buffered log entries
func (l *SugaredLogger) Sync() error {
	return l.logger.Sync()
}

// RequestLogger creates a logger for HTTP request logging
func RequestLogger(logger *zap.Logger, requestID string) *zap.Logger {
	return logger.With(
		zap.String("request_id", requestID),
	)
}

// WithContext creates a logger with context fields
func WithContext(logger *zap.Logger, userID, ipAddress string) *zap.Logger {
	return logger.With(
		zap.String("user_id", userID),
		zap.String("ip_address", ipAddress),
	)
}

// LogSlowQuery logs a slow database query
func LogSlowQuery(logger *zap.Logger, query string, duration int64) {
	if duration > 1000 { // >1 second
		logger.Warn("slow query detected",
			zap.String("query", query),
			zap.Int64("duration_ms", duration),
		)
	}
}

// LogFileOperation logs a file operation
func LogFileOperation(logger *zap.Logger, operation, fileHash string, sizeBytes int64, duration int64) {
	logger.Info("file operation completed",
		zap.String("operation", operation),
		zap.String("file_hash", fileHash),
		zap.Int64("size_bytes", sizeBytes),
		zap.Int64("duration_ms", duration),
	)
}

// LogSecurityEvent logs a security-related event
func LogSecurityEvent(logger *zap.Logger, event, userID, ipAddress, details string) {
	logger.Warn("security event",
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("ip_address", ipAddress),
		zap.String("details", details),
	)
}

// LogQuotaExceeded logs when a user exceeds their quota
func LogQuotaExceeded(logger *zap.Logger, userID string, requestedBytes, availableBytes int64) {
	logger.Warn("quota exceeded",
		zap.String("user_id", userID),
		zap.Int64("requested_bytes", requestedBytes),
		zap.Int64("available_bytes", availableBytes),
	)
}

// LogVirusDetected logs when a virus is detected in an uploaded file
func LogVirusDetected(logger *zap.Logger, fileHash, virusName, userID string) {
	logger.Error("virus detected",
		zap.String("file_hash", fileHash),
		zap.String("virus_name", virusName),
		zap.String("user_id", userID),
	)
}

// LogStorageFailover logs when storage failover occurs
func LogStorageFailover(logger *zap.Logger, fromEndpoint, toEndpoint, reason string) {
	logger.Warn("storage failover",
		zap.String("from_endpoint", fromEndpoint),
		zap.String("to_endpoint", toEndpoint),
		zap.String("reason", reason),
	)
}

// Metrics wraps common logging metrics
type Metrics struct {
	logger *zap.Logger
}

// NewMetrics creates a new metrics logger
func NewMetrics(logger *zap.Logger) *Metrics {
	return &Metrics{logger: logger}
}

// RecordUpload records an upload metric
func (m *Metrics) RecordUpload(userID string, sizeBytes int64, duration int64, success bool) {
	m.logger.Info("upload_metric",
		zap.String("user_id", userID),
		zap.Int64("size_bytes", sizeBytes),
		zap.Int64("duration_ms", duration),
		zap.Bool("success", success),
	)
}

// RecordDownload records a download metric
func (m *Metrics) RecordDownload(userID string, sizeBytes int64, duration int64, success bool) {
	m.logger.Info("download_metric",
		zap.String("user_id", userID),
		zap.Int64("size_bytes", sizeBytes),
		zap.Int64("duration_ms", duration),
		zap.Bool("success", success),
	)
}

// RecordDeduplication records a deduplication event
func (m *Metrics) RecordDeduplication(fileHash string, savedBytes int64) {
	m.logger.Info("deduplication_metric",
		zap.String("file_hash", fileHash),
		zap.Int64("saved_bytes", savedBytes),
	)
}

// ParseLogLevel parses a log level string
func ParseLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}
