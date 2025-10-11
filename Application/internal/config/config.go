package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	Log       LogConfig        `json:"log"`
	Listeners []ListenerConfig `json:"listeners"`
	Plugins   []PluginConfig   `json:"plugins"`
	Database  DatabaseConfig   `json:"database"`
	Services  ServicesConfig   `json:"services"`
	WebSocket WebSocketConfig  `json:"websocket"`
	Version   string           `json:"version,omitempty"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	LogPath         string `json:"log_path"`
	LogfileBaseName string `json:"logfile_base_name"`
	LogSizeLimit    int64  `json:"log_size_limit"`
	Level           string `json:"level,omitempty"` // debug, info, warn, error
}

// ListenerConfig represents HTTP listener configuration
type ListenerConfig struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	HTTPS    bool   `json:"https"`
	CertFile string `json:"cert_file,omitempty"`
	KeyFile  string `json:"key_file,omitempty"`
}

// PluginConfig represents plugin configuration
type PluginConfig struct {
	Name         string                 `json:"name"`
	Dependencies []string               `json:"dependencies"`
	Config       map[string]interface{} `json:"config"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type             string `json:"type"` // sqlite or postgres
	SQLitePath       string `json:"sqlite_path,omitempty"`
	PostgresHost     string `json:"postgres_host,omitempty"`
	PostgresPort     int    `json:"postgres_port,omitempty"`
	PostgresUser     string `json:"postgres_user,omitempty"`
	PostgresPassword string `json:"postgres_password,omitempty"`
	PostgresDatabase string `json:"postgres_database,omitempty"`
	PostgresSSLMode  string `json:"postgres_ssl_mode,omitempty"`
}

// ServicesConfig represents external services configuration
type ServicesConfig struct {
	Authentication ServiceEndpoint            `json:"authentication"`
	Permissions    ServiceEndpoint            `json:"permissions"`
	Lokalisation   *ServiceEndpoint           `json:"lokalisation,omitempty"`
	Extensions     map[string]ServiceEndpoint `json:"extensions,omitempty"`
}

// ServiceEndpoint represents an external service endpoint
type ServiceEndpoint struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Timeout int    `json:"timeout,omitempty"` // in seconds
}

// WebSocketConfig represents WebSocket configuration
type WebSocketConfig struct {
	Enabled           bool     `json:"enabled"`
	Path              string   `json:"path"`
	ReadBufferSize    int      `json:"readBufferSize"`
	WriteBufferSize   int      `json:"writeBufferSize"`
	MaxMessageSize    int64    `json:"maxMessageSize"`
	WriteWaitSeconds  int      `json:"writeWaitSeconds"`
	PongWaitSeconds   int      `json:"pongWaitSeconds"`
	PingPeriodSeconds int      `json:"pingPeriodSeconds"`
	MaxClients        int      `json:"maxClients"`
	RequireAuth       bool     `json:"requireAuth"`
	AllowOrigins      []string `json:"allowOrigins"`
	EnableCompression bool     `json:"enableCompression"`
	HandshakeTimeout  int      `json:"handshakeTimeout"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	config.applyDefaults()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// applyDefaults applies default values to missing configuration
func (c *Config) applyDefaults() {
	if c.Log.LogPath == "" {
		c.Log.LogPath = "/tmp/htCoreLogs"
	}
	if c.Log.LogSizeLimit == 0 {
		c.Log.LogSizeLimit = 100000000 // 100MB
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}

	if c.Database.Type == "" {
		c.Database.Type = "sqlite"
	}
	if c.Database.Type == "sqlite" && c.Database.SQLitePath == "" {
		c.Database.SQLitePath = "Database/Definition.sqlite"
	}
	if c.Database.Type == "postgres" && c.Database.PostgresSSLMode == "" {
		c.Database.PostgresSSLMode = "disable"
	}

	// Set default timeouts for services
	if c.Services.Authentication.Timeout == 0 {
		c.Services.Authentication.Timeout = 30
	}
	if c.Services.Permissions.Timeout == 0 {
		c.Services.Permissions.Timeout = 30
	}
	if c.Services.Lokalisation != nil && c.Services.Lokalisation.Timeout == 0 {
		c.Services.Lokalisation.Timeout = 30
	}
	for name, ext := range c.Services.Extensions {
		if ext.Timeout == 0 {
			ext.Timeout = 30
			c.Services.Extensions[name] = ext
		}
	}

	// Set default WebSocket configuration
	if c.WebSocket.Path == "" {
		c.WebSocket.Path = "/ws"
	}
	if c.WebSocket.ReadBufferSize == 0 {
		c.WebSocket.ReadBufferSize = 1024
	}
	if c.WebSocket.WriteBufferSize == 0 {
		c.WebSocket.WriteBufferSize = 1024
	}
	if c.WebSocket.MaxMessageSize == 0 {
		c.WebSocket.MaxMessageSize = 512 * 1024 // 512KB
	}
	if c.WebSocket.WriteWaitSeconds == 0 {
		c.WebSocket.WriteWaitSeconds = 10
	}
	if c.WebSocket.PongWaitSeconds == 0 {
		c.WebSocket.PongWaitSeconds = 60
	}
	if c.WebSocket.PingPeriodSeconds == 0 {
		c.WebSocket.PingPeriodSeconds = 54 // Must be less than pongWait
	}
	if c.WebSocket.MaxClients == 0 {
		c.WebSocket.MaxClients = 1000
	}
	if c.WebSocket.HandshakeTimeout == 0 {
		c.WebSocket.HandshakeTimeout = 10
	}
	if len(c.WebSocket.AllowOrigins) == 0 {
		c.WebSocket.AllowOrigins = []string{"*"}
	}
	// RequireAuth defaults to true, Enabled defaults to false
	// No need to set defaults for these booleans
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Listeners) == 0 {
		return fmt.Errorf("at least one listener must be configured")
	}

	for i, listener := range c.Listeners {
		if listener.Address == "" {
			return fmt.Errorf("listener %d: address is required", i)
		}
		if listener.Port <= 0 || listener.Port > 65535 {
			return fmt.Errorf("listener %d: invalid port %d", i, listener.Port)
		}
		if listener.HTTPS {
			if listener.CertFile == "" {
				return fmt.Errorf("listener %d: cert_file is required for HTTPS", i)
			}
			if listener.KeyFile == "" {
				return fmt.Errorf("listener %d: key_file is required for HTTPS", i)
			}
		}
	}

	if c.Database.Type != "sqlite" && c.Database.Type != "postgres" {
		return fmt.Errorf("database type must be 'sqlite' or 'postgres', got '%s'", c.Database.Type)
	}

	if c.Database.Type == "sqlite" && c.Database.SQLitePath == "" {
		return fmt.Errorf("sqlite_path is required when using sqlite database")
	}

	if c.Database.Type == "postgres" {
		if c.Database.PostgresHost == "" {
			return fmt.Errorf("postgres_host is required when using postgres database")
		}
		if c.Database.PostgresPort <= 0 || c.Database.PostgresPort > 65535 {
			return fmt.Errorf("invalid postgres_port: %d", c.Database.PostgresPort)
		}
		if c.Database.PostgresUser == "" {
			return fmt.Errorf("postgres_user is required when using postgres database")
		}
		if c.Database.PostgresDatabase == "" {
			return fmt.Errorf("postgres_database is required when using postgres database")
		}
	}

	// Validate service endpoints
	if c.Services.Authentication.Enabled && c.Services.Authentication.URL == "" {
		return fmt.Errorf("authentication service URL is required when enabled")
	}
	if c.Services.Permissions.Enabled && c.Services.Permissions.URL == "" {
		return fmt.Errorf("permissions service URL is required when enabled")
	}

	return nil
}

// GetPrimaryListener returns the first configured listener
func (c *Config) GetPrimaryListener() *ListenerConfig {
	if len(c.Listeners) > 0 {
		return &c.Listeners[0]
	}
	return nil
}

// GetListenerAddress returns the full address of the primary listener
func (c *Config) GetListenerAddress() string {
	listener := c.GetPrimaryListener()
	if listener == nil {
		return ":8080"
	}
	return fmt.Sprintf("%s:%d", listener.Address, listener.Port)
}

// GetWebSocketConfig converts config WebSocketConfig to models.WebSocketConfig
func (c *Config) GetWebSocketConfig() WebSocketConfig {
	return c.WebSocket
}

// IsWebSocketEnabled returns whether WebSocket is enabled
func (c *Config) IsWebSocketEnabled() bool {
	return c.WebSocket.Enabled
}
