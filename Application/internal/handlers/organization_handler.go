package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
)

// OrganizationCreate handles creating a new organization
func (h *Handler) OrganizationCreate(c *gin.Context, req *models.Request) {
	// Parse the organization data from request
	var organization models.Organization
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal organization data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &organization); err != nil {
		logger.Error("Failed to unmarshal organization data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if organization.Title == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization title is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	organization.ID = uuid.New().String()
	organization.Created = time.Now().Unix()
	organization.Modified = organization.Created
	organization.Deleted = false

	// TODO: Store organization in database
	// This will be implemented when database layer is updated
	logger.Info("Organization created", "id", organization.ID, "title", organization.Title)

	response := models.NewSuccessResponse(organization)
	c.JSON(http.StatusOK, response)
}

// OrganizationRead handles reading a single organization by ID
func (h *Handler) OrganizationRead(c *gin.Context, req *models.Request) {
	// Get organization ID from request data
	organizationID, ok := req.Data["id"].(string)
	if !ok || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// TODO: Retrieve organization from database
	// This will be implemented when database layer is updated
	logger.Info("Organization read requested", "id", organizationID)

	// For now, return a placeholder response
	response := models.NewErrorResponse(models.ErrorCodeInternalError, "Organization read not yet implemented")
	c.JSON(http.StatusNotImplemented, response)
}

// OrganizationList handles listing all organizations
func (h *Handler) OrganizationList(c *gin.Context, req *models.Request) {
	// TODO: Retrieve organizations from database with pagination
	// This will be implemented when database layer is updated
	logger.Info("Organization list requested")

	// For now, return empty list
	organizations := []models.Organization{}
	response := models.NewSuccessResponse(organizations)
	c.JSON(http.StatusOK, response)
}

// OrganizationModify handles updating an existing organization
func (h *Handler) OrganizationModify(c *gin.Context, req *models.Request) {
	// Parse the organization data from request
	var organization models.Organization
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal organization data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &organization); err != nil {
		logger.Error("Failed to unmarshal organization data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if organization.ID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update timestamp
	organization.Modified = time.Now().Unix()

	// TODO: Update organization in database
	// This will be implemented when database layer is updated
	logger.Info("Organization modified", "id", organization.ID)

	response := models.NewSuccessResponse(organization)
	c.JSON(http.StatusOK, response)
}

// OrganizationRemove handles soft-deleting an organization
func (h *Handler) OrganizationRemove(c *gin.Context, req *models.Request) {
	// Get organization ID from request data
	organizationID, ok := req.Data["id"].(string)
	if !ok || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// TODO: Soft-delete organization in database (set deleted=true)
	// This will be implemented when database layer is updated
	logger.Info("Organization removed", "id", organizationID)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":      organizationID,
		"deleted": true,
	})
	c.JSON(http.StatusOK, response)
}

// OrganizationAssignAccount handles assigning an organization to an account
func (h *Handler) OrganizationAssignAccount(c *gin.Context, req *models.Request) {
	// Parse the mapping data from request
	var mapping models.OrganizationAccountMapping
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal mapping data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &mapping); err != nil {
		logger.Error("Failed to unmarshal mapping data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid mapping data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if mapping.OrganizationID == "" || mapping.AccountID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID and Account ID are required")
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
	logger.Info("Organization assigned to account",
		"organizationId", mapping.OrganizationID,
		"accountId", mapping.AccountID)

	response := models.NewSuccessResponse(mapping)
	c.JSON(http.StatusOK, response)
}

// OrganizationListAccounts handles listing all accounts for an organization
func (h *Handler) OrganizationListAccounts(c *gin.Context, req *models.Request) {
	// Get organization ID from request data
	organizationID, ok := req.Data["organizationId"].(string)
	if !ok || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// TODO: Retrieve accounts from database for this organization
	// This will be implemented when database layer is updated
	logger.Info("Organization accounts list requested", "organizationId", organizationID)

	// For now, return empty list
	accounts := []models.Account{}
	response := models.NewSuccessResponse(accounts)
	c.JSON(http.StatusOK, response)
}
