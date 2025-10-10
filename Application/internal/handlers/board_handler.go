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

// handleBoardCreate creates a new board
func (h *Handler) handleBoardCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionCreate)
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

	// Parse board data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	board := &models.Board{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Validate board
	if !board.IsValid() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid board data",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO board (id, title, description, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		board.ID,
		board.Title,
		board.Description,
		board.Created,
		board.Modified,
		board.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create board", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create board",
			"",
		))
		return
	}

	logger.Info("Board created",
		zap.String("board_id", board.ID),
		zap.String("title", board.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"board": board,
	})
	c.JSON(http.StatusCreated, response)
}

// handleBoardRead reads a single board by ID
func (h *Handler) handleBoardRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get board ID from request
	boardID, ok := req.Data["id"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	// Query board from database
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM board
		WHERE id = ? AND deleted = 0
	`

	var board models.Board
	err := h.db.QueryRow(c.Request.Context(), query, boardID).Scan(
		&board.ID,
		&board.Title,
		&board.Description,
		&board.Created,
		&board.Modified,
		&board.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Board not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read board", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read board",
			"",
		))
		return
	}

	logger.Info("Board read",
		zap.String("board_id", board.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"board": board,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardList lists all boards
func (h *Handler) handleBoardList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted boards ordered by modified date
	query := `
		SELECT id, title, description, created, modified, deleted
		FROM board
		WHERE deleted = 0
		ORDER BY modified DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list boards", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list boards",
			"",
		))
		return
	}
	defer rows.Close()

	boards := make([]models.Board, 0)
	for rows.Next() {
		var board models.Board
		err := rows.Scan(
			&board.ID,
			&board.Title,
			&board.Description,
			&board.Created,
			&board.Modified,
			&board.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan board", zap.Error(err))
			continue
		}
		boards = append(boards, board)
	}

	logger.Info("Boards listed",
		zap.Int("count", len(boards)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"boards": boards,
		"count":  len(boards),
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardModify updates an existing board
func (h *Handler) handleBoardModify(c *gin.Context, req *models.Request) {
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
	boardID, ok := req.Data["id"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	// Check if board exists
	checkQuery := `SELECT COUNT(*) FROM board WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, boardID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Board not found",
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
	query := "UPDATE board SET "
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
	args = append(args, boardID)

	_, err = h.db.Exec(c.Request.Context(), query, args...)
	if err != nil {
		logger.Error("Failed to update board", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update board",
			"",
		))
		return
	}

	logger.Info("Board updated",
		zap.String("board_id", boardID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      boardID,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardRemove soft-deletes a board
func (h *Handler) handleBoardRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "board", models.PermissionDelete)
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
	boardID, ok := req.Data["id"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	// Soft delete the board
	query := `UPDATE board SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), boardID)
	if err != nil {
		logger.Error("Failed to delete board", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete board",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Board not found",
			"",
		))
		return
	}

	logger.Info("Board deleted",
		zap.String("board_id", boardID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      boardID,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardAddTicket adds a ticket to a board
func (h *Handler) handleBoardAddTicket(c *gin.Context, req *models.Request) {
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

	// Parse data from request
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
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

	// Check if board exists
	boardQuery := `SELECT COUNT(*) FROM board WHERE id = ? AND deleted = 0`
	var boardCount int
	err = h.db.QueryRow(c.Request.Context(), boardQuery, boardID).Scan(&boardCount)
	if err != nil || boardCount == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Board not found",
			"",
		))
		return
	}

	// Check if mapping already exists
	checkQuery := `SELECT COUNT(*) FROM ticket_board_mapping WHERE ticket_id = ? AND board_id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, ticketID, boardID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check existing mapping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check existing mapping",
			"",
		))
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeAlreadyExists,
			"Ticket is already assigned to this board",
			"",
		))
		return
	}

	// Create mapping
	mapping := &models.TicketBoardMapping{
		ID:       uuid.New().String(),
		TicketID: ticketID,
		BoardID:  boardID,
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
		Deleted:  false,
	}

	// Insert into database
	query := `
		INSERT INTO ticket_board_mapping (id, ticket_id, board_id, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		mapping.ID,
		mapping.TicketID,
		mapping.BoardID,
		mapping.Created,
		mapping.Modified,
		mapping.Deleted,
	)

	if err != nil {
		logger.Error("Failed to add ticket to board", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add ticket to board",
			"",
		))
		return
	}

	logger.Info("Ticket added to board",
		zap.String("ticket_id", ticketID),
		zap.String("board_id", boardID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mapping": mapping,
	})
	c.JSON(http.StatusCreated, response)
}

// handleBoardRemoveTicket removes a ticket from a board
func (h *Handler) handleBoardRemoveTicket(c *gin.Context, req *models.Request) {
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

	// Parse data from request
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
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
	query := `UPDATE ticket_board_mapping SET deleted = 1, modified = ? WHERE ticket_id = ? AND board_id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), ticketID, boardID)
	if err != nil {
		logger.Error("Failed to remove ticket from board", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove ticket from board",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Ticket not found on this board",
			"",
		))
		return
	}

	logger.Info("Ticket removed from board",
		zap.String("ticket_id", ticketID),
		zap.String("board_id", boardID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"removed":  true,
		"ticketId": ticketID,
		"boardId":  boardID,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardListTickets lists all tickets on a board
func (h *Handler) handleBoardListTickets(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get board ID from request
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	// Query all ticket IDs mapped to this board
	query := `
		SELECT id, ticket_id, board_id, created, modified, deleted
		FROM ticket_board_mapping
		WHERE board_id = ? AND deleted = 0
		ORDER BY created DESC
	`

	rows, err := h.db.Query(c.Request.Context(), query, boardID)
	if err != nil {
		logger.Error("Failed to list board tickets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list board tickets",
			"",
		))
		return
	}
	defer rows.Close()

	mappings := make([]models.TicketBoardMapping, 0)
	for rows.Next() {
		var mapping models.TicketBoardMapping
		err := rows.Scan(
			&mapping.ID,
			&mapping.TicketID,
			&mapping.BoardID,
			&mapping.Created,
			&mapping.Modified,
			&mapping.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan mapping", zap.Error(err))
			continue
		}
		mappings = append(mappings, mapping)
	}

	logger.Info("Board tickets listed",
		zap.String("board_id", boardID),
		zap.Int("count", len(mappings)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mappings": mappings,
		"count":    len(mappings),
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardSetMetadata sets metadata for a board
func (h *Handler) handleBoardSetMetadata(c *gin.Context, req *models.Request) {
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

	// Parse data from request
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	property, ok := req.Data["property"].(string)
	if !ok || property == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing property",
			"",
		))
		return
	}

	value := getStringFromData(req.Data, "value")

	// Check if board exists
	boardQuery := `SELECT COUNT(*) FROM board WHERE id = ? AND deleted = 0`
	var boardCount int
	err = h.db.QueryRow(c.Request.Context(), boardQuery, boardID).Scan(&boardCount)
	if err != nil || boardCount == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Board not found",
			"",
		))
		return
	}

	// Check if metadata already exists for this property
	checkQuery := `SELECT id FROM board_meta_data WHERE board_id = ? AND property = ? AND deleted = 0`
	var existingID string
	err = h.db.QueryRow(c.Request.Context(), checkQuery, boardID, property).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new metadata
		metadata := &models.BoardMetaData{
			ID:       uuid.New().String(),
			BoardID:  boardID,
			Property: property,
			Value:    value,
			Created:  time.Now().Unix(),
			Modified: time.Now().Unix(),
			Deleted:  false,
		}

		query := `
			INSERT INTO board_meta_data (id, board_id, property, value, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`

		_, err = h.db.Exec(c.Request.Context(), query,
			metadata.ID,
			metadata.BoardID,
			metadata.Property,
			metadata.Value,
			metadata.Created,
			metadata.Modified,
			metadata.Deleted,
		)

		if err != nil {
			logger.Error("Failed to set board metadata", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to set board metadata",
				"",
			))
			return
		}

		logger.Info("Board metadata created",
			zap.String("board_id", boardID),
			zap.String("property", property),
			zap.String("username", username),
		)

		response := models.NewSuccessResponse(map[string]interface{}{
			"metadata": metadata,
		})
		c.JSON(http.StatusCreated, response)
	} else if err == nil {
		// Update existing metadata
		updateQuery := `UPDATE board_meta_data SET value = ?, modified = ? WHERE id = ?`
		_, err = h.db.Exec(c.Request.Context(), updateQuery, value, time.Now().Unix(), existingID)

		if err != nil {
			logger.Error("Failed to update board metadata", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrorCodeInternalError,
				"Failed to update board metadata",
				"",
			))
			return
		}

		logger.Info("Board metadata updated",
			zap.String("board_id", boardID),
			zap.String("property", property),
			zap.String("username", username),
		)

		response := models.NewSuccessResponse(map[string]interface{}{
			"updated":  true,
			"id":       existingID,
			"property": property,
			"value":    value,
		})
		c.JSON(http.StatusOK, response)
	} else {
		logger.Error("Failed to check existing metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to check existing metadata",
			"",
		))
	}
}

// handleBoardGetMetadata gets a specific metadata property for a board
func (h *Handler) handleBoardGetMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Parse data from request
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	property, ok := req.Data["property"].(string)
	if !ok || property == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing property",
			"",
		))
		return
	}

	// Query metadata from database
	query := `
		SELECT id, board_id, property, value, created, modified, deleted
		FROM board_meta_data
		WHERE board_id = ? AND property = ? AND deleted = 0
	`

	var metadata models.BoardMetaData
	err := h.db.QueryRow(c.Request.Context(), query, boardID, property).Scan(
		&metadata.ID,
		&metadata.BoardID,
		&metadata.Property,
		&metadata.Value,
		&metadata.Created,
		&metadata.Modified,
		&metadata.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Metadata not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to get board metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get board metadata",
			"",
		))
		return
	}

	logger.Info("Board metadata retrieved",
		zap.String("board_id", boardID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardListMetadata lists all metadata for a board
func (h *Handler) handleBoardListMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get board ID from request
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	// Query all metadata for this board
	query := `
		SELECT id, board_id, property, value, created, modified, deleted
		FROM board_meta_data
		WHERE board_id = ? AND deleted = 0
		ORDER BY property ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query, boardID)
	if err != nil {
		logger.Error("Failed to list board metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list board metadata",
			"",
		))
		return
	}
	defer rows.Close()

	metadata := make([]models.BoardMetaData, 0)
	for rows.Next() {
		var meta models.BoardMetaData
		err := rows.Scan(
			&meta.ID,
			&meta.BoardID,
			&meta.Property,
			&meta.Value,
			&meta.Created,
			&meta.Modified,
			&meta.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan metadata", zap.Error(err))
			continue
		}
		metadata = append(metadata, meta)
	}

	logger.Info("Board metadata listed",
		zap.String("board_id", boardID),
		zap.Int("count", len(metadata)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
		"count":    len(metadata),
	})
	c.JSON(http.StatusOK, response)
}

// handleBoardRemoveMetadata removes a specific metadata property from a board
func (h *Handler) handleBoardRemoveMetadata(c *gin.Context, req *models.Request) {
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

	// Parse data from request
	boardID, ok := req.Data["boardId"].(string)
	if !ok || boardID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing board ID",
			"",
		))
		return
	}

	property, ok := req.Data["property"].(string)
	if !ok || property == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing property",
			"",
		))
		return
	}

	// Soft delete the metadata
	query := `UPDATE board_meta_data SET deleted = 1, modified = ? WHERE board_id = ? AND property = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), boardID, property)
	if err != nil {
		logger.Error("Failed to remove board metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove board metadata",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Metadata not found",
			"",
		))
		return
	}

	logger.Info("Board metadata removed",
		zap.String("board_id", boardID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted":  true,
		"boardId":  boardID,
		"property": property,
	})
	c.JSON(http.StatusOK, response)
}
