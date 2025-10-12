package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/websocket"
)

// handleNotificationSchemeCreate creates a notification scheme
func (h *Handler) handleNotificationSchemeCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "notification_scheme", models.PermissionCreate)
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	if !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permission",
			"",
		))
		return
	}

	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	description := getStringFromData(req.Data, "description")
	projectID := getStringFromData(req.Data, "projectId")

	schemeID := uuid.New().String()
	now := time.Now().Unix()

	var query string
	var args []interface{}

	if projectID != "" {
		query = `INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, 0)`
		args = []interface{}{schemeID, title, description, projectID, now, now}
	} else {
		query = `INSERT INTO notification_scheme (id, title, description, project_id, created, modified, deleted) VALUES (?, ?, ?, NULL, ?, ?, 0)`
		args = []interface{}{schemeID, title, description, now, now}
	}

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to create notification scheme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create notification scheme",
			"",
		))
		return
	}

	logger.Info("Notification scheme created", zap.String("scheme_id", schemeID), zap.String("username", username))

	h.publisher.PublishEntityEvent("notificationSchemeCreate", "notification_scheme", schemeID, username,
		map[string]interface{}{"id": schemeID, "title": title}, websocket.NewProjectContext(projectID, []string{"READ"}))

	response := models.NewSuccessResponse(map[string]interface{}{"id": schemeID, "title": title, "projectId": projectID})
	c.JSON(http.StatusCreated, response)
}

// handleNotificationSchemeRead reads a notification scheme
func (h *Handler) handleNotificationSchemeRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	schemeID, ok := req.Data["schemeId"].(string)
	if !ok || schemeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing schemeId", ""))
		return
	}

	query := `SELECT id, title, description, project_id, created, modified FROM notification_scheme WHERE id = ? AND deleted = 0`
	var scheme models.NotificationScheme
	var projectID sql.NullString

	err := h.db.QueryRow(c.Request.Context(), query, schemeID).Scan(&scheme.ID, &scheme.Title, &scheme.Description, &projectID, &scheme.Created, &scheme.Modified)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Notification scheme not found", ""))
		return
	}
	if err != nil {
		logger.Error("Failed to read notification scheme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to read notification scheme", ""))
		return
	}

	if projectID.Valid && projectID.String != "" {
		scheme.ProjectID = &projectID.String
	}

	logger.Info("Notification scheme read", zap.String("scheme_id", schemeID), zap.String("username", username))

	response := models.NewSuccessResponse(map[string]interface{}{
		"id": scheme.ID, "title": scheme.Title, "description": scheme.Description, "projectId": projectID.String,
		"created": scheme.Created, "modified": scheme.Modified,
	})
	c.JSON(http.StatusOK, response)
}

// handleNotificationSchemeList lists notification schemes
func (h *Handler) handleNotificationSchemeList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	projectID := getStringFromData(req.Data, "projectId")
	var query string
	var args []interface{}

	if projectID != "" {
		query = `SELECT id, title, description, project_id, created, modified FROM notification_scheme WHERE deleted = 0 AND (project_id = ? OR project_id IS NULL) ORDER BY title ASC`
		args = []interface{}{projectID}
	} else {
		query = `SELECT id, title, description, project_id, created, modified FROM notification_scheme WHERE deleted = 0 ORDER BY title ASC`
	}

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list notification schemes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list notification schemes", ""))
		return
	}
	defer rows.Close()

	schemes := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title, description string
		var projectIDVal sql.NullString
		var created, modified int64

		if err := rows.Scan(&id, &title, &description, &projectIDVal, &created, &modified); err != nil {
			continue
		}

		schemes = append(schemes, map[string]interface{}{
			"id": id, "title": title, "description": description, "projectId": projectIDVal.String,
			"isGlobal": !projectIDVal.Valid || projectIDVal.String == "", "created": created, "modified": modified,
		})
	}

	logger.Info("Notification schemes listed", zap.Int("count", len(schemes)), zap.String("username", username))
	response := models.NewSuccessResponse(map[string]interface{}{"schemes": schemes, "count": len(schemes)})
	c.JSON(http.StatusOK, response)
}

// handleNotificationSchemeModify updates a notification scheme
func (h *Handler) handleNotificationSchemeModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "notification_scheme", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrorCodeForbidden, "Insufficient permission", ""))
		return
	}

	schemeID, ok := req.Data["schemeId"].(string)
	if !ok || schemeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing schemeId", ""))
		return
	}

	updates := make(map[string]interface{})
	if title, ok := req.Data["title"].(string); ok && title != "" {
		updates["title"] = title
	}
	if description, ok := req.Data["description"].(string); ok {
		updates["description"] = description
	}
	updates["modified"] = time.Now().Unix()

	if len(updates) == 1 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "No fields to update", ""))
		return
	}

	query := "UPDATE notification_scheme SET "
	args := make([]interface{}, 0)
	first := true
	for key, value := range updates {
		if !first {
			query += ", "
		}
		query += key + " = ?"
		args = append(args, value)
		first = false
	}
	query += " WHERE id = ?"
	args = append(args, schemeID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update notification scheme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to update notification scheme", ""))
		return
	}

	logger.Info("Notification scheme updated", zap.String("scheme_id", schemeID), zap.String("username", username))
	h.publisher.PublishEntityEvent("notificationSchemeModify", "notification_scheme", schemeID, username, updates, websocket.NewProjectContext("", []string{"READ"}))

	response := models.NewSuccessResponse(map[string]interface{}{"updated": true, "schemeId": schemeID})
	c.JSON(http.StatusOK, response)
}

// handleNotificationSchemeRemove soft-deletes a notification scheme
func (h *Handler) handleNotificationSchemeRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "notification_scheme", models.PermissionDelete)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrorCodeForbidden, "Insufficient permission", ""))
		return
	}

	schemeID, ok := req.Data["schemeId"].(string)
	if !ok || schemeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing schemeId", ""))
		return
	}

	query := `UPDATE notification_scheme SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), schemeID)
	if err != nil {
		logger.Error("Failed to remove notification scheme", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to remove notification scheme", ""))
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Notification scheme not found", ""))
		return
	}

	h.db.Exec(c.Request.Context(), "UPDATE notification_rule SET deleted = 1 WHERE notification_scheme_id = ?", schemeID)

	logger.Info("Notification scheme removed", zap.String("scheme_id", schemeID), zap.String("username", username))
	h.publisher.PublishEntityEvent("notificationSchemeRemove", "notification_scheme", schemeID, username, map[string]interface{}{"schemeId": schemeID}, websocket.NewProjectContext("", []string{"READ"}))

	response := models.NewSuccessResponse(map[string]interface{}{"removed": true, "schemeId": schemeID})
	c.JSON(http.StatusOK, response)
}

// handleNotificationSchemeAddRule adds a rule to a notification scheme
func (h *Handler) handleNotificationSchemeAddRule(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "notification_scheme", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrorCodeForbidden, "Insufficient permission", ""))
		return
	}

	schemeID, ok := req.Data["schemeId"].(string)
	if !ok || schemeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing schemeId", ""))
		return
	}

	eventID, ok := req.Data["eventId"].(string)
	if !ok || eventID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing eventId", ""))
		return
	}

	recipientType, ok := req.Data["recipientType"].(string)
	if !ok || recipientType == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing recipientType", ""))
		return
	}

	recipientID := getStringFromData(req.Data, "recipientId")

	ruleID := uuid.New().String()
	now := time.Now().Unix()

	query := `INSERT INTO notification_rule (id, notification_scheme_id, notification_event_id, recipient_type, recipient_id, created, deleted) VALUES (?, ?, ?, ?, ?, ?, 0)`
	var recipientIDPtr interface{}
	if recipientID != "" {
		recipientIDPtr = recipientID
	}

	_, err = h.db.Exec(c.Request.Context(), query, ruleID, schemeID, eventID, recipientType, recipientIDPtr, now)
	if err != nil {
		logger.Error("Failed to add notification rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to add notification rule", ""))
		return
	}

	logger.Info("Notification rule added", zap.String("rule_id", ruleID), zap.String("scheme_id", schemeID), zap.String("username", username))
	h.publisher.PublishEntityEvent("notificationSchemeAddRule", "notification_scheme", schemeID, username,
		map[string]interface{}{"ruleId": ruleID, "schemeId": schemeID, "eventId": eventID, "recipientType": recipientType},
		websocket.NewProjectContext("", []string{"READ"}))

	response := models.NewSuccessResponse(map[string]interface{}{"added": true, "ruleId": ruleID, "schemeId": schemeID})
	c.JSON(http.StatusCreated, response)
}

// handleNotificationSchemeRemoveRule removes a rule from a notification scheme
func (h *Handler) handleNotificationSchemeRemoveRule(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "notification_scheme", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrorCodeForbidden, "Insufficient permission", ""))
		return
	}

	ruleID, ok := req.Data["ruleId"].(string)
	if !ok || ruleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing ruleId", ""))
		return
	}

	query := `UPDATE notification_rule SET deleted = 1 WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, ruleID)
	if err != nil {
		logger.Error("Failed to remove notification rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to remove notification rule", ""))
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Notification rule not found", ""))
		return
	}

	logger.Info("Notification rule removed", zap.String("rule_id", ruleID), zap.String("username", username))
	h.publisher.PublishEntityEvent("notificationSchemeRemoveRule", "notification_scheme", ruleID, username, map[string]interface{}{"ruleId": ruleID}, websocket.NewProjectContext("", []string{"READ"}))

	response := models.NewSuccessResponse(map[string]interface{}{"removed": true, "ruleId": ruleID})
	c.JSON(http.StatusOK, response)
}

// handleNotificationSchemeListRules lists all rules for a notification scheme
func (h *Handler) handleNotificationSchemeListRules(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	schemeID, ok := req.Data["schemeId"].(string)
	if !ok || schemeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing schemeId", ""))
		return
	}

	query := `SELECT id, notification_event_id, recipient_type, recipient_id, created FROM notification_rule WHERE notification_scheme_id = ? AND deleted = 0 ORDER BY created ASC`
	rows, err := h.db.Query(c.Request.Context(), query, schemeID)
	if err != nil {
		logger.Error("Failed to list notification rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list notification rules", ""))
		return
	}
	defer rows.Close()

	rules := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eventID, recipientType string
		var recipientID sql.NullString
		var created int64

		if err := rows.Scan(&id, &eventID, &recipientType, &recipientID, &created); err != nil {
			continue
		}

		rules = append(rules, map[string]interface{}{
			"id": id, "eventId": eventID, "recipientType": recipientType, "recipientId": recipientID.String, "created": created,
		})
	}

	logger.Info("Notification rules listed", zap.String("scheme_id", schemeID), zap.Int("count", len(rules)), zap.String("username", username))
	response := models.NewSuccessResponse(map[string]interface{}{"schemeId": schemeID, "rules": rules, "count": len(rules)})
	c.JSON(http.StatusOK, response)
}

// handleNotificationEventList lists all available notification event types
func (h *Handler) handleNotificationEventList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	query := `SELECT id, event_type, title, description, created FROM notification_event WHERE deleted = 0 ORDER BY title ASC`
	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list notification events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list notification events", ""))
		return
	}
	defer rows.Close()

	events := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eventType, title, description string
		var created int64

		if err := rows.Scan(&id, &eventType, &title, &description, &created); err != nil {
			continue
		}

		events = append(events, map[string]interface{}{
			"id": id, "eventType": eventType, "title": title, "description": description, "created": created,
		})
	}

	logger.Info("Notification events listed", zap.Int("count", len(events)), zap.String("username", username))
	response := models.NewSuccessResponse(map[string]interface{}{"events": events, "count": len(events)})
	c.JSON(http.StatusOK, response)
}

// handleNotificationSend sends a notification manually (for testing/manual triggers)
func (h *Handler) handleNotificationSend(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrorCodeUnauthorized, "Unauthorized", ""))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "notification", models.PermissionCreate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrorCodeForbidden, "Insufficient permission", ""))
		return
	}

	recipientID, ok := req.Data["recipientId"].(string)
	if !ok || recipientID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing recipientId", ""))
		return
	}

	message, ok := req.Data["message"].(string)
	if !ok || message == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrorCodeMissingData, "Missing message", ""))
		return
	}

	logger.Info("Notification sent manually", zap.String("recipient_id", recipientID), zap.String("sender", username))

	h.publisher.PublishEntityEvent("notificationSend", "notification", recipientID, username,
		map[string]interface{}{"recipientId": recipientID, "message": message},
		websocket.NewProjectContext("", []string{"READ"}))

	response := models.NewSuccessResponse(map[string]interface{}{"sent": true, "recipientId": recipientID})
	c.JSON(http.StatusOK, response)
}
