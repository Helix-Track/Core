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
	Storage  StorageConfig  `json:"storage"`
	Security SecurityConfig `json:"security"`
	Logging  LoggingConfig  `json:"logging"`
	Metrics  MetricsConfig  `json:"metrics"`
}

// ServiceConfig holds service-level configuration
type ServiceConfig struct {
	Name           string          `json:"name"`
	Port           int             `json:"port"`
	PortRange      []int           `json:"port_range"`
	Environment    string          `json:"environment"` // development, staging, production
	ReadTimeout    int             `json:"read_timeout"`  // seconds
	WriteTimeout   int             `json:"write_timeout"` // seconds
	MaxHeaderBytes int             `json:"max_header_bytes"`
	Discovery      DiscoveryConfig `json:"discovery"`
}

// DiscoveryConfig holds service discovery configuration
type DiscoveryConfig struct {
	Enabled       bool   `json:"enabled"`
	Provider      string `json:"provider"` // consul, etcd, etc.
	ConsulAddress string `json:"consul_address"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver             string `json:"driver"` // postgres, sqlite3
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Database           string `json:"database"`
	User               string `json:"user"`
	Password           string `json:"password"`
	SSLMode            string `json:"ssl_mode"`
	MaxConnections     int    `json:"max_connections"`
	IdleConnections    int    `json:"idle_connections"`
	ConnectionTimeout  int    `json:"connection_timeout"` // seconds
	ConnectionLifetime int    `json:"connection_lifetime"` // seconds
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Endpoints       []StorageEndpoint `json:"endpoints"`
	ReplicationMode string            `json:"replication_mode"` // synchronous, asynchronous, hybrid
	Cleanup         CleanupConfig     `json:"cleanup"`
}

// StorageEndpoint defines a storage endpoint
type StorageEndpoint struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"` // local, s3, minio, custom
	Role          string                 `json:"role"` // primary, backup, mirror
	Priority      int                    `json:"priority"`
	Enabled       bool                   `json:"enabled"`
	MaxSizeGB     int                    `json:"max_size_gb"`
	AdapterConfig map[string]interface{} `json:"adapter_config"`
}

// CleanupConfig holds cleanup job configuration
type CleanupConfig struct {
	OrphanRetentionDays int    `json:"orphan_retention_days"`
	JobSchedule         string `json:"job_schedule"` // cron format
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret        string              `json:"jwt_secret"`
	JWTIssuer        string              `json:"jwt_issuer"`
	AllowedMimeTypes []string            `json:"allowed_mime_types"`
	MaxFileSizeMB    int64               `json:"max_file_size_mb"`
	VirusScanning    VirusScanningConfig `json:"virus_scanning"`
	RateLimiting     RateLimitingConfig  `json:"rate_limiting"`
	ImageValidation  ImageValidationConfig `json:"image_validation"`
}

// VirusScanningConfig holds virus scanning configuration
type VirusScanningConfig struct {
	Enabled        bool   `json:"enabled"`
	ClamdSocket    string `json:"clamd_socket"`
	ClamdHost      string `json:"clamd_host"`
	ClamdPort      int    `json:"clamd_port"`
	MaxScanSizeMB  int64  `json:"max_scan_size_mb"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	PerIPRequestsPerMinute    int `json:"per_ip_requests_per_minute"`
	PerUserRequestsPerMinute  int `json:"per_user_requests_per_minute"`
	PerIPUploadsPerMinute     int `json:"per_ip_uploads_per_minute"`
	PerUserUploadsPerMinute   int `json:"per_user_uploads_per_minute"`
	GlobalRequestsPerMinute   int `json:"global_requests_per_minute"`
	MaxConnectionsPerIP       int `json:"max_connections_per_ip"`
	MaxConnectionsGlobal      int `json:"max_connections_global"`
}

// ImageValidationConfig holds image-specific validation configuration
type ImageValidationConfig struct {
	MaxWidthPixels  int  `json:"max_width_pixels"`
	MaxHeightPixels int  `json:"max_height_pixels"`
	AutoCompress    bool `json:"auto_compress"`
	CompressionQuality int `json:"compression_quality"` // 1-100
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
	Output string `json:"output"` // stdout, stderr, file path
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled        bool   `json:"enabled"`
	PrometheusPort int    `json:"prometheus_port"`
	Path           string `json:"path"`
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
	}

	// Storage validation
	if len(c.Storage.Endpoints) == 0 {
		return fmt.Errorf("at least one storage endpoint is required")
	}

	primaryCount := 0
	for _, endpoint := range c.Storage.Endpoints {
		if endpoint.Role == "primary" {
			primaryCount++
		}
	}
	if primaryCount == 0 {
		return fmt.Errorf("at least one primary storage endpoint is required")
	}
	if primaryCount > 1 {
		return fmt.Errorf("only one primary storage endpoint is allowed")
	}

	// Security validation
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.Security.MaxFileSizeMB <= 0 {
		return fmt.Errorf("max file size must be positive")
	}

	if len(c.Security.AllowedMimeTypes) == 0 {
		return fmt.Errorf("at least one allowed MIME type is required")
	}

	return nil
}

// setDefaults sets default values for optional fields
func (c *Config) setDefaults() {
	// Service defaults
	if c.Service.Name == "" {
		c.Service.Name = "attachments-service"
	}
	if c.Service.Environment == "" {
		c.Service.Environment = "development"
	}
	if c.Service.ReadTimeout == 0 {
		c.Service.ReadTimeout = 60
	}
	if c.Service.WriteTimeout == 0 {
		c.Service.WriteTimeout = 300
	}
	if c.Service.MaxHeaderBytes == 0 {
		c.Service.MaxHeaderBytes = 8192 // 8 KB
	}

	// Database defaults
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

	// Storage defaults
	if c.Storage.ReplicationMode == "" {
		c.Storage.ReplicationMode = "hybrid"
	}
	if c.Storage.Cleanup.OrphanRetentionDays == 0 {
		c.Storage.Cleanup.OrphanRetentionDays = 30
	}
	if c.Storage.Cleanup.JobSchedule == "" {
		c.Storage.Cleanup.JobSchedule = "0 2 * * *" // 2 AM daily
	}

	// Security defaults
	if c.Security.JWTIssuer == "" {
		c.Security.JWTIssuer = "helixtrack-auth"
	}
	if c.Security.MaxFileSizeMB == 0 {
		c.Security.MaxFileSizeMB = 100
	}
	if c.Security.VirusScanning.MaxScanSizeMB == 0 {
		c.Security.VirusScanning.MaxScanSizeMB = 100
	}
	if c.Security.VirusScanning.TimeoutSeconds == 0 {
		c.Security.VirusScanning.TimeoutSeconds = 60
	}
	if c.Security.RateLimiting.PerIPRequestsPerMinute == 0 {
		c.Security.RateLimiting.PerIPRequestsPerMinute = 100
	}
	if c.Security.RateLimiting.PerUserRequestsPerMinute == 0 {
		c.Security.RateLimiting.PerUserRequestsPerMinute = 1000
	}
	if c.Security.RateLimiting.PerIPUploadsPerMinute == 0 {
		c.Security.RateLimiting.PerIPUploadsPerMinute = 10
	}
	if c.Security.RateLimiting.PerUserUploadsPerMinute == 0 {
		c.Security.RateLimiting.PerUserUploadsPerMinute = 100
	}
	if c.Security.RateLimiting.GlobalRequestsPerMinute == 0 {
		c.Security.RateLimiting.GlobalRequestsPerMinute = 10000
	}
	if c.Security.RateLimiting.MaxConnectionsPerIP == 0 {
		c.Security.RateLimiting.MaxConnectionsPerIP = 100
	}
	if c.Security.RateLimiting.MaxConnectionsGlobal == 0 {
		c.Security.RateLimiting.MaxConnectionsGlobal = 1000
	}
	if c.Security.ImageValidation.MaxWidthPixels == 0 {
		c.Security.ImageValidation.MaxWidthPixels = 10000
	}
	if c.Security.ImageValidation.MaxHeightPixels == 0 {
		c.Security.ImageValidation.MaxHeightPixels = 10000
	}
	if c.Security.ImageValidation.CompressionQuality == 0 {
		c.Security.ImageValidation.CompressionQuality = 85
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

	// Metrics defaults
	if c.Metrics.PrometheusPort == 0 {
		c.Metrics.PrometheusPort = 9090
	}
	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
			c.Host,
			c.Port,
			c.User,
			c.Password,
			c.Database,
			c.SSLMode,
			c.ConnectionTimeout,
		)
	case "sqlite3":
		return c.Database // file path
	default:
		return ""
	}
}
