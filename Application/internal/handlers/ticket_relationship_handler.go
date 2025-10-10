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

// ===== Ticket Relationship Type CRUD Operations =====

// handleTicketRelationshipTypeCreate creates a new ticket relationship type
func (h *Handler) handleTicketRelationshipTypeCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_relationship_type", models.PermissionCreate)
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

	relType := &models.TicketRelationshipType{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	query := `
		INSERT INTO ticket_relationship_type (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		relType.ID,
		relType.Title,
		relType.Description,
		relType.Created,
		relType.Modified,
		relType.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create ticket relationship type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create ticket relationship type",
			"",
		))
		return
	}

	logger.Info("Ticket relationship type created",
		zap.String("relationship_type_id", relType.ID),
		zap.String("title", relType.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"relationship_type": relType,
	})
	c.JSON(http.StatusCreated, response)
}

// handleTicketRelationshipTypeRead reads a single ticket relationship type by ID
func (h *Handler) handleTicketRelationshipTypeRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	relationshipTypeID, ok := req.Data["id"].(string)
	if !ok || relationshipTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing relationship type ID",
			"",
		))
		return
	}

	query := `
		SELECT id, title, description, created, modified, deleted
		FROM ticket_relationship_type
		WHERE id = ? AND deleted = 0
	`

	var relType models.TicketRelationshipType
	err := h.db.QueryRow(c.Request.Context(), query, relationshipTypeID).Scan(
		&relType.ID,
		&relType.Title,
		&relType.Description,
		&relType.Created,
		&relType.Modified,
		&relType.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket relationship type not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read ticket relationship type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read ticket relationship type",
			"",
		))
		return
	}

	logger.Info("Ticket relationship type read",
		zap.String("relationship_type_id", relType.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"relationship_type": relType,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketRelationshipTypeList lists all ticket relationship types
func (h *Handler) handleTicketRelationshipTypeList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	query := `
		SELECT id, title, description, created, modified, deleted
		FROM ticket_relationship_type
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list ticket relationship types", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list ticket relationship types",
			"",
		))
		return
	}
	defer rows.Close()

	relTypes := make([]models.TicketRelationshipType, 0)
	for rows.Next() {
		var relType models.TicketRelationshipType
		err := rows.Scan(
			&relType.ID,
			&relType.Title,
			&relType.Description,
			&relType.Created,
			&relType.Modified,
			&relType.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan ticket relationship type", zap.Error(err))
			continue
		}
		relTypes = append(relTypes, relType)
	}

	logger.Info("Ticket relationship types listed",
		zap.Int("count", len(relTypes)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"relationship_types": relTypes,
		"count":              len(relTypes),
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketRelationshipTypeModify updates an existing ticket relationship type
func (h *Handler) handleTicketRelationshipTypeModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_relationship_type", models.PermissionUpdate)
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

	relationshipTypeID, ok := req.Data["id"].(string)
	if !ok || relationshipTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing relationship type ID",
			"",
		))
		return
	}

	checkQuery := `SELECT COUNT(*) FROM ticket_relationship_type WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, relationshipTypeID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket relationship type not found",
			"",
		))
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
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"No fields to update",
			"",
		))
		return
	}

	query := "UPDATE ticket_relationship_type SET "
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
	args = append(args, relationshipTypeID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update ticket relationship type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update ticket relationship type",
			"",
		))
		return
	}

	logger.Info("Ticket relationship type updated",
		zap.String("relationship_type_id", relationshipTypeID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      relationshipTypeID,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketRelationshipTypeRemove soft-deletes a ticket relationship type
func (h *Handler) handleTicketRelationshipTypeRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_relationship_type", models.PermissionDelete)
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

	relationshipTypeID, ok := req.Data["id"].(string)
	if !ok || relationshipTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing relationship type ID",
			"",
		))
		return
	}

	query := `UPDATE ticket_relationship_type SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), relationshipTypeID)
	if err != nil {
		logger.Error("Failed to delete ticket relationship type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete ticket relationship type",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket relationship type not found",
			"",
		))
		return
	}

	logger.Info("Ticket relationship type deleted",
		zap.String("relationship_type_id", relationshipTypeID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      relationshipTypeID,
	})
	c.JSON(http.StatusOK, response)
}

// ===== Ticket Relationship Operations =====

// handleTicketRelationshipCreate creates a new relationship between two tickets
func (h *Handler) handleTicketRelationshipCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_relationship", models.PermissionCreate)
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

	ticketID, ok := req.Data["ticket_id"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket_id",
			"",
		))
		return
	}

	childTicketID, ok := req.Data["child_ticket_id"].(string)
	if !ok || childTicketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing child_ticket_id",
			"",
		))
		return
	}

	relationshipTypeID, ok := req.Data["ticket_relationship_type_id"].(string)
	if !ok || relationshipTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket_relationship_type_id",
			"",
		))
		return
	}

	relationship := &models.TicketRelationship{
		ID:                       uuid.New().String(),
		TicketID:                 ticketID,
		ChildTicketID:            childTicketID,
		TicketRelationshipTypeID: relationshipTypeID,
		Created:                  time.Now().Unix(),
		Modified:                 time.Now().Unix(),
		Deleted:                  false,
	}

	query := `
		INSERT INTO ticket_relationship (id, ticket_id, child_ticket_id, ticket_relationship_type_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		relationship.ID,
		relationship.TicketID,
		relationship.ChildTicketID,
		relationship.TicketRelationshipTypeID,
		relationship.Created,
		relationship.Modified,
		relationship.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create ticket relationship", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create ticket relationship",
			"",
		))
		return
	}

	logger.Info("Ticket relationship created",
		zap.String("relationship_id", relationship.ID),
		zap.String("ticket_id", ticketID),
		zap.String("child_ticket_id", childTicketID),
		zap.String("relationship_type_id", relationshipTypeID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"relationship": relationship,
	})
	c.JSON(http.StatusCreated, response)
}

// handleTicketRelationshipRemove removes a relationship between two tickets
func (h *Handler) handleTicketRelationshipRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket_relationship", models.PermissionDelete)
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

	relationshipID, ok := req.Data["id"].(string)
	if !ok || relationshipID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing relationship ID",
			"",
		))
		return
	}

	query := `UPDATE ticket_relationship SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), relationshipID)
	if err != nil {
		logger.Error("Failed to remove ticket relationship", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove ticket relationship",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket relationship not found",
			"",
		))
		return
	}

	logger.Info("Ticket relationship removed",
		zap.String("relationship_id", relationshipID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      relationshipID,
	})
	c.JSON(http.StatusOK, response)
}

// handleTicketRelationshipList lists all relationships for a ticket
func (h *Handler) handleTicketRelationshipList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	ticketID, ok := req.Data["ticket_id"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket_id",
			"",
		))
		return
	}

	query := `
		SELECT id, ticket_id, child_ticket_id, ticket_relationship_type_id, created, modified, deleted
		FROM ticket_relationship
		WHERE (ticket_id = ? OR child_ticket_id = ?) AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID, ticketID)
	if err != nil {
		logger.Error("Failed to list ticket relationships", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list ticket relationships",
			"",
		))
		return
	}
	defer rows.Close()

	relationships := make([]models.TicketRelationship, 0)
	for rows.Next() {
		var relationship models.TicketRelationship
		err := rows.Scan(
			&relationship.ID,
			&relationship.TicketID,
			&relationship.ChildTicketID,
			&relationship.TicketRelationshipTypeID,
			&relationship.Created,
			&relationship.Modified,
			&relationship.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan ticket relationship", zap.Error(err))
			continue
		}
		relationships = append(relationships, relationship)
	}

	logger.Info("Ticket relationships listed",
		zap.String("ticket_id", ticketID),
		zap.Int("count", len(relationships)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"relationships": relationships,
		"count":         len(relationships),
	})
	c.JSON(http.StatusOK, response)
}
