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

// handleProjectCategoryCreate creates a new project category
func (h *Handler) handleProjectCategoryCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_category", models.PermissionCreate)
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

	// Extract and validate data
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	category := &models.ProjectCategory{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO project_category (id, title, description, created, modified, deleted)
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
		logger.Error("Failed to create project category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create project category",
			"",
		))
		return
	}

	logger.Info("Project category created",
		zap.String("category_id", category.ID),
		zap.String("title", category.Title),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectCategoryCreate",
		"project_category",
		category.ID,
		username,
		map[string]interface{}{
			"id":          category.ID,
			"title":       category.Title,
			"description": category.Description,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"category": category,
	})
	c.JSON(http.StatusCreated, response)
}

// handleProjectCategoryRead reads a single project category by ID
func (h *Handler) handleProjectCategoryRead(c *gin.Context, req *models.Request) {
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
		FROM project_category
		WHERE id = ? AND deleted = 0
	`

	var category models.ProjectCategory
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
			"Project category not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read project category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read project category",
			"",
		))
		return
	}

	logger.Info("Project category read",
		zap.String("category_id", category.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"category": category,
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectCategoryList lists all project categories
func (h *Handler) handleProjectCategoryList(c *gin.Context, req *models.Request) {
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
		FROM project_category
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list project categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list project categories",
			"",
		))
		return
	}
	defer rows.Close()

	categories := make([]models.ProjectCategory, 0)
	for rows.Next() {
		var category models.ProjectCategory
		err := rows.Scan(
			&category.ID,
			&category.Title,
			&category.Description,
			&category.Created,
			&category.Modified,
			&category.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan project category", zap.Error(err))
			continue
		}
		categories = append(categories, category)
	}

	logger.Info("Project categories listed",
		zap.Int("count", len(categories)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectCategoryModify updates an existing project category
func (h *Handler) handleProjectCategoryModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_category", models.PermissionUpdate)
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
	checkQuery := `SELECT COUNT(*) FROM project_category WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, categoryID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project category not found",
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
	query := "UPDATE project_category SET "
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
		logger.Error("Failed to update project category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update project category",
			"",
		))
		return
	}

	logger.Info("Project category updated",
		zap.String("category_id", categoryID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectCategoryModify",
		"project_category",
		categoryID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      categoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectCategoryRemove soft-deletes a project category
func (h *Handler) handleProjectCategoryRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_category", models.PermissionDelete)
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

	// Read category data before deleting (for event publishing)
	readQuery := `SELECT id, title, description FROM project_category WHERE id = ? AND deleted = 0`
	var category models.ProjectCategory
	err = h.db.QueryRow(c.Request.Context(), readQuery, categoryID).Scan(
		&category.ID,
		&category.Title,
		&category.Description,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project category not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to read project category before deletion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read project category",
			"",
		))
		return
	}

	// Soft delete the category
	query := `UPDATE project_category SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), categoryID)
	if err != nil {
		logger.Error("Failed to delete project category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete project category",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project category not found",
			"",
		))
		return
	}

	logger.Info("Project category deleted",
		zap.String("category_id", categoryID),
		zap.String("username", username),
	)

	// Publish event with full data
	h.publisher.PublishEntityEvent(
		"projectCategoryRemove",
		"project_category",
		categoryID,
		username,
		map[string]interface{}{
			"id":          category.ID,
			"title":       category.Title,
			"description": category.Description,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      categoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectCategoryAssign assigns a category to a project
func (h *Handler) handleProjectCategoryAssign(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project", models.PermissionUpdate)
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

	// Get project ID
	projectID, ok := req.Data["projectId"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing projectId",
			"",
		))
		return
	}

	// Get category ID
	categoryID, ok := req.Data["categoryId"].(string)
	if !ok || categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing categoryId",
			"",
		))
		return
	}

	// Check if project exists
	checkQuery := `SELECT COUNT(*) FROM project WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, projectID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project not found",
			"",
		))
		return
	}

	// Check if category exists
	err = h.db.QueryRow(c.Request.Context(), checkQuery, categoryID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project category not found",
			"",
		))
		return
	}

	// Assign category to project
	query := `UPDATE project SET project_category_id = ? WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), query, categoryID, projectID)
	if err != nil {
		logger.Error("Failed to assign category to project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign category to project",
			"",
		))
		return
	}

	logger.Info("Category assigned to project",
		zap.String("project_id", projectID),
		zap.String("category_id", categoryID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectCategoryAssign",
		"project",
		projectID,
		username,
		map[string]interface{}{
			"projectId":  projectID,
			"categoryId": categoryID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"assigned":   true,
		"projectId":  projectID,
		"categoryId": categoryID,
	})
	c.JSON(http.StatusOK, response)
}
