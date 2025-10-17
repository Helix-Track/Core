package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"helixtrack.ru/chat/internal/models"
)

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*models.Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Expand environment variables
	expandEnvVars(&config)

	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// expandEnvVars expands environment variables in config values
func expandEnvVars(config *models.Config) {
	// Database password
	if strings.HasPrefix(config.Database.Password, "${") && strings.HasSuffix(config.Database.Password, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(config.Database.Password, "${"), "}")
		config.Database.Password = os.Getenv(envVar)
	}

	// JWT secret
	if strings.HasPrefix(config.JWT.Secret, "${") && strings.HasSuffix(config.JWT.Secret, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(config.JWT.Secret, "${"), "}")
		config.JWT.Secret = os.Getenv(envVar)
	}

	// Cert file
	if strings.HasPrefix(config.Server.CertFile, "${") && strings.HasSuffix(config.Server.CertFile, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(config.Server.CertFile, "${"), "}")
		config.Server.CertFile = os.Getenv(envVar)
	}

	// Key file
	if strings.HasPrefix(config.Server.KeyFile, "${") && strings.HasSuffix(config.Server.KeyFile, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(config.Server.KeyFile, "${"), "}")
		config.Server.KeyFile = os.Getenv(envVar)
	}
}

// validateConfig validates the configuration
func validateConfig(config *models.Config) error {
	// Server validation
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Server.HTTPS {
		if config.Server.CertFile == "" {
			return fmt.Errorf("cert_file is required when HTTPS is enabled")
		}
		if config.Server.KeyFile == "" {
			return fmt.Errorf("key_file is required when HTTPS is enabled")
		}
	}

	// Database validation
	if config.Database.Type != "postgres" && config.Database.Type != "postgresql" {
		return fmt.Errorf("unsupported database type: %s (only postgres supported)", config.Database.Type)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if config.Database.MaxConnections < 1 {
		config.Database.MaxConnections = 25 // Default
	}

	if config.Database.ConnectionTimeout < 1 {
		config.Database.ConnectionTimeout = 30 // Default
	}

	// JWT validation
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if config.JWT.ExpiryHours < 1 {
		config.JWT.ExpiryHours = 24 // Default
	}

	// Logger validation
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[config.Logger.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", config.Logger.Level)
	}

	if config.Logger.LogPath == "" {
		config.Logger.LogPath = "/tmp/htChatLogs"
	}

	if config.Logger.LogfileBaseName == "" {
		config.Logger.LogfileBaseName = "htChat"
	}

	if config.Logger.LogSizeLimit < 1 {
		config.Logger.LogSizeLimit = 100000000 // Default 100MB
	}

	// Security validation
	if config.Security.RateLimitPerSecond < 1 {
		config.Security.RateLimitPerSecond = 100 // Default
	}

	if config.Security.RateLimitBurst < 1 {
		config.Security.RateLimitBurst = 200 // Default
	}

	if config.Security.MaxMessageSize < 1 {
		config.Security.MaxMessageSize = 524288 // Default 512KB
	}

	if config.Security.MaxAttachmentSize < 1 {
		config.Security.MaxAttachmentSize = 104857600 // Default 100MB
	}

	if len(config.Security.AllowedOrigins) == 0 {
		config.Security.AllowedOrigins = []string{"*"} // Default allow all
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *models.Config {
	return &models.Config{
		Server: models.ServerConfig{
			Address:        "0.0.0.0",
			Port:           9090,
			HTTPS:          false,
			EnableHTTP3:    false,
			ReadTimeout:    30,
			WriteTimeout:   30,
			MaxHeaderBytes: 1048576, // 1MB
		},
		Database: models.DatabaseConfig{
			Type:              "postgres",
			Host:              "localhost",
			Port:              5432,
			Database:          "helixtrack_chat",
			User:              "chat_user",
			Password:          "password",
			SSLMode:           "disable",
			MaxConnections:    25,
			ConnectionTimeout: 30,
		},
		JWT: models.JWTConfig{
			Secret:      "change-this-secret-key",
			Issuer:      "helixtrack-chat",
			Audience:    "helixtrack",
			ExpiryHours: 24,
		},
		Logger: models.LoggerConfig{
			LogPath:         "/tmp/htChatLogs",
			LogfileBaseName: "htChat",
			LogSizeLimit:    100000000, // 100MB
			Level:           "info",
		},
		Security: models.SecurityConfig{
			EnableDDOSProtection: true,
			RateLimitPerSecond:   100,
			RateLimitBurst:       200,
			MaxMessageSize:       524288,      // 512KB
			MaxAttachmentSize:    104857600,   // 100MB
			AllowedOrigins:       []string{"*"},
		},
	}
}
