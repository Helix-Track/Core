package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := `{
		"service": {
			"name": "test-service",
			"port": 8085,
			"port_range": [8085, 8095],
			"environment": "testing"
		},
		"database": {
			"driver": "postgres",
			"host": "localhost",
			"port": 5432,
			"database": "testdb",
			"user": "testuser",
			"password": "testpass",
			"encryption_key": "test-encryption-key"
		},
		"cache": {
			"in_memory": {
				"enabled": true
			},
			"redis": {
				"enabled": false
			}
		},
		"security": {
			"jwt_secret": "test-secret"
		},
		"logging": {
			"level": "info"
		}
	}`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-service", cfg.Service.Name)
	assert.Equal(t, 8085, cfg.Service.Port)
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "test-secret", cfg.Security.JWTSecret)
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/config.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(configPath, []byte("invalid json{"), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestLoad_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Config missing required JWT secret
	configData := `{
		"service": {
			"port": 8085
		},
		"database": {
			"driver": "postgres"
		},
		"security": {},
		"cache": {
			"in_memory": {"enabled": false},
			"redis": {"enabled": false}
		}
	}`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "original-host",
			Password: "original-pass",
		},
		Security: SecurityConfig{
			JWTSecret: "original-secret",
		},
		Cache: CacheConfig{
			Redis: RedisCacheConfig{
				Password: "original-redis-pass",
			},
		},
	}

	// Set environment variables
	os.Setenv("DB_HOST", "env-host")
	os.Setenv("DB_PASSWORD", "env-pass")
	os.Setenv("JWT_SECRET", "env-secret")
	os.Setenv("DB_ENCRYPTION_KEY", "env-key")
	os.Setenv("REDIS_PASSWORD", "env-redis-pass")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DB_ENCRYPTION_KEY")
		os.Unsetenv("REDIS_PASSWORD")
	}()

	cfg.applyEnvOverrides()

	assert.Equal(t, "env-host", cfg.Database.Host)
	assert.Equal(t, "env-pass", cfg.Database.Password)
	assert.Equal(t, "env-secret", cfg.Security.JWTSecret)
	assert.Equal(t, "env-key", cfg.Database.EncryptionKey)
	assert.Equal(t, "env-redis-pass", cfg.Cache.Redis.Password)
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port:      8085,
			PortRange: []int{8085, 8095},
		},
		Database: DatabaseConfig{
			Driver:        "postgres",
			Host:          "localhost",
			Database:      "testdb",
			EncryptionKey: "test-key",
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
		Cache: CacheConfig{
			Redis: RedisCacheConfig{
				Enabled:   false,
				Addresses: []string{},
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port: 99999, // Invalid port
		},
		Database: DatabaseConfig{
			Driver: "postgres",
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid service port")
}

func TestValidate_InvalidPortRange(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port:      8085,
			PortRange: []int{8095, 8085}, // Start > end
		},
		Database: DatabaseConfig{
			Driver: "postgres",
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port range")
}

func TestValidate_UnsupportedDatabaseDriver(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port: 8085,
		},
		Database: DatabaseConfig{
			Driver: "mysql", // Unsupported
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database driver")
}

func TestValidate_PostgresMissingHost(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port: 8085,
		},
		Database: DatabaseConfig{
			Driver: "postgres",
			Host:   "", // Missing
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database host is required")
}

func TestValidate_PostgresMissingDatabase(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port: 8085,
		},
		Database: DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Database: "", // Missing
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database name is required")
}

func TestValidate_PostgresMissingEncryptionKey(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port: 8085,
		},
		Database: DatabaseConfig{
			Driver:        "postgres",
			Host:          "localhost",
			Database:      "testdb",
			EncryptionKey: "", // Missing
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database encryption key is required")
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port: 8085,
		},
		Database: DatabaseConfig{
			Driver: "sqlite3",
		},
		Security: SecurityConfig{
			JWTSecret: "", // Missing
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret is required")
}

func TestValidate_RedisMissingAddresses(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{
			Port: 8085,
		},
		Database: DatabaseConfig{
			Driver: "sqlite3",
		},
		Security: SecurityConfig{
			JWTSecret: "test-secret",
		},
		Cache: CacheConfig{
			Redis: RedisCacheConfig{
				Enabled:   true,
				Addresses: []string{}, // Missing
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis addresses required")
}

func TestSetDefaults(t *testing.T) {
	cfg := &Config{
		Service: ServiceConfig{},
		Database: DatabaseConfig{
			Driver: "postgres",
		},
		Cache: CacheConfig{
			InMemory: InMemoryCacheConfig{
				Enabled: true,
			},
			Redis: RedisCacheConfig{
				Enabled: true,
			},
		},
		Security: SecurityConfig{},
		Logging:  LoggingConfig{},
	}

	cfg.setDefaults()

	// Service defaults
	assert.Equal(t, "localization-service", cfg.Service.Name)
	assert.Equal(t, "development", cfg.Service.Environment)
	assert.Equal(t, 30, cfg.Service.ReadTimeout)
	assert.Equal(t, 30, cfg.Service.WriteTimeout)
	assert.Equal(t, 8192, cfg.Service.MaxHeaderBytes)

	// Database defaults
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, 50, cfg.Database.MaxConnections)
	assert.Equal(t, 10, cfg.Database.IdleConnections)
	assert.Equal(t, 30, cfg.Database.ConnectionTimeout)
	assert.Equal(t, 3600, cfg.Database.ConnectionLifetime)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	// In-Memory Cache defaults
	assert.Equal(t, 1024, cfg.Cache.InMemory.MaxSizeMB)
	assert.Equal(t, 3600, cfg.Cache.InMemory.DefaultTTL)
	assert.Equal(t, 300, cfg.Cache.InMemory.CleanupInterval)

	// Redis Cache defaults
	assert.Equal(t, 3, cfg.Cache.Redis.MaxRetries)
	assert.Equal(t, 10, cfg.Cache.Redis.PoolSize)
	assert.Equal(t, 14400, cfg.Cache.Redis.DefaultTTL)

	// Security defaults
	assert.Equal(t, "helixtrack-auth", cfg.Security.JWTIssuer)
	assert.Equal(t, 1000, cfg.Security.RateLimiting.PerIPRequestsPerMinute)
	assert.Equal(t, 5000, cfg.Security.RateLimiting.PerUserRequestsPerMinute)
	assert.Equal(t, 100000, cfg.Security.RateLimiting.GlobalRequestsPerMinute)
	assert.Equal(t, []string{"admin", "superadmin"}, cfg.Security.AdminRoles)

	// Logging defaults
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
}

func TestGetDSN_Postgres(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:            "postgres",
		Host:              "localhost",
		Port:              5432,
		User:              "testuser",
		Password:          "testpass",
		Database:          "testdb",
		SSLMode:           "disable",
		ConnectionTimeout: 30,
		EncryptionKey:     "test-key",
	}

	dsn := cfg.GetDSN()

	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=testuser")
	assert.Contains(t, dsn, "password=testpass")
	assert.Contains(t, dsn, "dbname=testdb")
	assert.Contains(t, dsn, "sslmode=disable")
	assert.Contains(t, dsn, "connect_timeout=30")
	assert.Contains(t, dsn, "encryption_key=test-key")
}

func TestGetDSN_Postgres_NoEncryptionKey(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:            "postgres",
		Host:              "localhost",
		Port:              5432,
		User:              "testuser",
		Password:          "testpass",
		Database:          "testdb",
		SSLMode:           "disable",
		ConnectionTimeout: 30,
		EncryptionKey:     "",
	}

	dsn := cfg.GetDSN()

	assert.Contains(t, dsn, "host=localhost")
	assert.NotContains(t, dsn, "encryption_key")
}

func TestGetDSN_SQLite(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:   "sqlite3",
		Database: "/path/to/database.db",
	}

	dsn := cfg.GetDSN()
	assert.Equal(t, "/path/to/database.db", dsn)
}

func TestGetDSN_UnsupportedDriver(t *testing.T) {
	cfg := DatabaseConfig{
		Driver: "unknown",
	}

	dsn := cfg.GetDSN()
	assert.Equal(t, "", dsn)
}
