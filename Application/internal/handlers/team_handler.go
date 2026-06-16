package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
	"go.uber.org/zap"
)

// TeamCreate handles creating a new team
func (h *Handler) TeamCreate(c *gin.Context, req *models.Request) {
	// Parse the team data from request
	var team models.Team
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal team data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid team data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &team); err != nil {
		logger.Error("Failed to unmarshal team data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid team data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if team.Title == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team title is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	team.ID = uuid.New().String()
	team.Created = time.Now().Unix()
	team.Modified = team.Created
	team.Deleted = false

	// Store team in database
	query := `
		INSERT INTO team (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		team.ID,
		team.Title,
		team.Description,
		team.Created,
		team.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to create team", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to create team", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Team created", zap.String("id", team.ID), zap.String("title", team.Title))

	response := models.NewSuccessResponse(map[string]interface{}{"team": team})
	c.JSON(http.StatusOK, response)
}

// TeamRead handles reading a single team by ID
func (h *Handler) TeamRead(c *gin.Context, req *models.Request) {
	// Get team ID from request data
	teamID, ok := req.Data["id"].(string)
	if !ok || teamID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve team from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM team
		WHERE id = ? AND deleted = 0
	`

	var team models.Team
	err := h.db.QueryRow(context.Background(), query, teamID).Scan(
		&team.ID,
		&team.Title,
		&team.Description,
		&team.Created,
		&team.Modified,
		&team.Deleted,
	)

	if err != nil {
		logger.Error("Failed to read team", zap.Error(err), zap.String("id", teamID))
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Team not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Team read", zap.String("id", team.ID), zap.String("title", team.Title))

	response := models.NewSuccessResponse(map[string]interface{}{
		"team": team,
	})
	c.JSON(http.StatusOK, response)
}

// TeamList handles listing all teams
func (h *Handler) TeamList(c *gin.Context, req *models.Request) {
	// Retrieve teams from database with pagination
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM team
		WHERE deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		logger.Error("Failed to list teams", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list teams", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		err := rows.Scan(
			&team.ID,
			&team.Title,
			&team.Description,
			&team.Created,
			&team.Modified,
			&team.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan team row", zap.Error(err))
			continue
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating team rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list teams", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Team list retrieved", zap.Int("count", len(teams)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"teams": teams,
		"count": len(teams),
	})
	c.JSON(http.StatusOK, response)
}

// TeamModify handles updating an existing team
func (h *Handler) TeamModify(c *gin.Context, req *models.Request) {
	// Parse the team data from request
	var team models.Team
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal team data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid team data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &team); err != nil {
		logger.Error("Failed to unmarshal team data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid team data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if team.ID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update timestamp
	team.Modified = time.Now().Unix()

	// Update team in database
	query := `
		UPDATE team
		SET title = ?, description = ?, modified = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		team.Title,
		team.Description,
		team.Modified,
		team.ID,
	)

	if err != nil {
		logger.Error("Failed to modify team", zap.Error(err), zap.String("id", team.ID))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to modify team", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Team not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Team modified", zap.String("id", team.ID))

	response := models.NewSuccessResponse(map[string]interface{}{"team": team})
	c.JSON(http.StatusOK, response)
}

// TeamRemove handles soft-deleting a team
func (h *Handler) TeamRemove(c *gin.Context, req *models.Request) {
	// Get team ID from request data
	teamID, ok := req.Data["id"].(string)
	if !ok || teamID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Soft-delete team in database (set deleted=true)
	query := `
		UPDATE team
		SET deleted = 1, modified = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		time.Now().Unix(),
		teamID,
	)

	if err != nil {
		logger.Error("Failed to remove team", zap.Error(err), zap.String("id", teamID))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to remove team", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Team not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Team removed", zap.String("id", teamID))

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":      teamID,
		"deleted": true,
	})
	c.JSON(http.StatusOK, response)
}

// TeamAssignOrganization handles assigning a team to an organization
func (h *Handler) TeamAssignOrganization(c *gin.Context, req *models.Request) {
	// Parse the mapping data from request
	var mapping models.TeamOrganizationMapping
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &mapping); err != nil {
		logger.Error("Failed to unmarshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if mapping.TeamID == "" || mapping.OrganizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID and Organization ID are required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	mapping.ID = uuid.New().String()
	mapping.Created = time.Now().Unix()
	mapping.Modified = mapping.Created
	mapping.Deleted = false

	// Store mapping in database
	query := `
		INSERT INTO team_organization_mapping (id, team_id, organization_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		mapping.ID,
		mapping.TeamID,
		mapping.OrganizationID,
		mapping.Created,
		mapping.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to assign team to organization", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to assign team to organization", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Team assigned to organization",
		zap.String("teamId", mapping.TeamID),
		zap.String("organizationId", mapping.OrganizationID))

	response := models.NewSuccessResponse(map[string]interface{}{"mapping": mapping})
	c.JSON(http.StatusOK, response)
}

// TeamUnassignOrganization handles unassigning a team from an organization
func (h *Handler) TeamUnassignOrganization(c *gin.Context, req *models.Request) {
	// Get team ID and organization ID from request data
	teamID, ok1 := req.Data["teamId"].(string)
	organizationID, ok2 := req.Data["organizationId"].(string)

	if !ok1 || !ok2 || teamID == "" || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID and Organization ID are required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Remove mapping from database (soft delete)
	query := `
		UPDATE team_organization_mapping
		SET deleted = 1, modified = ?
		WHERE team_id = ? AND organization_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		time.Now().Unix(),
		teamID,
		organizationID,
	)

	if err != nil {
		logger.Error("Failed to unassign team from organization", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to unassign team from organization", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Team-organization mapping not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Team unassigned from organization",
		zap.String("teamId", teamID),
		zap.String("organizationId", organizationID))

	response := models.NewSuccessResponse(map[string]interface{}{
		"teamId":         teamID,
		"organizationId": organizationID,
		"unassigned":     true,
	})
	c.JSON(http.StatusOK, response)
}

// TeamListOrganizations handles listing all organizations for a team
func (h *Handler) TeamListOrganizations(c *gin.Context, req *models.Request) {
	// Get team ID from request data
	teamID, ok := req.Data["teamId"].(string)
	if !ok || teamID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve organizations from database for this team
	query := `
		SELECT o.id, o.title, o.description, o.created, o.modified, o.deleted
		FROM organization o
		INNER JOIN team_organization_mapping tom ON o.id = tom.organization_id
		WHERE tom.team_id = ? AND tom.deleted = 0 AND o.deleted = 0
		ORDER BY o.created DESC
	`

	rows, err := h.db.Query(context.Background(), query, teamID)
	if err != nil {
		logger.Error("Failed to list team organizations", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list team organizations", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	// Initialise to a non-nil slice so an empty result serialises as a JSON
	// array ([]) rather than null (consistent with other list endpoints).
	organizations := make([]models.Organization, 0)
	for rows.Next() {
		var organization models.Organization
		err := rows.Scan(
			&organization.ID,
			&organization.Title,
			&organization.Description,
			&organization.Created,
			&organization.Modified,
			&organization.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan organization row", zap.Error(err))
			continue
		}
		organizations = append(organizations, organization)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating organization rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list team organizations", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Team organizations list retrieved", zap.String("teamId", teamID), zap.Int("count", len(organizations)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"organizations": organizations,
		"count":        len(organizations),
	})
	c.JSON(http.StatusOK, response)
}

// TeamAssignProject handles assigning a team to a project
func (h *Handler) TeamAssignProject(c *gin.Context, req *models.Request) {
	// Parse the mapping data from request
	var mapping models.TeamProjectMapping
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &mapping); err != nil {
		logger.Error("Failed to unmarshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if mapping.TeamID == "" || mapping.ProjectID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID and Project ID are required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	mapping.ID = uuid.New().String()
	mapping.Created = time.Now().Unix()
	mapping.Modified = mapping.Created
	mapping.Deleted = false

	// Store mapping in database
	query := `
		INSERT INTO team_project_mapping (id, team_id, project_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		mapping.ID,
		mapping.TeamID,
		mapping.ProjectID,
		mapping.Created,
		mapping.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to assign team to project", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to assign team to project", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Team assigned to project",
		zap.String("teamId", mapping.TeamID),
		zap.String("projectId", mapping.ProjectID))

	response := models.NewSuccessResponse(map[string]interface{}{"mapping": mapping})
	c.JSON(http.StatusOK, response)
}

// TeamUnassignProject handles unassigning a team from a project
func (h *Handler) TeamUnassignProject(c *gin.Context, req *models.Request) {
	// Get team ID and project ID from request data
	teamID, ok1 := req.Data["teamId"].(string)
	projectID, ok2 := req.Data["projectId"].(string)

	if !ok1 || !ok2 || teamID == "" || projectID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID and Project ID are required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Remove mapping from database (soft delete)
	query := `
		UPDATE team_project_mapping
		SET deleted = 1, modified = ?
		WHERE team_id = ? AND project_id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		time.Now().Unix(),
		teamID,
		projectID,
	)

	if err != nil {
		logger.Error("Failed to unassign team from project", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to unassign team from project", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Team-project mapping not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Team unassigned from project",
		zap.String("teamId", teamID),
		zap.String("projectId", projectID))

	response := models.NewSuccessResponse(map[string]interface{}{
		"teamId":     teamID,
		"projectId":  projectID,
		"unassigned": true,
	})
	c.JSON(http.StatusOK, response)
}

// TeamListProjects handles listing all projects for a team
func (h *Handler) TeamListProjects(c *gin.Context, req *models.Request) {
	// Get team ID from request data
	teamID, ok := req.Data["teamId"].(string)
	if !ok || teamID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve projects from database for this team
	query := `
		SELECT p.id, p.identifier, p.title, p.description, p.workflow_id, p.created, p.modified, p.deleted, p.version
		FROM project p
		INNER JOIN team_project_mapping tpm ON p.id = tpm.project_id
		WHERE tpm.team_id = ? AND tpm.deleted = 0 AND p.deleted = 0
		ORDER BY p.created DESC
	`

	rows, err := h.db.Query(context.Background(), query, teamID)
	if err != nil {
		logger.Error("Failed to list team projects", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list team projects", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var projects []interface{}
	for rows.Next() {
		var project map[string]interface{} = make(map[string]interface{})
		var id, identifier, title, description, workflowID string
		var created, modified int64
		var deleted bool
		var version int
		
		err := rows.Scan(
			&id,
			&identifier,
			&title,
			&description,
			&workflowID,
			&created,
			&modified,
			&deleted,
			&version,
		)
		if err != nil {
			logger.Error("Failed to scan project row", zap.Error(err))
			continue
		}
		
		project["id"] = id
		project["identifier"] = identifier
		project["title"] = title
		project["description"] = description
		project["workflow_id"] = workflowID
		project["created"] = created
		project["modified"] = modified
		project["deleted"] = deleted
		project["version"] = version
		
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating project rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list team projects", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Team projects list retrieved", zap.String("teamId", teamID), zap.Int("count", len(projects)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	})
	c.JSON(http.StatusOK, response)
}

// UserAssignOrganization handles assigning a user to an organization
func (h *Handler) UserAssignOrganization(c *gin.Context, req *models.Request) {
	// Parse the mapping data from request
	var mapping models.UserOrganizationMapping
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &mapping); err != nil {
		logger.Error("Failed to unmarshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if mapping.UserID == "" || mapping.OrganizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "User ID and Organization ID are required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	mapping.ID = uuid.New().String()
	mapping.Created = time.Now().Unix()
	mapping.Modified = mapping.Created
	mapping.Deleted = false

	// Store mapping in database
	query := `
		INSERT INTO user_organization_mapping (id, user_id, organization_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		mapping.ID,
		mapping.UserID,
		mapping.OrganizationID,
		mapping.Created,
		mapping.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to assign user to organization", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to assign user to organization", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("User assigned to organization",
		zap.String("userId", mapping.UserID),
		zap.String("organizationId", mapping.OrganizationID))

	response := models.NewSuccessResponse(map[string]interface{}{"mapping": mapping})
	c.JSON(http.StatusOK, response)
}

// UserListOrganizations handles listing all organizations for a user
func (h *Handler) UserListOrganizations(c *gin.Context, req *models.Request) {
	// Get user ID from request data
	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "User ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve organizations from database for this user
	query := `
		SELECT o.id, o.title, o.description, o.created, o.modified, o.deleted
		FROM organization o
		INNER JOIN user_organization_mapping uom ON o.id = uom.organization_id
		WHERE uom.user_id = ? AND uom.deleted = 0 AND o.deleted = 0
		ORDER BY o.created DESC
	`

	rows, err := h.db.Query(context.Background(), query, userID)
	if err != nil {
		logger.Error("Failed to list user organizations", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list user organizations", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	// Initialise to a non-nil slice so an empty result serialises as a JSON
	// array ([]) rather than null (consistent with other list endpoints).
	organizations := make([]models.Organization, 0)
	for rows.Next() {
		var organization models.Organization
		err := rows.Scan(
			&organization.ID,
			&organization.Title,
			&organization.Description,
			&organization.Created,
			&organization.Modified,
			&organization.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan organization row", zap.Error(err))
			continue
		}
		organizations = append(organizations, organization)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating organization rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list user organizations", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("User organizations list retrieved", zap.String("userId", userID), zap.Int("count", len(organizations)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"organizations": organizations,
		"count":        len(organizations),
	})
	c.JSON(http.StatusOK, response)
}

// OrganizationListUsers handles listing all users in an organization
func (h *Handler) OrganizationListUsers(c *gin.Context, req *models.Request) {
	// Get organization ID from request data
	organizationID, ok := req.Data["organizationId"].(string)
	if !ok || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve users from database for this organization
	query := `
		SELECT u.id, u.username, u.email, u.name, u.role, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_organization_mapping uom ON u.id = uom.user_id
		WHERE uom.organization_id = ? AND uom.deleted = 0 AND u.deleted = 0
		ORDER BY u.created_at DESC
	`

	rows, err := h.db.Query(context.Background(), query, organizationID)
	if err != nil {
		logger.Error("Failed to list organization users", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list organization users", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var users []interface{}
	for rows.Next() {
		var user map[string]interface{} = make(map[string]interface{})
		var id, username, email, name, role string
		var createdAt, updatedAt int64
		
		err := rows.Scan(
			&id,
			&username,
			&email,
			&name,
			&role,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan user row", zap.Error(err))
			continue
		}
		
		user["id"] = id
		user["username"] = username
		user["email"] = email
		user["name"] = name
		user["role"] = role
		user["created_at"] = createdAt
		user["updated_at"] = updatedAt
		
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating user rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list organization users", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Organization users list retrieved", zap.String("organizationId", organizationID), zap.Int("count", len(users)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
	c.JSON(http.StatusOK, response)
}

// UserAssignTeam handles assigning a user to a team
func (h *Handler) UserAssignTeam(c *gin.Context, req *models.Request) {
	// Parse the mapping data from request
	var mapping models.UserTeamMapping
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &mapping); err != nil {
		logger.Error("Failed to unmarshal mapping data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if mapping.UserID == "" || mapping.TeamID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "User ID and Team ID are required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	mapping.ID = uuid.New().String()
	mapping.Created = time.Now().Unix()
	mapping.Modified = mapping.Created
	mapping.Deleted = false

	// Store mapping in database
	query := `
		INSERT INTO user_team_mapping (id, user_id, team_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		mapping.ID,
		mapping.UserID,
		mapping.TeamID,
		mapping.Created,
		mapping.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to assign user to team", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to assign user to team", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("User assigned to team",
		zap.String("userId", mapping.UserID),
		zap.String("teamId", mapping.TeamID))

	response := models.NewSuccessResponse(map[string]interface{}{"mapping": mapping})
	c.JSON(http.StatusOK, response)
}

// UserListTeams handles listing all teams for a user
func (h *Handler) UserListTeams(c *gin.Context, req *models.Request) {
	// Get user ID from request data
	userID, ok := req.Data["userId"].(string)
	if !ok || userID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "User ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve teams from database for this user
	query := `
		SELECT t.id, t.title, t.description, t.created, t.modified, t.deleted
		FROM team t
		INNER JOIN user_team_mapping utm ON t.id = utm.team_id
		WHERE utm.user_id = ? AND utm.deleted = 0 AND t.deleted = 0
		ORDER BY t.created DESC
	`

	rows, err := h.db.Query(context.Background(), query, userID)
	if err != nil {
		logger.Error("Failed to list user teams", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list user teams", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		err := rows.Scan(
			&team.ID,
			&team.Title,
			&team.Description,
			&team.Created,
			&team.Modified,
			&team.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan team row", zap.Error(err))
			continue
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating team rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list user teams", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("User teams list retrieved", zap.String("userId", userID), zap.Int("count", len(teams)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"teams": teams,
		"count": len(teams),
	})
	c.JSON(http.StatusOK, response)
}

// TeamListUsers handles listing all users in a team
func (h *Handler) TeamListUsers(c *gin.Context, req *models.Request) {
	// Get team ID from request data
	teamID, ok := req.Data["teamId"].(string)
	if !ok || teamID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Team ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve users from database for this team
	query := `
		SELECT u.id, u.username, u.email, u.name, u.role, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_team_mapping utm ON u.id = utm.user_id
		WHERE utm.team_id = ? AND utm.deleted = 0 AND u.deleted = 0
		ORDER BY u.created_at DESC
	`

	rows, err := h.db.Query(context.Background(), query, teamID)
	if err != nil {
		logger.Error("Failed to list team users", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list team users", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var users []interface{}
	for rows.Next() {
		var user map[string]interface{} = make(map[string]interface{})
		var id, username, email, name, role string
		var createdAt, updatedAt int64
		
		err := rows.Scan(
			&id,
			&username,
			&email,
			&name,
			&role,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan user row", zap.Error(err))
			continue
		}
		
		user["id"] = id
		user["username"] = username
		user["email"] = email
		user["name"] = name
		user["role"] = role
		user["created_at"] = createdAt
		user["updated_at"] = updatedAt
		
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating user rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list team users", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Team users list retrieved", zap.String("teamId", teamID), zap.Int("count", len(users)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
	c.JSON(http.StatusOK, response)
}
