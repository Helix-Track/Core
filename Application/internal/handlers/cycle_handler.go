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

// handleCycleCreate creates a new cycle (Sprint/Milestone/Release)
func (h *Handler) handleCycleCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "cycle", models.PermissionCreate)
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

	// Parse cycle data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	cycleType, ok := req.Data["type"].(float64) // JSON numbers are float64
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing type",
			"",
		))
		return
	}

	cycle := &models.Cycle{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		CycleID:     getStringFromData(req.Data, "cycleId"), // Parent cycle ID
		Type:        int(cycleType),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Validate cycle type
	if !cycle.IsValidType() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid cycle type (must be 10, 100, or 1000)",
			"",
		))
		return
	}

	// If parent cycle is specified, validate it exists and type hierarchy
	if cycle.CycleID != "" {
		var parentType int
		checkQuery := `SELECT type FROM cycle WHERE id = ? AND deleted = 0`
		err := h.db.QueryRow(c.Request.Context(), checkQuery, cycle.CycleID).Scan(&parentType)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeEntityNotFound,
				"Parent cycle not found",
				"",
			))
			return
		}
		if err != nil {
			logger.Error("Failed to check parent cycle", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to validate parent cycle",
				"",
			))
			return
		}

		// Validate parent type hierarchy
		if !cycle.IsValidParent(parentType) {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid parent cycle type (must be greater than current type)",
				"",
			))
			return
		}
	}

	// Insert into database
	query := `
		INSERT INTO cycle (id, title, description, cycle_id, type, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		cycle.ID,
		cycle.Title,
		cycle.Description,
		cycle.CycleID,
		cycle.Type,
		cycle.Created,
		cycle.Modified,
		cycle.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create cycle", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create cycle",
			"",
		))
		return
	}

	logger.Info("Cycle created",
		zap.String("cycle_id", cycle.ID),
		zap.String("title", cycle.Title),
		zap.String("type", cycle.GetTypeName()),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"cycle": cycle,
	})
	c.JSON(http.StatusCreated, response)
}

// handleCycleRead reads a single cycle by ID
func (h *Handler) handleCycleRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get cycle ID from request
	cycleID, ok := req.Data["id"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	// Query cycle from database
	query := `
		SELECT id, title, description, cycle_id, type, created, modified, deleted
		FROM cycle
		WHERE id = ? AND deleted = 0
	`

	var cycle models.Cycle
	err := h.db.QueryRow(c.Request.Context(), query, cycleID).Scan(
		&cycle.ID,
		&cycle.Title,
		&cycle.Description,
		&cycle.CycleID,
		&cycle.Type,
		&cycle.Created,
		&cycle.Modified,
		&cycle.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Cycle not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read cycle", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read cycle",
			"",
		))
		return
	}

	logger.Info("Cycle read",
		zap.String("cycle_id", cycle.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"cycle": cycle,
	})
	c.JSON(http.StatusOK, response)
}

// handleCycleList lists all cycles with optional filtering
func (h *Handler) handleCycleList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Build query with optional filters
	query := `
		SELECT id, title, description, cycle_id, type, created, modified, deleted
		FROM cycle
		WHERE deleted = 0
	`
	args := make([]interface{}, 0)

	// Filter by type if provided
	if cycleType, ok := req.Data["type"].(float64); ok {
		query += " AND type = ?"
		args = append(args, int(cycleType))
	}

	// Filter by parent cycle if provided
	if parentID, ok := req.Data["cycleId"].(string); ok && parentID != "" {
		query += " AND cycle_id = ?"
		args = append(args, parentID)
	}

	query += " ORDER BY type DESC, created DESC"

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to list cycles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list cycles",
			"",
		))
		return
	}
	defer rows.Close()

	cycles := make([]models.Cycle, 0)
	for rows.Next() {
		var cycle models.Cycle
		err := rows.Scan(
			&cycle.ID,
			&cycle.Title,
			&cycle.Description,
			&cycle.CycleID,
			&cycle.Type,
			&cycle.Created,
			&cycle.Modified,
			&cycle.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan cycle", zap.Error(err))
			continue
		}
		cycles = append(cycles, cycle)
	}

	logger.Info("Cycles listed",
		zap.Int("count", len(cycles)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"cycles": cycles,
		"count":  len(cycles),
	})
	c.JSON(http.StatusOK, response)
}

// handleCycleModify updates an existing cycle
func (h *Handler) handleCycleModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "cycle", models.PermissionUpdate)
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

	// Get cycle ID
	cycleID, ok := req.Data["id"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	// Check if cycle exists and get current type
	var currentType int
	checkQuery := `SELECT type FROM cycle WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, cycleID).Scan(&currentType)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Cycle not found",
			"",
		))
		return
	}
	if err != nil {
		logger.Error("Failed to check cycle", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to validate cycle",
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
	if cycleType, ok := req.Data["type"].(float64); ok {
		typeInt := int(cycleType)
		if typeInt != models.CycleTypeSprint && typeInt != models.CycleTypeMilestone && typeInt != models.CycleTypeRelease {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid cycle type (must be 10, 100, or 1000)",
				"",
			))
			return
		}
		updates["type"] = typeInt
	}
	if cycleID, ok := req.Data["cycleId"].(string); ok {
		updates["cycle_id"] = cycleID
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
	query := "UPDATE cycle SET "
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
	args = append(args, cycleID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update cycle", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update cycle",
			"",
		))
		return
	}

	logger.Info("Cycle updated",
		zap.String("cycle_id", cycleID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      cycleID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCycleRemove soft-deletes a cycle
func (h *Handler) handleCycleRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "cycle", models.PermissionDelete)
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

	// Get cycle ID
	cycleID, ok := req.Data["id"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	// Soft delete the cycle
	query := `UPDATE cycle SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), cycleID)
	if err != nil {
		logger.Error("Failed to delete cycle", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete cycle",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Cycle not found",
			"",
		))
		return
	}

	logger.Info("Cycle deleted",
		zap.String("cycle_id", cycleID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      cycleID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCycleAssignProject assigns a cycle to a project
func (h *Handler) handleCycleAssignProject(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "cycle", models.PermissionUpdate)
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

	// Get cycle ID and project ID
	cycleID, ok := req.Data["cycleId"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	projectID, ok := req.Data["projectId"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project ID",
			"",
		))
		return
	}

	// Check if cycle exists
	var count int
	checkQuery := `SELECT COUNT(*) FROM cycle WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, cycleID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Cycle not found",
			"",
		))
		return
	}

	// Check if project exists
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
	checkQuery = `SELECT COUNT(*) FROM cycle_project_mapping WHERE cycle_id = ? AND project_id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, cycleID, projectID).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeEntityAlreadyExists,
			"Cycle already assigned to project",
			"",
		))
		return
	}

	// Create mapping
	mapping := &models.CycleProjectMapping{
		ID:        uuid.New().String(),
		CycleID:   cycleID,
		ProjectID: projectID,
		Created:   time.Now().Unix(),
		Modified:  time.Now().Unix(),
		Deleted:   false,
	}

	query := `
		INSERT INTO cycle_project_mapping (id, cycle_id, project_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.CycleID,
		mapping.ProjectID,
		mapping.Created,
		mapping.Modified,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to assign cycle to project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to assign cycle to project",
			"",
		))
		return
	}

	logger.Info("Cycle assigned to project",
		zap.String("cycle_id", cycleID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleCycleUnassignProject unassigns a cycle from a project
func (h *Handler) handleCycleUnassignProject(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "cycle", models.PermissionUpdate)
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

	// Get cycle ID and project ID
	cycleID, ok := req.Data["cycleId"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	projectID, ok := req.Data["projectId"].(string)
	if !ok || projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing project ID",
			"",
		))
		return
	}

	// Soft delete the mapping
	query := `
		UPDATE cycle_project_mapping
		SET deleted = 1, modified = ?
		WHERE cycle_id = ? AND project_id = ? AND deleted = 0
	`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), cycleID, projectID)
	if err != nil {
		logger.Error("Failed to unassign cycle from project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unassign cycle from project",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Cycle-project mapping not found",
			"",
		))
		return
	}

	logger.Info("Cycle unassigned from project",
		zap.String("cycle_id", cycleID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"unassigned": true,
		"cycleId":    cycleID,
		"projectId":  projectID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCycleListProjects lists all projects assigned to a cycle
func (h *Handler) handleCycleListProjects(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get cycle ID
	cycleID, ok := req.Data["cycleId"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	// Query projects assigned to cycle
	query := `
		SELECT p.id, p.title, p.description, p.created, p.modified
		FROM project p
		INNER JOIN cycle_project_mapping cpm ON p.id = cpm.project_id
		WHERE cpm.cycle_id = ? AND cpm.deleted = 0 AND p.deleted = 0
		ORDER BY p.title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, cycleID)
	if err != nil {
		logger.Error("Failed to list cycle projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list cycle projects",
			"",
		))
		return
	}
	defer rows.Close()

	projects := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title, description string
		var created, modified int64
		err := rows.Scan(&id, &title, &description, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan project", zap.Error(err))
			continue
		}
		projects = append(projects, map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": description,
			"created":     created,
			"modified":    modified,
		})
	}

	logger.Info("Cycle projects listed",
		zap.String("cycle_id", cycleID),
		zap.Int("count", len(projects)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	})
	c.JSON(http.StatusOK, response)
}

// handleCycleAddTicket adds a ticket to a cycle
func (h *Handler) handleCycleAddTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "cycle", models.PermissionUpdate)
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

	// Get cycle ID and ticket ID
	cycleID, ok := req.Data["cycleId"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket ID",
			"",
		))
		return
	}

	// Check if cycle exists
	var count int
	checkQuery := `SELECT COUNT(*) FROM cycle WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, cycleID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Cycle not found",
			"",
		))
		return
	}

	// Check if ticket exists
	checkQuery = `SELECT COUNT(*) FROM ticket WHERE id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket not found",
			"",
		))
		return
	}

	// Check if mapping already exists
	checkQuery = `SELECT COUNT(*) FROM ticket_cycle_mapping WHERE ticket_id = ? AND cycle_id = ? AND deleted = 0`
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketID, cycleID).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrorCodeEntityAlreadyExists,
			"Ticket already in cycle",
			"",
		))
		return
	}

	// Create mapping
	mapping := &models.TicketCycleMapping{
		ID:       uuid.New().String(),
		TicketID: ticketID,
		CycleID:  cycleID,
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
		Deleted:  false,
	}

	query := `
		INSERT INTO ticket_cycle_mapping (id, ticket_id, cycle_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.TicketID,
		mapping.CycleID,
		mapping.Created,
		mapping.Modified,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add ticket to cycle", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add ticket to cycle",
			"",
		))
		return
	}

	logger.Info("Ticket added to cycle",
		zap.String("ticket_id", ticketID),
		zap.String("cycle_id", cycleID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleCycleRemoveTicket removes a ticket from a cycle
func (h *Handler) handleCycleRemoveTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "cycle", models.PermissionUpdate)
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

	// Get cycle ID and ticket ID
	cycleID, ok := req.Data["cycleId"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	ticketID, ok := req.Data["ticketId"].(string)
	if !ok || ticketID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing ticket ID",
			"",
		))
		return
	}

	// Soft delete the mapping
	query := `
		UPDATE ticket_cycle_mapping
		SET deleted = 1, modified = ?
		WHERE ticket_id = ? AND cycle_id = ? AND deleted = 0
	`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), ticketID, cycleID)
	if err != nil {
		logger.Error("Failed to remove ticket from cycle", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove ticket from cycle",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Ticket-cycle mapping not found",
			"",
		))
		return
	}

	logger.Info("Ticket removed from cycle",
		zap.String("ticket_id", ticketID),
		zap.String("cycle_id", cycleID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":  true,
		"ticketId": ticketID,
		"cycleId":  cycleID,
	})
	c.JSON(http.StatusOK, response)
}

// handleCycleListTickets lists all tickets in a cycle
func (h *Handler) handleCycleListTickets(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get cycle ID
	cycleID, ok := req.Data["cycleId"].(string)
	if !ok || cycleID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing cycle ID",
			"",
		))
		return
	}

	// Query tickets in cycle
	query := `
		SELECT t.id, t.title, t.description, t.status, t.created, t.modified
		FROM ticket t
		INNER JOIN ticket_cycle_mapping tcm ON t.id = tcm.ticket_id
		WHERE tcm.cycle_id = ? AND tcm.deleted = 0 AND t.deleted = 0
		ORDER BY t.created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, cycleID)
	if err != nil {
		logger.Error("Failed to list cycle tickets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list cycle tickets",
			"",
		))
		return
	}
	defer rows.Close()

	tickets := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title, description, status string
		var created, modified int64
		err := rows.Scan(&id, &title, &description, &status, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan ticket", zap.Error(err))
			continue
		}
		tickets = append(tickets, map[string]interface{}{
			"id":          id,
			"title":       title,
			"description": description,
			"status":      status,
			"created":     created,
			"modified":    modified,
		})
	}

	logger.Info("Cycle tickets listed",
		zap.String("cycle_id", cycleID),
		zap.Int("count", len(tickets)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"tickets": tickets,
		"count":   len(tickets),
	})
	c.JSON(http.StatusOK, response)
}
