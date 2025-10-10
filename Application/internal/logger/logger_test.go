package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/config"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name        string
		config      config.LogConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: config.LogConfig{
				LogPath:         t.TempDir(),
				LogfileBaseName: "test",
				LogSizeLimit:    10485760, // 10MB
				Level:           "info",
			},
			expectError: false,
		},
		{
			name: "Debug level",
			config: config.LogConfig{
				LogPath:         t.TempDir(),
				LogfileBaseName: "test",
				LogSizeLimit:    10485760,
				Level:           "debug",
			},
			expectError: false,
		},
		{
			name: "Warn level",
			config: config.LogConfig{
				LogPath:         t.TempDir(),
				LogfileBaseName: "test",
				LogSizeLimit:    10485760,
				Level:           "warn",
			},
			expectError: false,
		},
		{
			name: "Error level",
			config: config.LogConfig{
				LogPath:         t.TempDir(),
				LogfileBaseName: "test",
				LogSizeLimit:    10485760,
				Level:           "error",
			},
			expectError: false,
		},
		{
			name: "Invalid level",
			config: config.LogConfig{
				LogPath:         t.TempDir(),
				LogfileBaseName: "test",
				LogSizeLimit:    10485760,
				Level:           "invalid",
			},
			expectError: true,
		},
		{
			name: "Empty basename",
			config: config.LogConfig{
				LogPath:         t.TempDir(),
				LogfileBaseName: "",
				LogSizeLimit:    10485760,
				Level:           "info",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger
			globalLogger = nil
			sugared = nil

			err := Initialize(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, globalLogger)
				assert.NotNil(t, sugared)

				// Verify log file was created
				expectedLogFile := filepath.Join(tt.config.LogPath, "htCore.log")
				if tt.config.LogfileBaseName != "" {
					expectedLogFile = filepath.Join(tt.config.LogPath, tt.config.LogfileBaseName+".log")
				}

				// Write a test log to ensure file is created
				Info("Test log message")
				Sync()

				_, err := os.Stat(expectedLogFile)
				assert.NoError(t, err, "Log file should exist")
			}
		})
	}
}

func TestGet(t *testing.T) {
	// Reset global logger
	globalLogger = nil
	sugared = nil

	logger := Get()
	assert.NotNil(t, logger)

	// Should return the same instance on subsequent calls
	logger2 := Get()
	assert.Equal(t, logger, logger2)
}

func TestGetSugared(t *testing.T) {
	// Reset global logger
	globalLogger = nil
	sugared = nil

	logger := GetSugared()
	assert.NotNil(t, logger)

	// Should return the same instance on subsequent calls
	logger2 := GetSugared()
	assert.Equal(t, logger, logger2)
}

func TestLoggingFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.LogConfig{
		LogPath:         tmpDir,
		LogfileBaseName: "test",
		LogSizeLimit:    10485760,
		Level:           "debug",
	}

	err := Initialize(cfg)
	require.NoError(t, err)

	// Test all logging functions (should not panic)
	t.Run("Debug", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Debug("debug message", zap.String("key", "value"))
		})
	})

	t.Run("Info", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Info("info message", zap.String("key", "value"))
		})
	})

	t.Run("Warn", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Warn("warn message", zap.String("key", "value"))
		})
	})

	t.Run("Error", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Error("error message", zap.String("key", "value"))
		})
	})

	t.Run("Debugf", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Debugf("debug %s %d", "test", 123)
		})
	})

	t.Run("Infof", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Infof("info %s %d", "test", 123)
		})
	})

	t.Run("Warnf", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Warnf("warn %s %d", "test", 123)
		})
	})

	t.Run("Errorf", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Errorf("error %s %d", "test", 123)
		})
	})

	// Sync logs
	err = Sync()
	assert.NoError(t, err)

	// Verify log file exists and has content
	logFilePath := filepath.Join(tmpDir, "test.log")
	stat, err := os.Stat(logFilePath)
	require.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(0), "Log file should contain data")
}

func TestSync(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.LogConfig{
		LogPath:         tmpDir,
		LogfileBaseName: "test",
		LogSizeLimit:    10485760,
		Level:           "info",
	}

	err := Initialize(cfg)
	require.NoError(t, err)

	Info("test message")

	err = Sync()
	assert.NoError(t, err)
}

func TestSync_Uninitialized(t *testing.T) {
	// Reset global logger
	globalLogger = nil
	sugared = nil

	err := Sync()
	assert.NoError(t, err, "Sync should not error when logger is nil")
}

func TestInitialize_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "nested", "log", "directory")

	cfg := config.LogConfig{
		LogPath:         logPath,
		LogfileBaseName: "test",
		LogSizeLimit:    10485760,
		Level:           "info",
	}

	err := Initialize(cfg)
	assert.NoError(t, err)

	// Verify directory was created
	stat, err := os.Stat(logPath)
	require.NoError(t, err)
	assert.True(t, stat.IsDir())

	// Write a log and verify file is created
	Info("test message")
	Sync()

	logFilePath := filepath.Join(logPath, "test.log")
	_, err = os.Stat(logFilePath)
	assert.NoError(t, err)
}
