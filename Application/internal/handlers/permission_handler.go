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

// handlePermissionCreate creates a new permission
func (h *Handler) handlePermissionCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionCreate)
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

	// Parse permission data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	value, ok := req.Data["value"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing value",
			"",
		))
		return
	}

	permission := &models.Permission{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Value:       int(value),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Validate permission value
	if !permission.IsValidPermissionValue() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid permission value (must be 1=READ, 2=CREATE, 3=UPDATE, or 5=DELETE)",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO permission (id, title, description, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		permission.ID,
		permission.Title,
		permission.Description,
		permission.Value,
		permission.Created,
		permission.Modified,
		permission.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create permission",
			"",
		))
		return
	}

	logger.Info("Permission created",
		zap.String("permission_id", permission.ID),
		zap.String("title", permission.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"permission": permission,
	})
	c.JSON(http.StatusCreated, response)
}

// handlePermissionRead reads a single permission by ID
func (h *Handler) handlePermissionRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get permission ID from request
	permissionID, ok := req.Data["id"].(string)
	if !ok || permissionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permission ID",
			"",
		))
		return
	}

	// Query permission from database
	query := `
		SELECT id, title, description, value, created, modified, deleted
		FROM permission
		WHERE id = ? AND deleted = 0
	`

	var permission models.Permission
	err := h.db.QueryRow(c.Request.Context(), query, permissionID).Scan(
		&permission.ID,
		&permission.Title,
		&permission.Description,
		&permission.Value,
		&permission.Created,
		&permission.Modified,
		&permission.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read permission",
			"",
		))
		return
	}

	logger.Info("Permission read",
		zap.String("permission_id", permission.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"permission": permission,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionList lists all permissions
func (h *Handler) handlePermissionList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted permissions
	query := `
		SELECT id, title, description, value, created, modified, deleted
		FROM permission
		WHERE deleted = 0
		ORDER BY value ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list permissions",
			"",
		))
		return
	}
	defer rows.Close()

	permissions := make([]models.Permission, 0)
	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(
			&permission.ID,
			&permission.Title,
			&permission.Description,
			&permission.Value,
			&permission.Created,
			&permission.Modified,
			&permission.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan permission", zap.Error(err))
			continue
		}
		permissions = append(permissions, permission)
	}

	logger.Info("Permissions listed",
		zap.Int("count", len(permissions)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"permissions": permissions,
		"count":       len(permissions),
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionModify updates an existing permission
func (h *Handler) handlePermissionModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionUpdate)
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

	// Get permission ID
	permissionID, ok := req.Data["id"].(string)
	if !ok || permissionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permission ID",
			"",
		))
		return
	}

	// Check if permission exists
	checkQuery := `SELECT COUNT(*) FROM permission WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, permissionID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission not found",
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
	if value, ok := req.Data["value"].(float64); ok {
		valueInt := int(value)
		// Validate permission value
		if valueInt != models.PermissionRead && valueInt != models.PermissionCreate &&
			valueInt != models.PermissionUpdate && valueInt != models.PermissionDelete {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid permission value",
				"",
			))
			return
		}
		updates["value"] = valueInt
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
	query := "UPDATE permission SET "
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
	args = append(args, permissionID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update permission",
			"",
		))
		return
	}

	logger.Info("Permission updated",
		zap.String("permission_id", permissionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      permissionID,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionRemove soft-deletes a permission
func (h *Handler) handlePermissionRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionDelete)
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

	// Get permission ID
	permissionID, ok := req.Data["id"].(string)
	if !ok || permissionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permission ID",
			"",
		))
		return
	}

	// Soft delete the permission
	query := `UPDATE permission SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), permissionID)
	if err != nil {
		logger.Error("Failed to delete permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete permission",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission not found",
			"",
		))
		return
	}

	logger.Info("Permission deleted",
		zap.String("permission_id", permissionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      permissionID,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionContextCreate creates a new permission context
func (h *Handler) handlePermissionContextCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionCreate)
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

	// Parse context data from request
	context, ok := req.Data["context"].(string)
	if !ok || context == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing context",
			"",
		))
		return
	}

	permContext := &models.PermissionContext{
		ID:       uuid.New().String(),
		Context:  context,
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
		Deleted:  false,
	}

	// Validate context
	if !permContext.IsValidContext() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid context (must be: node, account, organization, team, or project)",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO permission_context (id, context, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		permContext.ID,
		permContext.Context,
		permContext.Created,
		permContext.Modified,
		permContext.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create permission context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create permission context",
			"",
		))
		return
	}

	logger.Info("Permission context created",
		zap.String("context_id", permContext.ID),
		zap.String("context", permContext.Context),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"permissionContext": permContext,
	})
	c.JSON(http.StatusCreated, response)
}

// handlePermissionContextRead reads a single permission context by ID
func (h *Handler) handlePermissionContextRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get context ID from request
	contextID, ok := req.Data["id"].(string)
	if !ok || contextID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing context ID",
			"",
		))
		return
	}

	// Query context from database
	query := `
		SELECT id, context, created, modified, deleted
		FROM permission_context
		WHERE id = ? AND deleted = 0
	`

	var permContext models.PermissionContext
	err := h.db.QueryRow(c.Request.Context(), query, contextID).Scan(
		&permContext.ID,
		&permContext.Context,
		&permContext.Created,
		&permContext.Modified,
		&permContext.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission context not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read permission context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read permission context",
			"",
		))
		return
	}

	logger.Info("Permission context read",
		zap.String("context_id", permContext.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"permissionContext": permContext,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionContextList lists all permission contexts
func (h *Handler) handlePermissionContextList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted contexts
	query := `
		SELECT id, context, created, modified, deleted
		FROM permission_context
		WHERE deleted = 0
		ORDER BY context ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list permission contexts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list permission contexts",
			"",
		))
		return
	}
	defer rows.Close()

	contexts := make([]models.PermissionContext, 0)
	for rows.Next() {
		var permContext models.PermissionContext
		err := rows.Scan(
			&permContext.ID,
			&permContext.Context,
			&permContext.Created,
			&permContext.Modified,
			&permContext.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan permission context", zap.Error(err))
			continue
		}
		contexts = append(contexts, permContext)
	}

	logger.Info("Permission contexts listed",
		zap.Int("count", len(contexts)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"permissionContexts": contexts,
		"count":              len(contexts),
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionContextModify updates an existing permission context
func (h *Handler) handlePermissionContextModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionUpdate)
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

	// Get context ID
	contextID, ok := req.Data["id"].(string)
	if !ok || contextID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing context ID",
			"",
		))
		return
	}

	// Get new context value
	context, ok := req.Data["context"].(string)
	if !ok || context == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing context value",
			"",
		))
		return
	}

	// Validate context
	validContexts := []string{"node", "account", "organization", "team", "project"}
	valid := false
	for _, ctx := range validContexts {
		if context == ctx {
			valid = true
			break
		}
	}

	if !valid {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid context value",
			"",
		))
		return
	}

	// Update context
	query := `UPDATE permission_context SET context = ?, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, context, time.Now().Unix(), contextID)
	if err != nil {
		logger.Error("Failed to update permission context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update permission context",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission context not found",
			"",
		))
		return
	}

	logger.Info("Permission context updated",
		zap.String("context_id", contextID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      contextID,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionContextRemove soft-deletes a permission context
func (h *Handler) handlePermissionContextRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionDelete)
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

	// Get context ID
	contextID, ok := req.Data["id"].(string)
	if !ok || contextID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing context ID",
			"",
		))
		return
	}

	// Soft delete the context
	query := `UPDATE permission_context SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), contextID)
	if err != nil {
		logger.Error("Failed to delete permission context", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete permission context",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission context not found",
			"",
		))
		return
	}

	logger.Info("Permission context deleted",
		zap.String("context_id", contextID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      contextID,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionAssignUser assigns a permission to a user within a context
func (h *Handler) handlePermissionAssignUser(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionCreate)
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

	// Parse mapping data
	permissionID, ok := req.Data["permissionId"].(string)
	if !ok || permissionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permissionId",
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

	contextID, ok := req.Data["permissionContextId"].(string)
	if !ok || contextID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permissionContextId",
			"",
		))
		return
	}

	mapping := &models.PermissionUserMapping{
		ID:                  uuid.New().String(),
		PermissionID:        permissionID,
		UserID:              userID,
		PermissionContextID: contextID,
		Created:             time.Now().Unix(),
		Deleted:             false,
	}

	// Insert into database
	query := `
		INSERT INTO permission_user_mapping (id, permission_id, user_id, permission_context_id, created, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.PermissionID,
		mapping.UserID,
		mapping.PermissionContextID,
		mapping.Created,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to assign permission to user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign permission to user",
			"",
		))
		return
	}

	logger.Info("Permission assigned to user",
		zap.String("mapping_id", mapping.ID),
		zap.String("permission_id", permissionID),
		zap.String("user_id", userID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handlePermissionUnassignUser removes a permission from a user
func (h *Handler) handlePermissionUnassignUser(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionDelete)
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

	// Get mapping ID
	mappingID, ok := req.Data["id"].(string)
	if !ok || mappingID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing mapping ID",
			"",
		))
		return
	}

	// Soft delete the mapping
	query := `UPDATE permission_user_mapping SET deleted = 1 WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, mappingID)
	if err != nil {
		logger.Error("Failed to unassign permission from user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unassign permission from user",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission mapping not found",
			"",
		))
		return
	}

	logger.Info("Permission unassigned from user",
		zap.String("mapping_id", mappingID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      mappingID,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionAssignTeam assigns a permission to a team within a context
func (h *Handler) handlePermissionAssignTeam(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionCreate)
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

	// Parse mapping data
	permissionID, ok := req.Data["permissionId"].(string)
	if !ok || permissionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permissionId",
			"",
		))
		return
	}

	teamID, ok := req.Data["teamId"].(string)
	if !ok || teamID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing teamId",
			"",
		))
		return
	}

	contextID, ok := req.Data["permissionContextId"].(string)
	if !ok || contextID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permissionContextId",
			"",
		))
		return
	}

	mapping := &models.PermissionTeamMapping{
		ID:                  uuid.New().String(),
		PermissionID:        permissionID,
		TeamID:              teamID,
		PermissionContextID: contextID,
		Created:             time.Now().Unix(),
		Deleted:             false,
	}

	// Insert into database
	query := `
		INSERT INTO permission_team_mapping (id, permission_id, team_id, permission_context_id, created, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.PermissionID,
		mapping.TeamID,
		mapping.PermissionContextID,
		mapping.Created,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to assign permission to team", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign permission to team",
			"",
		))
		return
	}

	logger.Info("Permission assigned to team",
		zap.String("mapping_id", mapping.ID),
		zap.String("permission_id", permissionID),
		zap.String("team_id", teamID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handlePermissionUnassignTeam removes a permission from a team
func (h *Handler) handlePermissionUnassignTeam(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "permission", models.PermissionDelete)
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

	// Get mapping ID
	mappingID, ok := req.Data["id"].(string)
	if !ok || mappingID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing mapping ID",
			"",
		))
		return
	}

	// Soft delete the mapping
	query := `UPDATE permission_team_mapping SET deleted = 1 WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, mappingID)
	if err != nil {
		logger.Error("Failed to unassign permission from team", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unassign permission from team",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Permission mapping not found",
			"",
		))
		return
	}

	logger.Info("Permission unassigned from team",
		zap.String("mapping_id", mappingID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      mappingID,
	})
	c.JSON(http.StatusOK, response)
}

// handlePermissionCheck checks if a user has a specific permission
func (h *Handler) handlePermissionCheck(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get parameters
	targetUser, ok := req.Data["userId"].(string)
	if !ok || targetUser == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing userId",
			"",
		))
		return
	}

	resource, ok := req.Data["resource"].(string)
	if !ok || resource == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing resource",
			"",
		))
		return
	}

	permValue, ok := req.Data["permission"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing permission value",
			"",
		))
		return
	}

	// Use permission service to check
	allowed, err := h.permService.CheckPermission(c.Request.Context(), targetUser, resource, int(permValue))
	if err != nil {
		logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodePermissionServiceError,
			"Permission check failed",
			"",
		))
		return
	}

	logger.Info("Permission checked",
		zap.String("target_user", targetUser),
		zap.String("resource", resource),
		zap.Int("permission", int(permValue)),
		zap.Bool("allowed", allowed),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"allowed": allowed,
	})
	c.JSON(http.StatusOK, response)
}
