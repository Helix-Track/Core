package handlers

import (
	"context"
	"fmt"
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

// handleCreateProject creates a new project
func (h *Handler) handleCreateProject(c *gin.Context, req *models.Request) {
	// Extract project data from request
	projectData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		// Try direct data fields
		projectData = req.Data
	}

	name, _ := projectData["name"].(string)
	if name == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project name",
			"",
		))
		return
	}

	key, _ := projectData["key"].(string)
	if key == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project key",
			"",
		))
		return
	}

	description, _ := projectData["description"].(string)
	projectType, _ := projectData["type"].(string)
	if projectType == "" {
		projectType = "software"
	}

	// Get default workflow ID
	var workflowID string
	err := h.db.QueryRow(context.Background(), "SELECT id FROM workflow LIMIT 1").Scan(&workflowID)
	if err != nil {
		logger.Error("Failed to get default workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create project",
			"",
		))
		return
	}

	// Create project
	projectID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO project (id, identifier, title, description, workflow_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		projectID,
		key,
		name,
		description,
		workflowID,
		now,
		now,
		0,
	)

	if err != nil {
		logger.Error("Failed to create project", zap.Error(err))
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeEntityAlreadyExists,
			"Project with this identifier already exists",
			"",
		))
		return
	}

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Publish project created event
	h.publisher.PublishEntityEvent(
		models.ActionCreate,
		"project",
		projectID,
		username,
		map[string]interface{}{
			"id":          projectID,
			"identifier":  key,
			"title":       name,
			"description": description,
			"type":        projectType,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"project": map[string]interface{}{
			"id":          projectID,
			"identifier":  key,
			"title":       name,
			"description": description,
			"type":        projectType,
			"created":     now,
			"modified":    now,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleModifyProject updates an existing project
func (h *Handler) handleModifyProject(c *gin.Context, req *models.Request) {
	projectData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		projectData = req.Data
	}

	projectID, _ := projectData["id"].(string)
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project ID",
			"",
		))
		return
	}

	// Check if project exists
	var exists int
	err := h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM project WHERE id = ? AND deleted = 0",
		projectID).Scan(&exists)

	if err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project not found",
			"",
		))
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}

	if title, ok := projectData["title"].(string); ok && title != "" {
		updates = append(updates, "title = ?")
		args = append(args, title)
	}

	if desc, ok := projectData["description"].(string); ok {
		updates = append(updates, "description = ?")
		args = append(args, desc)
	}

	if identifier, ok := projectData["identifier"].(string); ok && identifier != "" {
		updates = append(updates, "identifier = ?")
		args = append(args, identifier)
	}

	// Always update modified timestamp
	updates = append(updates, "modified = ?")
	args = append(args, time.Now().Unix())

	// Add project ID to args
	args = append(args, projectID)

	query := fmt.Sprintf("UPDATE project SET %s WHERE id = ?",
		joinWithComma(updates))

	_, err = h.db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.Error("Failed to update project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update project",
			"",
		))
		return
	}

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Publish project updated event
	h.publisher.PublishEntityEvent(
		models.ActionModify,
		"project",
		projectID,
		username,
		projectData,
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"project": map[string]interface{}{
			"id":      projectID,
			"updated": true,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleRemoveProject soft-deletes a project
func (h *Handler) handleRemoveProject(c *gin.Context, req *models.Request) {
	projectData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		projectData = req.Data
	}

	projectID, _ := projectData["id"].(string)
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project ID",
			"",
		))
		return
	}

	// Check if project exists before deletion
	var exists int
	err := h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM project WHERE id = ? AND deleted = 0",
		projectID).Scan(&exists)

	if err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project not found",
			"",
		))
		return
	}

	query := "UPDATE project SET deleted = 1, modified = ? WHERE id = ?"
	_, err = h.db.Exec(context.Background(), query, time.Now().Unix(), projectID)
	if err != nil {
		logger.Error("Failed to delete project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete project",
			"",
		))
		return
	}

	// Get username from context
	username, _ := middleware.GetUsername(c)

	// Publish project deleted event
	h.publisher.PublishEntityEvent(
		models.ActionRemove,
		"project",
		projectID,
		username,
		map[string]interface{}{
			"id": projectID,
		},
		websocket.NewProjectContext(projectID, []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"project": map[string]interface{}{
			"id":      projectID,
			"deleted": true,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleReadProject retrieves a single project
func (h *Handler) handleReadProject(c *gin.Context, req *models.Request) {
	projectData, ok := req.Data["data"].(map[string]interface{})
	if !ok {
		projectData = req.Data
	}

	projectID, _ := projectData["id"].(string)
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project ID",
			"",
		))
		return
	}

	query := `
		SELECT id, identifier, title, description, workflow_id, created, modified
		FROM project
		WHERE id = ? AND deleted = 0
	`

	var id, identifier, title, description, workflowID string
	var created, modified int64

	err := h.db.QueryRow(context.Background(), query, projectID).Scan(
		&id, &identifier, &title, &description, &workflowID, &created, &modified)

	if err != nil {
		logger.Error("Project not found", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Project not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"project": map[string]interface{}{
			"id":          id,
			"identifier":  identifier,
			"title":       title,
			"description": description,
			"workflowId":  workflowID,
			"created":     created,
			"modified":    modified,
		},
	})

	c.JSON(http.StatusOK, response)
}

// handleListProjects retrieves all projects
func (h *Handler) handleListProjects(c *gin.Context, req *models.Request) {
	query := `
		SELECT id, identifier, title, description, workflow_id, created, modified
		FROM project
		WHERE deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		logger.Error("Failed to list projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list projects",
			"",
		))
		return
	}
	defer rows.Close()

	projects := []map[string]interface{}{}

	for rows.Next() {
		var id, identifier, title, description, workflowID string
		var created, modified int64

		err := rows.Scan(&id, &identifier, &title, &description, &workflowID, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan project", zap.Error(err))
			continue
		}

		projects = append(projects, map[string]interface{}{
			"id":          id,
			"identifier":  identifier,
			"title":       title,
			"description": description,
			"workflowId":  workflowID,
			"created":     created,
			"modified":    modified,
		})
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"items": projects,
		"total": len(projects),
	})

	c.JSON(http.StatusOK, response)
}

// Helper function to join strings with comma
func joinWithComma(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
