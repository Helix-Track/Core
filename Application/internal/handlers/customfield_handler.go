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
	"helixtrack.ru/core/internal/websocket"
)

// handleCustomFieldCreate creates a new custom field definition
func (h *Handler) handleCustomFieldCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "customfield", models.PermissionCreate)
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

	// Parse custom field data from request
	fieldName, ok := req.Data["fieldName"].(string)
	if !ok || fieldName == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing fieldName",
			"",
		))
		return
	}

	fieldType, ok := req.Data["fieldType"].(string)
	if !ok || fieldType == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing fieldType",
			"",
		))
		return
	}

	customField := &models.CustomField{
		ID:        uuid.New().String(),
		FieldName: fieldName,
		FieldType: fieldType,
		Description: getStringFromData(req.Data, "description"),
		IsRequired:  getBoolFromData(req.Data, "isRequired"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Handle optional projectId
	if projectID, ok := req.Data["projectId"].(string); ok && projectID != "" {
		customField.ProjectID = &projectID
	}

	// Handle optional defaultValue
	if defaultValue, ok := req.Data["defaultValue"].(string); ok && defaultValue != "" {
		customField.DefaultValue = &defaultValue
	}

	// Handle optional configuration (JSON object)
	if configuration, ok := req.Data["configuration"]; ok && configuration != nil {
		configJSON, err := json.Marshal(configuration)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid configuration format",
				"",
			))
			return
		}
		configStr := string(configJSON)
		customField.Configuration = &configStr
	}

	// Validate field type
	if !customField.IsValidFieldType() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid field type",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		customField.ID,
		customField.FieldName,
		customField.FieldType,
		customField.Description,
		customField.ProjectID,
		customField.IsRequired,
		customField.DefaultValue,
		customField.Configuration,
		customField.Created,
		customField.Modified,
		customField.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create custom field", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create custom field",
			"",
		))
		return
	}

	logger.Info("Custom field created",
		zap.String("custom_field_id", customField.ID),
		zap.String("field_name", customField.FieldName),
		zap.String("username", username),
	)

	// Publish custom field created event
	projectContext := ""
	if customField.ProjectID != nil {
		projectContext = *customField.ProjectID
	}
	h.publisher.PublishEntityEvent(
		models.ActionCreate,
		"customfield",
		customField.ID,
		username,
		map[string]interface{}{
			"id":          customField.ID,
			"field_name":  customField.FieldName,
			"field_type":  customField.FieldType,
			"description": customField.Description,
			"project_id":  customField.ProjectID,
			"is_required": customField.IsRequired,
		},
		websocket.NewProjectContext(projectContext, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"customField": customField,
	})
	c.JSON(http.StatusCreated, response)
}

// handleCustomFieldRead reads a single custom field by ID
func (h *Handler) handleCustomFieldRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get custom field ID from request
	customFieldID, ok := req.Data["id"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing custom field ID",
			"",
		))
		return
	}

	// Query custom field from database
	query := `
		SELECT id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted
		FROM custom_field
		WHERE id = ? AND deleted = 0
	`

	var customField models.CustomField
	err := h.db.QueryRow(c.Request.Context(), query, customFieldID).Scan(
		&customField.ID,
		&customField.FieldName,
		&customField.FieldType,
		&customField.Description,
		&customField.ProjectID,
		&customField.IsRequired,
		&customField.DefaultValue,
		&customField.Configuration,
		&customField.Created,
		&customField.Modified,
		&customField.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read custom field", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read custom field",
			"",
		))
		return
	}

	logger.Info("Custom field read",
		zap.String("custom_field_id", customField.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"customField": customField,
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldList lists all custom fields (optionally filtered by project_id)
func (h *Handler) handleCustomFieldList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check for optional projectId filter
	projectID, hasProjectFilter := req.Data["projectId"].(string)

	var query string
	var args []interface{}

	if hasProjectFilter && projectID != "" {
		// List custom fields for specific project + global fields
		query = `
			SELECT id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted
			FROM custom_field
			WHERE deleted = 0 AND (project_id = ? OR project_id IS NULL)
			ORDER BY field_name ASC
		`
		args = []interface{}{projectID}
	} else {
		// List all custom fields
		query = `
			SELECT id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted
			FROM custom_field
			WHERE deleted = 0
			ORDER BY field_name ASC
		`
		args = []interface{}{}
	}

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list custom fields", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list custom fields",
			"",
		))
		return
	}
	defer rows.Close()

	customFields := make([]models.CustomField, 0)
	for rows.Next() {
		var customField models.CustomField
		err := rows.Scan(
			&customField.ID,
			&customField.FieldName,
			&customField.FieldType,
			&customField.Description,
			&customField.ProjectID,
			&customField.IsRequired,
			&customField.DefaultValue,
			&customField.Configuration,
			&customField.Created,
			&customField.Modified,
			&customField.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan custom field", zap.Error(err))
			continue
		}
		customFields = append(customFields, customField)
	}

	logger.Info("Custom fields listed",
		zap.Int("count", len(customFields)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"customFields": customFields,
		"count":        len(customFields),
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldModify updates an existing custom field
func (h *Handler) handleCustomFieldModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "customfield", models.PermissionUpdate)
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

	// Get custom field ID
	customFieldID, ok := req.Data["id"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing custom field ID",
			"",
		))
		return
	}

	// Check if custom field exists
	checkQuery := `SELECT COUNT(*) FROM custom_field WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, customFieldID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field not found",
			"",
		))
		return
	}

	// Build update query dynamically based on provided fields
	updates := make(map[string]interface{})

	if fieldName, ok := req.Data["fieldName"].(string); ok && fieldName != "" {
		updates["field_name"] = fieldName
	}
	if description, ok := req.Data["description"].(string); ok {
		updates["description"] = description
	}
	if isRequired, ok := req.Data["isRequired"].(bool); ok {
		updates["is_required"] = isRequired
	}
	if defaultValue, ok := req.Data["defaultValue"].(string); ok {
		updates["default_value"] = defaultValue
	}
	if configuration, ok := req.Data["configuration"]; ok && configuration != nil {
		configJSON, err := json.Marshal(configuration)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid configuration format",
				"",
			))
			return
		}
		updates["configuration"] = string(configJSON)
	}

	updates["modified"] = time.Now().Unix()

	if len(updates) == 1 { // Only modified was set
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"No fields to update",
			"",
		))
		return
	}

	// Build and execute update query
	query := "UPDATE custom_field SET "
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
	args = append(args, customFieldID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update custom field", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update custom field",
			"",
		))
		return
	}

	logger.Info("Custom field updated",
		zap.String("custom_field_id", customFieldID),
		zap.String("username", username),
	)

	// Get project context for event publishing
	var projectID *string
	contextQuery := `SELECT project_id FROM custom_field WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), contextQuery, customFieldID).Scan(&projectID)

	projectContext := ""
	if err == nil && projectID != nil {
		projectContext = *projectID
	}

	// Publish custom field updated event
	h.publisher.PublishEntityEvent(
		models.ActionModify,
		"customfield",
		customFieldID,
		username,
		updates,
		websocket.NewProjectContext(projectContext, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      customFieldID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldRemove soft-deletes a custom field
func (h *Handler) handleCustomFieldRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "customfield", models.PermissionDelete)
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

	// Get custom field ID
	customFieldID, ok := req.Data["id"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing custom field ID",
			"",
		))
		return
	}

	// Get project context before deletion for event publishing
	var projectID *string
	contextQuery := `SELECT project_id FROM custom_field WHERE id = ? AND deleted = 0`
	_ = h.db.QueryRow(c.Request.Context(), contextQuery, customFieldID).Scan(&projectID)

	// Soft delete the custom field
	query := `UPDATE custom_field SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), customFieldID)
	if err != nil {
		logger.Error("Failed to delete custom field", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete custom field",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field not found",
			"",
		))
		return
	}

	logger.Info("Custom field deleted",
		zap.String("custom_field_id", customFieldID),
		zap.String("username", username),
	)

	// Publish custom field deleted event
	projectContext := ""
	if projectID != nil {
		projectContext = *projectID
	}
	h.publisher.PublishEntityEvent(
		models.ActionRemove,
		"customfield",
		customFieldID,
		username,
		map[string]interface{}{
			"id":         customFieldID,
			"project_id": projectID,
		},
		websocket.NewProjectContext(projectContext, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      customFieldID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldOptionCreate creates an option for a select/multi-select field
func (h *Handler) handleCustomFieldOptionCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "customfield", models.PermissionCreate)
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

	// Parse option data from request
	customFieldID, ok := req.Data["customFieldId"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing customFieldId",
			"",
		))
		return
	}

	value, ok := req.Data["value"].(string)
	if !ok || value == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing value",
			"",
		))
		return
	}

	displayValue, ok := req.Data["displayValue"].(string)
	if !ok || displayValue == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing displayValue",
			"",
		))
		return
	}

	// Verify custom field exists and is a select type
	var fieldType string
	checkQuery := `SELECT field_type FROM custom_field WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, customFieldID).Scan(&fieldType)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to verify custom field", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to verify custom field",
			"",
		))
		return
	}

	// Validate that field type requires options
	tempField := &models.CustomField{FieldType: fieldType}
	if !tempField.RequiresOptions() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Custom field type does not support options",
			"",
		))
		return
	}

	option := &models.CustomFieldOption{
		ID:            uuid.New().String(),
		CustomFieldID: customFieldID,
		Value:         value,
		DisplayValue:  displayValue,
		Position:      getIntFromData(req.Data, "position"),
		IsDefault:     getBoolFromData(req.Data, "isDefault"),
		Created:       time.Now().Unix(),
		Modified:      time.Now().Unix(),
		Deleted:       false,
	}

	// Insert into database
	query := `
		INSERT INTO custom_field_option (id, custom_field_id, value, display_value, position, is_default, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		option.ID,
		option.CustomFieldID,
		option.Value,
		option.DisplayValue,
		option.Position,
		option.IsDefault,
		option.Created,
		option.Modified,
		option.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create custom field option", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create custom field option",
			"",
		))
		return
	}

	logger.Info("Custom field option created",
		zap.String("option_id", option.ID),
		zap.String("custom_field_id", customFieldID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"option": option,
	})
	c.JSON(http.StatusCreated, response)
}

// handleCustomFieldOptionModify updates an option
func (h *Handler) handleCustomFieldOptionModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "customfield", models.PermissionUpdate)
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

	// Get option ID
	optionID, ok := req.Data["id"].(string)
	if !ok || optionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing option ID",
			"",
		))
		return
	}

	// Check if option exists
	checkQuery := `SELECT COUNT(*) FROM custom_field_option WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, optionID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field option not found",
			"",
		))
		return
	}

	// Build update query dynamically based on provided fields
	updates := make(map[string]interface{})

	if value, ok := req.Data["value"].(string); ok && value != "" {
		updates["value"] = value
	}
	if displayValue, ok := req.Data["displayValue"].(string); ok && displayValue != "" {
		updates["display_value"] = displayValue
	}
	if position, ok := req.Data["position"].(float64); ok {
		updates["position"] = int(position)
	}
	if isDefault, ok := req.Data["isDefault"].(bool); ok {
		updates["is_default"] = isDefault
	}

	updates["modified"] = time.Now().Unix()

	if len(updates) == 1 { // Only modified was set
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"No fields to update",
			"",
		))
		return
	}

	// Build and execute update query
	query := "UPDATE custom_field_option SET "
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
	args = append(args, optionID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update custom field option", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update custom field option",
			"",
		))
		return
	}

	logger.Info("Custom field option updated",
		zap.String("option_id", optionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      optionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldOptionRemove soft-deletes an option
func (h *Handler) handleCustomFieldOptionRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "customfield", models.PermissionDelete)
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

	// Get option ID
	optionID, ok := req.Data["id"].(string)
	if !ok || optionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing option ID",
			"",
		))
		return
	}

	// Soft delete the option
	query := `UPDATE custom_field_option SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), optionID)
	if err != nil {
		logger.Error("Failed to delete custom field option", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete custom field option",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field option not found",
			"",
		))
		return
	}

	logger.Info("Custom field option deleted",
		zap.String("option_id", optionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      optionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldOptionList lists all options for a custom field
func (h *Handler) handleCustomFieldOptionList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get custom field ID from request
	customFieldID, ok := req.Data["customFieldId"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing customFieldId",
			"",
		))
		return
	}

	// Query all non-deleted options ordered by position
	query := `
		SELECT id, custom_field_id, value, display_value, position, is_default, created, modified, deleted
		FROM custom_field_option
		WHERE custom_field_id = ? AND deleted = 0
		ORDER BY position ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, customFieldID)
	if err != nil {
		logger.Error("Failed to list custom field options", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list custom field options",
			"",
		))
		return
	}
	defer rows.Close()

	options := make([]models.CustomFieldOption, 0)
	for rows.Next() {
		var option models.CustomFieldOption
		err := rows.Scan(
			&option.ID,
			&option.CustomFieldID,
			&option.Value,
			&option.DisplayValue,
			&option.Position,
			&option.IsDefault,
			&option.Created,
			&option.Modified,
			&option.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan custom field option", zap.Error(err))
			continue
		}
		options = append(options, option)
	}

	logger.Info("Custom field options listed",
		zap.Int("count", len(options)),
		zap.String("custom_field_id", customFieldID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"options": options,
		"count":   len(options),
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldValueSet sets/updates a custom field value for a ticket
func (h *Handler) handleCustomFieldValueSet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket", models.PermissionUpdate)
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

	// Parse data from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	customFieldID, ok := req.Data["customFieldId"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing customFieldId",
			"",
		))
		return
	}

	// Value can be null/empty for clearing a field
	var valuePtr *string
	if value, ok := req.Data["value"].(string); ok && value != "" {
		valuePtr = &value
	}

	// Check if value already exists
	checkQuery := `SELECT id FROM ticket_custom_field_value WHERE ticket_id = ? AND custom_field_id = ? AND deleted = 0`
	var existingID string
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketID, customFieldID).Scan(&existingID)

	now := time.Now().Unix()

	if err == sql.ErrNoRows {
		// Create new value
		newID := uuid.New().String()
		insertQuery := `
			INSERT INTO ticket_custom_field_value (id, ticket_id, custom_field_id, value, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err = h.db.Exec(c.Request.Context(), insertQuery, newID, ticketID, customFieldID, valuePtr, now, now, false)
		if err != nil {
			logger.Error("Failed to create custom field value", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to create custom field value",
				"",
			))
			return
		}

		logger.Info("Custom field value created",
			zap.String("ticket_id", ticketID),
			zap.String("custom_field_id", customFieldID),
			zap.String("username", username),
		)

		response := models.NewSuccessResponse(map[string]interface{}{
			"created": true,
			"id":      newID,
		})
		c.JSON(http.StatusCreated, response)
	} else if err != nil {
		logger.Error("Failed to check existing custom field value", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check existing custom field value",
			"",
		))
		return
	} else {
		// Update existing value
		updateQuery := `UPDATE ticket_custom_field_value SET value = ?, modified = ? WHERE id = ?`
		_, err = h.db.Exec(c.Request.Context(), updateQuery, valuePtr, now, existingID)
		if err != nil {
			logger.Error("Failed to update custom field value", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to update custom field value",
				"",
			))
			return
		}

		logger.Info("Custom field value updated",
			zap.String("ticket_id", ticketID),
			zap.String("custom_field_id", customFieldID),
			zap.String("username", username),
		)

		response := models.NewSuccessResponse(map[string]interface{}{
			"updated": true,
			"id":      existingID,
		})
		c.JSON(http.StatusOK, response)
	}
}

// handleCustomFieldValueGet gets a custom field value for a ticket
func (h *Handler) handleCustomFieldValueGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Parse data from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	customFieldID, ok := req.Data["customFieldId"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing customFieldId",
			"",
		))
		return
	}

	// Query custom field value
	query := `
		SELECT id, ticket_id, custom_field_id, value, created, modified, deleted
		FROM ticket_custom_field_value
		WHERE ticket_id = ? AND custom_field_id = ? AND deleted = 0
	`

	var fieldValue models.TicketCustomFieldValue
	err := h.db.QueryRow(c.Request.Context(), query, ticketID, customFieldID).Scan(
		&fieldValue.ID,
		&fieldValue.TicketID,
		&fieldValue.CustomFieldID,
		&fieldValue.Value,
		&fieldValue.Created,
		&fieldValue.Modified,
		&fieldValue.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field value not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read custom field value", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read custom field value",
			"",
		))
		return
	}

	logger.Info("Custom field value read",
		zap.String("ticket_id", ticketID),
		zap.String("custom_field_id", customFieldID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"fieldValue": fieldValue,
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldValueList lists all custom field values for a ticket
func (h *Handler) handleCustomFieldValueList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get ticket ID from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	// Query all non-deleted custom field values for the ticket
	query := `
		SELECT id, ticket_id, custom_field_id, value, created, modified, deleted
		FROM ticket_custom_field_value
		WHERE ticket_id = ? AND deleted = 0
		ORDER BY created ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list custom field values", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list custom field values",
			"",
		))
		return
	}
	defer rows.Close()

	fieldValues := make([]models.TicketCustomFieldValue, 0)
	for rows.Next() {
		var fieldValue models.TicketCustomFieldValue
		err := rows.Scan(
			&fieldValue.ID,
			&fieldValue.TicketID,
			&fieldValue.CustomFieldID,
			&fieldValue.Value,
			&fieldValue.Created,
			&fieldValue.Modified,
			&fieldValue.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan custom field value", zap.Error(err))
			continue
		}
		fieldValues = append(fieldValues, fieldValue)
	}

	logger.Info("Custom field values listed",
		zap.Int("count", len(fieldValues)),
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"fieldValues": fieldValues,
		"count":       len(fieldValues),
	})
	c.JSON(http.StatusOK, response)
}

// handleCustomFieldValueRemove removes a custom field value from a ticket
func (h *Handler) handleCustomFieldValueRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket", models.PermissionUpdate)
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

	// Parse data from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	customFieldID, ok := req.Data["customFieldId"].(string)
	if !ok || customFieldID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing customFieldId",
			"",
		))
		return
	}

	// Soft delete the custom field value
	query := `UPDATE ticket_custom_field_value SET deleted = 1, modified = ? WHERE ticket_id = ? AND custom_field_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), ticketID, customFieldID)
	if err != nil {
		logger.Error("Failed to delete custom field value", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete custom field value",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Custom field value not found",
			"",
		))
		return
	}

	logger.Info("Custom field value deleted",
		zap.String("ticket_id", ticketID),
		zap.String("custom_field_id", customFieldID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
	})
	c.JSON(http.StatusOK, response)
}

// Helper function to safely get int from map
func getIntFromData(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return 0
}
