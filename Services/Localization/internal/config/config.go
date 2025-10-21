package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the complete service configuration
type Config struct {
	Service  ServiceConfig  `json:"service"`
	Database DatabaseConfig `json:"database"`
	Cache    CacheConfig    `json:"cache"`
	Security SecurityConfig `json:"security"`
	Logging  LoggingConfig  `json:"logging"`
}

// ServiceConfig holds service-level configuration
type ServiceConfig struct {
	Name           string          `json:"name"`
	Port           int             `json:"port"`
	PortRange      []int           `json:"port_range"`
	Environment    string          `json:"environment"`    // development, staging, production
	ReadTimeout    int             `json:"read_timeout"`   // seconds
	WriteTimeout   int             `json:"write_timeout"`  // seconds
	MaxHeaderBytes int             `json:"max_header_bytes"`
	TLSCertFile    string          `json:"tls_cert_file"`  // Path to TLS certificate (for HTTP/3)
	TLSKeyFile     string          `json:"tls_key_file"`   // Path to TLS private key (for HTTP/3)
	Discovery      DiscoveryConfig `json:"discovery"`
}

// DiscoveryConfig holds service discovery configuration
type DiscoveryConfig struct {
	Enabled       bool   `json:"enabled"`
	Provider      string `json:"provider"`        // consul, etcd
	ConsulAddress string `json:"consul_address"`
	EtcdEndpoints []string `json:"etcd_endpoints"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver             string `json:"driver"`              // postgres, sqlite3
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Database           string `json:"database"`
	User               string `json:"user"`
	Password           string `json:"password"`
	SSLMode            string `json:"ssl_mode"`
	MaxConnections     int    `json:"max_connections"`
	IdleConnections    int    `json:"idle_connections"`
	ConnectionTimeout  int    `json:"connection_timeout"`  // seconds
	ConnectionLifetime int    `json:"connection_lifetime"` // seconds
	EncryptionKey      string `json:"encryption_key"`      // SQL Cipher key
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	InMemory InMemoryCacheConfig `json:"in_memory"`
	Redis    RedisCacheConfig    `json:"redis"`
}

// InMemoryCacheConfig holds in-memory cache configuration
type InMemoryCacheConfig struct {
	Enabled          bool `json:"enabled"`
	MaxSizeMB        int  `json:"max_size_mb"`
	DefaultTTL       int  `json:"default_ttl"`        // seconds
	CleanupInterval  int  `json:"cleanup_interval"`   // seconds
}

// RedisCacheConfig holds Redis cache configuration
type RedisCacheConfig struct {
	Enabled      bool     `json:"enabled"`
	Addresses    []string `json:"addresses"`
	Password     string   `json:"password"`
	Database     int      `json:"database"`
	MaxRetries   int      `json:"max_retries"`
	PoolSize     int      `json:"pool_size"`
	DefaultTTL   int      `json:"default_ttl"` // seconds
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret     string             `json:"jwt_secret"`
	JWTIssuer     string             `json:"jwt_issuer"`
	RateLimiting  RateLimitingConfig `json:"rate_limiting"`
	AdminRoles    []string           `json:"admin_roles"`
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	PerIPRequestsPerMinute   int `json:"per_ip_requests_per_minute"`
	PerUserRequestsPerMinute int `json:"per_user_requests_per_minute"`
	GlobalRequestsPerMinute  int `json:"global_requests_per_minute"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
	Output string `json:"output"` // stdout, stderr, file path
}

// Load loads configuration from a JSON file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	cfg.setDefaults()

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		c.Database.Password = dbPassword
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		c.Security.JWTSecret = jwtSecret
	}
	if encryptionKey := os.Getenv("DB_ENCRYPTION_KEY"); encryptionKey != "" {
		c.Database.EncryptionKey = encryptionKey
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		c.Cache.Redis.Password = redisPassword
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Service validation
	if c.Service.Port < 1 || c.Service.Port > 65535 {
		return fmt.Errorf("invalid service port: %d", c.Service.Port)
	}

	if len(c.Service.PortRange) == 2 {
		if c.Service.PortRange[0] > c.Service.PortRange[1] {
			return fmt.Errorf("invalid port range: %v", c.Service.PortRange)
		}
	}

	// Database validation
	if c.Database.Driver != "postgres" && c.Database.Driver != "sqlite3" {
		return fmt.Errorf("unsupported database driver: %s", c.Database.Driver)
	}

	if c.Database.Driver == "postgres" {
		if c.Database.Host == "" {
			return fmt.Errorf("database host is required for postgres")
		}
		if c.Database.Database == "" {
			return fmt.Errorf("database name is required")
		}
		if c.Database.EncryptionKey == "" {
			return fmt.Errorf("database encryption key is required for postgres")
		}
	}

	// Security validation
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	// Redis validation
	if c.Cache.Redis.Enabled {
		if len(c.Cache.Redis.Addresses) == 0 {
			return fmt.Errorf("redis addresses required when redis is enabled")
		}
	}

	return nil
}

// setDefaults sets default values for optional fields
func (c *Config) setDefaults() {
	// Service defaults
	if c.Service.Name == "" {
		c.Service.Name = "localization-service"
	}
	if c.Service.Environment == "" {
		c.Service.Environment = "development"
	}
	if c.Service.ReadTimeout == 0 {
		c.Service.ReadTimeout = 30
	}
	if c.Service.WriteTimeout == 0 {
		c.Service.WriteTimeout = 30
	}
	if c.Service.MaxHeaderBytes == 0 {
		c.Service.MaxHeaderBytes = 8192 // 8 KB
	}

	// Database defaults
	if c.Database.Port == 0 {
		if c.Database.Driver == "postgres" {
			c.Database.Port = 5432
		}
	}
	if c.Database.MaxConnections == 0 {
		c.Database.MaxConnections = 50
	}
	if c.Database.IdleConnections == 0 {
		c.Database.IdleConnections = 10
	}
	if c.Database.ConnectionTimeout == 0 {
		c.Database.ConnectionTimeout = 30
	}
	if c.Database.ConnectionLifetime == 0 {
		c.Database.ConnectionLifetime = 3600
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}

	// In-Memory Cache defaults
	if c.Cache.InMemory.Enabled {
		if c.Cache.InMemory.MaxSizeMB == 0 {
			c.Cache.InMemory.MaxSizeMB = 1024 // 1 GB
		}
		if c.Cache.InMemory.DefaultTTL == 0 {
			c.Cache.InMemory.DefaultTTL = 3600 // 1 hour
		}
		if c.Cache.InMemory.CleanupInterval == 0 {
			c.Cache.InMemory.CleanupInterval = 300 // 5 minutes
		}
	}

	// Redis Cache defaults
	if c.Cache.Redis.Enabled {
		if c.Cache.Redis.MaxRetries == 0 {
			c.Cache.Redis.MaxRetries = 3
		}
		if c.Cache.Redis.PoolSize == 0 {
			c.Cache.Redis.PoolSize = 10
		}
		if c.Cache.Redis.DefaultTTL == 0 {
			c.Cache.Redis.DefaultTTL = 14400 // 4 hours
		}
	}

	// Security defaults
	if c.Security.JWTIssuer == "" {
		c.Security.JWTIssuer = "helixtrack-auth"
	}
	if c.Security.RateLimiting.PerIPRequestsPerMinute == 0 {
		c.Security.RateLimiting.PerIPRequestsPerMinute = 1000
	}
	if c.Security.RateLimiting.PerUserRequestsPerMinute == 0 {
		c.Security.RateLimiting.PerUserRequestsPerMinute = 5000
	}
	if c.Security.RateLimiting.GlobalRequestsPerMinute == 0 {
		c.Security.RateLimiting.GlobalRequestsPerMinute = 100000
	}
	if len(c.Security.AdminRoles) == 0 {
		c.Security.AdminRoles = []string{"admin", "superadmin"}
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
			c.Host,
			c.Port,
			c.User,
			c.Password,
			c.Database,
			c.SSLMode,
			c.ConnectionTimeout,
		)
		// Add encryption key parameter if provided
		if c.EncryptionKey != "" {
			dsn += fmt.Sprintf(" options='-c session_preload_libraries=pgcrypto -c encryption_key=%s'", c.EncryptionKey)
		}
		return dsn
	case "sqlite3":
		return c.Database // file path
	default:
		return ""
	}
}
