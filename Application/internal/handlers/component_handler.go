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
)

// handleComponentCreate creates a new component
func (h *Handler) handleComponentCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "component", models.PermissionCreate)
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

	// Parse component data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	component := &models.Component{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO component (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		component.ID,
		component.Title,
		component.Description,
		component.Created,
		component.Modified,
		component.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create component",
			"",
		))
		return
	}

	logger.Info("Component created",
		zap.String("component_id", component.ID),
		zap.String("title", component.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"component": component,
	})
	c.JSON(http.StatusCreated, response)
}

// handleComponentRead reads a single component by ID
func (h *Handler) handleComponentRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get component ID from request
	componentID, ok := req.Data["id"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
			"",
		))
		return
	}

	// Query component from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM component
		WHERE id = ? AND deleted = 0
	`

	var component models.Component
	err := h.db.QueryRow(c.Request.Context(), query, componentID).Scan(
		&component.ID,
		&component.Title,
		&component.Description,
		&component.Created,
		&component.Modified,
		&component.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Component not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read component",
			"",
		))
		return
	}

	logger.Info("Component read",
		zap.String("component_id", component.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"component": component,
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentList lists all components
func (h *Handler) handleComponentList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted components ordered by title
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM component
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list components", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list components",
			"",
		))
		return
	}
	defer rows.Close()

	components := make([]models.Component, 0)
	for rows.Next() {
		var component models.Component
		err := rows.Scan(
			&component.ID,
			&component.Title,
			&component.Description,
			&component.Created,
			&component.Modified,
			&component.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan component", zap.Error(err))
			continue
		}
		components = append(components, component)
	}

	logger.Info("Components listed",
		zap.Int("count", len(components)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"components": components,
		"count":      len(components),
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentModify updates an existing component
func (h *Handler) handleComponentModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "component", models.PermissionUpdate)
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

	// Get component ID
	componentID, ok := req.Data["id"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
			"",
		))
		return
	}

	// Check if component exists
	checkQuery := `SELECT COUNT(*) FROM component WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, componentID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Component not found",
			"",
		))
		return
	}

	// Build update query dynamically based on provided fields
	updates := make(map[string]interface{})

	if title, ok := req.Data["title"].(string); ok && title != "" {
		updates["title"] = title
	}
	if description, ok := req.Data["description"].(string); ok {
		updates["description"] = description
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
	query := "UPDATE component SET "
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
	args = append(args, componentID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update component",
			"",
		))
		return
	}

	logger.Info("Component updated",
		zap.String("component_id", componentID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      componentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentRemove soft-deletes a component
func (h *Handler) handleComponentRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "component", models.PermissionDelete)
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

	// Get component ID
	componentID, ok := req.Data["id"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
			"",
		))
		return
	}

	// Soft delete the component
	query := `UPDATE component SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), componentID)
	if err != nil {
		logger.Error("Failed to delete component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete component",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Component not found",
			"",
		))
		return
	}

	logger.Info("Component deleted",
		zap.String("component_id", componentID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      componentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentAddTicket adds a component to a ticket
func (h *Handler) handleComponentAddTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "component", models.PermissionUpdate)
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

	// Get component ID and ticket ID
	componentID, ok := req.Data["componentId"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
			"",
		))
		return
	}

	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket ID",
			"",
		))
		return
	}

	// Create mapping
	mappingID := uuid.New().String()
	query := `
		INSERT INTO component_ticket_mapping (id, component_id, ticket_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err = h.db.Exec(c.Request.Context(), query, mappingID, componentID, ticketID, now, now, false)
	if err != nil {
		logger.Error("Failed to add component to ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add component to ticket",
			"",
		))
		return
	}

	logger.Info("Component added to ticket",
		zap.String("component_id", componentID),
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":       true,
		"componentId": componentID,
		"ticketId":    ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentRemoveTicket removes a component from a ticket
func (h *Handler) handleComponentRemoveTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "component", models.PermissionUpdate)
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

	// Get component ID and ticket ID
	componentID, ok := req.Data["componentId"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
			"",
		))
		return
	}

	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket ID",
			"",
		))
		return
	}

	// Remove mapping (soft delete)
	query := `UPDATE component_ticket_mapping SET deleted = 1, modified = ? WHERE component_id = ? AND ticket_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), componentID, ticketID)
	if err != nil {
		logger.Error("Failed to remove component from ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove component from ticket",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Component-ticket mapping not found",
			"",
		))
		return
	}

	logger.Info("Component removed from ticket",
		zap.String("component_id", componentID),
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":     true,
		"componentId": componentID,
		"ticketId":    ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentListTickets lists all tickets for a component
func (h *Handler) handleComponentListTickets(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get component ID
	componentID, ok := req.Data["componentId"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
			"",
		))
		return
	}

	// Query tickets
	query := `
		SELECT ticket_id
		FROM component_ticket_mapping
		WHERE component_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, componentID)
	if err != nil {
		logger.Error("Failed to list tickets for component", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list tickets",
			"",
		))
		return
	}
	defer rows.Close()

	ticketIDs := make([]string, 0)
	for rows.Next() {
		var ticketID string
		if err := rows.Scan(&ticketID); err != nil {
			logger.Error("Failed to scan ticket ID", zap.Error(err))
			continue
		}
		ticketIDs = append(ticketIDs, ticketID)
	}

	logger.Info("Tickets listed for component",
		zap.String("component_id", componentID),
		zap.Int("count", len(ticketIDs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketIds": ticketIDs,
		"count":     len(ticketIDs),
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentSetMetadata sets metadata for a component
func (h *Handler) handleComponentSetMetadata(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "component", models.PermissionUpdate)
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

	// Get component ID, property, and value
	componentID, ok := req.Data["componentId"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
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

	value := getStringFromData(req.Data, "value")

	// Check if metadata exists, update or insert
	checkQuery := `SELECT id FROM component_meta_data WHERE component_id = ? AND property = ? AND deleted = 0`
	var existingID string
	err = h.db.QueryRow(c.Request.Context(), checkQuery, componentID, property).Scan(&existingID)

	now := time.Now().Unix()

	if err == sql.ErrNoRows {
		// Insert new metadata
		metadataID := uuid.New().String()
		insertQuery := `
			INSERT INTO component_meta_data (id, component_id, property, value, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err = h.db.Exec(c.Request.Context(), insertQuery, metadataID, componentID, property, value, now, now, false)
	} else {
		// Update existing metadata
		updateQuery := `UPDATE component_meta_data SET value = ?, modified = ? WHERE id = ?`
		_, err = h.db.Exec(c.Request.Context(), updateQuery, value, now, existingID)
	}

	if err != nil {
		logger.Error("Failed to set component metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to set metadata",
			"",
		))
		return
	}

	logger.Info("Component metadata set",
		zap.String("component_id", componentID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"set":         true,
		"componentId": componentID,
		"property":    property,
		"value":       value,
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentGetMetadata gets metadata for a component
func (h *Handler) handleComponentGetMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get component ID and property
	componentID, ok := req.Data["componentId"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
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

	// Query metadata
	query := `
		SELECT id, component_id, property, value, created, modified, deleted
		FROM component_meta_data
		WHERE component_id = ? AND property = ? AND deleted = 0
	`

	var metadata models.ComponentMetaData
	err := h.db.QueryRow(c.Request.Context(), query, componentID, property).Scan(
		&metadata.ID,
		&metadata.ComponentID,
		&metadata.Property,
		&metadata.Value,
		&metadata.Created,
		&metadata.Modified,
		&metadata.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Metadata not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to get component metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get metadata",
			"",
		))
		return
	}

	logger.Info("Component metadata retrieved",
		zap.String("component_id", componentID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentListMetadata lists all metadata for a component
func (h *Handler) handleComponentListMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get component ID
	componentID, ok := req.Data["componentId"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
			"",
		))
		return
	}

	// Query all metadata
	query := `
		SELECT id, component_id, property, value, created, modified, deleted
		FROM component_meta_data
		WHERE component_id = ? AND deleted = 0
		ORDER BY property ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, componentID)
	if err != nil {
		logger.Error("Failed to list component metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list metadata",
			"",
		))
		return
	}
	defer rows.Close()

	metadata := make([]models.ComponentMetaData, 0)
	for rows.Next() {
		var md models.ComponentMetaData
		err := rows.Scan(
			&md.ID,
			&md.ComponentID,
			&md.Property,
			&md.Value,
			&md.Created,
			&md.Modified,
			&md.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan metadata", zap.Error(err))
			continue
		}
		metadata = append(metadata, md)
	}

	logger.Info("Component metadata listed",
		zap.String("component_id", componentID),
		zap.Int("count", len(metadata)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
		"count":    len(metadata),
	})
	c.JSON(http.StatusOK, response)
}

// handleComponentRemoveMetadata removes metadata from a component
func (h *Handler) handleComponentRemoveMetadata(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "component", models.PermissionUpdate)
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

	// Get component ID and property
	componentID, ok := req.Data["componentId"].(string)
	if !ok || componentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing component ID",
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

	// Soft delete metadata
	query := `UPDATE component_meta_data SET deleted = 1, modified = ? WHERE component_id = ? AND property = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), componentID, property)
	if err != nil {
		logger.Error("Failed to remove component metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove metadata",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Metadata not found",
			"",
		))
		return
	}

	logger.Info("Component metadata removed",
		zap.String("component_id", componentID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":     true,
		"componentId": componentID,
		"property":    property,
	})
	c.JSON(http.StatusOK, response)
}
