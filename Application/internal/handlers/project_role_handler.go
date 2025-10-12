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

// handleProjectRoleCreate creates a new project role
func (h *Handler) handleProjectRoleCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_role", models.PermissionCreate)
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

	// Get role details
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
	projectID := getStringFromData(req.Data, "projectId")

	// If projectID is provided, verify the project exists
	if projectID != "" {
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
	}

	// Create project role
	roleID := uuid.New().String()
	now := time.Now().Unix()

	var query string
	var args []interface{}

	if projectID != "" {
		query = `
			INSERT INTO project_role (id, title, description, project_id, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, 0)
		`
		args = []interface{}{roleID, title, description, projectID, now, now}
	} else {
		query = `
			INSERT INTO project_role (id, title, description, project_id, created, modified, deleted)
			VALUES (?, ?, ?, NULL, ?, ?, 0)
		`
		args = []interface{}{roleID, title, description, now, now}
	}

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to create project role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create project role",
			"",
		))
		return
	}

	logger.Info("Project role created",
		zap.String("role_id", roleID),
		zap.String("title", title),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectRoleCreate",
		"project_role",
		roleID,
		username,
		map[string]interface{}{
			"id":          roleID,
			"title":       title,
			"description": description,
			"projectId":   projectID,
			"isGlobal":    projectID == "",
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":          roleID,
		"title":       title,
		"description": description,
		"projectId":   projectID,
		"isGlobal":    projectID == "",
	})
	c.JSON(http.StatusCreated, response)
}

// handleProjectRoleRead reads a project role by ID
func (h *Handler) handleProjectRoleRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get role ID
	roleID, ok := req.Data["roleId"].(string)
	if !ok || roleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing roleId",
			"",
		))
		return
	}

	// Query project role
	query := `
		SELECT id, title, description, project_id, created, modified
		FROM project_role
		WHERE id = ? AND deleted = 0
	`

	var role models.ProjectRole
	var projectID sql.NullString

	err := h.db.QueryRow(c.Request.Context(), query, roleID).Scan(
		&role.ID,
		&role.Title,
		&role.Description,
		&projectID,
		&role.Created,
		&role.Modified,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project role not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read project role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read project role",
			"",
		))
		return
	}

	if projectID.Valid && projectID.String != "" {
		role.ProjectID = &projectID.String
	}

	logger.Info("Project role read",
		zap.String("role_id", roleID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":          role.ID,
		"title":       role.Title,
		"description": role.Description,
		"projectId":   projectID.String,
		"isGlobal":    !projectID.Valid || projectID.String == "",
		"created":     role.Created,
		"modified":    role.Modified,
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectRoleList lists all project roles
func (h *Handler) handleProjectRoleList(c *gin.Context, req *models.Request) {
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
		// List roles for specific project (includes both project-specific and global roles)
		query = `
			SELECT id, title, description, project_id, created, modified
			FROM project_role
			WHERE deleted = 0 AND (project_id = ? OR project_id IS NULL)
			ORDER BY title ASC
		`
		args = []interface{}{projectID}
	} else {
		// List all roles
		query = `
			SELECT id, title, description, project_id, created, modified
			FROM project_role
			WHERE deleted = 0
			ORDER BY title ASC
		`
	}

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list project roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list project roles",
			"",
		))
		return
	}
	defer rows.Close()

	roles := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title, description string
		var projectIDVal sql.NullString
		var created, modified int64

		err := rows.Scan(&id, &title, &description, &projectIDVal, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan project role", zap.Error(err))
			continue
		}

		roles = append(roles, map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": description,
			"projectId":   projectIDVal.String,
			"isGlobal":    !projectIDVal.Valid || projectIDVal.String == "",
			"created":     created,
			"modified":    modified,
		})
	}

	logger.Info("Project roles listed",
		zap.Int("count", len(roles)),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"roles": roles,
		"count": len(roles),
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectRoleModify updates a project role
func (h *Handler) handleProjectRoleModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_role", models.PermissionUpdate)
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

	// Get role ID
	roleID, ok := req.Data["roleId"].(string)
	if !ok || roleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing roleId",
			"",
		))
		return
	}

	// Check if role exists
	checkQuery := `SELECT COUNT(*) FROM project_role WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, roleID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project role not found",
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
	query := "UPDATE project_role SET "
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
	args = append(args, roleID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update project role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update project role",
			"",
		))
		return
	}

	logger.Info("Project role updated",
		zap.String("role_id", roleID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectRoleModify",
		"project_role",
		roleID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"roleId":  roleID,
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectRoleRemove soft-deletes a project role
func (h *Handler) handleProjectRoleRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_role", models.PermissionDelete)
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

	// Get role ID
	roleID, ok := req.Data["roleId"].(string)
	if !ok || roleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing roleId",
			"",
		))
		return
	}

	// Soft delete the role
	query := `UPDATE project_role SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), roleID)
	if err != nil {
		logger.Error("Failed to remove project role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove project role",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project role not found",
			"",
		))
		return
	}

	// Also soft delete all user mappings for this role
	_, err = h.db.Exec(c.Request.Context(),
		"UPDATE project_role_user_mapping SET deleted = 1 WHERE project_role_id = ?",
		roleID,
	)
	if err != nil {
		logger.Error("Failed to remove role user mappings", zap.Error(err))
		// Don't fail the request, role removal succeeded
	}

	logger.Info("Project role removed",
		zap.String("role_id", roleID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectRoleRemove",
		"project_role",
		roleID,
		username,
		map[string]interface{}{
			"roleId": roleID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed": true,
		"roleId":  roleID,
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectRoleAssignUser assigns a user to a project role
func (h *Handler) handleProjectRoleAssignUser(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_role", models.PermissionUpdate)
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

	// Get parameters
	roleID, ok := req.Data["roleId"].(string)
	if !ok || roleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing roleId",
			"",
		))
		return
	}

	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing userId",
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

	// Verify role exists
	checkQuery := `SELECT COUNT(*) FROM project_role WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, roleID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project role not found",
			"",
		))
		return
	}

	// Verify project exists
	checkQuery = `SELECT COUNT(*) FROM project WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, projectID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project not found",
			"",
		))
		return
	}

	// Check if mapping already exists
	checkQuery = `
		SELECT COUNT(*) FROM project_role_user_mapping
		WHERE project_role_id = ? AND project_id = ? AND user_id = ? AND deleted = 0
	`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, roleID, projectID, userID).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeEntityAlreadyExists,
			"User already assigned to this role in this project",
			"",
		))
		return
	}

	// Create mapping
	mappingID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO project_role_user_mapping (id, project_role_id, project_id, user_id, created, deleted)
		VALUES (?, ?, ?, ?, ?, 0)
	`

	_, err = h.db.Exec(c.Request.Context(), query, mappingID, roleID, projectID, userID, now)
	if err != nil {
		logger.Error("Failed to assign user to role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign user to role",
			"",
		))
		return
	}

	logger.Info("User assigned to project role",
		zap.String("mapping_id", mappingID),
		zap.String("role_id", roleID),
		zap.String("project_id", projectID),
		zap.String("user_id", userID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectRoleAssignUser",
		"project_role",
		roleID,
		username,
		map[string]interface{}{
			"mappingId": mappingID,
			"roleId":    roleID,
			"projectId": projectID,
			"userId":    userID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"assigned":  true,
		"mappingId": mappingID,
		"roleId":    roleID,
		"projectId": projectID,
		"userId":    userID,
	})
	c.JSON(http.StatusCreated, response)
}

// handleProjectRoleUnassignUser removes a user from a project role
func (h *Handler) handleProjectRoleUnassignUser(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "project_role", models.PermissionUpdate)
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

	// Get parameters
	roleID, ok := req.Data["roleId"].(string)
	if !ok || roleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing roleId",
			"",
		))
		return
	}

	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing userId",
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
	query := `
		UPDATE project_role_user_mapping
		SET deleted = 1
		WHERE project_role_id = ? AND project_id = ? AND user_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, roleID, projectID, userID)
	if err != nil {
		logger.Error("Failed to unassign user from role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unassign user from role",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"User role assignment not found",
			"",
		))
		return
	}

	logger.Info("User unassigned from project role",
		zap.String("role_id", roleID),
		zap.String("project_id", projectID),
		zap.String("user_id", userID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"projectRoleUnassignUser",
		"project_role",
		roleID,
		username,
		map[string]interface{}{
			"roleId":    roleID,
			"projectId": projectID,
			"userId":    userID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"unassigned": true,
		"roleId":     roleID,
		"projectId":  projectID,
		"userId":     userID,
	})
	c.JSON(http.StatusOK, response)
}

// handleProjectRoleListUsers lists all users assigned to a project role
func (h *Handler) handleProjectRoleListUsers(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get role ID
	roleID, ok := req.Data["roleId"].(string)
	if !ok || roleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing roleId",
			"",
		))
		return
	}

	// Optional project filter
	projectID := getStringFromData(req.Data, "projectId")

	var query string
	var args []interface{}

	if projectID != "" {
		// List users for specific project
		query = `
			SELECT m.id, m.user_id, m.project_id, m.created
			FROM project_role_user_mapping m
			WHERE m.project_role_id = ? AND m.project_id = ? AND m.deleted = 0
			ORDER BY m.created DESC
		`
		args = []interface{}{roleID, projectID}
	} else {
		// List all users across all projects
		query = `
			SELECT m.id, m.user_id, m.project_id, m.created
			FROM project_role_user_mapping m
			WHERE m.project_role_id = ? AND m.deleted = 0
			ORDER BY m.created DESC
		`
		args = []interface{}{roleID}
	}

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list role users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list role users",
			"",
		))
		return
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0)
	for rows.Next() {
		var mappingID, userID, projectIDVal string
		var created int64

		err := rows.Scan(&mappingID, &userID, &projectIDVal, &created)
		if err != nil {
			logger.Error("Failed to scan user mapping", zap.Error(err))
			continue
		}

		users = append(users, map[string]interface{}{
			"mappingId": mappingID,
			"userId":    userID,
			"projectId": projectIDVal,
			"created":   created,
		})
	}

	logger.Info("Role users listed",
		zap.String("role_id", roleID),
		zap.String("project_id", projectID),
		zap.Int("count", len(users)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"roleId":  roleID,
		"users":   users,
		"count":   len(users),
	})
	c.JSON(http.StatusOK, response)
}
