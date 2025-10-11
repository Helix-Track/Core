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

// handleLabelCreate creates a new label
func (h *Handler) handleLabelCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionCreate)
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

	// Parse label data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	label := &models.Label{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Color:       getStringFromData(req.Data, "color"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO label (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		label.ID,
		label.Title,
		label.Description,
		label.Created,
		label.Modified,
		label.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create label",
			"",
		))
		return
	}

	logger.Info("Label created",
		zap.String("label_id", label.ID),
		zap.String("title", label.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"label": label,
	})
	c.JSON(http.StatusCreated, response)
}

// handleLabelRead reads a single label by ID
func (h *Handler) handleLabelRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get label ID from request
	labelID, ok := req.Data["id"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
			"",
		))
		return
	}

	// Query label from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM label
		WHERE id = ? AND deleted = 0
	`

	var label models.Label
	err := h.db.QueryRow(c.Request.Context(), query, labelID).Scan(
		&label.ID,
		&label.Title,
		&label.Description,
		&label.Created,
		&label.Modified,
		&label.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read label",
			"",
		))
		return
	}

	logger.Info("Label read",
		zap.String("label_id", label.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"label": label,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelList lists all labels
func (h *Handler) handleLabelList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted labels ordered by title
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM label
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list labels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list labels",
			"",
		))
		return
	}
	defer rows.Close()

	labels := make([]models.Label, 0)
	for rows.Next() {
		var label models.Label
		err := rows.Scan(
			&label.ID,
			&label.Title,
			&label.Description,
			&label.Created,
			&label.Modified,
			&label.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan label", zap.Error(err))
			continue
		}
		labels = append(labels, label)
	}

	logger.Info("Labels listed",
		zap.Int("count", len(labels)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"labels": labels,
		"count":  len(labels),
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelModify updates an existing label
func (h *Handler) handleLabelModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionUpdate)
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

	// Get label ID
	labelID, ok := req.Data["id"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
			"",
		))
		return
	}

	// Check if label exists
	checkQuery := `SELECT COUNT(*) FROM label WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, labelID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label not found",
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
	query := "UPDATE label SET "
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
	args = append(args, labelID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update label",
			"",
		))
		return
	}

	logger.Info("Label updated",
		zap.String("label_id", labelID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      labelID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelRemove soft-deletes a label
func (h *Handler) handleLabelRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionDelete)
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

	// Get label ID
	labelID, ok := req.Data["id"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
			"",
		))
		return
	}

	// Soft delete the label
	query := `UPDATE label SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), labelID)
	if err != nil {
		logger.Error("Failed to delete label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete label",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label not found",
			"",
		))
		return
	}

	logger.Info("Label deleted",
		zap.String("label_id", labelID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      labelID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelCategoryCreate creates a new label category
func (h *Handler) handleLabelCategoryCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionCreate)
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

	// Parse category data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	category := &models.LabelCategory{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO label_category (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		category.ID,
		category.Title,
		category.Description,
		category.Created,
		category.Modified,
		category.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create label category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create label category",
			"",
		))
		return
	}

	logger.Info("Label category created",
		zap.String("category_id", category.ID),
		zap.String("title", category.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"category": category,
	})
	c.JSON(http.StatusCreated, response)
}

// handleLabelCategoryRead reads a single label category by ID
func (h *Handler) handleLabelCategoryRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get category ID from request
	categoryID, ok := req.Data["id"].(string)
	if !ok || categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing category ID",
			"",
		))
		return
	}

	// Query category from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM label_category
		WHERE id = ? AND deleted = 0
	`

	var category models.LabelCategory
	err := h.db.QueryRow(c.Request.Context(), query, categoryID).Scan(
		&category.ID,
		&category.Title,
		&category.Description,
		&category.Created,
		&category.Modified,
		&category.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label category not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read label category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read label category",
			"",
		))
		return
	}

	logger.Info("Label category read",
		zap.String("category_id", category.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"category": category,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelCategoryList lists all label categories
func (h *Handler) handleLabelCategoryList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted categories ordered by title
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM label_category
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list label categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list label categories",
			"",
		))
		return
	}
	defer rows.Close()

	categories := make([]models.LabelCategory, 0)
	for rows.Next() {
		var category models.LabelCategory
		err := rows.Scan(
			&category.ID,
			&category.Title,
			&category.Description,
			&category.Created,
			&category.Modified,
			&category.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan label category", zap.Error(err))
			continue
		}
		categories = append(categories, category)
	}

	logger.Info("Label categories listed",
		zap.Int("count", len(categories)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelCategoryModify updates an existing label category
func (h *Handler) handleLabelCategoryModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionUpdate)
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

	// Get category ID
	categoryID, ok := req.Data["id"].(string)
	if !ok || categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing category ID",
			"",
		))
		return
	}

	// Check if category exists
	checkQuery := `SELECT COUNT(*) FROM label_category WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, categoryID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label category not found",
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
	query := "UPDATE label_category SET "
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
	args = append(args, categoryID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update label category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update label category",
			"",
		))
		return
	}

	logger.Info("Label category updated",
		zap.String("category_id", categoryID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      categoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelCategoryRemove soft-deletes a label category
func (h *Handler) handleLabelCategoryRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionDelete)
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

	// Get category ID
	categoryID, ok := req.Data["id"].(string)
	if !ok || categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing category ID",
			"",
		))
		return
	}

	// Soft delete the category
	query := `UPDATE label_category SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), categoryID)
	if err != nil {
		logger.Error("Failed to delete label category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete label category",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label category not found",
			"",
		))
		return
	}

	logger.Info("Label category deleted",
		zap.String("category_id", categoryID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      categoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelAddTicket adds a label to a ticket
func (h *Handler) handleLabelAddTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionUpdate)
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

	// Get label ID and ticket ID
	labelID, ok := req.Data["labelId"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
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
		INSERT INTO label_ticket_mapping (id, label_id, ticket_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err = h.db.Exec(c.Request.Context(), query, mappingID, labelID, ticketID, now, now, false)
	if err != nil {
		logger.Error("Failed to add label to ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add label to ticket",
			"",
		))
		return
	}

	logger.Info("Label added to ticket",
		zap.String("label_id", labelID),
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":    true,
		"labelId":  labelID,
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelRemoveTicket removes a label from a ticket
func (h *Handler) handleLabelRemoveTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionUpdate)
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

	// Get label ID and ticket ID
	labelID, ok := req.Data["labelId"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
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
	query := `UPDATE label_ticket_mapping SET deleted = 1, modified = ? WHERE label_id = ? AND ticket_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), labelID, ticketID)
	if err != nil {
		logger.Error("Failed to remove label from ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove label from ticket",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label-ticket mapping not found",
			"",
		))
		return
	}

	logger.Info("Label removed from ticket",
		zap.String("label_id", labelID),
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":  true,
		"labelId":  labelID,
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelListTickets lists all tickets for a label
func (h *Handler) handleLabelListTickets(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get label ID
	labelID, ok := req.Data["labelId"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
			"",
		))
		return
	}

	// Query tickets
	query := `
		SELECT ticket_id
		FROM label_ticket_mapping
		WHERE label_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, labelID)
	if err != nil {
		logger.Error("Failed to list tickets for label", zap.Error(err))
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

	logger.Info("Tickets listed for label",
		zap.String("label_id", labelID),
		zap.Int("count", len(ticketIDs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketIds": ticketIDs,
		"count":     len(ticketIDs),
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelAssignCategory assigns a label to a category
func (h *Handler) handleLabelAssignCategory(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionUpdate)
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

	// Get label ID and category ID
	labelID, ok := req.Data["labelId"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
			"",
		))
		return
	}

	categoryID, ok := req.Data["categoryId"].(string)
	if !ok || categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing category ID",
			"",
		))
		return
	}

	// Create mapping
	mappingID := uuid.New().String()
	query := `
		INSERT INTO label_label_category_mapping (id, label_id, label_category_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err = h.db.Exec(c.Request.Context(), query, mappingID, labelID, categoryID, now, now, false)
	if err != nil {
		logger.Error("Failed to assign label to category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign label to category",
			"",
		))
		return
	}

	logger.Info("Label assigned to category",
		zap.String("label_id", labelID),
		zap.String("category_id", categoryID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"assigned":   true,
		"labelId":    labelID,
		"categoryId": categoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelUnassignCategory unassigns a label from a category
func (h *Handler) handleLabelUnassignCategory(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "label", models.PermissionUpdate)
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

	// Get label ID and category ID
	labelID, ok := req.Data["labelId"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
			"",
		))
		return
	}

	categoryID, ok := req.Data["categoryId"].(string)
	if !ok || categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing category ID",
			"",
		))
		return
	}

	// Remove mapping (soft delete)
	query := `UPDATE label_label_category_mapping SET deleted = 1, modified = ? WHERE label_id = ? AND label_category_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), labelID, categoryID)
	if err != nil {
		logger.Error("Failed to unassign label from category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unassign label from category",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Label-category mapping not found",
			"",
		))
		return
	}

	logger.Info("Label unassigned from category",
		zap.String("label_id", labelID),
		zap.String("category_id", categoryID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"unassigned": true,
		"labelId":    labelID,
		"categoryId": categoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleLabelListCategories lists all categories for a label
func (h *Handler) handleLabelListCategories(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get label ID
	labelID, ok := req.Data["labelId"].(string)
	if !ok || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing label ID",
			"",
		))
		return
	}

	// Query categories
	query := `
		SELECT label_category_id
		FROM label_label_category_mapping
		WHERE label_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, labelID)
	if err != nil {
		logger.Error("Failed to list categories for label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list categories",
			"",
		))
		return
	}
	defer rows.Close()

	categoryIDs := make([]string, 0)
	for rows.Next() {
		var categoryID string
		if err := rows.Scan(&categoryID); err != nil {
			logger.Error("Failed to scan category ID", zap.Error(err))
			continue
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	logger.Info("Categories listed for label",
		zap.String("label_id", labelID),
		zap.Int("count", len(categoryIDs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"categoryIds": categoryIDs,
		"count":       len(categoryIDs),
	})
	c.JSON(http.StatusOK, response)
}
