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

// handleFilterSave creates a new filter or updates an existing one
func (h *Handler) handleFilterSave(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get user ID from context (should be available from JWT)
	userID := username // In a real system, this would be the actual user ID from JWT claims

	// Parse filter data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	query, ok := req.Data["query"].(string)
	if !ok || query == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing query",
			"",
		))
		return
	}

	// Validate query is valid JSON
	var queryObj interface{}
	if err := json.Unmarshal([]byte(query), &queryObj); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid query format (must be valid JSON)",
			"",
		))
		return
	}

	// Check if this is an update (ID provided) or create (no ID)
	filterID, hasID := req.Data["id"].(string)
	isUpdate := hasID && filterID != ""

	if isUpdate {
		// Update existing filter - check ownership first
		ownerCheckQuery := `SELECT owner_id FROM filter WHERE id = ? AND deleted = 0`
		var ownerID string
		err := h.db.QueryRow(c.Request.Context(), ownerCheckQuery, filterID).Scan(&ownerID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrorCodeEntityNotFound,
				"Filter not found",
				"",
			))
			return
		}
		if err != nil {
			logger.Error("Failed to check filter ownership", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to check filter ownership",
				"",
			))
			return
		}

		// Verify ownership
		if ownerID != userID {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrorCodeForbidden,
				"You can only modify your own filters",
				"",
			))
			return
		}

		// Update the filter
		updateQuery := `
			UPDATE filter
			SET title = ?, description = ?, query = ?, is_public = ?, is_favorite = ?, modified = ?
			WHERE id = ? AND deleted = 0
		`

		description := getStringFromData(req.Data, "description")
		isPublic := getBoolFromData(req.Data, "isPublic")
		isFavorite := getBoolFromData(req.Data, "isFavorite")

		_, err = h.db.Exec(c.Request.Context(), updateQuery,
			title,
			description,
			query,
			isPublic,
			isFavorite,
			time.Now().Unix(),
			filterID,
		)

		if err != nil {
			logger.Error("Failed to update filter", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to update filter",
				"",
			))
			return
		}

		logger.Info("Filter updated",
			zap.String("filter_id", filterID),
			zap.String("title", title),
			zap.String("username", username),
		)

		response := models.NewSuccessResponse(map[string]interface{}{
			"filter": map[string]interface{}{
				"id":          filterID,
				"title":       title,
				"description": description,
				"ownerId":     userID,
				"query":       query,
				"isPublic":    isPublic,
				"isFavorite":  isFavorite,
				"modified":    time.Now().Unix(),
			},
		})
		c.JSON(http.StatusOK, response)
		return
	}

	// Create new filter
	filter := &models.Filter{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		OwnerID:     userID,
		Query:       query,
		IsPublic:    getBoolFromData(req.Data, "isPublic"),
		IsFavorite:  getBoolFromData(req.Data, "isFavorite"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	insertQuery := `
		INSERT INTO filter (id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(c.Request.Context(), insertQuery,
		filter.ID,
		filter.Title,
		filter.Description,
		filter.OwnerID,
		filter.Query,
		filter.IsPublic,
		filter.IsFavorite,
		filter.Created,
		filter.Modified,
		filter.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create filter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create filter",
			"",
		))
		return
	}

	logger.Info("Filter created",
		zap.String("filter_id", filter.ID),
		zap.String("title", filter.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"filter": filter,
	})
	c.JSON(http.StatusCreated, response)
}

// handleFilterLoad loads a filter by ID
func (h *Handler) handleFilterLoad(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	userID := username // In a real system, this would be the actual user ID from JWT claims

	// Get filter ID from request
	filterID, ok := req.Data["id"].(string)
	if !ok || filterID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing filter ID",
			"",
		))
		return
	}

	// Query filter from database
	query := `
		SELECT id, title, description, owner_id, query, is_public, is_favorite, created, modified, deleted
		FROM filter
		WHERE id = ? AND deleted = 0
	`

	var filter models.Filter
	err := h.db.QueryRow(c.Request.Context(), query, filterID).Scan(
		&filter.ID,
		&filter.Title,
		&filter.Description,
		&filter.OwnerID,
		&filter.Query,
		&filter.IsPublic,
		&filter.IsFavorite,
		&filter.Created,
		&filter.Modified,
		&filter.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Filter not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to load filter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to load filter",
			"",
		))
		return
	}

	// Check if user has access to this filter
	// User has access if: they own it, it's public, or it's shared with them
	hasAccess := filter.OwnerID == userID || filter.IsPublic

	if !hasAccess {
		// Check if shared with user
		shareQuery := `
			SELECT COUNT(*)
			FROM filter_share_mapping
			WHERE filter_id = ? AND user_id = ? AND deleted = 0
		`
		var shareCount int
		err = h.db.QueryRow(c.Request.Context(), shareQuery, filterID, userID).Scan(&shareCount)
		if err == nil && shareCount > 0 {
			hasAccess = true
		}
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"You do not have access to this filter",
			"",
		))
		return
	}

	logger.Info("Filter loaded",
		zap.String("filter_id", filter.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"filter": filter,
	})
	c.JSON(http.StatusOK, response)
}

// handleFilterList lists all filters accessible to the user
func (h *Handler) handleFilterList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	userID := username // In a real system, this would be the actual user ID from JWT claims

	// Query all filters that the user has access to:
	// 1. Filters owned by the user
	// 2. Public filters
	// 3. Filters shared with the user
	query := `
		SELECT DISTINCT f.id, f.title, f.description, f.owner_id, f.query, f.is_public, f.is_favorite, f.created, f.modified, f.deleted
		FROM filter f
		LEFT JOIN filter_share_mapping fsm ON f.id = fsm.filter_id AND fsm.deleted = 0
		WHERE f.deleted = 0
		AND (
			f.owner_id = ?           -- User's own filters
			OR f.is_public = 1       -- Public filters
			OR fsm.user_id = ?       -- Shared with user
		)
		ORDER BY f.is_favorite DESC, f.modified DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, userID, userID)
	if err != nil {
		logger.Error("Failed to list filters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list filters",
			"",
		))
		return
	}
	defer rows.Close()

	filters := make([]models.Filter, 0)
	for rows.Next() {
		var filter models.Filter
		err := rows.Scan(
			&filter.ID,
			&filter.Title,
			&filter.Description,
			&filter.OwnerID,
			&filter.Query,
			&filter.IsPublic,
			&filter.IsFavorite,
			&filter.Created,
			&filter.Modified,
			&filter.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan filter", zap.Error(err))
			continue
		}
		filters = append(filters, filter)
	}

	logger.Info("Filters listed",
		zap.Int("count", len(filters)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"filters": filters,
		"count":   len(filters),
	})
	c.JSON(http.StatusOK, response)
}

// handleFilterShare shares a filter with a user, team, or project
func (h *Handler) handleFilterShare(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	userID := username // In a real system, this would be the actual user ID from JWT claims

	// Get filter ID
	filterID, ok := req.Data["filterId"].(string)
	if !ok || filterID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing filter ID",
			"",
		))
		return
	}

	// Check if filter exists and user owns it
	ownerCheckQuery := `SELECT owner_id FROM filter WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), ownerCheckQuery, filterID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Filter not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to check filter ownership", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check filter ownership",
			"",
		))
		return
	}

	// Verify ownership
	if ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"You can only share your own filters",
			"",
		))
		return
	}

	// Determine share type
	shareType, ok := req.Data["shareType"].(string)
	if !ok || shareType == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing share type (user, team, project, or public)",
			"",
		))
		return
	}

	// Handle public sharing separately
	if shareType == string(models.ShareTypePublic) {
		updateQuery := `UPDATE filter SET is_public = 1, modified = ? WHERE id = ?`
		_, err = h.db.Exec(c.Request.Context(), updateQuery, time.Now().Unix(), filterID)
		if err != nil {
			logger.Error("Failed to make filter public", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to make filter public",
				"",
			))
			return
		}

		logger.Info("Filter made public",
			zap.String("filter_id", filterID),
			zap.String("username", username),
		)

		response := models.NewSuccessResponse(map[string]interface{}{
			"shared":     true,
			"filterId":   filterID,
			"shareType":  shareType,
			"isPublic":   true,
		})
		c.JSON(http.StatusOK, response)
		return
	}

	// Create share mapping
	shareMapping := &models.FilterShareMapping{
		ID:       uuid.New().String(),
		FilterID: filterID,
		Created:  time.Now().Unix(),
		Deleted:  false,
	}

	// Set the appropriate ID based on share type
	switch models.ShareType(shareType) {
	case models.ShareTypeUser:
		shareUserID, ok := req.Data["userId"].(string)
		if !ok || shareUserID == "" {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeMissingData,
				"Missing user ID for user share",
				"",
			))
			return
		}
		shareMapping.UserID = &shareUserID

	case models.ShareTypeTeam:
		shareTeamID, ok := req.Data["teamId"].(string)
		if !ok || shareTeamID == "" {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeMissingData,
				"Missing team ID for team share",
				"",
			))
			return
		}
		shareMapping.TeamID = &shareTeamID

	case models.ShareTypeProject:
		shareProjectID, ok := req.Data["projectId"].(string)
		if !ok || shareProjectID == "" {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeMissingData,
				"Missing project ID for project share",
				"",
			))
			return
		}
		shareMapping.ProjectID = &shareProjectID

	default:
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid share type (must be user, team, project, or public)",
			"",
		))
		return
	}

	// Insert share mapping
	insertQuery := `
		INSERT INTO filter_share_mapping (id, filter_id, user_id, team_id, project_id, created, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), insertQuery,
		shareMapping.ID,
		shareMapping.FilterID,
		shareMapping.UserID,
		shareMapping.TeamID,
		shareMapping.ProjectID,
		shareMapping.Created,
		shareMapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to share filter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to share filter",
			"",
		))
		return
	}

	logger.Info("Filter shared",
		zap.String("filter_id", filterID),
		zap.String("share_type", shareType),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"shared":    true,
		"filterId":  filterID,
		"shareType": shareType,
		"shareId":   shareMapping.ID,
	})
	c.JSON(http.StatusOK, response)
}

// handleFilterModify modifies an existing filter
func (h *Handler) handleFilterModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	userID := username // In a real system, this would be the actual user ID from JWT claims

	// Get filter ID
	filterID, ok := req.Data["id"].(string)
	if !ok || filterID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing filter ID",
			"",
		))
		return
	}

	// Check if filter exists and user owns it
	ownerCheckQuery := `SELECT owner_id FROM filter WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), ownerCheckQuery, filterID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Filter not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to check filter ownership", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check filter ownership",
			"",
		))
		return
	}

	// Verify ownership
	if ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"You can only modify your own filters",
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
	if query, ok := req.Data["query"].(string); ok && query != "" {
		// Validate query is valid JSON
		var queryObj interface{}
		if err := json.Unmarshal([]byte(query), &queryObj); err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid query format (must be valid JSON)",
				"",
			))
			return
		}
		updates["query"] = query
	}
	if isPublic, ok := req.Data["isPublic"].(bool); ok {
		updates["is_public"] = isPublic
	}
	if isFavorite, ok := req.Data["isFavorite"].(bool); ok {
		updates["is_favorite"] = isFavorite
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
	query := "UPDATE filter SET "
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
	args = append(args, filterID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update filter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update filter",
			"",
		))
		return
	}

	logger.Info("Filter modified",
		zap.String("filter_id", filterID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      filterID,
	})
	c.JSON(http.StatusOK, response)
}

// handleFilterRemove soft-deletes a filter
func (h *Handler) handleFilterRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	userID := username // In a real system, this would be the actual user ID from JWT claims

	// Get filter ID
	filterID, ok := req.Data["id"].(string)
	if !ok || filterID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing filter ID",
			"",
		))
		return
	}

	// Check if filter exists and user owns it
	ownerCheckQuery := `SELECT owner_id FROM filter WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), ownerCheckQuery, filterID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Filter not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to check filter ownership", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check filter ownership",
			"",
		))
		return
	}

	// Verify ownership
	if ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"You can only delete your own filters",
			"",
		))
		return
	}

	// Soft delete the filter
	query := `UPDATE filter SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), filterID)
	if err != nil {
		logger.Error("Failed to delete filter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete filter",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Filter not found",
			"",
		))
		return
	}

	// Also soft-delete all share mappings for this filter
	shareDeleteQuery := `UPDATE filter_share_mapping SET deleted = 1 WHERE filter_id = ? AND deleted = 0`
	_, err = h.db.Exec(c.Request.Context(), shareDeleteQuery, filterID)
	if err != nil {
		logger.Warn("Failed to delete filter share mappings", zap.Error(err))
		// Don't fail the request if share deletion fails
	}

	logger.Info("Filter deleted",
		zap.String("filter_id", filterID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      filterID,
	})
	c.JSON(http.StatusOK, response)
}

// Helper function to safely get bool from map
func getBoolFromData(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}
