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

// OrganizationCreate handles creating a new organization
func (h *Handler) OrganizationCreate(c *gin.Context, req *models.Request) {
	// Parse the organization data from request
	var organization models.Organization
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal organization data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &organization); err != nil {
		logger.Error("Failed to unmarshal organization data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if organization.Title == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization title is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	organization.ID = uuid.New().String()
	organization.Created = time.Now().Unix()
	organization.Modified = organization.Created
	organization.Deleted = false

	// Store organization in database
	query := `
		INSERT INTO organization (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		organization.ID,
		organization.Title,
		organization.Description,
		organization.Created,
		organization.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to create organization", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to create organization", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Organization created", zap.String("id", organization.ID), zap.String("title", organization.Title))

	response := models.NewSuccessResponse(map[string]interface{}{"organization": organization})
	c.JSON(http.StatusOK, response)
}

// OrganizationRead handles reading a single organization by ID
func (h *Handler) OrganizationRead(c *gin.Context, req *models.Request) {
	// Get organization ID from request data
	organizationID, ok := req.Data["id"].(string)
	if !ok || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve organization from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM organization
		WHERE id = ? AND deleted = 0
	`

	var organization models.Organization
	err := h.db.QueryRow(context.Background(), query, organizationID).Scan(
		&organization.ID,
		&organization.Title,
		&organization.Description,
		&organization.Created,
		&organization.Modified,
		&organization.Deleted,
	)

	if err != nil {
		logger.Error("Failed to read organization", zap.Error(err), zap.String("id", organizationID))
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Organization not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Organization read", zap.String("id", organization.ID), zap.String("title", organization.Title))

	response := models.NewSuccessResponse(map[string]interface{}{
		"organization": organization,
	})
	c.JSON(http.StatusOK, response)
}

// OrganizationList handles listing all organizations
func (h *Handler) OrganizationList(c *gin.Context, req *models.Request) {
	// Retrieve organizations from database with pagination
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM organization
		WHERE deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		logger.Error("Failed to list organizations", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list organizations", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var organizations []models.Organization
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
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list organizations", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Organization list retrieved", zap.Int("count", len(organizations)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"organizations": organizations,
		"count":        len(organizations),
	})
	c.JSON(http.StatusOK, response)
}

// OrganizationModify handles updating an existing organization
func (h *Handler) OrganizationModify(c *gin.Context, req *models.Request) {
	// Parse the organization data from request
	var organization models.Organization
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal organization data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &organization); err != nil {
		logger.Error("Failed to unmarshal organization data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid organization data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if organization.ID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update timestamp
	organization.Modified = time.Now().Unix()

	// Update organization in database
	query := `
		UPDATE organization
		SET title = ?, description = ?, modified = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		organization.Title,
		organization.Description,
		organization.Modified,
		organization.ID,
	)

	if err != nil {
		logger.Error("Failed to modify organization", zap.Error(err), zap.String("id", organization.ID))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to modify organization", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Organization not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Organization modified", zap.String("id", organization.ID))

	response := models.NewSuccessResponse(map[string]interface{}{"organization": organization})
	c.JSON(http.StatusOK, response)
}

// OrganizationRemove handles soft-deleting an organization
func (h *Handler) OrganizationRemove(c *gin.Context, req *models.Request) {
	// Get organization ID from request data
	organizationID, ok := req.Data["id"].(string)
	if !ok || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Soft-delete organization in database (set deleted=true)
	query := `
		UPDATE organization
		SET deleted = 1, modified = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		time.Now().Unix(),
		organizationID,
	)

	if err != nil {
		logger.Error("Failed to remove organization", zap.Error(err), zap.String("id", organizationID))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to remove organization", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Organization not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Organization removed", zap.String("id", organizationID))

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
	if mapping.OrganizationID == "" || mapping.AccountID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID and Account ID are required", "")
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
		INSERT INTO organization_account_mapping (id, organization_id, account_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		mapping.ID,
		mapping.OrganizationID,
		mapping.AccountID,
		mapping.Created,
		mapping.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to assign organization to account", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to assign organization to account", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Organization assigned to account",
		zap.String("organizationId", mapping.OrganizationID),
		zap.String("accountId", mapping.AccountID))

	response := models.NewSuccessResponse(map[string]interface{}{"mapping": mapping})
	c.JSON(http.StatusOK, response)
}

// OrganizationListAccounts handles listing all accounts for an organization
func (h *Handler) OrganizationListAccounts(c *gin.Context, req *models.Request) {
	// Get organization ID from request data
	organizationID, ok := req.Data["organizationId"].(string)
	if !ok || organizationID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Organization ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve accounts from database for this organization
	query := `
		SELECT a.id, a.title, a.description, a.created, a.modified, a.deleted
		FROM account a
		INNER JOIN organization_account_mapping oam ON a.id = oam.account_id
		WHERE oam.organization_id = ? AND oam.deleted = 0 AND a.deleted = 0
		ORDER BY a.created DESC
	`

	rows, err := h.db.Query(context.Background(), query, organizationID)
	if err != nil {
		logger.Error("Failed to list organization accounts", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list organization accounts", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID,
			&account.Title,
			&account.Description,
			&account.Created,
			&account.Modified,
			&account.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan account row", zap.Error(err))
			continue
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating account rows", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list organization accounts", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Organization accounts list retrieved", zap.String("organizationId", organizationID), zap.Int("count", len(accounts)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"accounts": accounts,
		"count":    len(accounts),
	})
	c.JSON(http.StatusOK, response)
}
