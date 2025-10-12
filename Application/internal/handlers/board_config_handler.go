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
	"helixtrack.ru/core/internal/websocket"
)

// handleBoardConfigureColumns configures board columns (batch operation)
func (h *Handler) handleBoardConfigureColumns(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get board ID
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
			"",
		))
		return
	}

	// Verify board exists
	checkQuery := `SELECT COUNT(*) FROM board WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, boardID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Board not found",
			"",
		))
		return
	}

	// Get columns array
	columnsData, ok := req.Data["columns"].([]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing columns array",
			"",
		))
		return
	}

	// Delete existing columns
	_, err = h.db.Exec(c.Request.Context(),
		"UPDATE board_column SET deleted = 1 WHERE board_id = ?",
		boardID,
	)
	if err != nil {
		logger.Error("Failed to delete old columns", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to configure columns",
			"",
		))
		return
	}

	// Insert new columns
	now := time.Now().Unix()
	columnIDs := make([]string, 0, len(columnsData))

	for i, colData := range columnsData {
		colMap, ok := colData.(map[string]interface{})
		if !ok {
			continue
		}

		columnID := uuid.New().String()
		title := getStringFromData(colMap, "title")
		statusID := getStringFromData(colMap, "statusId")
		maxItems := 0
		if val, ok := colMap["maxItems"].(float64); ok {
			maxItems = int(val)
		}

		query := `
			INSERT INTO board_column (id, board_id, title, status_id, position, max_items, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
		`

		var statusIDPtr interface{}
		if statusID != "" {
			statusIDPtr = statusID
		}

		var maxItemsPtr interface{}
		if maxItems > 0 {
			maxItemsPtr = maxItems
		}

		_, err = h.db.Exec(c.Request.Context(), query,
			columnID, boardID, title, statusIDPtr, i, maxItemsPtr, now, now,
		)
		if err != nil {
			logger.Error("Failed to insert column", zap.Error(err))
			continue
		}

		columnIDs = append(columnIDs, columnID)
	}

	logger.Info("Board columns configured",
		zap.String("board_id", boardID),
		zap.Int("column_count", len(columnIDs)),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardConfigureColumns",
		"board",
		boardID,
		username,
		map[string]interface{}{
			"boardId":     boardID,
			"columnCount": len(columnIDs),
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"configured":  true,
		"boardId":     boardID,
		"columnCount": len(columnIDs),
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardAddColumn adds a column to a board
func (h *Handler) handleBoardAddColumn(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get parameters
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
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

	statusID := getStringFromData(req.Data, "statusId")
	position := 0
	if val, ok := req.Data["position"].(float64); ok {
		position = int(val)
	}
	maxItems := 0
	if val, ok := req.Data["maxItems"].(float64); ok {
		maxItems = int(val)
	}

	// Create column
	columnID := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO board_column (id, board_id, title, status_id, position, max_items, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
	`

	var statusIDPtr, maxItemsPtr interface{}
	if statusID != "" {
		statusIDPtr = statusID
	}
	if maxItems > 0 {
		maxItemsPtr = maxItems
	}

	_, err = h.db.Exec(c.Request.Context(), query,
		columnID, boardID, title, statusIDPtr, position, maxItemsPtr, now, now,
	)
	if err != nil {
		logger.Error("Failed to add column", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add column",
			"",
		))
		return
	}

	logger.Info("Board column added",
		zap.String("column_id", columnID),
		zap.String("board_id", boardID),
		zap.String("title", title),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardAddColumn",
		"board",
		boardID,
		username,
		map[string]interface{}{
			"columnId": columnID,
			"boardId":  boardID,
			"title":    title,
			"position": position,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":    true,
		"columnId": columnID,
		"boardId":  boardID,
	})
	c.JSON(http.StatusCreated, response)
}

// handleBoardRemoveColumn removes a column from a board
func (h *Handler) handleBoardRemoveColumn(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get column ID
	columnID, ok := req.Data["columnId"].(string)
	if !ok || columnID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing columnId",
			"",
		))
		return
	}

	// Soft delete column
	query := `UPDATE board_column SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), columnID)
	if err != nil {
		logger.Error("Failed to remove column", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove column",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Column not found",
			"",
		))
		return
	}

	logger.Info("Board column removed",
		zap.String("column_id", columnID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardRemoveColumn",
		"board",
		columnID,
		username,
		map[string]interface{}{
			"columnId": columnID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":  true,
		"columnId": columnID,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardModifyColumn updates a column
func (h *Handler) handleBoardModifyColumn(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get column ID
	columnID, ok := req.Data["columnId"].(string)
	if !ok || columnID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing columnId",
			"",
		))
		return
	}

	// Check if column exists
	checkQuery := `SELECT COUNT(*) FROM board_column WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, columnID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Column not found",
			"",
		))
		return
	}

	// Build update query
	updates := make(map[string]interface{})

	if title, ok := req.Data["title"].(string); ok && title != "" {
		updates["title"] = title
	}
	if statusID, ok := req.Data["statusId"].(string); ok {
		if statusID != "" {
			updates["status_id"] = statusID
		} else {
			updates["status_id"] = nil
		}
	}
	if position, ok := req.Data["position"].(float64); ok {
		updates["position"] = int(position)
	}
	if maxItems, ok := req.Data["maxItems"].(float64); ok {
		updates["max_items"] = int(maxItems)
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

	// Build and execute query
	query := "UPDATE board_column SET "
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
	args = append(args, columnID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update column", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update column",
			"",
		))
		return
	}

	logger.Info("Board column updated",
		zap.String("column_id", columnID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardModifyColumn",
		"board",
		columnID,
		username,
		updates,
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated":  true,
		"columnId": columnID,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardListColumns lists all columns for a board
func (h *Handler) handleBoardListColumns(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get board ID
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
			"",
		))
		return
	}

	// Query columns
	query := `
		SELECT id, title, status_id, position, max_items, created, modified
		FROM board_column
		WHERE board_id = ? AND deleted = 0
		ORDER BY position ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, boardID)
	if err != nil {
		logger.Error("Failed to list columns", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list columns",
			"",
		))
		return
	}
	defer rows.Close()

	columns := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title string
		var statusID sql.NullString
		var position, maxItems sql.NullInt64
		var created, modified int64

		err := rows.Scan(&id, &title, &statusID, &position, &maxItems, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan column", zap.Error(err))
			continue
		}

		column := map[string]interface{}{
			"id":       id,
			"title":    title,
			"statusId": statusID.String,
			"position": int(position.Int64),
			"maxItems": nil,
			"created":  created,
			"modified": modified,
		}

		if maxItems.Valid && maxItems.Int64 > 0 {
			column["maxItems"] = int(maxItems.Int64)
		}

		columns = append(columns, column)
	}

	logger.Info("Board columns listed",
		zap.String("board_id", boardID),
		zap.Int("count", len(columns)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"boardId": boardID,
		"columns": columns,
		"count":   len(columns),
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardAddSwimlane adds a swimlane to a board
func (h *Handler) handleBoardAddSwimlane(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get parameters
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
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

	query := getStringFromData(req.Data, "query")
	position := 0
	if val, ok := req.Data["position"].(float64); ok {
		position = int(val)
	}

	// Create swimlane
	swimlaneID := uuid.New().String()
	now := time.Now().Unix()

	insertQuery := `
		INSERT INTO board_swimlane (id, board_id, title, query, position, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	`

	var queryPtr interface{}
	if query != "" {
		queryPtr = query
	}

	_, err = h.db.Exec(c.Request.Context(), insertQuery,
		swimlaneID, boardID, title, queryPtr, position, now, now,
	)
	if err != nil {
		logger.Error("Failed to add swimlane", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add swimlane",
			"",
		))
		return
	}

	logger.Info("Board swimlane added",
		zap.String("swimlane_id", swimlaneID),
		zap.String("board_id", boardID),
		zap.String("title", title),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardAddSwimlane",
		"board",
		boardID,
		username,
		map[string]interface{}{
			"swimlaneId": swimlaneID,
			"boardId":    boardID,
			"title":      title,
			"position":   position,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":      true,
		"swimlaneId": swimlaneID,
		"boardId":    boardID,
	})
	c.JSON(http.StatusCreated, response)
}

// handleBoardRemoveSwimlane removes a swimlane from a board
func (h *Handler) handleBoardRemoveSwimlane(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get swimlane ID
	swimlaneID, ok := req.Data["swimlaneId"].(string)
	if !ok || swimlaneID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing swimlaneId",
			"",
		))
		return
	}

	// Soft delete swimlane
	query := `UPDATE board_swimlane SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), swimlaneID)
	if err != nil {
		logger.Error("Failed to remove swimlane", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove swimlane",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Swimlane not found",
			"",
		))
		return
	}

	logger.Info("Board swimlane removed",
		zap.String("swimlane_id", swimlaneID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardRemoveSwimlane",
		"board",
		swimlaneID,
		username,
		map[string]interface{}{
			"swimlaneId": swimlaneID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":    true,
		"swimlaneId": swimlaneID,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardListSwimlanes lists all swimlanes for a board
func (h *Handler) handleBoardListSwimlanes(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get board ID
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
			"",
		))
		return
	}

	// Query swimlanes
	query := `
		SELECT id, title, query, position, created, modified
		FROM board_swimlane
		WHERE board_id = ? AND deleted = 0
		ORDER BY position ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, boardID)
	if err != nil {
		logger.Error("Failed to list swimlanes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list swimlanes",
			"",
		))
		return
	}
	defer rows.Close()

	swimlanes := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title string
		var queryVal sql.NullString
		var position int
		var created, modified int64

		err := rows.Scan(&id, &title, &queryVal, &position, &created, &modified)
		if err != nil {
			logger.Error("Failed to scan swimlane", zap.Error(err))
			continue
		}

		swimlanes = append(swimlanes, map[string]interface{}{
			"id":       id,
			"title":    title,
			"query":    queryVal.String,
			"position": position,
			"created":  created,
			"modified": modified,
		})
	}

	logger.Info("Board swimlanes listed",
		zap.String("board_id", boardID),
		zap.Int("count", len(swimlanes)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"boardId":   boardID,
		"swimlanes": swimlanes,
		"count":     len(swimlanes),
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardAddQuickFilter adds a quick filter to a board
func (h *Handler) handleBoardAddQuickFilter(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get parameters
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
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

	query := getStringFromData(req.Data, "query")
	position := 0
	if val, ok := req.Data["position"].(float64); ok {
		position = int(val)
	}

	// Create quick filter
	filterID := uuid.New().String()
	now := time.Now().Unix()

	insertQuery := `
		INSERT INTO board_quick_filter (id, board_id, title, query, position, created, deleted)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`

	var queryPtr interface{}
	if query != "" {
		queryPtr = query
	}

	_, err = h.db.Exec(c.Request.Context(), insertQuery,
		filterID, boardID, title, queryPtr, position, now,
	)
	if err != nil {
		logger.Error("Failed to add quick filter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add quick filter",
			"",
		))
		return
	}

	logger.Info("Board quick filter added",
		zap.String("filter_id", filterID),
		zap.String("board_id", boardID),
		zap.String("title", title),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardAddQuickFilter",
		"board",
		boardID,
		username,
		map[string]interface{}{
			"filterId": filterID,
			"boardId":  boardID,
			"title":    title,
			"position": position,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"added":    true,
		"filterId": filterID,
		"boardId":  boardID,
	})
	c.JSON(http.StatusCreated, response)
}

// handleBoardRemoveQuickFilter removes a quick filter from a board
func (h *Handler) handleBoardRemoveQuickFilter(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get filter ID
	filterID, ok := req.Data["filterId"].(string)
	if !ok || filterID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing filterId",
			"",
		))
		return
	}

	// Soft delete quick filter
	query := `UPDATE board_quick_filter SET deleted = 1 WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, filterID)
	if err != nil {
		logger.Error("Failed to remove quick filter", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove quick filter",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Quick filter not found",
			"",
		))
		return
	}

	logger.Info("Board quick filter removed",
		zap.String("filter_id", filterID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardRemoveQuickFilter",
		"board",
		filterID,
		username,
		map[string]interface{}{
			"filterId": filterID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":  true,
		"filterId": filterID,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardListQuickFilters lists all quick filters for a board
func (h *Handler) handleBoardListQuickFilters(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get board ID
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
			"",
		))
		return
	}

	// Query quick filters
	query := `
		SELECT id, title, query, position, created
		FROM board_quick_filter
		WHERE board_id = ? AND deleted = 0
		ORDER BY position ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, boardID)
	if err != nil {
		logger.Error("Failed to list quick filters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list quick filters",
			"",
		))
		return
	}
	defer rows.Close()

	filters := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, title string
		var queryVal sql.NullString
		var position int
		var created int64

		err := rows.Scan(&id, &title, &queryVal, &position, &created)
		if err != nil {
			logger.Error("Failed to scan quick filter", zap.Error(err))
			continue
		}

		filters = append(filters, map[string]interface{}{
			"id":       id,
			"title":    title,
			"query":    queryVal.String,
			"position": position,
			"created":  created,
		})
	}

	logger.Info("Board quick filters listed",
		zap.String("board_id", boardID),
		zap.Int("count", len(filters)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"boardId": boardID,
		"filters": filters,
		"count":   len(filters),
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardSetType sets the board type (scrum/kanban)
func (h *Handler) handleBoardSetType(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionUpdate)
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

	// Get parameters
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing boardId",
			"",
		))
		return
	}

	boardType, ok := req.Data["boardType"].(string)
	if !ok || (boardType != "scrum" && boardType != "kanban") {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid boardType (must be 'scrum' or 'kanban')",
			"",
		))
		return
	}

	// Update board type
	query := `UPDATE board SET type = ?, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, boardType, time.Now().Unix(), boardID)
	if err != nil {
		logger.Error("Failed to set board type", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to set board type",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Board not found",
			"",
		))
		return
	}

	logger.Info("Board type set",
		zap.String("board_id", boardID),
		zap.String("board_type", boardType),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		"boardSetType",
		"board",
		boardID,
		username,
		map[string]interface{}{
			"boardId":   boardID,
			"boardType": boardType,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated":   true,
		"boardId":   boardID,
		"boardType": boardType,
	})
	c.JSON(http.StatusOK, response)
}
