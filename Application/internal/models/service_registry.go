package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ServiceType represents the type of service
type ServiceType string

const (
	ServiceTypeAuthentication ServiceType = "authentication"
	ServiceTypePermissions    ServiceType = "permissions"
	ServiceTypeLokalization   ServiceType = "lokalisation"
	ServiceTypeExtension      ServiceType = "extension"
)

// ServiceRole represents the role of a service in failover scenarios
type ServiceRole string

const (
	ServiceRolePrimary ServiceRole = "primary"
	ServiceRoleBackup  ServiceRole = "backup"
)

// ServiceStatus represents the current status of a service
type ServiceStatus string

const (
	ServiceStatusHealthy      ServiceStatus = "healthy"
	ServiceStatusUnhealthy    ServiceStatus = "unhealthy"
	ServiceStatusRegistering  ServiceStatus = "registering"
	ServiceStatusRotating     ServiceStatus = "rotating"
	ServiceStatusDecommission ServiceStatus = "decommissioned"
)

// ServiceRegistration represents a registered service in the system
type ServiceRegistration struct {
	ID                string        `json:"id"`                  // Unique service ID (UUID)
	Name              string        `json:"name"`                // Service name
	Type              ServiceType   `json:"type"`                // Service type
	Version           string        `json:"version"`             // Service version
	URL               string        `json:"url"`                 // Service base URL
	HealthCheckURL    string        `json:"health_check_url"`    // Health check endpoint
	PublicKey         string        `json:"public_key"`          // RSA public key for verification
	Signature         string        `json:"signature"`           // Service metadata signature
	Certificate       string        `json:"certificate"`         // TLS certificate (PEM format)
	Status            ServiceStatus `json:"status"`              // Current status
	Role              ServiceRole   `json:"role"`                // Service role (primary/backup)
	FailoverGroup     string        `json:"failover_group"`      // Failover group identifier
	IsActive          bool          `json:"is_active"`           // Currently active service for its group
	Priority          int           `json:"priority"`            // Service priority (higher = preferred)
	Metadata          string        `json:"metadata"`            // JSON metadata
	RegisteredBy      string        `json:"registered_by"`       // Username who registered
	RegisteredAt      time.Time     `json:"registered_at"`       // Registration timestamp
	LastHealthCheck   time.Time     `json:"last_health_check"`   // Last health check timestamp
	HealthCheckCount  int           `json:"health_check_count"`  // Total health checks performed
	FailedHealthCount int           `json:"failed_health_count"` // Failed health check count
	LastFailoverAt    time.Time     `json:"last_failover_at"`    // Last failover timestamp
	Deleted           bool          `json:"deleted"`             // Soft delete flag
}

// ServiceHealthCheck represents a health check record
type ServiceHealthCheck struct {
	ID            string        `json:"id"`              // Unique check ID
	ServiceID     string        `json:"service_id"`      // Service being checked
	Timestamp     time.Time     `json:"timestamp"`       // Check timestamp
	Status        ServiceStatus `json:"status"`          // Health status result
	ResponseTime  int64         `json:"response_time"`   // Response time in milliseconds
	StatusCode    int           `json:"status_code"`     // HTTP status code
	ErrorMessage  string        `json:"error_message"`   // Error message if unhealthy
	CheckedBy     string        `json:"checked_by"`      // System/user performing check
}

// ServiceRotationRequest represents a request to rotate a service
type ServiceRotationRequest struct {
	CurrentServiceID string              `json:"current_service_id"` // Service to be replaced
	NewService       ServiceRegistration `json:"new_service"`        // New service to replace with
	Reason           string              `json:"reason"`             // Reason for rotation
	RequestedBy      string              `json:"requested_by"`       // User requesting rotation
	AdminToken       string              `json:"admin_token"`        // Admin authorization token
	VerificationCode string              `json:"verification_code"`  // Additional verification code
}

// ServiceRotationResponse represents the response to a rotation request
type ServiceRotationResponse struct {
	Success          bool      `json:"success"`
	OldServiceID     string    `json:"old_service_id"`
	NewServiceID     string    `json:"new_service_id"`
	RotationTime     time.Time `json:"rotation_time"`
	VerificationHash string    `json:"verification_hash"` // Hash for audit trail
	Message          string    `json:"message"`
}

// ServiceDiscoveryRequest represents a request to discover services
type ServiceDiscoveryRequest struct {
	Type       ServiceType `json:"type"`        // Type of service to discover
	MinVersion string      `json:"min_version"` // Minimum version required
	OnlyHealthy bool       `json:"only_healthy"` // Return only healthy services
}

// ServiceDiscoveryResponse represents the response with discovered services
type ServiceDiscoveryResponse struct {
	Services   []ServiceRegistration `json:"services"`
	TotalCount int                   `json:"total_count"`
	Timestamp  time.Time             `json:"timestamp"`
}

// ComputeServiceSignature computes a signature for service metadata
func (s *ServiceRegistration) ComputeServiceSignature() string {
	data := s.ID + s.Name + string(s.Type) + s.Version + s.URL + s.PublicKey
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifySignature verifies the service signature matches the computed signature
func (s *ServiceRegistration) VerifySignature() bool {
	expectedSignature := s.ComputeServiceSignature()
	return s.Signature == expectedSignature
}

// IsHealthy returns whether the service is currently healthy
func (s *ServiceRegistration) IsHealthy() bool {
	return s.Status == ServiceStatusHealthy
}

// CanRotate returns whether the service can be rotated
func (s *ServiceRegistration) CanRotate() bool {
	return s.Status != ServiceStatusRotating && s.Status != ServiceStatusDecommission
}

// ServiceFailoverEvent represents a failover event
type ServiceFailoverEvent struct {
	ID              string        `json:"id"`               // Unique event ID
	FailoverGroup   string        `json:"failover_group"`   // Failover group
	ServiceType     ServiceType   `json:"service_type"`     // Type of service
	OldServiceID    string        `json:"old_service_id"`   // Previous active service
	NewServiceID    string        `json:"new_service_id"`   // New active service
	FailoverReason  string        `json:"failover_reason"`  // Reason for failover
	FailoverType    string        `json:"failover_type"`    // "failover" or "failback"
	Timestamp       time.Time     `json:"timestamp"`        // When failover occurred
	Automatic       bool          `json:"automatic"`        // Was it automatic or manual
}

// ServiceRegistrationRequest represents a request to register a new service
type ServiceRegistrationRequest struct {
	Name           string      `json:"name"`
	Type           ServiceType `json:"type"`
	Version        string      `json:"version"`
	URL            string      `json:"url"`
	HealthCheckURL string      `json:"health_check_url"`
	PublicKey      string      `json:"public_key"`
	Certificate    string      `json:"certificate"`
	Role           ServiceRole `json:"role"`            // primary or backup
	FailoverGroup  string      `json:"failover_group"`  // Failover group (optional)
	Priority       int         `json:"priority"`
	Metadata       string      `json:"metadata"`
	AdminToken     string      `json:"admin_token"` // Required for registration
}

// ServiceUpdateRequest represents a request to update service metadata
type ServiceUpdateRequest struct {
	ServiceID      string `json:"service_id"`
	Version        string `json:"version,omitempty"`
	URL            string `json:"url,omitempty"`
	HealthCheckURL string `json:"health_check_url,omitempty"`
	Priority       int    `json:"priority,omitempty"`
	Metadata       string `json:"metadata,omitempty"`
	AdminToken     string `json:"admin_token"` // Required for update
}

// ServiceDecommissionRequest represents a request to decommission a service
type ServiceDecommissionRequest struct {
	ServiceID  string `json:"service_id"`
	Reason     string `json:"reason"`
	AdminToken string `json:"admin_token"` // Required for decommission
}
