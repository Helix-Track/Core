package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/models"
)

// AccountCreate handles creating a new account
func (h *Handler) AccountCreate(c *gin.Context, req *models.Request) {
	// Parse the account data from request
	var account models.Account
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal account data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &account); err != nil {
		logger.Error("Failed to unmarshal account data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if account.Title == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account title is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Generate ID and timestamps
	account.ID = uuid.New().String()
	account.Created = time.Now().Unix()
	account.Modified = account.Created
	account.Deleted = false

	// Store account in database
	query := `
		INSERT INTO account (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(
		context.Background(),
		query,
		account.ID,
		account.Title,
		account.Description,
		account.Created,
		account.Modified,
		0, // deleted as integer (false)
	)

	if err != nil {
		logger.Error("Failed to create account", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to create account", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Account created", zap.String("id", account.ID), zap.String("title", account.Title))

	response := models.NewSuccessResponse(map[string]interface{}{
		"account": account,
	})
	c.JSON(http.StatusOK, response)
}

// AccountRead handles reading a single account by ID
func (h *Handler) AccountRead(c *gin.Context, req *models.Request) {
	// Get account ID from request data
	accountID, ok := req.Data["id"].(string)
	if !ok || accountID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Retrieve account from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM account
		WHERE id = ? AND deleted = 0
	`

	var account models.Account
	err := h.db.QueryRow(context.Background(), query, accountID).Scan(
		&account.ID,
		&account.Title,
		&account.Description,
		&account.Created,
		&account.Modified,
		&account.Deleted,
	)

	if err != nil {
		logger.Error("Failed to read account", zap.Error(err), zap.String("id", accountID))
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Account not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Account read", zap.String("id", account.ID), zap.String("title", account.Title))

	response := models.NewSuccessResponse(map[string]interface{}{
		"account": account,
	})
	c.JSON(http.StatusOK, response)
}

// AccountList handles listing all accounts
func (h *Handler) AccountList(c *gin.Context, req *models.Request) {
	// Retrieve accounts from database with pagination
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM account
		WHERE deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(context.Background(), query)
	if err != nil {
		logger.Error("Failed to list accounts", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list accounts", "")
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
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to list accounts", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logger.Info("Account list retrieved", zap.Int("count", len(accounts)))

	response := models.NewSuccessResponse(map[string]interface{}{
		"accounts": accounts,
		"count":    len(accounts),
	})
	c.JSON(http.StatusOK, response)
}

// AccountModify handles updating an existing account
func (h *Handler) AccountModify(c *gin.Context, req *models.Request) {
	// Parse the account data from request
	var account models.Account
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logger.Error("Failed to marshal account data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := json.Unmarshal(dataBytes, &account); err != nil {
		logger.Error("Failed to unmarshal account data", zap.Error(err))
		response := models.NewErrorResponse(models.ErrorCodeInvalidRequest, "Invalid account data format", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if account.ID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update timestamp
	account.Modified = time.Now().Unix()

	// Update account in database
	query := `
		UPDATE account
		SET title = ?, description = ?, modified = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		account.Title,
		account.Description,
		account.Modified,
		account.ID,
	)

	if err != nil {
		logger.Error("Failed to modify account", zap.Error(err), zap.String("id", account.ID))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to modify account", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Account not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Account modified", zap.String("id", account.ID))

	response := models.NewSuccessResponse(map[string]interface{}{
		"account": account,
	})
	c.JSON(http.StatusOK, response)
}

// AccountRemove handles soft-deleting an account
func (h *Handler) AccountRemove(c *gin.Context, req *models.Request) {
	// Get account ID from request data
	accountID, ok := req.Data["id"].(string)
	if !ok || accountID == "" {
		response := models.NewErrorResponse(models.ErrorCodeMissingData, "Account ID is required", "")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Soft-delete account in database (set deleted=true)
	query := `
		UPDATE account
		SET deleted = 1, modified = ?
		WHERE id = ? AND deleted = 0
	`

	result, err := h.db.Exec(
		context.Background(),
		query,
		time.Now().Unix(),
		accountID,
	)

	if err != nil {
		logger.Error("Failed to remove account", zap.Error(err), zap.String("id", accountID))
		response := models.NewErrorResponse(models.ErrorCodeInternalError, "Failed to remove account", "")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := models.NewErrorResponse(models.ErrorCodeEntityNotFound, "Account not found", "")
		c.JSON(http.StatusNotFound, response)
		return
	}

	logger.Info("Account removed", zap.String("id", accountID))

	response := models.NewSuccessResponse(map[string]interface{}{
		"id":      accountID,
		"deleted": true,
	})
	c.JSON(http.StatusOK, response)
}
