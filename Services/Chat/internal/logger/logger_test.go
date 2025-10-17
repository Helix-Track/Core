package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"helixtrack.ru/chat/internal/models"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      *models.LoggerConfig
		expectError bool
	}{
		{
			name: "valid config - debug level",
			config: &models.LoggerConfig{
				LogPath:         "/tmp/test_logs",
				LogfileBaseName: "test",
				LogSizeLimit:    10000000,
				Level:           "debug",
			},
			expectError: false,
		},
		{
			name: "valid config - info level",
			config: &models.LoggerConfig{
				LogPath:         "/tmp/test_logs",
				LogfileBaseName: "test",
				LogSizeLimit:    10000000,
				Level:           "info",
			},
			expectError: false,
		},
		{
			name: "valid config - warn level",
			config: &models.LoggerConfig{
				LogPath:         "/tmp/test_logs",
				LogfileBaseName: "test",
				LogSizeLimit:    10000000,
				Level:           "warn",
			},
			expectError: false,
		},
		{
			name: "valid config - error level",
			config: &models.LoggerConfig{
				LogPath:         "/tmp/test_logs",
				LogfileBaseName: "test",
				LogSizeLimit:    10000000,
				Level:           "error",
			},
			expectError: false,
		},
		{
			name: "invalid log level",
			config: &models.LoggerConfig{
				LogPath:         "/tmp/test_logs",
				LogfileBaseName: "test",
				LogSizeLimit:    10000000,
				Level:           "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			defer os.RemoveAll(tt.config.LogPath)

			err := InitLogger(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, Log)

				// Verify log file was created
				logFile := filepath.Join(tt.config.LogPath, tt.config.LogfileBaseName+".log")
				_, err := os.Stat(logFile)
				assert.NoError(t, err)

				// Sync and close
				assert.NoError(t, Sync())
			}
		})
	}
}

func TestLoggerFunctions(t *testing.T) {
	config := &models.LoggerConfig{
		LogPath:         "/tmp/test_logs_functions",
		LogfileBaseName: "test",
		LogSizeLimit:    10000000,
		Level:           "debug",
	}

	defer os.RemoveAll(config.LogPath)

	err := InitLogger(config)
	require.NoError(t, err)

	// Test all logging functions
	t.Run("Info", func(t *testing.T) {
		Info("test info message", zap.String("key", "value"))
	})

	t.Run("Debug", func(t *testing.T) {
		Debug("test debug message", zap.Int("count", 42))
	})

	t.Run("Warn", func(t *testing.T) {
		Warn("test warning message", zap.Bool("flag", true))
	})

	t.Run("Error", func(t *testing.T) {
		Error("test error message", zap.Error(assert.AnError))
	})

	t.Run("With", func(t *testing.T) {
		childLogger := With(zap.String("component", "test"))
		assert.NotNil(t, childLogger)
	})

	// Sync
	assert.NoError(t, Sync())

	// Verify log file has content
	logFile := filepath.Join(config.LogPath, config.LogfileBaseName+".log")
	info, err := os.Stat(logFile)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestLoggerNotInitialized(t *testing.T) {
	// Reset logger
	Log = nil

	// These should not panic
	Info("test")
	Debug("test")
	Warn("test")
	Error("test")

	logger := With(zap.String("test", "value"))
	assert.NotNil(t, logger)

	assert.NoError(t, Sync())
}

func TestLoggerDirectoryCreation(t *testing.T) {
	config := &models.LoggerConfig{
		LogPath:         "/tmp/nested/deep/log/directory",
		LogfileBaseName: "test",
		LogSizeLimit:    10000000,
		Level:           "info",
	}

	defer os.RemoveAll("/tmp/nested")

	err := InitLogger(config)
	assert.NoError(t, err)

	// Verify nested directory was created
	_, err = os.Stat(config.LogPath)
	assert.NoError(t, err)

	assert.NoError(t, Sync())
}
