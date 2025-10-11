package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

// handleAuditCreate creates a new audit log entry
func (h *Handler) handleAuditCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Parse audit data from request
	action, ok := req.Data["action"].(string)
	if !ok || action == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing action",
			"",
		))
		return
	}

	audit := &models.Audit{
		ID:         uuid.New().String(),
		Action:     action,
		UserID:     getStringFromData(req.Data, "userId"),
		EntityID:   getStringFromData(req.Data, "entityId"),
		EntityType: getStringFromData(req.Data, "entityType"),
		Details:    getStringFromData(req.Data, "details"),
		Created:    time.Now().Unix(),
		Modified:   time.Now().Unix(),
		Deleted:    false,
	}

	// Validate audit
	if !audit.IsValidAction() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid action",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO audit (id, action, user_id, entity_id, entity_type, details, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(c.Request.Context(), query,
		audit.ID,
		audit.Action,
		audit.UserID,
		audit.EntityID,
		audit.EntityType,
		audit.Details,
		audit.Created,
		audit.Modified,
		audit.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create audit entry", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create audit entry",
			"",
		))
		return
	}

	logger.Info("Audit entry created",
		zap.String("audit_id", audit.ID),
		zap.String("action", audit.Action),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"audit": audit,
	})
	c.JSON(http.StatusCreated, response)
}

// handleAuditRead reads a single audit entry by ID
func (h *Handler) handleAuditRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get audit ID from request
	auditID, ok := req.Data["id"].(string)
	if !ok || auditID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing audit ID",
			"",
		))
		return
	}

	// Query audit from database
	query := `
		SELECT id, action, user_id, entity_id, entity_type, details, created, modified, deleted
		FROM audit
		WHERE id = ? AND deleted = 0
	`

	var audit models.Audit
	err := h.db.QueryRow(c.Request.Context(), query, auditID).Scan(
		&audit.ID,
		&audit.Action,
		&audit.UserID,
		&audit.EntityID,
		&audit.EntityType,
		&audit.Details,
		&audit.Created,
		&audit.Modified,
		&audit.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Audit entry not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read audit entry", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read audit entry",
			"",
		))
		return
	}

	logger.Info("Audit entry read",
		zap.String("audit_id", audit.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"audit": audit,
	})
	c.JSON(http.StatusOK, response)
}

// handleAuditList lists all audit entries
func (h *Handler) handleAuditList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted audit entries, ordered by creation time (newest first)
	query := `
		SELECT id, action, user_id, entity_id, entity_type, details, created, modified, deleted
		FROM audit
		WHERE deleted = 0
		ORDER BY created DESC
		LIMIT 1000
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list audit entries", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list audit entries",
			"",
		))
		return
	}
	defer rows.Close()

	audits := make([]models.Audit, 0)
	for rows.Next() {
		var audit models.Audit
		err := rows.Scan(
			&audit.ID,
			&audit.Action,
			&audit.UserID,
			&audit.EntityID,
			&audit.EntityType,
			&audit.Details,
			&audit.Created,
			&audit.Modified,
			&audit.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan audit entry", zap.Error(err))
			continue
		}
		audits = append(audits, audit)
	}

	logger.Info("Audit entries listed",
		zap.Int("count", len(audits)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"audits": audits,
		"count":  len(audits),
	})
	c.JSON(http.StatusOK, response)
}

// handleAuditQuery queries audit entries with filters
func (h *Handler) handleAuditQuery(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Build query based on filters
	queryStr := `
		SELECT id, action, user_id, entity_id, entity_type, details, created, modified, deleted
		FROM audit
		WHERE deleted = 0
	`
	args := make([]interface{}, 0)

	// Add filters if provided
	if userID, ok := req.Data["userId"].(string); ok && userID != "" {
		queryStr += " AND user_id = ?"
		args = append(args, userID)
	}

	if action, ok := req.Data["action"].(string); ok && action != "" {
		queryStr += " AND action = ?"
		args = append(args, action)
	}

	if entityType, ok := req.Data["entityType"].(string); ok && entityType != "" {
		queryStr += " AND entity_type = ?"
		args = append(args, entityType)
	}

	if entityID, ok := req.Data["entityId"].(string); ok && entityID != "" {
		queryStr += " AND entity_id = ?"
		args = append(args, entityID)
	}

	// Add time range filters if provided
	if startTime, ok := req.Data["startTime"].(float64); ok {
		queryStr += " AND created >= ?"
		args = append(args, int64(startTime))
	}

	if endTime, ok := req.Data["endTime"].(float64); ok {
		queryStr += " AND created <= ?"
		args = append(args, int64(endTime))
	}

	// Add ordering and limit
	queryStr += " ORDER BY created DESC LIMIT 1000"

	rows, err := h.db.Query(c.Request.Context(), queryStr, args...)
	if err != nil {
		logger.Error("Failed to query audit entries", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to query audit entries",
			"",
		))
		return
	}
	defer rows.Close()

	audits := make([]models.Audit, 0)
	for rows.Next() {
		var audit models.Audit
		err := rows.Scan(
			&audit.ID,
			&audit.Action,
			&audit.UserID,
			&audit.EntityID,
			&audit.EntityType,
			&audit.Details,
			&audit.Created,
			&audit.Modified,
			&audit.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan audit entry", zap.Error(err))
			continue
		}
		audits = append(audits, audit)
	}

	logger.Info("Audit entries queried",
		zap.Int("count", len(audits)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"audits": audits,
		"count":  len(audits),
	})
	c.JSON(http.StatusOK, response)
}

// handleAuditAddMeta adds metadata to an audit entry
func (h *Handler) handleAuditAddMeta(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Parse metadata from request
	auditID, ok := req.Data["auditId"].(string)
	if !ok || auditID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing auditId",
			"",
		))
		return
	}

	property, ok := req.Data["property"].(string)
	if !ok || property == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing property",
			"",
		))
		return
	}

	// Value can be any type, convert to JSON string
	var valueStr string
	if value, ok := req.Data["value"]; ok {
		valueBytes, err := json.Marshal(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid value format",
				"",
			))
			return
		}
		valueStr = string(valueBytes)
	}

	metadata := &models.AuditMetaData{
		ID:       uuid.New().String(),
		AuditID:  auditID,
		Property: property,
		Value:    valueStr,
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
		Deleted:  false,
	}

	// Insert into database
	query := `
		INSERT INTO audit_metadata (id, audit_id, property, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(c.Request.Context(), query,
		metadata.ID,
		metadata.AuditID,
		metadata.Property,
		metadata.Value,
		metadata.Created,
		metadata.Modified,
		metadata.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add audit metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add audit metadata",
			"",
		))
		return
	}

	logger.Info("Audit metadata added",
		zap.String("metadata_id", metadata.ID),
		zap.String("audit_id", auditID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
	})
	c.JSON(http.StatusCreated, response)
}
