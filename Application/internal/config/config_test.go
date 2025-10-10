package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		expectError bool
		validate    func(*testing.T, *Config)
	}{
		{
			name: "Valid minimal configuration",
			configData: `{
				"log": {
					"log_path": "/tmp/htCoreLogs",
					"logfile_base_name": "",
					"log_size_limit": 100000000
				},
				"listeners": [
					{
						"address": "0.0.0.0",
						"port": 8080,
						"https": false
					}
				],
				"plugins": [],
				"database": {
					"type": "sqlite",
					"sqlite_path": "test.db"
				},
				"services": {
					"authentication": {
						"enabled": false,
						"url": ""
					},
					"permissions": {
						"enabled": false,
						"url": ""
					}
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "/tmp/htCoreLogs", c.Log.LogPath)
				assert.Equal(t, int64(100000000), c.Log.LogSizeLimit)
				assert.Len(t, c.Listeners, 1)
				assert.Equal(t, "0.0.0.0", c.Listeners[0].Address)
				assert.Equal(t, 8080, c.Listeners[0].Port)
				assert.False(t, c.Listeners[0].HTTPS)
				assert.Equal(t, "sqlite", c.Database.Type)
				assert.Equal(t, "test.db", c.Database.SQLitePath)
			},
		},
		{
			name: "PostgreSQL configuration",
			configData: `{
				"log": {"log_path": "/tmp/logs"},
				"listeners": [{"address": "0.0.0.0", "port": 8080, "https": false}],
				"plugins": [],
				"database": {
					"type": "postgres",
					"postgres_host": "localhost",
					"postgres_port": 5432,
					"postgres_user": "htcore",
					"postgres_password": "secret",
					"postgres_database": "htcore",
					"postgres_ssl_mode": "disable"
				},
				"services": {
					"authentication": {"enabled": false, "url": ""},
					"permissions": {"enabled": false, "url": ""}
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, "postgres", c.Database.Type)
				assert.Equal(t, "localhost", c.Database.PostgresHost)
				assert.Equal(t, 5432, c.Database.PostgresPort)
				assert.Equal(t, "htcore", c.Database.PostgresUser)
				assert.Equal(t, "secret", c.Database.PostgresPassword)
				assert.Equal(t, "htcore", c.Database.PostgresDatabase)
				assert.Equal(t, "disable", c.Database.PostgresSSLMode)
			},
		},
		{
			name: "HTTPS configuration",
			configData: `{
				"log": {"log_path": "/tmp/logs"},
				"listeners": [{
					"address": "0.0.0.0",
					"port": 8443,
					"https": true,
					"cert_file": "/path/to/cert.pem",
					"key_file": "/path/to/key.pem"
				}],
				"plugins": [],
				"database": {"type": "sqlite", "sqlite_path": "test.db"},
				"services": {
					"authentication": {"enabled": false, "url": ""},
					"permissions": {"enabled": false, "url": ""}
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, c *Config) {
				assert.True(t, c.Listeners[0].HTTPS)
				assert.Equal(t, "/path/to/cert.pem", c.Listeners[0].CertFile)
				assert.Equal(t, "/path/to/key.pem", c.Listeners[0].KeyFile)
			},
		},
		{
			name:        "Invalid JSON",
			configData:  `{invalid json}`,
			expectError: true,
		},
		{
			name: "Missing listeners",
			configData: `{
				"log": {"log_path": "/tmp/logs"},
				"listeners": [],
				"plugins": [],
				"database": {"type": "sqlite", "sqlite_path": "test.db"},
				"services": {
					"authentication": {"enabled": false, "url": ""},
					"permissions": {"enabled": false, "url": ""}
				}
			}`,
			expectError: true,
		},
		{
			name: "Invalid database type",
			configData: `{
				"log": {"log_path": "/tmp/logs"},
				"listeners": [{"address": "0.0.0.0", "port": 8080, "https": false}],
				"plugins": [],
				"database": {"type": "mysql"},
				"services": {
					"authentication": {"enabled": false, "url": ""},
					"permissions": {"enabled": false, "url": ""}
				}
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")
			err := os.WriteFile(configPath, []byte(tt.configData), 0644)
			require.NoError(t, err)

			// Load config
			config, err := LoadConfig(configPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, config)
				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

func TestConfig_ApplyDefaults(t *testing.T) {
	config := &Config{
		Listeners: []ListenerConfig{
			{Address: "0.0.0.0", Port: 8080, HTTPS: false},
		},
		Services: ServicesConfig{
			Authentication: ServiceEndpoint{Enabled: true, URL: "http://auth:8081"},
			Permissions:    ServiceEndpoint{Enabled: true, URL: "http://perm:8082"},
		},
	}

	config.applyDefaults()

	assert.Equal(t, "/tmp/htCoreLogs", config.Log.LogPath)
	assert.Equal(t, int64(100000000), config.Log.LogSizeLimit)
	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "sqlite", config.Database.Type)
	assert.Equal(t, "Database/Definition.sqlite", config.Database.SQLitePath)
	assert.Equal(t, 30, config.Services.Authentication.Timeout)
	assert.Equal(t, 30, config.Services.Permissions.Timeout)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			config: &Config{
				Listeners: []ListenerConfig{
					{Address: "0.0.0.0", Port: 8080, HTTPS: false},
				},
				Database: DatabaseConfig{
					Type:       "sqlite",
					SQLitePath: "test.db",
				},
			},
			expectError: false,
		},
		{
			name: "No listeners",
			config: &Config{
				Listeners: []ListenerConfig{},
				Database: DatabaseConfig{
					Type:       "sqlite",
					SQLitePath: "test.db",
				},
			},
			expectError: true,
			errorMsg:    "at least one listener must be configured",
		},
		{
			name: "Invalid port",
			config: &Config{
				Listeners: []ListenerConfig{
					{Address: "0.0.0.0", Port: 99999, HTTPS: false},
				},
				Database: DatabaseConfig{
					Type:       "sqlite",
					SQLitePath: "test.db",
				},
			},
			expectError: true,
		},
		{
			name: "HTTPS without cert",
			config: &Config{
				Listeners: []ListenerConfig{
					{Address: "0.0.0.0", Port: 8443, HTTPS: true},
				},
				Database: DatabaseConfig{
					Type:       "sqlite",
					SQLitePath: "test.db",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid database type",
			config: &Config{
				Listeners: []ListenerConfig{
					{Address: "0.0.0.0", Port: 8080, HTTPS: false},
				},
				Database: DatabaseConfig{
					Type: "mongodb",
				},
			},
			expectError: true,
		},
		{
			name: "PostgreSQL missing host",
			config: &Config{
				Listeners: []ListenerConfig{
					{Address: "0.0.0.0", Port: 8080, HTTPS: false},
				},
				Database: DatabaseConfig{
					Type: "postgres",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetPrimaryListener(t *testing.T) {
	config := &Config{
		Listeners: []ListenerConfig{
			{Address: "0.0.0.0", Port: 8080, HTTPS: false},
			{Address: "0.0.0.0", Port: 8443, HTTPS: true},
		},
	}

	listener := config.GetPrimaryListener()
	assert.NotNil(t, listener)
	assert.Equal(t, 8080, listener.Port)
}

func TestConfig_GetPrimaryListener_NoListeners(t *testing.T) {
	config := &Config{
		Listeners: []ListenerConfig{},
	}

	listener := config.GetPrimaryListener()
	assert.Nil(t, listener)
}

func TestConfig_GetListenerAddress(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "With listener",
			config: &Config{
				Listeners: []ListenerConfig{
					{Address: "127.0.0.1", Port: 9090, HTTPS: false},
				},
			},
			expected: "127.0.0.1:9090",
		},
		{
			name: "No listeners",
			config: &Config{
				Listeners: []ListenerConfig{},
			},
			expected: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := tt.config.GetListenerAddress()
			assert.Equal(t, tt.expected, addr)
		})
	}
}

func TestServiceEndpoint_Marshal(t *testing.T) {
	endpoint := ServiceEndpoint{
		Enabled: true,
		URL:     "http://localhost:8081",
		Timeout: 30,
	}

	data, err := json.Marshal(endpoint)
	require.NoError(t, err)
	assert.Contains(t, string(data), "http://localhost:8081")

	var decoded ServiceEndpoint
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, endpoint.Enabled, decoded.Enabled)
	assert.Equal(t, endpoint.URL, decoded.URL)
	assert.Equal(t, endpoint.Timeout, decoded.Timeout)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	config, err := LoadConfig("/nonexistent/path/config.json")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to read config file")
}
