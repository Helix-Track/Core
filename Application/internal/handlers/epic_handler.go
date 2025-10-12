package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
	"helixtrack.ru/core/internal/websocket"
)

// handleEpicCreate creates a new epic (ticket marked as epic)
func (h *Handler) handleEpicCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "ticket", models.PermissionCreate)
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

	// Check if ticket exists
	checkQuery := `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket not found",
			"",
		))
		return
	}

	// Get optional epic details
	epicName := getStringFromData(req.Data, "epicName")
	epicColor := getStringFromData(req.Data, "epicColor")
	if epicColor == "" {
		epicColor = models.EpicColorGhola // Default color (purple)
	}

	// Mark ticket as epic
	query := `
		UPDATE ticket
		SET is_epic = 1, epic_name = ?, epic_color = ?, modified = ?
		WHERE id = ? AND deleted = 0
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		epicName,
		epicColor,
		time.Now().Unix(),
		ticketID,
	)

	if err != nil {
		logger.Error("Failed to create epic", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create epic",
			"",
		))
		return
	}

	logger.Info("Epic created",
		zap.String("ticket_id", ticketID),
		zap.String("epic_name", epicName),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"epicCreate",
		"epic",
		ticketID,
		username,
		map[string]interface{}{
			"ticketId":  ticketID,
			"epicName":  epicName,
			"epicColor": epicColor,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketId":  ticketID,
		"epicName":  epicName,
		"epicColor": epicColor,
		"isEpic":    true,
	})
	c.JSON(http.StatusCreated, response)
}

// handleEpicRead reads an epic's details
func (h *Handler) handleEpicRead(c *gin.Context, req *models.Request) {
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

	// Query epic data
	query := `
		SELECT id, title, is_epic, epic_name, epic_color
		FROM ticket
		WHERE id = ? AND deleted = 0 AND is_epic = 1
	`

	var epic struct {
		ID        string
		Title     string
		IsEpic    bool
		EpicName  sql.NullString
		EpicColor sql.NullString
	}

	err := h.db.QueryRow(c.Request.Context(), query, ticketID).Scan(
		&epic.ID,
		&epic.Title,
		&epic.IsEpic,
		&epic.EpicName,
		&epic.EpicColor,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Epic not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read epic", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read epic",
			"",
		))
		return
	}

	logger.Info("Epic read",
		zap.String("ticket_id", epic.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketId":  epic.ID,
		"title":     epic.Title,
		"isEpic":    epic.IsEpic,
		"epicName":  epic.EpicName.String,
		"epicColor": epic.EpicColor.String,
	})
	c.JSON(http.StatusOK, response)
}

// handleEpicList lists all epics
func (h *Handler) handleEpicList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all epics
	query := `
		SELECT id, title, epic_name, epic_color, created, modified
		FROM ticket
		WHERE deleted = 0 AND is_epic = 1
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list epics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list epics",
			"",
		))
		return
	}
	defer rows.Close()

	epics := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title string
		var epicName, epicColor sql.NullString
		var created, modified int64

		err := rows.Scan(&id, &title, &epicName, &epicColor, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan epic", zap.Error(err))
			continue
		}

		epics = append(epics, map[string]interface{}{
			"ticketId":  id,
			"title":     title,
			"epicName":  epicName.String,
			"epicColor": epicColor.String,
			"created":   created,
			"modified":  modified,
		})
	}

	logger.Info("Epics listed",
		zap.Int("count", len(epics)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"epics": epics,
		"count": len(epics),
	})
	c.JSON(http.StatusOK, response)
}

// handleEpicModify updates an epic's details
func (h *Handler) handleEpicModify(c *gin.Context, req *models.Request) {
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

	// Get ticket ID
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	// Check if epic exists
	checkQuery := `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0 AND is_epic = 1`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Epic not found",
			"",
		))
		return
	}

	// Build update query dynamically
	updates := make(map[string]interface{})

	if epicName, ok := req.Data["epicName"].(string); ok {
		updates["epic_name"] = epicName
	}
	if epicColor, ok := req.Data["epicColor"].(string); ok {
		updates["epic_color"] = epicColor
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
	query := "UPDATE ticket SET "
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
	args = append(args, ticketID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update epic", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update epic",
			"",
		))
		return
	}

	logger.Info("Epic updated",
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"epicModify",
		"epic",
		ticketID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated":  true,
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleEpicRemove removes epic status from a ticket
func (h *Handler) handleEpicRemove(c *gin.Context, req *models.Request) {
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

	// Get ticket ID
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	// Remove epic status and clear epic-specific fields
	query := `
		UPDATE ticket
		SET is_epic = 0, epic_name = NULL, epic_color = NULL, modified = ?
		WHERE id = ? AND deleted = 0 AND is_epic = 1
	`

	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), ticketID)
	if err != nil {
		logger.Error("Failed to remove epic", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove epic",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Epic not found",
			"",
		))
		return
	}

	// Also unlink all stories from this epic
	_, err = h.db.Exec(c.Request.Context(),
		"UPDATE ticket SET epic_id = NULL WHERE epic_id = ?",
		ticketID,
	)
	if err != nil {
		logger.Error("Failed to unlink stories from epic", zap.Error(err))
		// Don't fail the request, epic removal succeeded
	}

	logger.Info("Epic removed",
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"epicRemove",
		"epic",
		ticketID,
		username,
		map[string]interface{}{
			"ticketId": ticketID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":  true,
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleEpicAddStory adds a story (ticket) to an epic
func (h *Handler) handleEpicAddStory(c *gin.Context, req *models.Request) {
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

	// Get epic ID (ticket marked as epic)
	epicID, ok := req.Data["epicId"].(string)
	if !ok || epicID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing epicId",
			"",
		))
		return
	}

	// Get story ID (ticket to be added)
	storyID, ok := req.Data["storyId"].(string)
	if !ok || storyID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing storyId",
			"",
		))
		return
	}

	// Verify epic exists and is marked as epic
	checkQuery := `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0 AND is_epic = 1`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, epicID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Epic not found",
			"",
		))
		return
	}

	// Verify story exists
	checkQuery = `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, storyID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Story ticket not found",
			"",
		))
		return
	}

	// Link story to epic
	query := `UPDATE ticket SET epic_id = ?, modified = ? WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), query, epicID, time.Now().Unix(), storyID)
	if err != nil {
		logger.Error("Failed to add story to epic", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add story to epic",
			"",
		))
		return
	}

	logger.Info("Story added to epic",
		zap.String("epic_id", epicID),
		zap.String("story_id", storyID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"epicAddStory",
		"epic",
		epicID,
		username,
		map[string]interface{}{
			"epicId":  epicID,
			"storyId": storyID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":   true,
		"epicId":  epicID,
		"storyId": storyID,
	})
	c.JSON(http.StatusOK, response)
}

// handleEpicRemoveStory removes a story from an epic
func (h *Handler) handleEpicRemoveStory(c *gin.Context, req *models.Request) {
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

	// Get story ID
	storyID, ok := req.Data["storyId"].(string)
	if !ok || storyID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing storyId",
			"",
		))
		return
	}

	// Unlink story from epic
	query := `UPDATE ticket SET epic_id = NULL, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), storyID)
	if err != nil {
		logger.Error("Failed to remove story from epic", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove story from epic",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Story not found",
			"",
		))
		return
	}

	logger.Info("Story removed from epic",
		zap.String("story_id", storyID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"epicRemoveStory",
		"epic",
		storyID,
		username,
		map[string]interface{}{
			"storyId": storyID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed": true,
		"storyId": storyID,
	})
	c.JSON(http.StatusOK, response)
}

// handleEpicListStories lists all stories in an epic
func (h *Handler) handleEpicListStories(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get epic ID
	epicID, ok := req.Data["epicId"].(string)
	if !ok || epicID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing epicId",
			"",
		))
		return
	}

	// Query stories linked to this epic
	query := `
		SELECT id, title, ticket_key, status_id, created, modified
		FROM ticket
		WHERE epic_id = ? AND deleted = 0
		ORDER BY created ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, epicID)
	if err != nil {
		logger.Error("Failed to list epic stories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list epic stories",
			"",
		))
		return
	}
	defer rows.Close()

	stories := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title, ticketKey string
		var statusID sql.NullString
		var created, modified int64

		err := rows.Scan(&id, &title, &ticketKey, &statusID, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan story", zap.Error(err))
			continue
		}

		stories = append(stories, map[string]interface{}{
			"id":        id,
			"title":     title,
			"ticketKey": ticketKey,
			"statusId":  statusID.String,
			"created":   created,
			"modified":  modified,
		})
	}

	logger.Info("Epic stories listed",
		zap.String("epic_id", epicID),
		zap.Int("count", len(stories)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"epicId":  epicID,
		"stories": stories,
		"count":   len(stories),
	})
	c.JSON(http.StatusOK, response)
}
