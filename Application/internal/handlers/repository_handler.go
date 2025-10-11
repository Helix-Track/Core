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

// ===== Repository CRUD Operations =====

// handleRepositoryCreate creates a new repository
func (h *Handler) handleRepositoryCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository", models.PermissionCreate)
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

	// Parse repository data
	repository, ok := req.Data["repository"].(string)
	if !ok || repository == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository",
			"",
		))
		return
	}

	repositoryTypeID, ok := req.Data["repository_type_id"].(string)
	if !ok || repositoryTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository_type_id",
			"",
		))
		return
	}

	repo := &models.Repository{
		ID:               uuid.New().String(),
		Repository:       repository,
		Description:      getStringFromData(req.Data, "description"),
		RepositoryTypeID: repositoryTypeID,
		Created:          time.Now().Unix(),
		Modified:         time.Now().Unix(),
		Deleted:          false,
	}

	// Insert into database
	query := `
		INSERT INTO repository (id, repository, description, repository_type_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		repo.ID,
		repo.Repository,
		repo.Description,
		repo.RepositoryTypeID,
		repo.Created,
		repo.Modified,
		repo.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create repository",
			"",
		))
		return
	}

	logger.Info("Repository created",
		zap.String("repository_id", repo.ID),
		zap.String("repository", repo.Repository),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"repository": repo,
	})
	c.JSON(http.StatusCreated, response)
}

// handleRepositoryRead reads a single repository by ID
func (h *Handler) handleRepositoryRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	repositoryID, ok := req.Data["id"].(string)
	if !ok || repositoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository ID",
			"",
		))
		return
	}

	query := `
		SELECT id, repository, description, repository_type_id, created, modified, deleted
		FROM repository
		WHERE id = ? AND deleted = 0
	`

	var repo models.Repository
	err := h.db.QueryRow(c.Request.Context(), query, repositoryID).Scan(
		&repo.ID,
		&repo.Repository,
		&repo.Description,
		&repo.RepositoryTypeID,
		&repo.Created,
		&repo.Modified,
		&repo.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Repository not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read repository",
			"",
		))
		return
	}

	logger.Info("Repository read",
		zap.String("repository_id", repo.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"repository": repo,
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryList lists all repositories
func (h *Handler) handleRepositoryList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	query := `
		SELECT id, repository, description, repository_type_id, created, modified, deleted
		FROM repository
		WHERE deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list repositories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list repositories",
			"",
		))
		return
	}
	defer rows.Close()

	repositories := make([]models.Repository, 0)
	for rows.Next() {
		var repo models.Repository
		err := rows.Scan(
			&repo.ID,
			&repo.Repository,
			&repo.Description,
			&repo.RepositoryTypeID,
			&repo.Created,
			&repo.Modified,
			&repo.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan repository", zap.Error(err))
			continue
		}
		repositories = append(repositories, repo)
	}

	logger.Info("Repositories listed",
		zap.Int("count", len(repositories)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"repositories": repositories,
		"count":        len(repositories),
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryModify updates an existing repository
func (h *Handler) handleRepositoryModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository", models.PermissionUpdate)
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

	repositoryID, ok := req.Data["id"].(string)
	if !ok || repositoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository ID",
			"",
		))
		return
	}

	// Check if repository exists
	checkQuery := `SELECT COUNT(*) FROM repository WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, repositoryID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Repository not found",
			"",
		))
		return
	}

	// Build update query dynamically
	updates := make(map[string]interface{})

	if repository, ok := req.Data["repository"].(string); ok && repository != "" {
		updates["repository"] = repository
	}
	if description, ok := req.Data["description"].(string); ok {
		updates["description"] = description
	}
	if repositoryTypeID, ok := req.Data["repository_type_id"].(string); ok && repositoryTypeID != "" {
		updates["repository_type_id"] = repositoryTypeID
	}

	updates["modified"] = time.Now().Unix()

	if len(updates) == 1 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"No fields to update",
			"",
		))
		return
	}

	query := "UPDATE repository SET "
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
	args = append(args, repositoryID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update repository",
			"",
		))
		return
	}

	logger.Info("Repository updated",
		zap.String("repository_id", repositoryID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      repositoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryRemove soft-deletes a repository
func (h *Handler) handleRepositoryRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository", models.PermissionDelete)
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

	repositoryID, ok := req.Data["id"].(string)
	if !ok || repositoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository ID",
			"",
		))
		return
	}

	query := `UPDATE repository SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), repositoryID)
	if err != nil {
		logger.Error("Failed to delete repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete repository",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Repository not found",
			"",
		))
		return
	}

	logger.Info("Repository deleted",
		zap.String("repository_id", repositoryID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      repositoryID,
	})
	c.JSON(http.StatusOK, response)
}

// ===== Repository Type CRUD Operations =====

// handleRepositoryTypeCreate creates a new repository type
func (h *Handler) handleRepositoryTypeCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository_type", models.PermissionCreate)
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

	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	repoType := &models.RepositoryType{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	query := `
		INSERT INTO repository_type (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		repoType.ID,
		repoType.Title,
		repoType.Description,
		repoType.Created,
		repoType.Modified,
		repoType.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create repository type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create repository type",
			"",
		))
		return
	}

	logger.Info("Repository type created",
		zap.String("repository_type_id", repoType.ID),
		zap.String("title", repoType.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"repository_type": repoType,
	})
	c.JSON(http.StatusCreated, response)
}

// handleRepositoryTypeRead reads a single repository type by ID
func (h *Handler) handleRepositoryTypeRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	repositoryTypeID, ok := req.Data["id"].(string)
	if !ok || repositoryTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository type ID",
			"",
		))
		return
	}

	query := `
		SELECT id, title, description, created, modified, deleted
		FROM repository_type
		WHERE id = ? AND deleted = 0
	`

	var repoType models.RepositoryType
	err := h.db.QueryRow(c.Request.Context(), query, repositoryTypeID).Scan(
		&repoType.ID,
		&repoType.Title,
		&repoType.Description,
		&repoType.Created,
		&repoType.Modified,
		&repoType.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Repository type not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read repository type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read repository type",
			"",
		))
		return
	}

	logger.Info("Repository type read",
		zap.String("repository_type_id", repoType.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"repository_type": repoType,
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryTypeList lists all repository types
func (h *Handler) handleRepositoryTypeList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	query := `
		SELECT id, title, description, created, modified, deleted
		FROM repository_type
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list repository types", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list repository types",
			"",
		))
		return
	}
	defer rows.Close()

	repoTypes := make([]models.RepositoryType, 0)
	for rows.Next() {
		var repoType models.RepositoryType
		err := rows.Scan(
			&repoType.ID,
			&repoType.Title,
			&repoType.Description,
			&repoType.Created,
			&repoType.Modified,
			&repoType.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan repository type", zap.Error(err))
			continue
		}
		repoTypes = append(repoTypes, repoType)
	}

	logger.Info("Repository types listed",
		zap.Int("count", len(repoTypes)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"repository_types": repoTypes,
		"count":            len(repoTypes),
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryTypeModify updates an existing repository type
func (h *Handler) handleRepositoryTypeModify(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository_type", models.PermissionUpdate)
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

	repositoryTypeID, ok := req.Data["id"].(string)
	if !ok || repositoryTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository type ID",
			"",
		))
		return
	}

	checkQuery := `SELECT COUNT(*) FROM repository_type WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, repositoryTypeID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Repository type not found",
			"",
		))
		return
	}

	updates := make(map[string]interface{})

	if title, ok := req.Data["title"].(string); ok && title != "" {
		updates["title"] = title
	}
	if description, ok := req.Data["description"].(string); ok {
		updates["description"] = description
	}

	updates["modified"] = time.Now().Unix()

	if len(updates) == 1 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"No fields to update",
			"",
		))
		return
	}

	query := "UPDATE repository_type SET "
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
	args = append(args, repositoryTypeID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update repository type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update repository type",
			"",
		))
		return
	}

	logger.Info("Repository type updated",
		zap.String("repository_type_id", repositoryTypeID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      repositoryTypeID,
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryTypeRemove soft-deletes a repository type
func (h *Handler) handleRepositoryTypeRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository_type", models.PermissionDelete)
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

	repositoryTypeID, ok := req.Data["id"].(string)
	if !ok || repositoryTypeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository type ID",
			"",
		))
		return
	}

	query := `UPDATE repository_type SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), repositoryTypeID)
	if err != nil {
		logger.Error("Failed to delete repository type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete repository type",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Repository type not found",
			"",
		))
		return
	}

	logger.Info("Repository type deleted",
		zap.String("repository_type_id", repositoryTypeID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      repositoryTypeID,
	})
	c.JSON(http.StatusOK, response)
}

// ===== Repository-Project Mapping Operations =====

// handleRepositoryAssignProject assigns a repository to a project
func (h *Handler) handleRepositoryAssignProject(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository", models.PermissionUpdate)
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

	repositoryID, ok := req.Data["repository_id"].(string)
	if !ok || repositoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository_id",
			"",
		))
		return
	}

	projectID, ok := req.Data["project_id"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project_id",
			"",
		))
		return
	}

	mapping := &models.RepositoryProjectMapping{
		ID:           uuid.New().String(),
		RepositoryID: repositoryID,
		ProjectID:    projectID,
		Created:      time.Now().Unix(),
		Modified:     time.Now().Unix(),
		Deleted:      false,
	}

	query := `
		INSERT INTO repository_project_mapping (id, repository_id, project_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.RepositoryID,
		mapping.ProjectID,
		mapping.Created,
		mapping.Modified,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to assign repository to project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign repository to project",
			"",
		))
		return
	}

	logger.Info("Repository assigned to project",
		zap.String("repository_id", repositoryID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleRepositoryUnassignProject unassigns a repository from a project
func (h *Handler) handleRepositoryUnassignProject(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository", models.PermissionUpdate)
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

	repositoryID, ok := req.Data["repository_id"].(string)
	if !ok || repositoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository_id",
			"",
		))
		return
	}

	projectID, ok := req.Data["project_id"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project_id",
			"",
		))
		return
	}

	query := `
		UPDATE repository_project_mapping
		SET deleted = 1, modified = ?
		WHERE repository_id = ? AND project_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), repositoryID, projectID)
	if err != nil {
		logger.Error("Failed to unassign repository from project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unassign repository from project",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Repository-project mapping not found",
			"",
		))
		return
	}

	logger.Info("Repository unassigned from project",
		zap.String("repository_id", repositoryID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"unassigned": true,
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryListProjects lists all projects for a repository
func (h *Handler) handleRepositoryListProjects(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	repositoryID, ok := req.Data["repository_id"].(string)
	if !ok || repositoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository_id",
			"",
		))
		return
	}

	query := `
		SELECT rpm.id, rpm.repository_id, rpm.project_id, rpm.created, rpm.modified, rpm.deleted
		FROM repository_project_mapping rpm
		WHERE rpm.repository_id = ? AND rpm.deleted = 0
		ORDER BY rpm.created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, repositoryID)
	if err != nil {
		logger.Error("Failed to list projects for repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list projects for repository",
			"",
		))
		return
	}
	defer rows.Close()

	mappings := make([]models.RepositoryProjectMapping, 0)
	for rows.Next() {
		var mapping models.RepositoryProjectMapping
		err := rows.Scan(
			&mapping.ID,
			&mapping.RepositoryID,
			&mapping.ProjectID,
			&mapping.Created,
			&mapping.Modified,
			&mapping.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan repository-project mapping", zap.Error(err))
			continue
		}
		mappings = append(mappings, mapping)
	}

	logger.Info("Projects listed for repository",
		zap.String("repository_id", repositoryID),
		zap.Int("count", len(mappings)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mappings": mappings,
		"count":    len(mappings),
	})
	c.JSON(http.StatusOK, response)
}

// ===== Repository Commit-Ticket Mapping Operations =====

// handleRepositoryAddCommit adds a commit to a ticket
func (h *Handler) handleRepositoryAddCommit(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository", models.PermissionCreate)
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

	repositoryID, ok := req.Data["repository_id"].(string)
	if !ok || repositoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing repository_id",
			"",
		))
		return
	}

	ticketID, ok := req.Data["ticket_id"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket_id",
			"",
		))
		return
	}

	commitHash, ok := req.Data["commit_hash"].(string)
	if !ok || commitHash == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing commit_hash",
			"",
		))
		return
	}

	mapping := &models.RepositoryCommitTicketMapping{
		ID:           uuid.New().String(),
		RepositoryID: repositoryID,
		TicketID:     ticketID,
		CommitHash:   commitHash,
		Created:      time.Now().Unix(),
		Modified:     time.Now().Unix(),
		Deleted:      false,
	}

	query := `
		INSERT INTO repository_commit_ticket_mapping (id, repository_id, ticket_id, commit_hash, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.RepositoryID,
		mapping.TicketID,
		mapping.CommitHash,
		mapping.Created,
		mapping.Modified,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add commit to ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add commit to ticket",
			"",
		))
		return
	}

	logger.Info("Commit added to ticket",
		zap.String("repository_id", repositoryID),
		zap.String("ticket_id", ticketID),
		zap.String("commit_hash", commitHash),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleRepositoryRemoveCommit removes a commit from a ticket
func (h *Handler) handleRepositoryRemoveCommit(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "repository", models.PermissionDelete)
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

	commitHash, ok := req.Data["commit_hash"].(string)
	if !ok || commitHash == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing commit_hash",
			"",
		))
		return
	}

	query := `
		UPDATE repository_commit_ticket_mapping
		SET deleted = 1, modified = ?
		WHERE commit_hash = ? AND deleted = 0
	`

	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), commitHash)
	if err != nil {
		logger.Error("Failed to remove commit from ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove commit from ticket",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Commit mapping not found",
			"",
		))
		return
	}

	logger.Info("Commit removed from ticket",
		zap.String("commit_hash", commitHash),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed": true,
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryListCommits lists all commits for a ticket
func (h *Handler) handleRepositoryListCommits(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	ticketID, ok := req.Data["ticket_id"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket_id",
			"",
		))
		return
	}

	query := `
		SELECT id, repository_id, ticket_id, commit_hash, created, modified, deleted
		FROM repository_commit_ticket_mapping
		WHERE ticket_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, ticketID)
	if err != nil {
		logger.Error("Failed to list commits for ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list commits for ticket",
			"",
		))
		return
	}
	defer rows.Close()

	commits := make([]models.RepositoryCommitTicketMapping, 0)
	for rows.Next() {
		var commit models.RepositoryCommitTicketMapping
		err := rows.Scan(
			&commit.ID,
			&commit.RepositoryID,
			&commit.TicketID,
			&commit.CommitHash,
			&commit.Created,
			&commit.Modified,
			&commit.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan commit mapping", zap.Error(err))
			continue
		}
		commits = append(commits, commit)
	}

	logger.Info("Commits listed for ticket",
		zap.String("ticket_id", ticketID),
		zap.Int("count", len(commits)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"commits": commits,
		"count":   len(commits),
	})
	c.JSON(http.StatusOK, response)
}

// handleRepositoryGetCommit gets a specific commit by commit hash
func (h *Handler) handleRepositoryGetCommit(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	commitHash, ok := req.Data["commit_hash"].(string)
	if !ok || commitHash == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing commit_hash",
			"",
		))
		return
	}

	query := `
		SELECT id, repository_id, ticket_id, commit_hash, created, modified, deleted
		FROM repository_commit_ticket_mapping
		WHERE commit_hash = ? AND deleted = 0
	`

	var commit models.RepositoryCommitTicketMapping
	err := h.db.QueryRow(c.Request.Context(), query, commitHash).Scan(
		&commit.ID,
		&commit.RepositoryID,
		&commit.TicketID,
		&commit.CommitHash,
		&commit.Created,
		&commit.Modified,
		&commit.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Commit not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to get commit", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get commit",
			"",
		))
		return
	}

	logger.Info("Commit retrieved",
		zap.String("commit_hash", commitHash),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"commit": commit,
	})
	c.JSON(http.StatusOK, response)
}
