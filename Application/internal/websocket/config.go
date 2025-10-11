package websocket

import (
	"time"

	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/models"
)

// ConfigToModel converts config.WebSocketConfig to models.WebSocketConfig
func ConfigToModel(cfg config.WebSocketConfig) models.WebSocketConfig {
	return models.WebSocketConfig{
		Enabled:           cfg.Enabled,
		Path:              cfg.Path,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		MaxMessageSize:    cfg.MaxMessageSize,
		WriteWait:         time.Duration(cfg.WriteWaitSeconds) * time.Second,
		PongWait:          time.Duration(cfg.PongWaitSeconds) * time.Second,
		PingPeriod:        time.Duration(cfg.PingPeriodSeconds) * time.Second,
		MaxClients:        cfg.MaxClients,
		RequireAuth:       cfg.RequireAuth,
		AllowOrigins:      cfg.AllowOrigins,
		EnableCompression: cfg.EnableCompression,
		HandshakeTimeout:  time.Duration(cfg.HandshakeTimeout) * time.Second,
	}
}
