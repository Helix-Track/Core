package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

/*
Activity Stream Handlers - Phase 3

Activity streams provide a view of recent activities and changes in the system.
Based on the audit table with V3 enhancements (is_public, activity_type).

Handlers:
  1. handleActivityStreamGet - Get global activity stream
  2. handleActivityStreamGetByProject - Get project-specific activity
  3. handleActivityStreamGetByUser - Get user-specific activity
  4. handleActivityStreamGetByTicket - Get ticket-specific activity
  5. handleActivityStreamFilter - Filter activity by type
*/

// handleActivityStreamGet retrieves the global activity stream
func (h *Handler) handleActivityStreamGet(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "activity", models.PermissionRead)
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

	// Get pagination parameters
	limit := 50
	if limitVal, ok := req.Data["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
		if limit > 1000 {
			limit = 1000 // Cap at 1000 for performance
		}
	}

	offset := 0
	if offsetVal, ok := req.Data["offset"].(float64); ok && offsetVal > 0 {
		offset = int(offsetVal)
	}

	// Query global activity stream (only public activities)
	query := `
		SELECT id, action, user_id, entity_id, entity_type, details, is_public, activity_type, created, modified, deleted
		FROM audit
		WHERE deleted = 0 AND is_public = 1
		ORDER BY created DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.db.Query(c.Request.Context(), query, limit, offset)
	if err != nil {
		logger.Error("Failed to query activity stream", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to query activity stream",
			"",
		))
		return
	}
	defer rows.Close()

	activities := []models.Audit{}
	for rows.Next() {
		var activity models.Audit
		var details, activityType *string

		err := rows.Scan(
			&activity.ID,
			&activity.Action,
			&activity.UserID,
			&activity.EntityID,
			&activity.EntityType,
			&details,
			&activity.IsPublic,
			&activityType,
			&activity.Created,
			&activity.Modified,
			&activity.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan activity", zap.Error(err))
			continue
		}

		if details != nil {
			activity.Details = *details
		}
		if activityType != nil {
			activity.ActivityType = *activityType
		}

		activities = append(activities, activity)
	}

	logger.Info("Activity stream retrieved",
		zap.String("username", username),
		zap.Int("count", len(activities)),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
		"limit":      limit,
		"offset":     offset,
	}))
}

// handleActivityStreamGetByProject retrieves project-specific activity stream
func (h *Handler) handleActivityStreamGetByProject(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
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

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "activity", models.PermissionRead)
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

	// Get pagination parameters
	limit := 50
	if limitVal, ok := req.Data["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
		if limit > 1000 {
			limit = 1000
		}
	}

	offset := 0
	if offsetVal, ok := req.Data["offset"].(float64); ok && offsetVal > 0 {
		offset = int(offsetVal)
	}

	// Query project activity stream
	// We need to join with entity tables to filter by project
	query := `
		SELECT DISTINCT a.id, a.action, a.user_id, a.entity_id, a.entity_type, a.details, a.is_public, a.activity_type, a.created, a.modified, a.deleted
		FROM audit a
		LEFT JOIN ticket t ON a.entity_type = 'ticket' AND a.entity_id = t.id
		LEFT JOIN project p ON (a.entity_type = 'project' AND a.entity_id = p.id) OR t.project_id = p.id
		WHERE a.deleted = 0 AND a.is_public = 1
			AND (p.id = ? OR a.entity_type = 'project' AND a.entity_id = ?)
		ORDER BY a.created DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.db.Query(c.Request.Context(), query, projectID, projectID, limit, offset)
	if err != nil {
		logger.Error("Failed to query project activity stream", zap.Error(err), zap.String("projectId", projectID))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to query project activity stream",
			"",
		))
		return
	}
	defer rows.Close()

	activities := []models.Audit{}
	for rows.Next() {
		var activity models.Audit
		var details, activityType *string

		err := rows.Scan(
			&activity.ID,
			&activity.Action,
			&activity.UserID,
			&activity.EntityID,
			&activity.EntityType,
			&details,
			&activity.IsPublic,
			&activityType,
			&activity.Created,
			&activity.Modified,
			&activity.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan activity", zap.Error(err))
			continue
		}

		if details != nil {
			activity.Details = *details
		}
		if activityType != nil {
			activity.ActivityType = *activityType
		}

		activities = append(activities, activity)
	}

	logger.Info("Project activity stream retrieved",
		zap.String("username", username),
		zap.String("projectId", projectID),
		zap.Int("count", len(activities)),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"activities": activities,
		"projectId":  projectID,
		"count":      len(activities),
		"limit":      limit,
		"offset":     offset,
	}))
}

// handleActivityStreamGetByUser retrieves user-specific activity stream
func (h *Handler) handleActivityStreamGetByUser(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get user ID
	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing userId",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "activity", models.PermissionRead)
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

	// Get pagination parameters
	limit := 50
	if limitVal, ok := req.Data["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
		if limit > 1000 {
			limit = 1000
		}
	}

	offset := 0
	if offsetVal, ok := req.Data["offset"].(float64); ok && offsetVal > 0 {
		offset = int(offsetVal)
	}

	// Query user activity stream
	query := `
		SELECT id, action, user_id, entity_id, entity_type, details, is_public, activity_type, created, modified, deleted
		FROM audit
		WHERE deleted = 0 AND is_public = 1 AND user_id = ?
		ORDER BY created DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.db.Query(c.Request.Context(), query, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to query user activity stream", zap.Error(err), zap.String("userId", userID))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to query user activity stream",
			"",
		))
		return
	}
	defer rows.Close()

	activities := []models.Audit{}
	for rows.Next() {
		var activity models.Audit
		var details, activityType *string

		err := rows.Scan(
			&activity.ID,
			&activity.Action,
			&activity.UserID,
			&activity.EntityID,
			&activity.EntityType,
			&details,
			&activity.IsPublic,
			&activityType,
			&activity.Created,
			&activity.Modified,
			&activity.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan activity", zap.Error(err))
			continue
		}

		if details != nil {
			activity.Details = *details
		}
		if activityType != nil {
			activity.ActivityType = *activityType
		}

		activities = append(activities, activity)
	}

	logger.Info("User activity stream retrieved",
		zap.String("username", username),
		zap.String("userId", userID),
		zap.Int("count", len(activities)),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"activities": activities,
		"userId":     userID,
		"count":      len(activities),
		"limit":      limit,
		"offset":     offset,
	}))
}

// handleActivityStreamGetByTicket retrieves ticket-specific activity stream
func (h *Handler) handleActivityStreamGetByTicket(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
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

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "activity", models.PermissionRead)
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

	// Get pagination parameters
	limit := 50
	if limitVal, ok := req.Data["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
		if limit > 1000 {
			limit = 1000
		}
	}

	offset := 0
	if offsetVal, ok := req.Data["offset"].(float64); ok && offsetVal > 0 {
		offset = int(offsetVal)
	}

	// Query ticket activity stream
	query := `
		SELECT id, action, user_id, entity_id, entity_type, details, is_public, activity_type, created, modified, deleted
		FROM audit
		WHERE deleted = 0 AND is_public = 1 AND entity_type = 'ticket' AND entity_id = ?
		ORDER BY created DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID, limit, offset)
	if err != nil {
		logger.Error("Failed to query ticket activity stream", zap.Error(err), zap.String("ticketId", ticketID))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to query ticket activity stream",
			"",
		))
		return
	}
	defer rows.Close()

	activities := []models.Audit{}
	for rows.Next() {
		var activity models.Audit
		var details, activityType *string

		err := rows.Scan(
			&activity.ID,
			&activity.Action,
			&activity.UserID,
			&activity.EntityID,
			&activity.EntityType,
			&details,
			&activity.IsPublic,
			&activityType,
			&activity.Created,
			&activity.Modified,
			&activity.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan activity", zap.Error(err))
			continue
		}

		if details != nil {
			activity.Details = *details
		}
		if activityType != nil {
			activity.ActivityType = *activityType
		}

		activities = append(activities, activity)
	}

	logger.Info("Ticket activity stream retrieved",
		zap.String("username", username),
		zap.String("ticketId", ticketID),
		zap.Int("count", len(activities)),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"activities": activities,
		"ticketId":   ticketID,
		"count":      len(activities),
		"limit":      limit,
		"offset":     offset,
	}))
}

// handleActivityStreamFilter filters activity stream by activity type
func (h *Handler) handleActivityStreamFilter(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get activity type filter
	activityType, ok := req.Data["activityType"].(string)
	if !ok || activityType == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing activityType",
			"",
		))
		return
	}

	// Check permissions
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "activity", models.PermissionRead)
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

	// Get optional filter parameters
	var projectID, userID, entityType string
	if val, ok := req.Data["projectId"].(string); ok {
		projectID = val
	}
	if val, ok := req.Data["userId"].(string); ok {
		userID = val
	}
	if val, ok := req.Data["entityType"].(string); ok {
		entityType = val
	}

	// Get pagination parameters
	limit := 50
	if limitVal, ok := req.Data["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
		if limit > 1000 {
			limit = 1000
		}
	}

	offset := 0
	if offsetVal, ok := req.Data["offset"].(float64); ok && offsetVal > 0 {
		offset = int(offsetVal)
	}

	// Build dynamic query based on filters
	query := `
		SELECT DISTINCT a.id, a.action, a.user_id, a.entity_id, a.entity_type, a.details, a.is_public, a.activity_type, a.created, a.modified, a.deleted
		FROM audit a
	`

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "a.deleted = 0", "a.is_public = 1", "a.activity_type = ?")
	args = append(args, activityType)

	if userID != "" {
		conditions = append(conditions, "a.user_id = ?")
		args = append(args, userID)
	}

	if entityType != "" {
		conditions = append(conditions, "a.entity_type = ?")
		args = append(args, entityType)
	}

	if projectID != "" {
		query += `
		LEFT JOIN ticket t ON a.entity_type = 'ticket' AND a.entity_id = t.id
		LEFT JOIN project p ON (a.entity_type = 'project' AND a.entity_id = p.id) OR t.project_id = p.id
		`
		conditions = append(conditions, "(p.id = ? OR (a.entity_type = 'project' AND a.entity_id = ?))")
		args = append(args, projectID, projectID)
	}

	query += " WHERE "
	for i, cond := range conditions {
		if i > 0 {
			query += " AND "
		}
		query += cond
	}

	query += " ORDER BY a.created DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to filter activity stream", zap.Error(err), zap.String("activityType", activityType))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeDatabaseError,
			"Failed to filter activity stream",
			"",
		))
		return
	}
	defer rows.Close()

	activities := []models.Audit{}
	for rows.Next() {
		var activity models.Audit
		var details, actType *string

		err := rows.Scan(
			&activity.ID,
			&activity.Action,
			&activity.UserID,
			&activity.EntityID,
			&activity.EntityType,
			&details,
			&activity.IsPublic,
			&actType,
			&activity.Created,
			&activity.Modified,
			&activity.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan activity", zap.Error(err))
			continue
		}

		if details != nil {
			activity.Details = *details
		}
		if actType != nil {
			activity.ActivityType = *actType
		}

		activities = append(activities, activity)
	}

	logger.Info("Activity stream filtered",
		zap.String("username", username),
		zap.String("activityType", activityType),
		zap.Int("count", len(activities)),
	)

	c.JSON(http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"activities":   activities,
		"activityType": activityType,
		"count":        len(activities),
		"limit":        limit,
		"offset":       offset,
		"filters": map[string]interface{}{
			"projectId":  projectID,
			"userId":     userID,
			"entityType": entityType,
		},
	}))
}
