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

// handleSecurityLevelCreate creates a new security level
func (h *Handler) handleSecurityLevelCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "security_level", models.PermissionCreate)
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

	// Get security level details
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
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

	level, ok := req.Data["level"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing level",
			"",
		))
		return
	}

	// Validate level range
	if int(level) < models.SecurityLevelNone || int(level) > models.SecurityLevelSecret {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid security level (must be 0-5)",
			"",
		))
		return
	}

	description := getStringFromData(req.Data, "description")

	// Verify project exists
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

	// Create security level
	securityLevelID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO security_level (id, title, description, project_id, level, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		securityLevelID,
		title,
		description,
		projectID,
		int(level),
		now,
		now,
	)

	if err != nil {
		logger.Error("Failed to create security level", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create security level",
			"",
		))
		return
	}

	logger.Info("Security level created",
		zap.String("security_level_id", securityLevelID),
		zap.String("title", title),
		zap.Int("level", int(level)),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"securityLevelCreate",
		"security_level",
		securityLevelID,
		username,
		map[string]interface{}{
			"id":          securityLevelID,
			"title":       title,
			"description": description,
			"projectId":   projectID,
			"level":       int(level),
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":          securityLevelID,
		"title":       title,
		"description": description,
		"projectId":   projectID,
		"level":       int(level),
	})
	c.JSON(http.StatusCreated, response)
}

// handleSecurityLevelRead reads a security level by ID
func (h *Handler) handleSecurityLevelRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get security level ID
	securityLevelID, ok := req.Data["securityLevelId"].(string)
	if !ok || securityLevelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing securityLevelId",
			"",
		))
		return
	}

	// Query security level
	query := `
		SELECT id, title, description, project_id, level, created, modified
		FROM security_level
		WHERE id = ? AND deleted = 0
	`

	var secLevel models.SecurityLevel

	err := h.db.QueryRow(c.Request.Context(), query, securityLevelID).Scan(
		&secLevel.ID,
		&secLevel.Title,
		&secLevel.Description,
		&secLevel.ProjectID,
		&secLevel.Level,
		&secLevel.Created,
		&secLevel.Modified,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Security level not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read security level", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read security level",
			"",
		))
		return
	}

	logger.Info("Security level read",
		zap.String("security_level_id", securityLevelID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":          secLevel.ID,
		"title":       secLevel.Title,
		"description": secLevel.Description,
		"projectId":   secLevel.ProjectID,
		"level":       secLevel.Level,
		"created":     secLevel.Created,
		"modified":    secLevel.Modified,
	})
	c.JSON(http.StatusOK, response)
}

// handleSecurityLevelList lists all security levels
func (h *Handler) handleSecurityLevelList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Optional filter by projectId
	projectID := getStringFromData(req.Data, "projectId")

	var query string
	var args []interface{}

	if projectID != "" {
		query = `
			SELECT id, title, description, project_id, level, created, modified
			FROM security_level
			WHERE deleted = 0 AND project_id = ?
			ORDER BY level ASC, title ASC
		`
		args = []interface{}{projectID}
	} else {
		query = `
			SELECT id, title, description, project_id, level, created, modified
			FROM security_level
			WHERE deleted = 0
			ORDER BY level ASC, title ASC
		`
	}

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list security levels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list security levels",
			"",
		))
		return
	}
	defer rows.Close()

	securityLevels := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title, description, projectIDVal string
		var level int
		var created, modified int64

		err := rows.Scan(&id, &title, &description, &projectIDVal, &level, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan security level", zap.Error(err))
			continue
		}

		securityLevels = append(securityLevels, map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": description,
			"projectId":   projectIDVal,
			"level":       level,
			"created":     created,
			"modified":    modified,
		})
	}

	logger.Info("Security levels listed",
		zap.Int("count", len(securityLevels)),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"securityLevels": securityLevels,
		"count":          len(securityLevels),
	})
	c.JSON(http.StatusOK, response)
}

// handleSecurityLevelModify updates a security level
func (h *Handler) handleSecurityLevelModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "security_level", models.PermissionUpdate)
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

	// Get security level ID
	securityLevelID, ok := req.Data["securityLevelId"].(string)
	if !ok || securityLevelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing securityLevelId",
			"",
		))
		return
	}

	// Check if security level exists
	checkQuery := `SELECT COUNT(*) FROM security_level WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, securityLevelID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Security level not found",
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
	if level, ok := req.Data["level"].(float64); ok {
		if int(level) < models.SecurityLevelNone || int(level) > models.SecurityLevelSecret {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid security level (must be 0-5)",
				"",
			))
			return
		}
		updates["level"] = int(level)
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
	query := "UPDATE security_level SET "
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
	args = append(args, securityLevelID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update security level", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update security level",
			"",
		))
		return
	}

	logger.Info("Security level updated",
		zap.String("security_level_id", securityLevelID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"securityLevelModify",
		"security_level",
		securityLevelID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated":         true,
		"securityLevelId": securityLevelID,
	})
	c.JSON(http.StatusOK, response)
}

// handleSecurityLevelRemove soft-deletes a security level
func (h *Handler) handleSecurityLevelRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "security_level", models.PermissionDelete)
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

	// Get security level ID
	securityLevelID, ok := req.Data["securityLevelId"].(string)
	if !ok || securityLevelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing securityLevelId",
			"",
		))
		return
	}

	// Soft delete the security level
	query := `UPDATE security_level SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), securityLevelID)
	if err != nil {
		logger.Error("Failed to remove security level", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove security level",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Security level not found",
			"",
		))
		return
	}

	// Also soft delete all permission mappings
	_, err = h.db.Exec(c.Request.Context(),
		"UPDATE security_level_permission_mapping SET deleted = 1 WHERE security_level_id = ?",
		securityLevelID,
	)
	if err != nil {
		logger.Error("Failed to remove security level permissions", zap.Error(err))
		// Don't fail the request, security level removal succeeded
	}

	logger.Info("Security level removed",
		zap.String("security_level_id", securityLevelID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"securityLevelRemove",
		"security_level",
		securityLevelID,
		username,
		map[string]interface{}{
			"securityLevelId": securityLevelID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":         true,
		"securityLevelId": securityLevelID,
	})
	c.JSON(http.StatusOK, response)
}

// handleSecurityLevelGrant grants access to a user, team, or role for a security level
func (h *Handler) handleSecurityLevelGrant(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "security_level", models.PermissionUpdate)
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

	// Get security level ID
	securityLevelID, ok := req.Data["securityLevelId"].(string)
	if !ok || securityLevelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing securityLevelId",
			"",
		))
		return
	}

	// Verify security level exists
	checkQuery := `SELECT COUNT(*) FROM security_level WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, securityLevelID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Security level not found",
			"",
		))
		return
	}

	// Get recipient (user, team, or role) - exactly one must be specified
	userID := getStringFromData(req.Data, "userId")
	teamID := getStringFromData(req.Data, "teamId")
	projectRoleID := getStringFromData(req.Data, "projectRoleId")

	// Validate that exactly one recipient is specified
	recipientCount := 0
	if userID != "" {
		recipientCount++
	}
	if teamID != "" {
		recipientCount++
	}
	if projectRoleID != "" {
		recipientCount++
	}

	if recipientCount == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Must specify userId, teamId, or projectRoleId",
			"",
		))
		return
	}

	if recipientCount > 1 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Can only specify one of userId, teamId, or projectRoleId",
			"",
		))
		return
	}

	// Check for duplicate grant
	var duplicateQuery string
	var duplicateArgs []interface{}

	if userID != "" {
		duplicateQuery = `
			SELECT COUNT(*) FROM security_level_permission_mapping
			WHERE security_level_id = ? AND user_id = ? AND deleted = 0
		`
		duplicateArgs = []interface{}{securityLevelID, userID}
	} else if teamID != "" {
		duplicateQuery = `
			SELECT COUNT(*) FROM security_level_permission_mapping
			WHERE security_level_id = ? AND team_id = ? AND deleted = 0
		`
		duplicateArgs = []interface{}{securityLevelID, teamID}
	} else {
		duplicateQuery = `
			SELECT COUNT(*) FROM security_level_permission_mapping
			WHERE security_level_id = ? AND project_role_id = ? AND deleted = 0
		`
		duplicateArgs = []interface{}{securityLevelID, projectRoleID}
	}

	err = h.db.QueryRow(c.Request.Context(), duplicateQuery, duplicateArgs...).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeEntityAlreadyExists,
			"Access already granted to this recipient",
			"",
		))
		return
	}

	// Create permission mapping
	mappingID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO security_level_permission_mapping (
			id, security_level_id, user_id, team_id, project_role_id, created, deleted
		) VALUES (?, ?, ?, ?, ?, ?, 0)
	`

	var userIDPtr, teamIDPtr, projectRoleIDPtr interface{}
	if userID != "" {
		userIDPtr = userID
	}
	if teamID != "" {
		teamIDPtr = teamID
	}
	if projectRoleID != "" {
		projectRoleIDPtr = projectRoleID
	}

	_, err = h.db.Exec(c.Request.Context(), query,
		mappingID,
		securityLevelID,
		userIDPtr,
		teamIDPtr,
		projectRoleIDPtr,
		now,
	)

	if err != nil {
		logger.Error("Failed to grant security level access", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to grant access",
			"",
		))
		return
	}

	logger.Info("Security level access granted",
		zap.String("mapping_id", mappingID),
		zap.String("security_level_id", securityLevelID),
		zap.String("user_id", userID),
		zap.String("team_id", teamID),
		zap.String("project_role_id", projectRoleID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"securityLevelGrant",
		"security_level",
		securityLevelID,
		username,
		map[string]interface{}{
			"mappingId":       mappingID,
			"securityLevelId": securityLevelID,
			"userId":          userID,
			"teamId":          teamID,
			"projectRoleId":   projectRoleID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"granted":         true,
		"mappingId":       mappingID,
		"securityLevelId": securityLevelID,
		"userId":          userID,
		"teamId":          teamID,
		"projectRoleId":   projectRoleID,
	})
	c.JSON(http.StatusCreated, response)
}

// handleSecurityLevelRevoke revokes access from a user, team, or role
func (h *Handler) handleSecurityLevelRevoke(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "security_level", models.PermissionUpdate)
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

	// Get security level ID
	securityLevelID, ok := req.Data["securityLevelId"].(string)
	if !ok || securityLevelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing securityLevelId",
			"",
		))
		return
	}

	// Get recipient (user, team, or role)
	userID := getStringFromData(req.Data, "userId")
	teamID := getStringFromData(req.Data, "teamId")
	projectRoleID := getStringFromData(req.Data, "projectRoleId")

	// Build revoke query based on recipient
	var query string
	var args []interface{}

	if userID != "" {
		query = `
			UPDATE security_level_permission_mapping
			SET deleted = 1
			WHERE security_level_id = ? AND user_id = ? AND deleted = 0
		`
		args = []interface{}{securityLevelID, userID}
	} else if teamID != "" {
		query = `
			UPDATE security_level_permission_mapping
			SET deleted = 1
			WHERE security_level_id = ? AND team_id = ? AND deleted = 0
		`
		args = []interface{}{securityLevelID, teamID}
	} else if projectRoleID != "" {
		query = `
			UPDATE security_level_permission_mapping
			SET deleted = 1
			WHERE security_level_id = ? AND project_role_id = ? AND deleted = 0
		`
		args = []interface{}{securityLevelID, projectRoleID}
	} else {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Must specify userId, teamId, or projectRoleId",
			"",
		))
		return
	}

	result, err := h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to revoke security level access", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to revoke access",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Access grant not found",
			"",
		))
		return
	}

	logger.Info("Security level access revoked",
		zap.String("security_level_id", securityLevelID),
		zap.String("user_id", userID),
		zap.String("team_id", teamID),
		zap.String("project_role_id", projectRoleID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"securityLevelRevoke",
		"security_level",
		securityLevelID,
		username,
		map[string]interface{}{
			"securityLevelId": securityLevelID,
			"userId":          userID,
			"teamId":          teamID,
			"projectRoleId":   projectRoleID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"revoked":         true,
		"securityLevelId": securityLevelID,
		"userId":          userID,
		"teamId":          teamID,
		"projectRoleId":   projectRoleID,
	})
	c.JSON(http.StatusOK, response)
}

// handleSecurityLevelCheck checks if a user has access to a security level
func (h *Handler) handleSecurityLevelCheck(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get security level ID
	securityLevelID, ok := req.Data["securityLevelId"].(string)
	if !ok || securityLevelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing securityLevelId",
			"",
		))
		return
	}

	// Get user ID to check (defaults to current user)
	userID := getStringFromData(req.Data, "userId")
	if userID == "" {
		userID = username
	}

	// Check if user has direct access
	query := `
		SELECT COUNT(*) FROM security_level_permission_mapping
		WHERE security_level_id = ? AND user_id = ? AND deleted = 0
	`

	var count int
	err := h.db.QueryRow(c.Request.Context(), query, securityLevelID, userID).Scan(&count)
	hasAccess := err == nil && count > 0

	// If no direct access, check team and role access
	// This is a simplified implementation - you may need to join with team_user and project_role_user_mapping tables
	if !hasAccess {
		// Check team access
		teamQuery := `
			SELECT COUNT(*)
			FROM security_level_permission_mapping slpm
			INNER JOIN team_user tu ON slpm.team_id = tu.team_id
			WHERE slpm.security_level_id = ? AND tu.user_id = ?
			AND slpm.deleted = 0 AND tu.deleted = 0
		`
		err = h.db.QueryRow(c.Request.Context(), teamQuery, securityLevelID, userID).Scan(&count)
		hasAccess = err == nil && count > 0
	}

	if !hasAccess {
		// Check role access
		roleQuery := `
			SELECT COUNT(*)
			FROM security_level_permission_mapping slpm
			INNER JOIN project_role_user_mapping prum ON slpm.project_role_id = prum.project_role_id
			WHERE slpm.security_level_id = ? AND prum.user_id = ?
			AND slpm.deleted = 0 AND prum.deleted = 0
		`
		err = h.db.QueryRow(c.Request.Context(), roleQuery, securityLevelID, userID).Scan(&count)
		hasAccess = err == nil && count > 0
	}

	logger.Info("Security level access checked",
		zap.String("security_level_id", securityLevelID),
		zap.String("user_id", userID),
		zap.Bool("has_access", hasAccess),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"securityLevelId": securityLevelID,
		"userId":          userID,
		"hasAccess":       hasAccess,
	})
	c.JSON(http.StatusOK, response)
}
