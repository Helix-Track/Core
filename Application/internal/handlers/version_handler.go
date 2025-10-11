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

// handleVersionCreate creates a new version
func (h *Handler) handleVersionCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionCreate)
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

	// Parse version data from request
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

	version := &models.Version{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		ProjectID:   projectID,
		Released:    false,
		Archived:    false,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Handle optional start_date
	if startDate, ok := req.Data["startDate"].(float64); ok {
		timestamp := int64(startDate)
		version.StartDate = &timestamp
	}

	// Handle optional release_date
	if releaseDate, ok := req.Data["releaseDate"].(float64); ok {
		timestamp := int64(releaseDate)
		version.ReleaseDate = &timestamp
	}

	// Insert into database
	query := `
		INSERT INTO version (id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		version.ID,
		version.Title,
		version.Description,
		version.ProjectID,
		version.StartDate,
		version.ReleaseDate,
		version.Released,
		version.Archived,
		version.Created,
		version.Modified,
		version.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create version",
			"",
		))
		return
	}

	logger.Info("Version created",
		zap.String("version_id", version.ID),
		zap.String("title", version.Title),
		zap.String("project_id", version.ProjectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"version": version,
	})
	c.JSON(http.StatusCreated, response)
}

// handleVersionRead reads a single version by ID
func (h *Handler) handleVersionRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get version ID from request
	versionID, ok := req.Data["id"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing version ID",
			"",
		))
		return
	}

	// Query version from database
	query := `
		SELECT id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted
		FROM version
		WHERE id = ? AND deleted = 0
	`

	var version models.Version
	err := h.db.QueryRow(c.Request.Context(), query, versionID).Scan(
		&version.ID,
		&version.Title,
		&version.Description,
		&version.ProjectID,
		&version.StartDate,
		&version.ReleaseDate,
		&version.Released,
		&version.Archived,
		&version.Created,
		&version.Modified,
		&version.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Version not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read version",
			"",
		))
		return
	}

	logger.Info("Version read",
		zap.String("version_id", version.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"version": version,
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionList lists all versions, optionally filtered by project_id
func (h *Handler) handleVersionList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Check if filtering by project_id
	projectID, hasProjectFilter := req.Data["projectId"].(string)

	var query string
	var args []interface{}

	if hasProjectFilter && projectID != "" {
		// Filter by project_id
		query = `
			SELECT id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted
			FROM version
			WHERE project_id = ? AND deleted = 0
			ORDER BY created DESC
		`
		args = append(args, projectID)
	} else {
		// List all versions
		query = `
			SELECT id, title, description, project_id, start_date, release_date, released, archived, created, modified, deleted
			FROM version
			WHERE deleted = 0
			ORDER BY created DESC
		`
	}

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list versions",
			"",
		))
		return
	}
	defer rows.Close()

	versions := make([]models.Version, 0)
	for rows.Next() {
		var version models.Version
		err := rows.Scan(
			&version.ID,
			&version.Title,
			&version.Description,
			&version.ProjectID,
			&version.StartDate,
			&version.ReleaseDate,
			&version.Released,
			&version.Archived,
			&version.Created,
			&version.Modified,
			&version.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan version", zap.Error(err))
			continue
		}
		versions = append(versions, version)
	}

	logger.Info("Versions listed",
		zap.Int("count", len(versions)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"versions": versions,
		"count":    len(versions),
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionModify updates an existing version
func (h *Handler) handleVersionModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionUpdate)
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

	// Get version ID
	versionID, ok := req.Data["id"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing version ID",
			"",
		))
		return
	}

	// Check if version exists
	checkQuery := `SELECT COUNT(*) FROM version WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, versionID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Version not found",
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
	if startDate, ok := req.Data["startDate"].(float64); ok {
		timestamp := int64(startDate)
		updates["start_date"] = &timestamp
	}
	if releaseDate, ok := req.Data["releaseDate"].(float64); ok {
		timestamp := int64(releaseDate)
		updates["release_date"] = &timestamp
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
	query := "UPDATE version SET "
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
	args = append(args, versionID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update version",
			"",
		))
		return
	}

	logger.Info("Version updated",
		zap.String("version_id", versionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      versionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionRemove soft-deletes a version
func (h *Handler) handleVersionRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionDelete)
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

	// Get version ID
	versionID, ok := req.Data["id"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing version ID",
			"",
		))
		return
	}

	// Soft delete the version
	query := `UPDATE version SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), versionID)
	if err != nil {
		logger.Error("Failed to delete version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete version",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Version not found",
			"",
		))
		return
	}

	logger.Info("Version deleted",
		zap.String("version_id", versionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      versionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionRelease marks a version as released
func (h *Handler) handleVersionRelease(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionUpdate)
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

	// Get version ID
	versionID, ok := req.Data["id"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing version ID",
			"",
		))
		return
	}

	// Mark version as released and set release_date to now if not already set
	now := time.Now().Unix()
	query := `
		UPDATE version
		SET released = 1,
		    release_date = COALESCE(release_date, ?),
		    modified = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, now, now, versionID)
	if err != nil {
		logger.Error("Failed to release version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to release version",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Version not found",
			"",
		))
		return
	}

	logger.Info("Version released",
		zap.String("version_id", versionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"released": true,
		"id":       versionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionArchive archives a version
func (h *Handler) handleVersionArchive(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionUpdate)
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

	// Get version ID
	versionID, ok := req.Data["id"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing version ID",
			"",
		))
		return
	}

	// Mark version as archived
	query := `UPDATE version SET archived = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), versionID)
	if err != nil {
		logger.Error("Failed to archive version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to archive version",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Version not found",
			"",
		))
		return
	}

	logger.Info("Version archived",
		zap.String("version_id", versionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"archived": true,
		"id":       versionID,
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionAddAffected adds an affected version mapping to a ticket
func (h *Handler) handleVersionAddAffected(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionCreate)
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

	// Parse ticket and version IDs from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	versionID, ok := req.Data["versionId"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing versionId",
			"",
		))
		return
	}

	mapping := &models.TicketVersionMapping{
		ID:        uuid.New().String(),
		TicketID:  ticketID,
		VersionID: versionID,
		Created:   time.Now().Unix(),
		Deleted:   false,
	}

	// Insert into database
	query := `
		INSERT INTO ticket_affected_version_mapping (id, ticket_id, version_id, created, deleted)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.TicketID,
		mapping.VersionID,
		mapping.Created,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add affected version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add affected version",
			"",
		))
		return
	}

	logger.Info("Affected version added",
		zap.String("mapping_id", mapping.ID),
		zap.String("ticket_id", mapping.TicketID),
		zap.String("version_id", mapping.VersionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleVersionRemoveAffected removes an affected version mapping from a ticket
func (h *Handler) handleVersionRemoveAffected(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionDelete)
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

	// Parse ticket and version IDs from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	versionID, ok := req.Data["versionId"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing versionId",
			"",
		))
		return
	}

	// Soft delete the mapping
	query := `
		UPDATE ticket_affected_version_mapping
		SET deleted = 1
		WHERE ticket_id = ? AND version_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, ticketID, versionID)
	if err != nil {
		logger.Error("Failed to remove affected version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove affected version",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Affected version mapping not found",
			"",
		))
		return
	}

	logger.Info("Affected version removed",
		zap.String("ticket_id", ticketID),
		zap.String("version_id", versionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed": true,
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionListAffected lists all affected versions for a ticket
func (h *Handler) handleVersionListAffected(c *gin.Context, req *models.Request) {
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

	// Query all affected versions for the ticket
	query := `
		SELECT v.id, v.title, v.description, v.project_id, v.start_date, v.release_date, v.released, v.archived, v.created, v.modified, v.deleted
		FROM version v
		INNER JOIN ticket_affected_version_mapping m ON v.id = m.version_id
		WHERE m.ticket_id = ? AND m.deleted = 0 AND v.deleted = 0
		ORDER BY v.created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list affected versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list affected versions",
			"",
		))
		return
	}
	defer rows.Close()

	versions := make([]models.Version, 0)
	for rows.Next() {
		var version models.Version
		err := rows.Scan(
			&version.ID,
			&version.Title,
			&version.Description,
			&version.ProjectID,
			&version.StartDate,
			&version.ReleaseDate,
			&version.Released,
			&version.Archived,
			&version.Created,
			&version.Modified,
			&version.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan affected version", zap.Error(err))
			continue
		}
		versions = append(versions, version)
	}

	logger.Info("Affected versions listed",
		zap.String("ticket_id", ticketID),
		zap.Int("count", len(versions)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"versions": versions,
		"count":    len(versions),
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionAddFix adds a fix version mapping to a ticket
func (h *Handler) handleVersionAddFix(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionCreate)
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

	// Parse ticket and version IDs from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	versionID, ok := req.Data["versionId"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing versionId",
			"",
		))
		return
	}

	mapping := &models.TicketVersionMapping{
		ID:        uuid.New().String(),
		TicketID:  ticketID,
		VersionID: versionID,
		Created:   time.Now().Unix(),
		Deleted:   false,
	}

	// Insert into database
	query := `
		INSERT INTO ticket_fix_version_mapping (id, ticket_id, version_id, created, deleted)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.TicketID,
		mapping.VersionID,
		mapping.Created,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add fix version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add fix version",
			"",
		))
		return
	}

	logger.Info("Fix version added",
		zap.String("mapping_id", mapping.ID),
		zap.String("ticket_id", mapping.TicketID),
		zap.String("version_id", mapping.VersionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleVersionRemoveFix removes a fix version mapping from a ticket
func (h *Handler) handleVersionRemoveFix(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "version", models.PermissionDelete)
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

	// Parse ticket and version IDs from request
	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticketId",
			"",
		))
		return
	}

	versionID, ok := req.Data["versionId"].(string)
	if !ok || versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing versionId",
			"",
		))
		return
	}

	// Soft delete the mapping
	query := `
		UPDATE ticket_fix_version_mapping
		SET deleted = 1
		WHERE ticket_id = ? AND version_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, ticketID, versionID)
	if err != nil {
		logger.Error("Failed to remove fix version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove fix version",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Fix version mapping not found",
			"",
		))
		return
	}

	logger.Info("Fix version removed",
		zap.String("ticket_id", ticketID),
		zap.String("version_id", versionID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed": true,
	})
	c.JSON(http.StatusOK, response)
}

// handleVersionListFix lists all fix versions for a ticket
func (h *Handler) handleVersionListFix(c *gin.Context, req *models.Request) {
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

	// Query all fix versions for the ticket
	query := `
		SELECT v.id, v.title, v.description, v.project_id, v.start_date, v.release_date, v.released, v.archived, v.created, v.modified, v.deleted
		FROM version v
		INNER JOIN ticket_fix_version_mapping m ON v.id = m.version_id
		WHERE m.ticket_id = ? AND m.deleted = 0 AND v.deleted = 0
		ORDER BY v.created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list fix versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list fix versions",
			"",
		))
		return
	}
	defer rows.Close()

	versions := make([]models.Version, 0)
	for rows.Next() {
		var version models.Version
		err := rows.Scan(
			&version.ID,
			&version.Title,
			&version.Description,
			&version.ProjectID,
			&version.StartDate,
			&version.ReleaseDate,
			&version.Released,
			&version.Archived,
			&version.Created,
			&version.Modified,
			&version.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan fix version", zap.Error(err))
			continue
		}
		versions = append(versions, version)
	}

	logger.Info("Fix versions listed",
		zap.String("ticket_id", ticketID),
		zap.Int("count", len(versions)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"versions": versions,
		"count":    len(versions),
	})
	c.JSON(http.StatusOK, response)
}
