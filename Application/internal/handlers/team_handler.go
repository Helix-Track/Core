package handlers

import (
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

	// TODO: Store team in database
	// This will be implemented when database layer is updated
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

	// TODO: Retrieve team from database
	// This will be implemented when database layer is updated
	logger.Info("Team read requested", zap.String("id", teamID))

	// For now, return a placeholder response
	response := models.NewErrorResponse(models.ErrorCodeInternalError, "Team read not yet implemented", "")
	c.JSON(http.StatusNotImplemented, response)
}

// TeamList handles listing all teams
func (h *Handler) TeamList(c *gin.Context, req *models.Request) {
	// TODO: Retrieve teams from database with pagination
	// This will be implemented when database layer is updated
	logger.Info("Team list requested")

	// For now, return empty list
	teams := []models.Team{}
	response := models.NewSuccessResponse(map[string]interface{}{"teams": teams, "count": len(teams)})
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

	// TODO: Update team in database
	// This will be implemented when database layer is updated
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

	// TODO: Soft-delete team in database (set deleted=true)
	// This will be implemented when database layer is updated
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

	// TODO: Store mapping in database
	// This will be implemented when database layer is updated
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

	// TODO: Remove mapping from database (soft delete)
	// This will be implemented when database layer is updated
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

	// TODO: Retrieve organizations from database for this team
	// This will be implemented when database layer is updated
	logger.Info("Team organizations list requested", zap.String("teamId", teamID))

	// For now, return empty list
	organizations := []models.Organization{}
	response := models.NewSuccessResponse(map[string]interface{}{"organizations": organizations, "count": len(organizations)})
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

	// TODO: Store mapping in database
	// This will be implemented when database layer is updated
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

	// TODO: Remove mapping from database (soft delete)
	// This will be implemented when database layer is updated
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

	// TODO: Retrieve projects from database for this team
	// This will be implemented when database layer is updated
	logger.Info("Team projects list requested", zap.String("teamId", teamID))

	// For now, return empty list
	projects := []interface{}{} // Will be replaced with proper Project model
	response := models.NewSuccessResponse(map[string]interface{}{"projects": projects, "count": len(projects)})
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

	// TODO: Store mapping in database
	// This will be implemented when database layer is updated
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

	// TODO: Retrieve organizations from database for this user
	// This will be implemented when database layer is updated
	logger.Info("User organizations list requested", zap.String("userId", userID))

	// For now, return empty list
	organizations := []models.Organization{}
	response := models.NewSuccessResponse(map[string]interface{}{"organizations": organizations, "count": len(organizations)})
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

	// TODO: Retrieve users from database for this organization
	// This will be implemented when database layer is updated
	logger.Info("Organization users list requested", zap.String("organizationId", organizationID))

	// For now, return empty list
	users := []interface{}{} // Will be replaced with proper User model
	response := models.NewSuccessResponse(map[string]interface{}{"users": users, "count": len(users)})
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

	// TODO: Store mapping in database
	// This will be implemented when database layer is updated
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

	// TODO: Retrieve teams from database for this user
	// This will be implemented when database layer is updated
	logger.Info("User teams list requested", zap.String("userId", userID))

	// For now, return empty list
	teams := []models.Team{}
	response := models.NewSuccessResponse(map[string]interface{}{"teams": teams, "count": len(teams)})
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

	// TODO: Retrieve users from database for this team
	// This will be implemented when database layer is updated
	logger.Info("Team users list requested", zap.String("teamId", teamID))

	// For now, return empty list
	users := []interface{}{} // Will be replaced with proper User model
	response := models.NewSuccessResponse(map[string]interface{}{"users": users, "count": len(users)})
	c.JSON(http.StatusOK, response)
}
