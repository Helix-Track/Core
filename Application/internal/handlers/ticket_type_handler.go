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

// handleTicketTypeCreate creates a new ticket type
func (h *Handler) handleTicketTypeCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_type", models.PermissionCreate)
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

	// Parse ticket type data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	ticketType := &models.TicketType{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Icon:        getStringFromData(req.Data, "icon"),
		Color:       getStringFromData(req.Data, "color"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO ticket_type (id, title, description, icon, color, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		ticketType.ID,
		ticketType.Title,
		ticketType.Description,
		ticketType.Icon,
		ticketType.Color,
		ticketType.Created,
		ticketType.Modified,
		ticketType.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create ticket type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create ticket type",
			"",
		))
		return
	}

	logger.Info("Ticket type created",
		zap.String("ticket_type_id", ticketType.ID),
		zap.String("title", ticketType.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketType": ticketType,
	})
	c.JSON(http.StatusCreated, response)
}

// handleTicketTypeRead reads a single ticket type by ID
func (h *Handler) handleTicketTypeRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get ticket type ID from request
	ticketTypeID, ok := req.Data["id"].(string)
	if !ok || ticketTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket type ID",
			"",
		))
		return
	}

	// Query ticket type from database
	query := `
		SELECT id, title, description, icon, color, created, modified, deleted
		FROM ticket_type
		WHERE id = ? AND deleted = 0
	`

	var ticketType models.TicketType
	err := h.db.QueryRow(c.Request.Context(), query, ticketTypeID).Scan(
		&ticketType.ID,
		&ticketType.Title,
		&ticketType.Description,
		&ticketType.Icon,
		&ticketType.Color,
		&ticketType.Created,
		&ticketType.Modified,
		&ticketType.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket type not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read ticket type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read ticket type",
			"",
		))
		return
	}

	logger.Info("Ticket type read",
		zap.String("ticket_type_id", ticketType.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketType": ticketType,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketTypeList lists all ticket types
func (h *Handler) handleTicketTypeList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted ticket types ordered by title
	query := `
		SELECT id, title, description, icon, color, created, modified, deleted
		FROM ticket_type
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list ticket types", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list ticket types",
			"",
		))
		return
	}
	defer rows.Close()

	ticketTypes := make([]models.TicketType, 0)
	for rows.Next() {
		var ticketType models.TicketType
		err := rows.Scan(
			&ticketType.ID,
			&ticketType.Title,
			&ticketType.Description,
			&ticketType.Icon,
			&ticketType.Color,
			&ticketType.Created,
			&ticketType.Modified,
			&ticketType.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan ticket type", zap.Error(err))
			continue
		}
		ticketTypes = append(ticketTypes, ticketType)
	}

	logger.Info("Ticket types listed",
		zap.Int("count", len(ticketTypes)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketTypes": ticketTypes,
		"count":       len(ticketTypes),
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketTypeModify updates an existing ticket type
func (h *Handler) handleTicketTypeModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_type", models.PermissionUpdate)
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

	// Get ticket type ID
	ticketTypeID, ok := req.Data["id"].(string)
	if !ok || ticketTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket type ID",
			"",
		))
		return
	}

	// Check if ticket type exists
	checkQuery := `SELECT COUNT(*) FROM ticket_type WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketTypeID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket type not found",
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
	if icon, ok := req.Data["icon"].(string); ok {
		updates["icon"] = icon
	}
	if color, ok := req.Data["color"].(string); ok {
		updates["color"] = color
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
	query := "UPDATE ticket_type SET "
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
	args = append(args, ticketTypeID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update ticket type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update ticket type",
			"",
		))
		return
	}

	logger.Info("Ticket type updated",
		zap.String("ticket_type_id", ticketTypeID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      ticketTypeID,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketTypeRemove soft-deletes a ticket type
func (h *Handler) handleTicketTypeRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_type", models.PermissionDelete)
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

	// Get ticket type ID
	ticketTypeID, ok := req.Data["id"].(string)
	if !ok || ticketTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket type ID",
			"",
		))
		return
	}

	// Soft delete the ticket type
	query := `UPDATE ticket_type SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), ticketTypeID)
	if err != nil {
		logger.Error("Failed to delete ticket type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete ticket type",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket type not found",
			"",
		))
		return
	}

	logger.Info("Ticket type deleted",
		zap.String("ticket_type_id", ticketTypeID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      ticketTypeID,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketTypeAssign assigns a ticket type to a project
func (h *Handler) handleTicketTypeAssign(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_type", models.PermissionCreate)
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
	ticketTypeID, ok := req.Data["ticketTypeId"].(string)
	if !ok || ticketTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketTypeId",
			"",
		))
		return
	}

	projectID, ok := req.Data["projectId"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing projectId",
			"",
		))
		return
	}

	// Check if mapping already exists
	checkQuery := `SELECT COUNT(*) FROM ticket_type_project_mapping WHERE ticket_type_id = ? AND project_id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketTypeID, projectID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check existing mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check existing mapping",
			"",
		))
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeAlreadyExists,
			"Ticket type already assigned to this project",
			"",
		))
		return
	}

	mapping := &models.TicketTypeProjectMapping{
		ID:           uuid.New().String(),
		TicketTypeID: ticketTypeID,
		ProjectID:    projectID,
		Created:      time.Now().Unix(),
		Deleted:      false,
	}

	// Insert into database
	query := `
		INSERT INTO ticket_type_project_mapping (id, ticket_type_id, project_id, created, deleted)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.TicketTypeID,
		mapping.ProjectID,
		mapping.Created,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to assign ticket type to project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign ticket type to project",
			"",
		))
		return
	}

	logger.Info("Ticket type assigned to project",
		zap.String("ticket_type_id", ticketTypeID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"assigned": true,
		"mapping":  mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleTicketTypeUnassign unassigns a ticket type from a project
func (h *Handler) handleTicketTypeUnassign(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_type", models.PermissionDelete)
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
	ticketTypeID, ok := req.Data["ticketTypeId"].(string)
	if !ok || ticketTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketTypeId",
			"",
		))
		return
	}

	projectID, ok := req.Data["projectId"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing projectId",
			"",
		))
		return
	}

	// Soft delete the mapping
	query := `UPDATE ticket_type_project_mapping SET deleted = 1 WHERE ticket_type_id = ? AND project_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, ticketTypeID, projectID)
	if err != nil {
		logger.Error("Failed to unassign ticket type from project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unassign ticket type from project",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket type assignment not found",
			"",
		))
		return
	}

	logger.Info("Ticket type unassigned from project",
		zap.String("ticket_type_id", ticketTypeID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"unassigned": true,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketTypeListByProject lists ticket types assigned to a project
func (h *Handler) handleTicketTypeListByProject(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get project ID from request
	projectID, ok := req.Data["projectId"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing projectId",
			"",
		))
		return
	}

	// Query ticket types assigned to the project
	query := `
		SELECT tt.id, tt.title, tt.description, tt.icon, tt.color, tt.created, tt.modified, tt.deleted
		FROM ticket_type tt
		INNER JOIN ticket_type_project_mapping ttpm ON tt.id = ttpm.ticket_type_id
		WHERE ttpm.project_id = ? AND tt.deleted = 0 AND ttpm.deleted = 0
		ORDER BY tt.title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, projectID)
	if err != nil {
		logger.Error("Failed to list ticket types by project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list ticket types by project",
			"",
		))
		return
	}
	defer rows.Close()

	ticketTypes := make([]models.TicketType, 0)
	for rows.Next() {
		var ticketType models.TicketType
		err := rows.Scan(
			&ticketType.ID,
			&ticketType.Title,
			&ticketType.Description,
			&ticketType.Icon,
			&ticketType.Color,
			&ticketType.Created,
			&ticketType.Modified,
			&ticketType.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan ticket type", zap.Error(err))
			continue
		}
		ticketTypes = append(ticketTypes, ticketType)
	}

	logger.Info("Ticket types by project listed",
		zap.String("project_id", projectID),
		zap.Int("count", len(ticketTypes)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketTypes": ticketTypes,
		"count":       len(ticketTypes),
		"projectId":   projectID,
	})
	c.JSON(http.StatusOK, response)
}
