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

// handleDashboardCreate creates a new dashboard
func (h *Handler) handleDashboardCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "dashboard", models.PermissionCreate)
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

	// Get dashboard details
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
	isPublic := getBoolFromData(req.Data, "isPublic")
	isFavorite := getBoolFromData(req.Data, "isFavorite")
	layout := getStringFromData(req.Data, "layout")

	// Create dashboard
	dashboardID := uuid.New().String()
	now := time.Now().Unix()

	var query string
	var args []interface{}

	if layout != "" {
		query = `
			INSERT INTO dashboard (id, title, description, owner_id, is_public, is_favorite, layout, created, modified, deleted, version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1)
		`
		args = []interface{}{dashboardID, title, description, username, isPublic, isFavorite, layout, now, now}
	} else {
		query = `
			INSERT INTO dashboard (id, title, description, owner_id, is_public, is_favorite, layout, created, modified, deleted, version)
			VALUES (?, ?, ?, ?, ?, ?, NULL, ?, ?, 0, 1)
		`
		args = []interface{}{dashboardID, title, description, username, isPublic, isFavorite, now, now}
	}

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to create dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create dashboard",
			"",
		))
		return
	}

	logger.Info("Dashboard created",
		zap.String("dashboard_id", dashboardID),
		zap.String("title", title),
		zap.Bool("is_public", isPublic),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardCreate",
		"dashboard",
		dashboardID,
		username,
		map[string]interface{}{
			"id":          dashboardID,
			"title":       title,
			"description": description,
			"ownerId":     username,
			"isPublic":    isPublic,
			"isFavorite":  isFavorite,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":          dashboardID,
		"title":       title,
		"description": description,
		"ownerId":     username,
		"isPublic":    isPublic,
		"isFavorite":  isFavorite,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDashboardRead reads a dashboard by ID
func (h *Handler) handleDashboardRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Query dashboard
	query := `
		SELECT id, title, description, owner_id, is_public, is_favorite, layout, created, modified
		FROM dashboard
		WHERE id = ? AND deleted = 0
	`

	var dashboard models.Dashboard
	var layout sql.NullString

	err := h.db.QueryRow(c.Request.Context(), query, dashboardID).Scan(
		&dashboard.ID,
		&dashboard.Title,
		&dashboard.Description,
		&dashboard.OwnerID,
		&dashboard.IsPublic,
		&dashboard.IsFavorite,
		&layout,
		&dashboard.Created,
		&dashboard.Modified,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read dashboard",
			"",
		))
		return
	}

	if layout.Valid && layout.String != "" {
		dashboard.Layout = &layout.String
	}

	// Check access: owner, public, or shared
	hasAccess := dashboard.OwnerID == username || dashboard.IsPublic
	if !hasAccess {
		// Check if shared with user
		shareQuery := `
			SELECT COUNT(*) FROM dashboard_share_mapping
			WHERE dashboard_id = ? AND user_id = ? AND deleted = 0
		`
		var count int
		err = h.db.QueryRow(c.Request.Context(), shareQuery, dashboardID, username).Scan(&count)
		hasAccess = err == nil && count > 0
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Access denied to this dashboard",
			"",
		))
		return
	}

	logger.Info("Dashboard read",
		zap.String("dashboard_id", dashboardID),
		zap.String("username", username),
	)

	layoutStr := ""
	if dashboard.Layout != nil {
		layoutStr = *dashboard.Layout
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":          dashboard.ID,
		"title":       dashboard.Title,
		"description": dashboard.Description,
		"ownerId":     dashboard.OwnerID,
		"isPublic":    dashboard.IsPublic,
		"isFavorite":  dashboard.IsFavorite,
		"layout":      layoutStr,
		"created":     dashboard.Created,
		"modified":    dashboard.Modified,
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardList lists dashboards accessible to the user
func (h *Handler) handleDashboardList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// List dashboards owned by user, public dashboards, or shared with user
	query := `
		SELECT DISTINCT d.id, d.title, d.description, d.owner_id, d.is_public, d.is_favorite, d.created, d.modified
		FROM dashboard d
		LEFT JOIN dashboard_share_mapping dsm ON d.id = dsm.dashboard_id AND dsm.deleted = 0 AND dsm.user_id = ?
		WHERE d.deleted = 0 AND (d.owner_id = ? OR d.is_public = 1 OR dsm.id IS NOT NULL)
		ORDER BY d.is_favorite DESC, d.modified DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, username, username)
	if err != nil {
		logger.Error("Failed to list dashboards", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list dashboards",
			"",
		))
		return
	}
	defer rows.Close()

	dashboards := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title, description, ownerID string
		var isPublic, isFavorite bool
		var created, modified int64

		err := rows.Scan(&id, &title, &description, &ownerID, &isPublic, &isFavorite, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan dashboard", zap.Error(err))
			continue
		}

		dashboards = append(dashboards, map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": description,
			"ownerId":     ownerID,
			"isPublic":    isPublic,
			"isFavorite":  isFavorite,
			"created":     created,
			"modified":    modified,
		})
	}

	logger.Info("Dashboards listed",
		zap.Int("count", len(dashboards)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"dashboards": dashboards,
		"count":      len(dashboards),
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardModify updates a dashboard
func (h *Handler) handleDashboardModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Check if dashboard exists and user is owner
	checkQuery := `SELECT owner_id FROM dashboard WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, dashboardID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to check dashboard ownership", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check dashboard",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can modify it",
			"",
		))
		return
	}

	// Build update query dynamically
	updates := make(map[string]interface{})

	if title, ok := req.Data["title"].(string); ok && title != "" {
		updates["title"] = title
	}
	if description, ok := req.Data["description"].(string); ok {
		updates["description"] = description
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
	query := "UPDATE dashboard SET "
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
	args = append(args, dashboardID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update dashboard",
			"",
		))
		return
	}

	logger.Info("Dashboard updated",
		zap.String("dashboard_id", dashboardID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardModify",
		"dashboard",
		dashboardID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated":     true,
		"dashboardId": dashboardID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardRemove soft-deletes a dashboard
func (h *Handler) handleDashboardRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Check if dashboard exists and user is owner
	checkQuery := `SELECT owner_id FROM dashboard WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, dashboardID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to check dashboard ownership", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check dashboard",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can delete it",
			"",
		))
		return
	}

	// Soft delete the dashboard
	query := `UPDATE dashboard SET deleted = 1, modified = ? WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), query, time.Now().Unix(), dashboardID)
	if err != nil {
		logger.Error("Failed to remove dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove dashboard",
			"",
		))
		return
	}

	// Also soft delete all widgets and shares
	_, err = h.db.Exec(c.Request.Context(),
		"UPDATE dashboard_widget SET deleted = 1 WHERE dashboard_id = ?",
		dashboardID,
	)
	_, err = h.db.Exec(c.Request.Context(),
		"UPDATE dashboard_share_mapping SET deleted = 1 WHERE dashboard_id = ?",
		dashboardID,
	)

	logger.Info("Dashboard removed",
		zap.String("dashboard_id", dashboardID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardRemove",
		"dashboard",
		dashboardID,
		username,
		map[string]interface{}{
			"dashboardId": dashboardID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":     true,
		"dashboardId": dashboardID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardShare shares a dashboard with a user, team, or project
func (h *Handler) handleDashboardShare(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Check if dashboard exists and user is owner
	checkQuery := `SELECT owner_id FROM dashboard WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, dashboardID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to check dashboard ownership", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check dashboard",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can share it",
			"",
		))
		return
	}

	// Get share recipient (user, team, or project)
	userID := getStringFromData(req.Data, "userId")
	teamID := getStringFromData(req.Data, "teamId")
	projectID := getStringFromData(req.Data, "projectId")

	// Validate that exactly one recipient is specified
	recipientCount := 0
	if userID != "" {
		recipientCount++
	}
	if teamID != "" {
		recipientCount++
	}
	if projectID != "" {
		recipientCount++
	}

	if recipientCount == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Must specify userId, teamId, or projectId",
			"",
		))
		return
	}

	if recipientCount > 1 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Can only specify one of userId, teamId, or projectId",
			"",
		))
		return
	}

	// Create share mapping
	shareID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO dashboard_share_mapping (id, dashboard_id, user_id, team_id, project_id, created, deleted)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`

	var userIDPtr, teamIDPtr, projectIDPtr interface{}
	if userID != "" {
		userIDPtr = userID
	}
	if teamID != "" {
		teamIDPtr = teamID
	}
	if projectID != "" {
		projectIDPtr = projectID
	}

	_, err = h.db.Exec(c.Request.Context(), query,
		shareID,
		dashboardID,
		userIDPtr,
		teamIDPtr,
		projectIDPtr,
		now,
	)

	if err != nil {
		logger.Error("Failed to share dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to share dashboard",
			"",
		))
		return
	}

	logger.Info("Dashboard shared",
		zap.String("share_id", shareID),
		zap.String("dashboard_id", dashboardID),
		zap.String("user_id", userID),
		zap.String("team_id", teamID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardShare",
		"dashboard",
		dashboardID,
		username,
		map[string]interface{}{
			"shareId":     shareID,
			"dashboardId": dashboardID,
			"userId":      userID,
			"teamId":      teamID,
			"projectId":   projectID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"shared":      true,
		"shareId":     shareID,
		"dashboardId": dashboardID,
		"userId":      userID,
		"teamId":      teamID,
		"projectId":   projectID,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDashboardUnshare removes dashboard sharing
func (h *Handler) handleDashboardUnshare(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Check if dashboard exists and user is owner
	checkQuery := `SELECT owner_id FROM dashboard WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, dashboardID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can unshare it",
			"",
		))
		return
	}

	// Get share recipient
	userID := getStringFromData(req.Data, "userId")
	teamID := getStringFromData(req.Data, "teamId")
	projectID := getStringFromData(req.Data, "projectId")

	// Build unshare query based on recipient
	var query string
	var args []interface{}

	if userID != "" {
		query = `
			UPDATE dashboard_share_mapping
			SET deleted = 1
			WHERE dashboard_id = ? AND user_id = ? AND deleted = 0
		`
		args = []interface{}{dashboardID, userID}
	} else if teamID != "" {
		query = `
			UPDATE dashboard_share_mapping
			SET deleted = 1
			WHERE dashboard_id = ? AND team_id = ? AND deleted = 0
		`
		args = []interface{}{dashboardID, teamID}
	} else if projectID != "" {
		query = `
			UPDATE dashboard_share_mapping
			SET deleted = 1
			WHERE dashboard_id = ? AND project_id = ? AND deleted = 0
		`
		args = []interface{}{dashboardID, projectID}
	} else {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Must specify userId, teamId, or projectId",
			"",
		))
		return
	}

	result, err := h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to unshare dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unshare dashboard",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Share not found",
			"",
		))
		return
	}

	logger.Info("Dashboard unshared",
		zap.String("dashboard_id", dashboardID),
		zap.String("user_id", userID),
		zap.String("team_id", teamID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardUnshare",
		"dashboard",
		dashboardID,
		username,
		map[string]interface{}{
			"dashboardId": dashboardID,
			"userId":      userID,
			"teamId":      teamID,
			"projectId":   projectID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"unshared":    true,
		"dashboardId": dashboardID,
		"userId":      userID,
		"teamId":      teamID,
		"projectId":   projectID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardAddWidget adds a widget to a dashboard
func (h *Handler) handleDashboardAddWidget(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Get widget type
	widgetType, ok := req.Data["widgetType"].(string)
	if !ok || widgetType == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing widgetType",
			"",
		))
		return
	}

	// Check if dashboard exists and user has access
	checkQuery := `SELECT owner_id FROM dashboard WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, dashboardID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can add widgets",
			"",
		))
		return
	}

	// Get optional widget properties
	title := getStringFromData(req.Data, "title")
	configuration := getStringFromData(req.Data, "configuration")

	// Position and size
	var posX, posY, width, height interface{}
	if posXVal, ok := req.Data["positionX"].(float64); ok {
		posX = int(posXVal)
	}
	if posYVal, ok := req.Data["positionY"].(float64); ok {
		posY = int(posYVal)
	}
	if widthVal, ok := req.Data["width"].(float64); ok {
		width = int(widthVal)
	}
	if heightVal, ok := req.Data["height"].(float64); ok {
		height = int(heightVal)
	}

	// Create widget
	widgetID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO dashboard_widget (
			id, dashboard_id, widget_type, title, position_x, position_y, width, height, configuration, created, modified, deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`

	var titlePtr, configPtr interface{}
	if title != "" {
		titlePtr = title
	}
	if configuration != "" {
		configPtr = configuration
	}

	_, err = h.db.Exec(c.Request.Context(), query,
		widgetID,
		dashboardID,
		widgetType,
		titlePtr,
		posX,
		posY,
		width,
		height,
		configPtr,
		now,
		now,
	)

	if err != nil {
		logger.Error("Failed to add widget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add widget",
			"",
		))
		return
	}

	logger.Info("Widget added to dashboard",
		zap.String("widget_id", widgetID),
		zap.String("dashboard_id", dashboardID),
		zap.String("widget_type", widgetType),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardAddWidget",
		"dashboard",
		dashboardID,
		username,
		map[string]interface{}{
			"widgetId":    widgetID,
			"dashboardId": dashboardID,
			"widgetType":  widgetType,
			"title":       title,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":       true,
		"widgetId":    widgetID,
		"dashboardId": dashboardID,
		"widgetType":  widgetType,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDashboardRemoveWidget removes a widget from a dashboard
func (h *Handler) handleDashboardRemoveWidget(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get widget ID
	widgetID, ok := req.Data["widgetId"].(string)
	if !ok || widgetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing widgetId",
			"",
		))
		return
	}

	// Check if widget exists and user owns the dashboard
	checkQuery := `
		SELECT dw.dashboard_id, d.owner_id
		FROM dashboard_widget dw
		INNER JOIN dashboard d ON dw.dashboard_id = d.id
		WHERE dw.id = ? AND dw.deleted = 0 AND d.deleted = 0
	`

	var dashboardID, ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, widgetID).Scan(&dashboardID, &ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Widget not found",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can remove widgets",
			"",
		))
		return
	}

	// Soft delete the widget
	query := `UPDATE dashboard_widget SET deleted = 1, modified = ? WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), query, time.Now().Unix(), widgetID)
	if err != nil {
		logger.Error("Failed to remove widget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove widget",
			"",
		))
		return
	}

	logger.Info("Widget removed from dashboard",
		zap.String("widget_id", widgetID),
		zap.String("dashboard_id", dashboardID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardRemoveWidget",
		"dashboard",
		dashboardID,
		username,
		map[string]interface{}{
			"widgetId":    widgetID,
			"dashboardId": dashboardID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":     true,
		"widgetId":    widgetID,
		"dashboardId": dashboardID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardModifyWidget updates a widget's properties
func (h *Handler) handleDashboardModifyWidget(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get widget ID
	widgetID, ok := req.Data["widgetId"].(string)
	if !ok || widgetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing widgetId",
			"",
		))
		return
	}

	// Check if widget exists and user owns the dashboard
	checkQuery := `
		SELECT dw.dashboard_id, d.owner_id
		FROM dashboard_widget dw
		INNER JOIN dashboard d ON dw.dashboard_id = d.id
		WHERE dw.id = ? AND dw.deleted = 0 AND d.deleted = 0
	`

	var dashboardID, ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, widgetID).Scan(&dashboardID, &ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Widget not found",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can modify widgets",
			"",
		))
		return
	}

	// Build update query dynamically
	updates := make(map[string]interface{})

	if title, ok := req.Data["title"].(string); ok {
		updates["title"] = title
	}
	if posX, ok := req.Data["positionX"].(float64); ok {
		updates["position_x"] = int(posX)
	}
	if posY, ok := req.Data["positionY"].(float64); ok {
		updates["position_y"] = int(posY)
	}
	if width, ok := req.Data["width"].(float64); ok {
		updates["width"] = int(width)
	}
	if height, ok := req.Data["height"].(float64); ok {
		updates["height"] = int(height)
	}
	if configuration, ok := req.Data["configuration"].(string); ok {
		updates["configuration"] = configuration
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
	query := "UPDATE dashboard_widget SET "
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
	args = append(args, widgetID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update widget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update widget",
			"",
		))
		return
	}

	logger.Info("Widget updated",
		zap.String("widget_id", widgetID),
		zap.String("dashboard_id", dashboardID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardModifyWidget",
		"dashboard",
		dashboardID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated":     true,
		"widgetId":    widgetID,
		"dashboardId": dashboardID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardListWidgets lists all widgets in a dashboard
func (h *Handler) handleDashboardListWidgets(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Check dashboard access
	checkQuery := `SELECT owner_id, is_public FROM dashboard WHERE id = ? AND deleted = 0`
	var ownerID string
	var isPublic bool
	err := h.db.QueryRow(c.Request.Context(), checkQuery, dashboardID).Scan(&ownerID, &isPublic)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}

	hasAccess := ownerID == username || isPublic
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Access denied to this dashboard",
			"",
		))
		return
	}

	// List widgets
	query := `
		SELECT id, widget_type, title, position_x, position_y, width, height, configuration, created, modified
		FROM dashboard_widget
		WHERE dashboard_id = ? AND deleted = 0
		ORDER BY position_y ASC, position_x ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, dashboardID)
	if err != nil {
		logger.Error("Failed to list widgets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list widgets",
			"",
		))
		return
	}
	defer rows.Close()

	widgets := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, widgetType string
		var title, configuration sql.NullString
		var posX, posY, width, height sql.NullInt64
		var created, modified int64

		err := rows.Scan(&id, &widgetType, &title, &posX, &posY, &width, &height, &configuration, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan widget", zap.Error(err))
			continue
		}

		widget := map[string]interface{}{
			"id":            id,
			"widgetType":    widgetType,
			"title":         title.String,
			"positionX":     nil,
			"positionY":     nil,
			"width":         nil,
			"height":        nil,
			"configuration": configuration.String,
			"created":       created,
			"modified":      modified,
		}

		if posX.Valid {
			widget["positionX"] = int(posX.Int64)
		}
		if posY.Valid {
			widget["positionY"] = int(posY.Int64)
		}
		if width.Valid {
			widget["width"] = int(width.Int64)
		}
		if height.Valid {
			widget["height"] = int(height.Int64)
		}

		widgets = append(widgets, widget)
	}

	logger.Info("Widgets listed",
		zap.String("dashboard_id", dashboardID),
		zap.Int("count", len(widgets)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"dashboardId": dashboardID,
		"widgets":     widgets,
		"count":       len(widgets),
	})
	c.JSON(http.StatusOK, response)
}

// handleDashboardSetLayout updates the dashboard layout configuration
func (h *Handler) handleDashboardSetLayout(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get dashboard ID
	dashboardID, ok := req.Data["dashboardId"].(string)
	if !ok || dashboardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing dashboardId",
			"",
		))
		return
	}

	// Get layout JSON
	layout, ok := req.Data["layout"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing layout",
			"",
		))
		return
	}

	// Check if dashboard exists and user is owner
	checkQuery := `SELECT owner_id FROM dashboard WHERE id = ? AND deleted = 0`
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(), checkQuery, dashboardID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Dashboard not found",
			"",
		))
		return
	}

	if ownerID != username {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Only dashboard owner can modify layout",
			"",
		))
		return
	}

	// Update layout
	query := `UPDATE dashboard SET layout = ?, modified = ? WHERE id = ?`
	_, err = h.db.Exec(c.Request.Context(), query, layout, time.Now().Unix(), dashboardID)
	if err != nil {
		logger.Error("Failed to set dashboard layout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to set layout",
			"",
		))
		return
	}

	logger.Info("Dashboard layout updated",
		zap.String("dashboard_id", dashboardID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"dashboardSetLayout",
		"dashboard",
		dashboardID,
		username,
		map[string]interface{}{
			"dashboardId": dashboardID,
			"layout":      layout,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated":     true,
		"dashboardId": dashboardID,
	})
	c.JSON(http.StatusOK, response)
}
