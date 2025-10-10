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

// AccountCreate handles creating a new account
func (h *Handler) AccountCreate(c *gin.Context, req *models.Request) {
	// Parse the account data from request
	var account models.Account
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal account data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &account); err != nil {
		logger.Error("Failed to unmarshal account data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if account.Title == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account title is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	account.ID = uuid.New().String()
	account.Created = time.Now().Unix()
	account.Modified = account.Created
	account.Deleted = false

	// TODO: Store account in database
	// This will be implemented when database layer is updated
	logger.Info("Account created", "id", account.ID, "title", account.Title)

	response := models.NewSuccessResponse(account)
	c.JSON(http.StatusOK, response)
}

// AccountRead handles reading a single account by ID
func (h *Handler) AccountRead(c *gin.Context, req *models.Request) {
	// Get account ID from request data
	accountID, ok := req.Data["id"].(string)
	if !ok || accountID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account ID is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// TODO: Retrieve account from database
	// This will be implemented when database layer is updated
	logger.Info("Account read requested", "id", accountID)

	// For now, return a placeholder response
	response := models.NewErrorResponse(models.ErrorCodeInternalError, "Account read not yet implemented")
	c.JSON(http.StatusNotImplemented, response)
}

// AccountList handles listing all accounts
func (h *Handler) AccountList(c *gin.Context, req *models.Request) {
	// TODO: Retrieve accounts from database with pagination
	// This will be implemented when database layer is updated
	logger.Info("Account list requested")

	// For now, return empty list
	accounts := []models.Account{}
	response := models.NewSuccessResponse(accounts)
	c.JSON(http.StatusOK, response)
}

// AccountModify handles updating an existing account
func (h *Handler) AccountModify(c *gin.Context, req *models.Request) {
	// Parse the account data from request
	var account models.Account
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal account data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &account); err != nil {
		logger.Error("Failed to unmarshal account data", "error", err)
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if account.ID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account ID is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update timestamp
	account.Modified = time.Now().Unix()

	// TODO: Update account in database
	// This will be implemented when database layer is updated
	logger.Info("Account modified", "id", account.ID)

	response := models.NewSuccessResponse(account)
	c.JSON(http.StatusOK, response)
}

// AccountRemove handles soft-deleting an account
func (h *Handler) AccountRemove(c *gin.Context, req *models.Request) {
	// Get account ID from request data
	accountID, ok := req.Data["id"].(string)
	if !ok || accountID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account ID is required")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// TODO: Soft-delete account in database (set deleted=true)
	// This will be implemented when database layer is updated
	logger.Info("Account removed", "id", accountID)

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":      accountID,
		"deleted": true,
	})
	c.JSON(http.StatusOK, response)
}
