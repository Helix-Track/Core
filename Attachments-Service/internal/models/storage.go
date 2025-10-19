package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// StorageEndpoint represents a storage endpoint configuration
type StorageEndpoint struct {
	ID            string                 `json:"id" db:"id"`
	Name          string                 `json:"name" db:"name"`
	Type          string                 `json:"type" db:"type"` // local, s3, minio, azure, gcs, custom
	Role          string                 `json:"role" db:"role"` // primary, backup, mirror
	AdapterConfig map[string]interface{} `json:"adapter_config" db:"adapter_config"`
	Priority      int                    `json:"priority" db:"priority"`
	Enabled       bool                   `json:"enabled" db:"enabled"`
	MaxSizeBytes  *int64                 `json:"max_size_bytes,omitempty" db:"max_size_bytes"`
	CurrentSize   int64                  `json:"current_size" db:"current_size"`
	Created       int64                  `json:"created" db:"created"`
	Modified      int64                  `json:"modified" db:"modified"`
}

// Storage endpoint types
const (
	StorageTypeLocal  = "local"
	StorageTypeS3     = "s3"
	StorageTypeMinIO  = "minio"
	StorageTypeAzure  = "azure"
	StorageTypeGCS    = "gcs"
	StorageTypeCustom = "custom"
)

// Storage endpoint roles
const (
	RolePrimary = "primary"
	RoleBackup  = "backup"
	RoleMirror  = "mirror"
)

// NewStorageEndpoint creates a new storage endpoint
func NewStorageEndpoint(id, name, endpointType, role string, config map[string]interface{}) *StorageEndpoint {
	now := time.Now().Unix()
	return &StorageEndpoint{
		ID:            id,
		Name:          name,
		Type:          endpointType,
		Role:          role,
		AdapterConfig: config,
		Priority:      1,
		Enabled:       true,
		CurrentSize:   0,
		Created:       now,
		Modified:      now,
	}
}

// Validate validates the storage endpoint
func (e *StorageEndpoint) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("id is required")
	}
	if e.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !isValidStorageType(e.Type) {
		return fmt.Errorf("invalid storage type: %s", e.Type)
	}
	if !isValidRole(e.Role) {
		return fmt.Errorf("invalid role: %s", e.Role)
	}
	if e.AdapterConfig == nil {
		return fmt.Errorf("adapter_config is required")
	}
	if e.Priority < 1 {
		return fmt.Errorf("priority must be >= 1")
	}
	if e.CurrentSize < 0 {
		return fmt.Errorf("current_size must be non-negative")
	}
	if e.MaxSizeBytes != nil && *e.MaxSizeBytes <= 0 {
		return fmt.Errorf("max_size_bytes must be positive")
	}
	if e.Created == 0 {
		return fmt.Errorf("created timestamp is required")
	}
	return nil
}

// GetAdapterConfigJSON returns adapter config as JSON string
func (e *StorageEndpoint) GetAdapterConfigJSON() (string, error) {
	data, err := json.Marshal(e.AdapterConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal adapter config: %w", err)
	}
	return string(data), nil
}

// SetAdapterConfigFromJSON sets adapter config from JSON string
func (e *StorageEndpoint) SetAdapterConfigFromJSON(jsonStr string) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return fmt.Errorf("failed to unmarshal adapter config: %w", err)
	}
	e.AdapterConfig = config
	return nil
}

// IsNearCapacity checks if endpoint is near capacity (>90%)
func (e *StorageEndpoint) IsNearCapacity() bool {
	if e.MaxSizeBytes == nil {
		return false
	}
	threshold := float64(*e.MaxSizeBytes) * 0.9
	return float64(e.CurrentSize) >= threshold
}

// GetUsagePercent returns storage usage percentage
func (e *StorageEndpoint) GetUsagePercent() float64 {
	if e.MaxSizeBytes == nil || *e.MaxSizeBytes == 0 {
		return 0
	}
	return (float64(e.CurrentSize) / float64(*e.MaxSizeBytes)) * 100
}

// UpdateSize updates the current size
func (e *StorageEndpoint) UpdateSize(delta int64) {
	e.CurrentSize += delta
	if e.CurrentSize < 0 {
		e.CurrentSize = 0
	}
	e.Modified = time.Now().Unix()
}

// Enable enables the endpoint
func (e *StorageEndpoint) Enable() {
	e.Enabled = true
	e.Modified = time.Now().Unix()
}

// Disable disables the endpoint
func (e *StorageEndpoint) Disable() {
	e.Enabled = false
	e.Modified = time.Now().Unix()
}

// StorageHealth represents health check data for a storage endpoint
type StorageHealth struct {
	EndpointID     string  `json:"endpoint_id" db:"endpoint_id"`
	CheckTime      int64   `json:"check_time" db:"check_time"`
	Status         string  `json:"status" db:"status"` // healthy, degraded, unhealthy
	LatencyMs      *int    `json:"latency_ms,omitempty" db:"latency_ms"`
	ErrorMessage   *string `json:"error_message,omitempty" db:"error_message"`
	AvailableBytes *int64  `json:"available_bytes,omitempty" db:"available_bytes"`
}

// Health status constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
)

// NewStorageHealth creates a new storage health record
func NewStorageHealth(endpointID, status string) *StorageHealth {
	return &StorageHealth{
		EndpointID: endpointID,
		CheckTime:  time.Now().Unix(),
		Status:     status,
	}
}

// Validate validates the storage health record
func (h *StorageHealth) Validate() error {
	if h.EndpointID == "" {
		return fmt.Errorf("endpoint_id is required")
	}
	if !isValidHealthStatus(h.Status) {
		return fmt.Errorf("invalid status: %s", h.Status)
	}
	if h.CheckTime == 0 {
		return fmt.Errorf("check_time is required")
	}
	if h.LatencyMs != nil && *h.LatencyMs < 0 {
		return fmt.Errorf("latency_ms must be non-negative")
	}
	if h.AvailableBytes != nil && *h.AvailableBytes < 0 {
		return fmt.Errorf("available_bytes must be non-negative")
	}
	return nil
}

// IsHealthy checks if the status is healthy
func (h *StorageHealth) IsHealthy() bool {
	return h.Status == HealthStatusHealthy
}

// Helper functions

func isValidStorageType(t string) bool {
	validTypes := []string{
		StorageTypeLocal, StorageTypeS3, StorageTypeMinIO,
		StorageTypeAzure, StorageTypeGCS, StorageTypeCustom,
	}
	for _, valid := range validTypes {
		if t == valid {
			return true
		}
	}
	return false
}

func isValidRole(r string) bool {
	validRoles := []string{RolePrimary, RoleBackup, RoleMirror}
	for _, valid := range validRoles {
		if r == valid {
			return true
		}
	}
	return false
}

func isValidHealthStatus(s string) bool {
	validStatuses := []string{
		HealthStatusHealthy, HealthStatusDegraded, HealthStatusUnhealthy,
	}
	for _, valid := range validStatuses {
		if s == valid {
			return true
		}
	}
	return false
}
