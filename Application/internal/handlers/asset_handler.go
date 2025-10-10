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

// handleAssetCreate creates a new asset
func (h *Handler) handleAssetCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionCreate)
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

	// Parse asset data from request
	url, ok := req.Data["url"].(string)
	if !ok || url == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing url",
			"",
		))
		return
	}

	asset := &models.Asset{
		ID:          uuid.New().String(),
		URL:         url,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Insert into database
	query := `
		INSERT INTO asset (id, url, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		asset.ID,
		asset.URL,
		asset.Description,
		asset.Created,
		asset.Modified,
		asset.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create asset",
			"",
		))
		return
	}

	logger.Info("Asset created",
		zap.String("asset_id", asset.ID),
		zap.String("url", asset.URL),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"asset": asset,
	})
	c.JSON(http.StatusCreated, response)
}

// handleAssetRead reads a single asset by ID
func (h *Handler) handleAssetRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get asset ID from request
	assetID, ok := req.Data["id"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	// Query asset from database
	query := `
		SELECT id, url, description, created, modified, deleted
		FROM asset
		WHERE id = ? AND deleted = 0
	`

	var asset models.Asset
	err := h.db.QueryRow(c.Request.Context(), query, assetID).Scan(
		&asset.ID,
		&asset.URL,
		&asset.Description,
		&asset.Created,
		&asset.Modified,
		&asset.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Asset not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read asset",
			"",
		))
		return
	}

	logger.Info("Asset read",
		zap.String("asset_id", asset.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"asset": asset,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetList lists all assets
func (h *Handler) handleAssetList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted assets ordered by created
	query := `
		SELECT id, url, description, created, modified, deleted
		FROM asset
		WHERE deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list assets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list assets",
			"",
		))
		return
	}
	defer rows.Close()

	assets := make([]models.Asset, 0)
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.URL,
			&asset.Description,
			&asset.Created,
			&asset.Modified,
			&asset.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan asset", zap.Error(err))
			continue
		}
		assets = append(assets, asset)
	}

	logger.Info("Assets listed",
		zap.Int("count", len(assets)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"assets": assets,
		"count":  len(assets),
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetModify updates an existing asset
func (h *Handler) handleAssetModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionUpdate)
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

	// Get asset ID
	assetID, ok := req.Data["id"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	// Check if asset exists
	checkQuery := `SELECT COUNT(*) FROM asset WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, assetID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Asset not found",
			"",
		))
		return
	}

	// Build update query dynamically based on provided fields
	updates := make(map[string]interface{})

	if url, ok := req.Data["url"].(string); ok && url != "" {
		updates["url"] = url
	}
	if description, ok := req.Data["description"].(string); ok {
		updates["description"] = description
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
	query := "UPDATE asset SET "
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
	args = append(args, assetID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update asset",
			"",
		))
		return
	}

	logger.Info("Asset updated",
		zap.String("asset_id", assetID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      assetID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetRemove soft-deletes an asset
func (h *Handler) handleAssetRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionDelete)
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

	// Get asset ID
	assetID, ok := req.Data["id"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	// Soft delete the asset
	query := `UPDATE asset SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), assetID)
	if err != nil {
		logger.Error("Failed to delete asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete asset",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Asset not found",
			"",
		))
		return
	}

	logger.Info("Asset deleted",
		zap.String("asset_id", assetID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      assetID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetAddTicket adds an asset to a ticket
func (h *Handler) handleAssetAddTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionUpdate)
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

	// Get asset ID and ticket ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
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

	// Create mapping
	mappingID := uuid.New().String()
	query := `
		INSERT INTO asset_ticket_mapping (id, asset_id, ticket_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err = h.db.Exec(c.Request.Context(), query, mappingID, assetID, ticketID, now, now, false)
	if err != nil {
		logger.Error("Failed to add asset to ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add asset to ticket",
			"",
		))
		return
	}

	logger.Info("Asset added to ticket",
		zap.String("asset_id", assetID),
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":    true,
		"assetId":  assetID,
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetRemoveTicket removes an asset from a ticket
func (h *Handler) handleAssetRemoveTicket(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionUpdate)
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

	// Get asset ID and ticket ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
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

	// Remove mapping (soft delete)
	query := `UPDATE asset_ticket_mapping SET deleted = 1, modified = ? WHERE asset_id = ? AND ticket_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), assetID, ticketID)
	if err != nil {
		logger.Error("Failed to remove asset from ticket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove asset from ticket",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Asset-ticket mapping not found",
			"",
		))
		return
	}

	logger.Info("Asset removed from ticket",
		zap.String("asset_id", assetID),
		zap.String("ticket_id", ticketID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":  true,
		"assetId":  assetID,
		"ticketId": ticketID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetListTickets lists all tickets for an asset
func (h *Handler) handleAssetListTickets(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get asset ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	// Query tickets
	query := `
		SELECT ticket_id
		FROM asset_ticket_mapping
		WHERE asset_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, assetID)
	if err != nil {
		logger.Error("Failed to list tickets for asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list tickets",
			"",
		))
		return
	}
	defer rows.Close()

	ticketIDs := make([]string, 0)
	for rows.Next() {
		var ticketID string
		if err := rows.Scan(&ticketID); err != nil {
			logger.Error("Failed to scan ticket ID", zap.Error(err))
			continue
		}
		ticketIDs = append(ticketIDs, ticketID)
	}

	logger.Info("Tickets listed for asset",
		zap.String("asset_id", assetID),
		zap.Int("count", len(ticketIDs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"ticketIds": ticketIDs,
		"count":     len(ticketIDs),
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetAddComment adds an asset to a comment
func (h *Handler) handleAssetAddComment(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionUpdate)
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

	// Get asset ID and comment ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	commentID, ok := req.Data["commentId"].(string)
	if !ok || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing comment ID",
			"",
		))
		return
	}

	// Create mapping
	mappingID := uuid.New().String()
	query := `
		INSERT INTO asset_comment_mapping (id, asset_id, comment_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err = h.db.Exec(c.Request.Context(), query, mappingID, assetID, commentID, now, now, false)
	if err != nil {
		logger.Error("Failed to add asset to comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add asset to comment",
			"",
		))
		return
	}

	logger.Info("Asset added to comment",
		zap.String("asset_id", assetID),
		zap.String("comment_id", commentID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":     true,
		"assetId":   assetID,
		"commentId": commentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetRemoveComment removes an asset from a comment
func (h *Handler) handleAssetRemoveComment(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionUpdate)
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

	// Get asset ID and comment ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	commentID, ok := req.Data["commentId"].(string)
	if !ok || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing comment ID",
			"",
		))
		return
	}

	// Remove mapping (soft delete)
	query := `UPDATE asset_comment_mapping SET deleted = 1, modified = ? WHERE asset_id = ? AND comment_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), assetID, commentID)
	if err != nil {
		logger.Error("Failed to remove asset from comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove asset from comment",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Asset-comment mapping not found",
			"",
		))
		return
	}

	logger.Info("Asset removed from comment",
		zap.String("asset_id", assetID),
		zap.String("comment_id", commentID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":   true,
		"assetId":   assetID,
		"commentId": commentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetListComments lists all comments for an asset
func (h *Handler) handleAssetListComments(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get asset ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	// Query comments
	query := `
		SELECT comment_id
		FROM asset_comment_mapping
		WHERE asset_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, assetID)
	if err != nil {
		logger.Error("Failed to list comments for asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list comments",
			"",
		))
		return
	}
	defer rows.Close()

	commentIDs := make([]string, 0)
	for rows.Next() {
		var commentID string
		if err := rows.Scan(&commentID); err != nil {
			logger.Error("Failed to scan comment ID", zap.Error(err))
			continue
		}
		commentIDs = append(commentIDs, commentID)
	}

	logger.Info("Comments listed for asset",
		zap.String("asset_id", assetID),
		zap.Int("count", len(commentIDs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"commentIds": commentIDs,
		"count":      len(commentIDs),
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetAddProject adds an asset to a project
func (h *Handler) handleAssetAddProject(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionUpdate)
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

	// Get asset ID and project ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
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

	// Create mapping
	mappingID := uuid.New().String()
	query := `
		INSERT INTO asset_project_mapping (id, asset_id, project_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err = h.db.Exec(c.Request.Context(), query, mappingID, assetID, projectID, now, now, false)
	if err != nil {
		logger.Error("Failed to add asset to project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add asset to project",
			"",
		))
		return
	}

	logger.Info("Asset added to project",
		zap.String("asset_id", assetID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":     true,
		"assetId":   assetID,
		"projectId": projectID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetRemoveProject removes an asset from a project
func (h *Handler) handleAssetRemoveProject(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "asset", models.PermissionUpdate)
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

	// Get asset ID and project ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
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

	// Remove mapping (soft delete)
	query := `UPDATE asset_project_mapping SET deleted = 1, modified = ? WHERE asset_id = ? AND project_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), assetID, projectID)
	if err != nil {
		logger.Error("Failed to remove asset from project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove asset from project",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Asset-project mapping not found",
			"",
		))
		return
	}

	logger.Info("Asset removed from project",
		zap.String("asset_id", assetID),
		zap.String("project_id", projectID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":   true,
		"assetId":   assetID,
		"projectId": projectID,
	})
	c.JSON(http.StatusOK, response)
}

// handleAssetListProjects lists all projects for an asset
func (h *Handler) handleAssetListProjects(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get asset ID
	assetID, ok := req.Data["assetId"].(string)
	if !ok || assetID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing asset ID",
			"",
		))
		return
	}

	// Query projects
	query := `
		SELECT project_id
		FROM asset_project_mapping
		WHERE asset_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, assetID)
	if err != nil {
		logger.Error("Failed to list projects for asset", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list projects",
			"",
		))
		return
	}
	defer rows.Close()

	projectIDs := make([]string, 0)
	for rows.Next() {
		var projectID string
		if err := rows.Scan(&projectID); err != nil {
			logger.Error("Failed to scan project ID", zap.Error(err))
			continue
		}
		projectIDs = append(projectIDs, projectID)
	}

	logger.Info("Projects listed for asset",
		zap.String("asset_id", assetID),
		zap.Int("count", len(projectIDs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"projectIds": projectIDs,
		"count":      len(projectIDs),
	})
	c.JSON(http.StatusOK, response)
}
