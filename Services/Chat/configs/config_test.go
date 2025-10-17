package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"helixtrack.ru/chat/internal/models"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		expectError bool
	}{
		{
			name:        "valid dev config",
			configPath:  "dev.json",
			expectError: false,
		},
		{
			name:        "valid prod config",
			configPath:  "prod.json",
			expectError: false,
		},
		{
			name:        "valid test config",
			configPath:  "test.json",
			expectError: false,
		},
		{
			name:        "non-existent file",
			configPath:  "nonexistent.json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := LoadConfig(tt.configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)

				// Validate required fields
				assert.NotEmpty(t, config.Server.Address)
				assert.Greater(t, config.Server.Port, 0)
				assert.NotEmpty(t, config.Database.Host)
				assert.NotEmpty(t, config.Database.Database)
				assert.NotEmpty(t, config.JWT.Secret)
				assert.NotEmpty(t, config.Logger.Level)
			}
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_DB_PASSWORD", "test_password_123")
	os.Setenv("TEST_JWT_SECRET", "test_jwt_secret_456")
	defer os.Unsetenv("TEST_DB_PASSWORD")
	defer os.Unsetenv("TEST_JWT_SECRET")

	config := GetDefaultConfig()
	config.Database.Password = "${TEST_DB_PASSWORD}"
	config.JWT.Secret = "${TEST_JWT_SECRET}"

	expandEnvVars(config)

	assert.Equal(t, "test_password_123", config.Database.Password)
	assert.Equal(t, "test_jwt_secret_456", config.JWT.Secret)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		modifyConfig  func(*models.Config)
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid config",
			modifyConfig: func(c *models.Config) {},
			expectError:  false,
		},
		{
			name: "invalid port - too low",
			modifyConfig: func(c *models.Config) {
				c.Server.Port = 0
			},
			expectError:   true,
			errorContains: "invalid server port",
		},
		{
			name: "invalid port - too high",
			modifyConfig: func(c *models.Config) {
				c.Server.Port = 70000
			},
			expectError:   true,
			errorContains: "invalid server port",
		},
		{
			name: "HTTPS without cert file",
			modifyConfig: func(c *models.Config) {
				c.Server.HTTPS = true
				c.Server.CertFile = ""
			},
			expectError:   true,
			errorContains: "cert_file is required",
		},
		{
			name: "HTTPS without key file",
			modifyConfig: func(c *models.Config) {
				c.Server.HTTPS = true
				c.Server.CertFile = "cert.crt"
				c.Server.KeyFile = ""
			},
			expectError:   true,
			errorContains: "key_file is required",
		},
		{
			name: "unsupported database type",
			modifyConfig: func(c *models.Config) {
				c.Database.Type = "mysql"
			},
			expectError:   true,
			errorContains: "unsupported database type",
		},
		{
			name: "empty database host",
			modifyConfig: func(c *models.Config) {
				c.Database.Host = ""
			},
			expectError:   true,
			errorContains: "database host is required",
		},
		{
			name: "empty database name",
			modifyConfig: func(c *models.Config) {
				c.Database.Database = ""
			},
			expectError:   true,
			errorContains: "database name is required",
		},
		{
			name: "empty database user",
			modifyConfig: func(c *models.Config) {
				c.Database.User = ""
			},
			expectError:   true,
			errorContains: "database user is required",
		},
		{
			name: "empty JWT secret",
			modifyConfig: func(c *models.Config) {
				c.JWT.Secret = ""
			},
			expectError:   true,
			errorContains: "JWT secret is required",
		},
		{
			name: "invalid log level",
			modifyConfig: func(c *models.Config) {
				c.Logger.Level = "invalid"
			},
			expectError:   true,
			errorContains: "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			tt.modifyConfig(config)

			err := validateConfig(config)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "0.0.0.0", config.Server.Address)
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "postgres", config.Database.Type)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "helixtrack-chat", config.JWT.Issuer)
	assert.Equal(t, "info", config.Logger.Level)
	assert.True(t, config.Security.EnableDDOSProtection)
}

func TestValidateConfigDefaults(t *testing.T) {
	config := &models.Config{
		Server: models.ServerConfig{
			Address: "0.0.0.0",
			Port:    9090,
		},
		Database: models.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Database: "test",
			User:     "test",
			Password: "test",
		},
		JWT: models.JWTConfig{
			Secret: "secret",
		},
		Logger: models.LoggerConfig{
			Level: "info",
		},
		Security: models.SecurityConfig{},
	}

	err := validateConfig(config)
	assert.NoError(t, err)

	// Check defaults were applied
	assert.Equal(t, 25, config.Database.MaxConnections)
	assert.Equal(t, 30, config.Database.ConnectionTimeout)
	assert.Equal(t, 24, config.JWT.ExpiryHours)
	assert.Equal(t, "/tmp/htChatLogs", config.Logger.LogPath)
	assert.Equal(t, "htChat", config.Logger.LogfileBaseName)
	assert.Equal(t, 100000000, config.Logger.LogSizeLimit)
	assert.Equal(t, 100, config.Security.RateLimitPerSecond)
	assert.Equal(t, 200, config.Security.RateLimitBurst)
	assert.Equal(t, 524288, config.Security.MaxMessageSize)
	assert.Equal(t, int64(104857600), config.Security.MaxAttachmentSize)
	assert.Equal(t, []string{"*"}, config.Security.AllowedOrigins)
}
