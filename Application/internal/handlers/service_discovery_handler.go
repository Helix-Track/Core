package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/security"
	"helixtrack.ru/core/internal/services"
)

// ServiceDiscoveryHandler handles service discovery operations
type ServiceDiscoveryHandler struct {
	db            database.Database
	signer        *security.ServiceSigner
	healthChecker *services.HealthChecker
}

// NewServiceDiscoveryHandler creates a new service discovery handler
func NewServiceDiscoveryHandler(db database.Database) (*ServiceDiscoveryHandler, error) {
	signer, err := security.NewServiceSigner()
	if err != nil {
		return nil, fmt.Errorf("failed to create service signer: %w", err)
	}

	healthChecker := services.NewHealthChecker(db, 1*time.Minute, 10*time.Second)

	return &ServiceDiscoveryHandler{
		db:            db,
		signer:        signer,
		healthChecker: healthChecker,
	}, nil
}

// StartHealthChecker starts the health checker background process
func (h *ServiceDiscoveryHandler) StartHealthChecker() error {
	return h.healthChecker.Start()
}

// StopHealthChecker stops the health checker background process
func (h *ServiceDiscoveryHandler) StopHealthChecker() {
	h.healthChecker.Stop()
}

// RegisterService handles service registration
func (h *ServiceDiscoveryHandler) RegisterService(c *gin.Context) {
	var req models.ServiceRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid registration request",
			"",
		))
		return
	}

	// Verify admin token
	if req.AdminToken == "" || len(req.AdminToken) < 32 {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid admin token",
			"",
		))
		return
	}

	// Get requesting user from context
	username, exists := c.Get("username")
	if !exists {
		username = "system"
	}

	// Create service registration
	service := &models.ServiceRegistration{
		ID:                uuid.New().String(),
		Name:              req.Name,
		Type:              req.Type,
		Version:           req.Version,
		URL:               req.URL,
		HealthCheckURL:    req.HealthCheckURL,
		PublicKey:         req.PublicKey,
		Certificate:       req.Certificate,
		Status:            models.ServiceStatusRegistering,
		Priority:          req.Priority,
		Metadata:          req.Metadata,
		RegisteredBy:      username.(string),
		RegisteredAt:      time.Now(),
		LastHealthCheck:   time.Time{},
		HealthCheckCount:  0,
		FailedHealthCount: 0,
		Deleted:           false,
	}

	// Sign the service registration
	if err := h.signer.SignServiceRegistration(service); err != nil {
		logger.Error("Failed to sign service registration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to sign service registration",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO service_registry (
			id, name, type, version, url, health_check_url, public_key, signature, certificate,
			status, priority, metadata, registered_by, registered_at, last_health_check,
			health_check_count, failed_health_count, deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(context.Background(), query,
		service.ID,
		service.Name,
		service.Type,
		service.Version,
		service.URL,
		service.HealthCheckURL,
		service.PublicKey,
		service.Signature,
		service.Certificate,
		service.Status,
		service.Priority,
		service.Metadata,
		service.RegisteredBy,
		service.RegisteredAt.Unix(),
		0, // last_health_check
		0, // health_check_count
		0, // failed_health_count
		0, // deleted
	)

	if err != nil {
		logger.Error("Failed to insert service registration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to register service",
			"",
		))
		return
	}

	// Perform immediate health check
	go h.healthChecker.CheckServiceNow(service.ID)

	logger.Info("Service registered successfully",
		zap.String("service_id", service.ID),
		zap.String("name", service.Name),
		zap.String("type", string(service.Type)),
		zap.String("url", service.URL),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"service": service,
	}))
}

// DiscoverServices handles service discovery requests
func (h *ServiceDiscoveryHandler) DiscoverServices(c *gin.Context) {
	var req models.ServiceDiscoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid discovery request",
			"",
		))
		return
	}

	ctx := context.Background()

	// Build query based on request
	query := `
		SELECT id, name, type, version, url, health_check_url, public_key, signature, certificate,
		       status, priority, metadata, registered_by, registered_at, last_health_check,
		       health_check_count, failed_health_count
		FROM service_registry
		WHERE deleted = 0
	`

	args := []interface{}{}

	if req.Type != "" {
		query += " AND type = ?"
		args = append(args, req.Type)
	}

	if req.OnlyHealthy {
		query += " AND status = ?"
		args = append(args, models.ServiceStatusHealthy)
	}

	// Order by priority (higher first) and then by health check count (more reliable)
	query += " ORDER BY priority DESC, health_check_count DESC"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to query services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to discover services",
			"",
		))
		return
	}
	defer rows.Close()

	var services []models.ServiceRegistration

	for rows.Next() {
		var service models.ServiceRegistration
		var registeredAt int64
		var lastHealthCheck sql.NullInt64
		var healthCheckURL, publicKey, signature, certificate, metadata sql.NullString

		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.Type,
			&service.Version,
			&service.URL,
			&healthCheckURL,
			&publicKey,
			&signature,
			&certificate,
			&service.Status,
			&service.Priority,
			&metadata,
			&service.RegisteredBy,
			&registeredAt,
			&lastHealthCheck,
			&service.HealthCheckCount,
			&service.FailedHealthCount,
		)

		if err != nil {
			logger.Error("Failed to scan service row", zap.Error(err))
			continue
		}

		// Convert nullable strings
		if healthCheckURL.Valid {
			service.HealthCheckURL = healthCheckURL.String
		}
		if publicKey.Valid {
			service.PublicKey = publicKey.String
		}
		if signature.Valid {
			service.Signature = signature.String
		}
		if certificate.Valid {
			service.Certificate = certificate.String
		}
		if metadata.Valid {
			service.Metadata = metadata.String
		}

		service.RegisteredAt = time.Unix(registeredAt, 0)
		if lastHealthCheck.Valid && lastHealthCheck.Int64 > 0 {
			service.LastHealthCheck = time.Unix(lastHealthCheck.Int64, 0)
		}

		// Filter by version if specified
		if req.MinVersion != "" && service.Version < req.MinVersion {
			continue
		}

		services = append(services, service)
	}

	response := models.ServiceDiscoveryResponse{
		Services:   services,
		TotalCount: len(services),
		Timestamp:  time.Now(),
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"discovery": response,
	}))
}

// RotateService handles secure service rotation
func (h *ServiceDiscoveryHandler) RotateService(c *gin.Context) {
	var req models.ServiceRotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid rotation request",
			"",
		))
		return
	}

	ctx := context.Background()

	// Verify admin token
	if req.AdminToken == "" || len(req.AdminToken) < 32 {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid admin token",
			"",
		))
		return
	}

	// Get old service
	var oldService models.ServiceRegistration
	err := h.getServiceByID(ctx, req.CurrentServiceID, &oldService)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Current service not found",
			"",
		))
		return
	}

	// Verify old service can be rotated
	if !oldService.CanRotate() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidAction,
			fmt.Sprintf("Service cannot be rotated in current status: %s", oldService.Status),
			"",
		))
		return
	}

	// Verify service types match
	if oldService.Type != req.NewService.Type {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidAction,
			"Service type mismatch",
			"",
		))
		return
	}

	// Set new service metadata
	req.NewService.ID = uuid.New().String()
	req.NewService.Status = models.ServiceStatusRegistering
	req.NewService.RegisteredAt = time.Now()
	req.NewService.RegisteredBy = req.RequestedBy

	// Sign new service
	if err := h.signer.SignServiceRegistration(&req.NewService); err != nil {
		logger.Error("Failed to sign new service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to sign new service",
			"",
		))
		return
	}

	// Verify rotation is legitimate
	if err := h.signer.VerifyServiceRotation(&oldService, &req.NewService, req.AdminToken); err != nil {
		logger.Warn("Service rotation verification failed",
			zap.Error(err),
			zap.String("old_service", oldService.ID),
			zap.String("requested_by", req.RequestedBy),
		)
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Service rotation verification failed: "+err.Error(),
			"",
		))
		return
	}

	// Start transaction
	// 1. Mark old service as decommissioned
	updateOldQuery := `
		UPDATE service_registry
		SET status = ?, deleted = 1
		WHERE id = ?
	`

	_, err = h.db.Exec(ctx, updateOldQuery, models.ServiceStatusDecommission, oldService.ID)
	if err != nil {
		logger.Error("Failed to decommission old service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to decommission old service",
			"",
		))
		return
	}

	// 2. Register new service
	insertNewQuery := `
		INSERT INTO service_registry (
			id, name, type, version, url, health_check_url, public_key, signature, certificate,
			status, priority, metadata, registered_by, registered_at, last_health_check,
			health_check_count, failed_health_count, deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(ctx, insertNewQuery,
		req.NewService.ID,
		req.NewService.Name,
		req.NewService.Type,
		req.NewService.Version,
		req.NewService.URL,
		req.NewService.HealthCheckURL,
		req.NewService.PublicKey,
		req.NewService.Signature,
		req.NewService.Certificate,
		req.NewService.Status,
		req.NewService.Priority,
		req.NewService.Metadata,
		req.NewService.RegisteredBy,
		req.NewService.RegisteredAt.Unix(),
		0, 0, 0, 0,
	)

	if err != nil {
		logger.Error("Failed to register new service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to register new service",
			"",
		))
		return
	}

	// 3. Record rotation in audit log
	rotationTime := time.Now()
	verificationHash := security.GenerateRotationCode(req.NewService.ID, req.AdminToken)

	auditQuery := `
		INSERT INTO service_rotation_audit (
			id, old_service_id, new_service_id, reason, requested_by, rotation_time,
			verification_hash, success, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(ctx, auditQuery,
		uuid.New().String(),
		oldService.ID,
		req.NewService.ID,
		req.Reason,
		req.RequestedBy,
		rotationTime.Unix(),
		verificationHash,
		1, // success
		"",
	)

	if err != nil {
		logger.Error("Failed to record rotation audit", zap.Error(err))
		// Don't fail the rotation for this
	}

	// Perform immediate health check on new service
	go h.healthChecker.CheckServiceNow(req.NewService.ID)

	logger.Info("Service rotated successfully",
		zap.String("old_service_id", oldService.ID),
		zap.String("new_service_id", req.NewService.ID),
		zap.String("type", string(oldService.Type)),
		zap.String("requested_by", req.RequestedBy),
	)

	response := models.ServiceRotationResponse{
		Success:          true,
		OldServiceID:     oldService.ID,
		NewServiceID:     req.NewService.ID,
		RotationTime:     rotationTime,
		VerificationHash: verificationHash,
		Message:          "Service rotated successfully",
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"rotation": response,
	}))
}

// DecommissionService handles service decommissioning
func (h *ServiceDiscoveryHandler) DecommissionService(c *gin.Context) {
	var req models.ServiceDecommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid decommission request",
			"",
		))
		return
	}

	// Verify admin token
	if req.AdminToken == "" || len(req.AdminToken) < 32 {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid admin token",
			"",
		))
		return
	}

	ctx := context.Background()

	// Update service status
	query := `
		UPDATE service_registry
		SET status = ?, deleted = 1
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(ctx, query, models.ServiceStatusDecommission, req.ServiceID)
	if err != nil {
		logger.Error("Failed to decommission service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to decommission service",
			"",
		))
		return
	}

	// Check if service was found
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Service not found",
			"",
		))
		return
	}

	logger.Info("Service decommissioned",
		zap.String("service_id", req.ServiceID),
		zap.String("reason", req.Reason),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"message":    "Service decommissioned successfully",
		"service_id": req.ServiceID,
	}))
}

// GetServiceHealth returns health information for a service
func (h *ServiceDiscoveryHandler) GetServiceHealth(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Service ID is required",
			"",
		))
		return
	}

	ctx := context.Background()

	// Get service info
	var service models.ServiceRegistration
	err := h.getServiceByID(ctx, serviceID, &service)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Service not found",
			"",
		))
		return
	}

	// Get health history
	history, err := h.healthChecker.GetServiceHealthHistory(serviceID, 20)
	if err != nil {
		logger.Error("Failed to get health history", zap.Error(err))
		history = []models.ServiceHealthCheck{}
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"service": map[string]interface{}{
			"id":                  service.ID,
			"name":                service.Name,
			"status":              service.Status,
			"last_health_check":   service.LastHealthCheck,
			"health_check_count":  service.HealthCheckCount,
			"failed_health_count": service.FailedHealthCount,
		},
		"health_history": history,
	}))
}

// ListServices returns all registered services
func (h *ServiceDiscoveryHandler) ListServices(c *gin.Context) {
	ctx := context.Background()

	query := `
		SELECT id, name, type, version, url, status, priority, registered_at, last_health_check,
		       health_check_count, failed_health_count
		FROM service_registry
		WHERE deleted = 0
		ORDER BY type, priority DESC, name
	`

	rows, err := h.db.Query(ctx, query)
	if err != nil {
		logger.Error("Failed to query services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list services",
			"",
		))
		return
	}
	defer rows.Close()

	var services []map[string]interface{}

	for rows.Next() {
		var id, name, svcType, version, url, status string
		var priority, healthCheckCount, failedHealthCount int
		var registeredAt, lastHealthCheck int64

		err := rows.Scan(&id, &name, &svcType, &version, &url, &status, &priority,
			&registeredAt, &lastHealthCheck, &healthCheckCount, &failedHealthCount)
		if err != nil {
			logger.Error("Failed to scan service row", zap.Error(err))
			continue
		}

		service := map[string]interface{}{
			"id":                  id,
			"name":                name,
			"type":                svcType,
			"version":             version,
			"url":                 url,
			"status":              status,
			"priority":            priority,
			"registered_at":       time.Unix(registeredAt, 0),
			"health_check_count":  healthCheckCount,
			"failed_health_count": failedHealthCount,
		}

		if lastHealthCheck > 0 {
			service["last_health_check"] = time.Unix(lastHealthCheck, 0)
		}

		services = append(services, service)
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"services": services,
		"total":    len(services),
	}))
}

// getServiceByID retrieves a service by ID
func (h *ServiceDiscoveryHandler) getServiceByID(ctx context.Context, serviceID string, service *models.ServiceRegistration) error {
	query := `
		SELECT id, name, type, version, url, health_check_url, public_key, signature, certificate,
		       status, priority, metadata, registered_by, registered_at, last_health_check,
		       health_check_count, failed_health_count
		FROM service_registry
		WHERE id = ? AND deleted = 0
	`

	var registeredAt int64
	var lastHealthCheck sql.NullInt64
	var healthCheckURL, publicKey, signature, certificate, metadata sql.NullString

	err := h.db.QueryRow(ctx, query, serviceID).Scan(
		&service.ID,
		&service.Name,
		&service.Type,
		&service.Version,
		&service.URL,
		&healthCheckURL,
		&publicKey,
		&signature,
		&certificate,
		&service.Status,
		&service.Priority,
		&metadata,
		&service.RegisteredBy,
		&registeredAt,
		&lastHealthCheck,
		&service.HealthCheckCount,
		&service.FailedHealthCount,
	)

	if err != nil {
		return err
	}

	// Convert nullable strings
	if healthCheckURL.Valid {
		service.HealthCheckURL = healthCheckURL.String
	}
	if publicKey.Valid {
		service.PublicKey = publicKey.String
	}
	if signature.Valid {
		service.Signature = signature.String
	}
	if certificate.Valid {
		service.Certificate = certificate.String
	}
	if metadata.Valid {
		service.Metadata = metadata.String
	}

	service.RegisteredAt = time.Unix(registeredAt, 0)
	if lastHealthCheck.Valid && lastHealthCheck.Int64 > 0 {
		service.LastHealthCheck = time.Unix(lastHealthCheck.Int64, 0)
	}

	return nil
}

// UpdateService handles service metadata updates
func (h *ServiceDiscoveryHandler) UpdateService(c *gin.Context) {
	var req models.ServiceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Invalid update request",
			"",
		))
		return
	}

	// Verify admin token
	if req.AdminToken == "" || len(req.AdminToken) < 32 {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Invalid admin token",
			"",
		))
		return
	}

	ctx := context.Background()

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}

	if req.Version != "" {
		updates = append(updates, "version = ?")
		args = append(args, req.Version)
	}
	if req.URL != "" {
		updates = append(updates, "url = ?")
		args = append(args, req.URL)
	}
	if req.HealthCheckURL != "" {
		updates = append(updates, "health_check_url = ?")
		args = append(args, req.HealthCheckURL)
	}
	if req.Priority != 0 {
		updates = append(updates, "priority = ?")
		args = append(args, req.Priority)
	}
	if req.Metadata != "" {
		updates = append(updates, "metadata = ?")
		args = append(args, req.Metadata)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"No fields to update",
			"",
		))
		return
	}

	args = append(args, req.ServiceID)

	query := fmt.Sprintf("UPDATE service_registry SET %s WHERE id = ? AND deleted = 0",
		joinStrings(updates, ", "))

	result, err := h.db.Exec(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to update service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update service",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Service not found",
			"",
		))
		return
	}

	logger.Info("Service updated",
		zap.String("service_id", req.ServiceID),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"message":    "Service updated successfully",
		"service_id": req.ServiceID,
	}))
}

// joinStrings is a helper to join strings
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
