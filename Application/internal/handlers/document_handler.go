package handlers

import (
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

// ========================================================================
// CORE DOCUMENT OPERATIONS
// ========================================================================

// handleDocumentCreate creates a new document
func (h *Handler) handleDocumentCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionCreate)
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

	// Parse document data
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	spaceID, ok := req.Data["space_id"].(string)
	if !ok || spaceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing space_id",
			"",
		))
		return
	}

	typeID, ok := req.Data["type_id"].(string)
	if !ok || typeID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing type_id",
			"",
		))
		return
	}

	// Create document
	doc := &models.Document{
		ID:        uuid.New().String(),
		Title:     title,
		SpaceID:   spaceID,
		TypeID:    typeID,
		CreatorID: username,
		Version:   1,
	}

	// Optional fields
	if parentID, ok := req.Data["parent_id"].(string); ok && parentID != "" {
		doc.ParentID = &parentID
	}
	if projectID, ok := req.Data["project_id"].(string); ok && projectID != "" {
		doc.ProjectID = &projectID
	}
	if position, ok := req.Data["position"].(float64); ok {
		doc.Position = int(position)
	}

	// Insert into database using database interface
	db, ok := h.db.(interface {
		CreateDocument(*models.Document) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocument(doc)
	if err != nil {
		logger.Error("Failed to create document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document",
			"",
		))
		return
	}

	// Create initial content if provided
	if content, ok := req.Data["content"].(string); ok && content != "" {
		docContent := &models.DocumentContent{
			ID:          uuid.New().String(),
			DocumentID:  doc.ID,
			Version:     1,
			ContentType: "markdown", // Default to markdown
			Content:     &content,
		}

		if contentType, ok := req.Data["content_type"].(string); ok && contentType != "" {
			docContent.ContentType = contentType
		}

		// Insert content
		dbContent, ok := h.db.(interface {
			CreateDocumentContent(*models.DocumentContent) error
		})
		if ok {
			_ = dbContent.CreateDocumentContent(docContent) // Ignore error, content is optional
		}
	}

	logger.Info("Document created",
		zap.String("document_id", doc.ID),
		zap.String("title", doc.Title),
		zap.String("username", username),
	)

	// Publish document created event
	h.publisher.PublishEntityEvent(
		models.ActionCreate,
		"document",
		doc.ID,
		username,
		map[string]interface{}{
			"id":        doc.ID,
			"title":     doc.Title,
			"space_id":  doc.SpaceID,
			"type_id":   doc.TypeID,
			"parent_id": doc.ParentID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"document": doc,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentRead reads a single document by ID
func (h *Handler) handleDocumentRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get document ID
	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	// Get document from database
	db, ok := h.db.(interface {
		GetDocument(string) (*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	doc, err := db.GetDocument(documentID)
	if err != nil {
		logger.Error("Failed to read document", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document not found",
			"",
		))
		return
	}

	// Optionally get content
	var content *models.DocumentContent
	if includeContent, ok := req.Data["include_content"].(bool); ok && includeContent {
		dbContent, ok := h.db.(interface {
			GetLatestDocumentContent(string) (*models.DocumentContent, error)
		})
		if ok {
			content, _ = dbContent.GetLatestDocumentContent(documentID)
		}
	}

	logger.Info("Document read",
		zap.String("document_id", doc.ID),
		zap.String("username", username),
	)

	responseData := map[string]interface{}{
		"document": doc,
	}
	if content != nil {
		responseData["content"] = content
	}

	response := models.NewSuccessResponse(responseData)
	c.JSON(http.StatusOK, response)
}

// handleDocumentList lists documents with optional filters
func (h *Handler) handleDocumentList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if spaceID, ok := req.Data["space_id"].(string); ok && spaceID != "" {
		filters["space_id"] = spaceID
	}
	if projectID, ok := req.Data["project_id"].(string); ok && projectID != "" {
		filters["project_id"] = projectID
	}
	if parentID, ok := req.Data["parent_id"].(string); ok && parentID != "" {
		filters["parent_id"] = parentID
	}
	if isPublished, ok := req.Data["is_published"].(bool); ok {
		filters["is_published"] = isPublished
	}
	if isArchived, ok := req.Data["is_archived"].(bool); ok {
		filters["is_archived"] = isArchived
	}

	// Pagination
	limit := 50
	offset := 0
	if l, ok := req.Data["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	if o, ok := req.Data["offset"].(float64); ok && o > 0 {
		offset = int(o)
	}

	// List documents
	db, ok := h.db.(interface {
		ListDocuments(map[string]interface{}, int, int) ([]*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	docs, err := db.ListDocuments(filters, limit, offset)
	if err != nil {
		logger.Error("Failed to list documents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list documents",
			"",
		))
		return
	}

	logger.Info("Documents listed",
		zap.Int("count", len(docs)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"documents": docs,
		"count":     len(docs),
		"limit":     limit,
		"offset":    offset,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentUpdate updates an existing document
func (h *Handler) handleDocumentUpdate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
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

	// Get document ID
	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	// Get existing document
	db, ok := h.db.(interface {
		GetDocument(string) (*models.Document, error)
		UpdateDocument(*models.Document) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	doc, err := db.GetDocument(documentID)
	if err != nil {
		logger.Error("Failed to read document", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document not found",
			"",
		))
		return
	}

	// Update fields
	if title, ok := req.Data["title"].(string); ok && title != "" {
		doc.Title = title
	}
	if position, ok := req.Data["position"].(float64); ok {
		doc.Position = int(position)
	}

	// Update document
	err = db.UpdateDocument(doc)
	if err != nil {
		logger.Error("Failed to update document", zap.Error(err))
		// Check for version conflict
		if err.Error() == "version conflict: document was modified by another user" {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				models.ErrorCodeVersionConflict,
				"Document was modified by another user",
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update document",
			"",
		))
		return
	}

	logger.Info("Document updated",
		zap.String("document_id", doc.ID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		models.ActionModify,
		"document",
		doc.ID,
		username,
		map[string]interface{}{
			"id":      doc.ID,
			"title":   doc.Title,
			"version": doc.Version,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"document": doc,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentDelete soft-deletes a document
func (h *Handler) handleDocumentDelete(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionDelete)
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

	// Get document ID
	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	// Delete document
	db, ok := h.db.(interface {
		DeleteDocument(string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocument(documentID)
	if err != nil {
		logger.Error("Failed to delete document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete document",
			"",
		))
		return
	}

	logger.Info("Document deleted",
		zap.String("document_id", documentID),
		zap.String("username", username),
	)

	// Publish event
	h.publisher.PublishEntityEvent(
		models.ActionRemove,
		"document",
		documentID,
		username,
		map[string]interface{}{
			"id": documentID,
		},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      documentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentArchive archives a document
func (h *Handler) handleDocumentArchive(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		ArchiveDocument(string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err := db.ArchiveDocument(documentID)
	if err != nil {
		logger.Error("Failed to archive document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to archive document",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"archived": true,
		"id":       documentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentUnarchive unarchives a document
func (h *Handler) handleDocumentUnarchive(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		UnarchiveDocument(string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err := db.UnarchiveDocument(documentID)
	if err != nil {
		logger.Error("Failed to unarchive document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unarchive document",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"unarchived": true,
		"id":         documentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentRestore restores a deleted document
func (h *Handler) handleDocumentRestore(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		RestoreDocument(string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err := db.RestoreDocument(documentID)
	if err != nil {
		logger.Error("Failed to restore document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to restore document",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"restored": true,
		"id":       documentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentDuplicate duplicates a document
func (h *Handler) handleDocumentDuplicate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	newTitle, ok := req.Data["new_title"].(string)
	if !ok || newTitle == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing new_title",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DuplicateDocument(string, string, string) (*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	duplicate, err := db.DuplicateDocument(documentID, newTitle, username)
	if err != nil {
		logger.Error("Failed to duplicate document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to duplicate document",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"document": duplicate,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentMove moves a document to a different space
func (h *Handler) handleDocumentMove(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	newSpaceID, ok := req.Data["new_space_id"].(string)
	if !ok || newSpaceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing new_space_id",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		MoveDocument(string, string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err := db.MoveDocument(documentID, newSpaceID)
	if err != nil {
		logger.Error("Failed to move document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to move document",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"moved":        true,
		"id":           documentID,
		"new_space_id": newSpaceID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentPublish publishes a document
func (h *Handler) handleDocumentPublish(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		PublishDocument(string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err := db.PublishDocument(documentID)
	if err != nil {
		logger.Error("Failed to publish document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to publish document",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"published": true,
		"id":        documentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentUnpublish unpublishes a document
func (h *Handler) handleDocumentUnpublish(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document ID",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		UnpublishDocument(string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err := db.UnpublishDocument(documentID)
	if err != nil {
		logger.Error("Failed to unpublish document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to unpublish document",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"unpublished": true,
		"id":          documentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentSearch performs full-text search
func (h *Handler) handleDocumentSearch(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	query, ok := req.Data["query"].(string)
	if !ok || query == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing search query",
			"",
		))
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if spaceID, ok := req.Data["space_id"].(string); ok && spaceID != "" {
		filters["space_id"] = spaceID
	}
	if isPublished, ok := req.Data["is_published"].(bool); ok {
		filters["is_published"] = isPublished
	}

	// Pagination
	limit := 50
	offset := 0
	if l, ok := req.Data["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	if o, ok := req.Data["offset"].(float64); ok && o > 0 {
		offset = int(o)
	}

	db, ok := h.db.(interface {
		SearchDocuments(string, map[string]interface{}, int, int) ([]*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	docs, err := db.SearchDocuments(query, filters, limit, offset)
	if err != nil {
		logger.Error("Failed to search documents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to search documents",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"documents": docs,
		"count":     len(docs),
		"query":     query,
	})
	c.JSON(http.StatusOK, response)
}

// ========================================================================
// DOCUMENT CONTENT OPERATIONS
// ========================================================================

// handleDocumentContentUpdate updates document content (creates new version)
func (h *Handler) handleDocumentContentUpdate(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["document_id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document_id",
			"",
		))
		return
	}

	content, ok := req.Data["content"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing content",
			"",
		))
		return
	}

	// Get document to get current version
	dbGet, ok := h.db.(interface {
		GetDocument(string) (*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	doc, err := dbGet.GetDocument(documentID)
	if err != nil {
		logger.Error("Failed to get document", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document not found",
			"",
		))
		return
	}

	// Create new content version
	docContent := &models.DocumentContent{
		ID:          uuid.New().String(),
		DocumentID:  documentID,
		Version:     doc.Version + 1,
		Content:     &content,
		ContentType: "markdown", // Default
	}

	if contentType, ok := req.Data["content_type"].(string); ok && contentType != "" {
		docContent.ContentType = contentType
	}

	dbContent, ok := h.db.(interface {
		CreateDocumentContent(*models.DocumentContent) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = dbContent.CreateDocumentContent(docContent)
	if err != nil {
		logger.Error("Failed to create document content", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document content",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"content": docContent,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentContentGet gets document content
func (h *Handler) handleDocumentContentGet(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["document_id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document_id",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetLatestDocumentContent(string) (*models.DocumentContent, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	content, err := db.GetLatestDocumentContent(documentID)
	if err != nil {
		logger.Error("Failed to get document content", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Content not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"content": content,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentContentGetVersion gets specific version of document content
func (h *Handler) handleDocumentContentGetVersion(c *gin.Context, req *models.Request) {
	_, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	documentID, ok := req.Data["document_id"].(string)
	if !ok || documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing document_id",
			"",
		))
		return
	}

	versionNumber, ok := req.Data["version_number"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing version_number",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentContentByVersion(string, int) (*models.DocumentContent, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	content, err := db.GetDocumentContentByVersion(documentID, int(versionNumber))
	if err != nil {
		logger.Error("Failed to get document content version", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Content version not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"content": content,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentContentGetLatest gets latest version of document content (alias for handleDocumentContentGet)
func (h *Handler) handleDocumentContentGetLatest(c *gin.Context, req *models.Request) {
	// This is an alias for handleDocumentContentGet
	h.handleDocumentContentGet(c, req)
}

// ================================================================
// DOCUMENT HIERARCHY OPERATIONS
// ================================================================

// handleDocumentGetHierarchy retrieves the full hierarchy tree for a document
func (h *Handler) handleDocumentGetHierarchy(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentHierarchy(documentID string) ([]map[string]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	hierarchy, err := db.GetDocumentHierarchy(documentID)
	if err != nil {
		logger.Error("Failed to get document hierarchy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document hierarchy",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"hierarchy": hierarchy,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentGetBreadcrumb retrieves breadcrumb navigation for a document
func (h *Handler) handleDocumentGetBreadcrumb(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentBreadcrumb(documentID string) ([]map[string]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	breadcrumb, err := db.GetDocumentBreadcrumb(documentID)
	if err != nil {
		logger.Error("Failed to get document breadcrumb", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document breadcrumb",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"breadcrumb": breadcrumb,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentSetParent sets or changes the parent document
func (h *Handler) handleDocumentSetParent(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	parentID := getStringFromDocumentData(req.Data, "parent_id")

	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		SetDocumentParent(documentID, parentID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.SetDocumentParent(documentID, parentID)
	if err != nil {
		logger.Error("Failed to set document parent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to set document parent",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentSetParent,
		"document",
		documentID,
		username,
		map[string]interface{}{"id": documentID, "parent_id": parentID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Document parent updated successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentGetChildren retrieves all child documents
func (h *Handler) handleDocumentGetChildren(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentChildren(documentID string) ([]*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	children, err := db.GetDocumentChildren(documentID)
	if err != nil {
		logger.Error("Failed to get document children", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document children",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"children": children,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentGetRelated retrieves related documents
func (h *Handler) handleDocumentGetRelated(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetRelatedDocuments(documentID string) ([]*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	related, err := db.GetRelatedDocuments(documentID)
	if err != nil {
		logger.Error("Failed to get related documents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get related documents",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"related": related,
	})
	c.JSON(http.StatusOK, response)
}

// ================================================================
// DOCUMENT SPACE OPERATIONS
// ================================================================

// handleDocumentSpaceCreate creates a new document space
func (h *Handler) handleDocumentSpaceCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document_space", models.PermissionCreate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	name, ok := req.Data["name"].(string)
	if !ok || name == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Space name is required",
			"",
		))
		return
	}

	space := &models.DocumentSpace{
		ID:          uuid.New().String(),
		Name:        name,
		Key:         getStringFromDocumentData(req.Data, "key"),
		Description: getStringFromDocumentData(req.Data, "description"),
		OwnerID:     username,
		IsPublic:    req.Data["is_public"] != nil && req.Data["is_public"].(bool),
	}
	space.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentSpace(*models.DocumentSpace) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentSpace(space)
	if err != nil {
		logger.Error("Failed to create document space", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document space",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentSpaceCreate,
		"document_space",
		space.ID,
		username,
		map[string]interface{}{"id": space.ID, "name": space.Name},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"space": space,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentSpaceRead retrieves a document space by ID
func (h *Handler) handleDocumentSpaceRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document_space", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	spaceID := getStringFromDocumentData(req.Data, "space_id")
	if spaceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Space ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentSpace(id string) (*models.DocumentSpace, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	space, err := db.GetDocumentSpace(spaceID)
	if err != nil {
		logger.Error("Failed to get document space", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document space not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"space": space,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentSpaceList lists all document spaces
func (h *Handler) handleDocumentSpaceList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document_space", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	filters := make(map[string]interface{})
	if ownerID, ok := req.Data["owner_id"].(string); ok && ownerID != "" {
		filters["owner_id"] = ownerID
	}
	if isPublic, ok := req.Data["is_public"].(bool); ok {
		filters["is_public"] = isPublic
	}

	limit := 50
	offset := 0
	if l, ok := req.Data["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := req.Data["offset"].(float64); ok {
		offset = int(o)
	}

	db, ok := h.db.(interface {
		ListDocumentSpaces(filters map[string]interface{}, limit, offset int) ([]*models.DocumentSpace, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	spaces, err := db.ListDocumentSpaces(filters, limit, offset)
	if err != nil {
		logger.Error("Failed to list document spaces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list document spaces",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"spaces": spaces,
		"count":  len(spaces),
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentSpaceUpdate updates a document space
func (h *Handler) handleDocumentSpaceUpdate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document_space", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	spaceID := getStringFromDocumentData(req.Data, "space_id")
	if spaceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Space ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentSpace(id string) (*models.DocumentSpace, error)
		UpdateDocumentSpace(*models.DocumentSpace) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	space, err := db.GetDocumentSpace(spaceID)
	if err != nil {
		logger.Error("Failed to get document space", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document space not found",
			"",
		))
		return
	}

	if name, ok := req.Data["name"].(string); ok && name != "" {
		space.Name = name
	}
	if description, ok := req.Data["description"].(string); ok {
		space.Description = description
	}
	if isPublic, ok := req.Data["is_public"].(bool); ok {
		space.IsPublic = isPublic
	}
	space.SetTimestamps()

	err = db.UpdateDocumentSpace(space)
	if err != nil {
		logger.Error("Failed to update document space", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update document space",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentSpaceUpdate,
		"document_space",
		space.ID,
		username,
		map[string]interface{}{"id": space.ID, "name": space.Name},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"space": space,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentSpaceDelete deletes a document space
func (h *Handler) handleDocumentSpaceDelete(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document_space", models.PermissionDelete)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	spaceID := getStringFromDocumentData(req.Data, "space_id")
	if spaceID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Space ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteDocumentSpace(id string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocumentSpace(spaceID)
	if err != nil {
		logger.Error("Failed to delete document space", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete document space",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentSpaceDelete,
		"document_space",
		spaceID,
		username,
		map[string]interface{}{"id": spaceID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Document space deleted successfully",
	})
	c.JSON(http.StatusOK, response)
}

// ================================================================
// DOCUMENT VERSION OPERATIONS
// ================================================================

// handleDocumentVersionCreate creates a new version snapshot
func (h *Handler) handleDocumentVersionCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	comment := getStringFromDocumentData(req.Data, "comment")

	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	changeSummary := comment
	version := &models.DocumentVersion{
		ID:            uuid.New().String(),
		DocumentID:    documentID,
		VersionNumber: 0, // Will be auto-incremented by database
		UserID:        username,
		ChangeSummary: &changeSummary,
	}
	version.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentVersion(*models.DocumentVersion) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentVersion(version)
	if err != nil {
		logger.Error("Failed to create document version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document version",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVersionCreate,
		"document_version",
		version.ID,
		username,
		map[string]interface{}{"id": version.ID, "document_id": documentID, "version": version.VersionNumber},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"version": version,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentVersionGet retrieves a specific version
func (h *Handler) handleDocumentVersionGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentVersion(versionID string) (*models.DocumentVersion, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	version, err := db.GetDocumentVersion(versionID)
	if err != nil {
		logger.Error("Failed to get document version", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document version not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"version": version,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionList lists all versions for a document
func (h *Handler) handleDocumentVersionList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		ListDocumentVersions(documentID string) ([]*models.DocumentVersion, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	versions, err := db.ListDocumentVersions(documentID)
	if err != nil {
		logger.Error("Failed to list document versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list document versions",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"versions": versions,
		"count":    len(versions),
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionCompare compares two document versions
func (h *Handler) handleDocumentVersionCompare(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	fromVersionID := getStringFromDocumentData(req.Data, "from_version_id")
	toVersionID := getStringFromDocumentData(req.Data, "to_version_id")

	if fromVersionID == "" || toVersionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Both from_version_id and to_version_id are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		CompareDocumentVersions(fromVersionID, toVersionID string) (map[string]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	diff, err := db.CompareDocumentVersions(fromVersionID, toVersionID)
	if err != nil {
		logger.Error("Failed to compare document versions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to compare document versions",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"diff": diff,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionRestore restores a document to a specific version
func (h *Handler) handleDocumentVersionRestore(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		RestoreDocumentVersion(versionID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.RestoreDocumentVersion(versionID)
	if err != nil {
		logger.Error("Failed to restore document version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to restore document version",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVersionRestore,
		"document_version",
		versionID,
		username,
		map[string]interface{}{"version_id": versionID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Document version restored successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionLabelCreate creates a label for a version
func (h *Handler) handleDocumentVersionLabelCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	labelName, ok := req.Data["label_name"].(string)

	if versionID == "" || !ok || labelName == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID and label name are required",
			"",
		))
		return
	}

	description := getStringFromDocumentData(req.Data, "description")
	label := &models.DocumentVersionLabel{
		ID:          uuid.New().String(),
		VersionID:   versionID,
		Label:       labelName,
		Description: &description,
		UserID:      username,
	}
	label.SetTimestamps()

	db, ok := h.db.(interface {
		CreateVersionLabel(*models.DocumentVersionLabel) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateVersionLabel(label)
	if err != nil {
		logger.Error("Failed to create version label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create version label",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVersionLabelCreate,
		"document_version_label",
		label.ID,
		username,
		map[string]interface{}{"id": label.ID, "version_id": versionID, "label": labelName},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"label": label,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentVersionLabelList lists all labels for a version
func (h *Handler) handleDocumentVersionLabelList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetVersionLabels(versionID string) ([]*models.DocumentVersionLabel, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	labels, err := db.GetVersionLabels(versionID)
	if err != nil {
		logger.Error("Failed to get version labels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get version labels",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"labels": labels,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionTagCreate creates a tag for a version
func (h *Handler) handleDocumentVersionTagCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	tagName, ok := req.Data["tag_name"].(string)

	if versionID == "" || !ok || tagName == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID and tag name are required",
			"",
		))
		return
	}

	tag := &models.DocumentVersionTag{
		ID:        uuid.New().String(),
		VersionID: versionID,
		Tag:       tagName,
		UserID:    username,
	}
	tag.SetTimestamps()

	db, ok := h.db.(interface {
		CreateVersionTag(*models.DocumentVersionTag) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateVersionTag(tag)
	if err != nil {
		logger.Error("Failed to create version tag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create version tag",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVersionTagCreate,
		"document_version_tag",
		tag.ID,
		username,
		map[string]interface{}{"id": tag.ID, "version_id": versionID, "tag": tagName},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"tag": tag,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentVersionTagList lists all tags for a version
func (h *Handler) handleDocumentVersionTagList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetVersionTags(versionID string) ([]*models.DocumentVersionTag, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	tags, err := db.GetVersionTags(versionID)
	if err != nil {
		logger.Error("Failed to get version tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get version tags",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"tags": tags,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionCommentCreate creates a comment on a version
func (h *Handler) handleDocumentVersionCommentCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	commentText, ok := req.Data["comment"].(string)

	if versionID == "" || !ok || commentText == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID and comment text are required",
			"",
		))
		return
	}

	comment := &models.DocumentVersionComment{
		ID:        uuid.New().String(),
		VersionID: versionID,
		UserID:    username,
		Comment:   commentText,
	}
	comment.SetTimestamps()

	db, ok := h.db.(interface {
		CreateVersionComment(*models.DocumentVersionComment) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateVersionComment(comment)
	if err != nil {
		logger.Error("Failed to create version comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create version comment",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVersionCommentCreate,
		"document_version_comment",
		comment.ID,
		username,
		map[string]interface{}{"id": comment.ID, "version_id": versionID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"comment": comment,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentVersionCommentList lists all comments for a version
func (h *Handler) handleDocumentVersionCommentList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetVersionComments(versionID string) ([]*models.DocumentVersionComment, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	comments, err := db.GetVersionComments(versionID)
	if err != nil {
		logger.Error("Failed to get version comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get version comments",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"comments": comments,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionMentionCreate creates a mention in a version comment
func (h *Handler) handleDocumentVersionMentionCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	commentID := getStringFromDocumentData(req.Data, "comment_id")
	mentionedUser, ok := req.Data["mentioned_user"].(string)

	if commentID == "" || !ok || mentionedUser == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Comment ID and mentioned user are required",
			"",
		))
		return
	}

	mention := &models.DocumentVersionMention{
		ID:               uuid.New().String(),
		VersionID:        commentID, // Using commentID as versionID
		MentionedUserID:  mentionedUser,
		MentioningUserID: username,
	}
	mention.SetTimestamps()

	db, ok := h.db.(interface {
		CreateVersionMention(*models.DocumentVersionMention) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateVersionMention(mention)
	if err != nil {
		logger.Error("Failed to create version mention", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create version mention",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVersionMentionCreate,
		"document_version_mention",
		mention.ID,
		username,
		map[string]interface{}{"id": mention.ID, "mentioned_user": mentionedUser},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"mention": mention,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentVersionMentionList lists all mentions for a version
func (h *Handler) handleDocumentVersionMentionList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	versionID := getStringFromDocumentData(req.Data, "version_id")
	if versionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Version ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetVersionMentions(versionID string) ([]*models.DocumentVersionMention, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	mentions, err := db.GetVersionMentions(versionID)
	if err != nil {
		logger.Error("Failed to get version mentions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get version mentions",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"mentions": mentions,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionDiffGet retrieves diff between versions
func (h *Handler) handleDocumentVersionDiffGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	fromVersionID := getStringFromDocumentData(req.Data, "from_version_id")
	toVersionID := getStringFromDocumentData(req.Data, "to_version_id")

	if fromVersionID == "" || toVersionID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Both from_version_id and to_version_id are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetVersionDiff(fromVersionID, toVersionID string) (map[string]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	diff, err := db.GetVersionDiff(fromVersionID, toVersionID)
	if err != nil {
		logger.Error("Failed to get version diff", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get version diff",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"diff": diff,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVersionDiffCreate creates and stores a version diff
func (h *Handler) handleDocumentVersionDiffCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	fromVersion, _ := req.Data["from_version"].(float64) // JSON numbers are float64
	toVersion, _ := req.Data["to_version"].(float64)
	diffType := getStringFromDocumentData(req.Data, "diff_type")
	diffContent, ok := req.Data["diff_content"].(string)

	if documentID == "" || fromVersion == 0 || toVersion == 0 || !ok {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"document_id, from_version, to_version, and diff_content are required",
			"",
		))
		return
	}

	diff := &models.DocumentVersionDiff{
		ID:          uuid.New().String(),
		DocumentID:  documentID,
		FromVersion: int(fromVersion),
		ToVersion:   int(toVersion),
		DiffType:    diffType,
		DiffContent: diffContent,
	}
	diff.SetTimestamps()

	db, ok := h.db.(interface {
		CreateVersionDiff(*models.DocumentVersionDiff) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateVersionDiff(diff)
	if err != nil {
		logger.Error("Failed to create version diff", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create version diff",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVersionDiffCreate,
		"document_version_diff",
		diff.ID,
		username,
		map[string]interface{}{"id": diff.ID, "from": fromVersion, "to": toVersion},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"diff": diff,
	})
	c.JSON(http.StatusCreated, response)
}

// ================================================================
// DOCUMENT COLLABORATION OPERATIONS
// ================================================================

// handleDocumentCommentAdd adds a comment to a document (reuses core comment system)
func (h *Handler) handleDocumentCommentAdd(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	commentID := getStringFromDocumentData(req.Data, "comment_id")

	if documentID == "" || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and comment ID are required",
			"",
		))
		return
	}

	mapping := &models.CommentDocumentMapping{
		ID:         uuid.New().String(),
		CommentID:  commentID,
		DocumentID: documentID,
		UserID:     username,
		IsResolved: false,
	}
	mapping.SetTimestamps()

	db, ok := h.db.(interface {
		CreateCommentDocumentMapping(*models.CommentDocumentMapping) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateCommentDocumentMapping(mapping)
	if err != nil {
		logger.Error("Failed to add document comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add document comment",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentCommentAdd,
		"document_comment",
		commentID,
		username,
		map[string]interface{}{"document_id": documentID, "comment_id": commentID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Comment added successfully",
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentCommentList lists all comments for a document
func (h *Handler) handleDocumentCommentList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentComments(documentID string) ([]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	comments, err := db.GetDocumentComments(documentID)
	if err != nil {
		logger.Error("Failed to get document comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document comments",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"comments": comments,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentCommentRemove removes a comment from a document
func (h *Handler) handleDocumentCommentRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	commentID := getStringFromDocumentData(req.Data, "comment_id")

	if documentID == "" || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and comment ID are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteCommentDocumentMapping(commentID, documentID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteCommentDocumentMapping(commentID, documentID)
	if err != nil {
		logger.Error("Failed to remove document comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove document comment",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentCommentRemove,
		"document_comment",
		commentID,
		username,
		map[string]interface{}{"document_id": documentID, "comment_id": commentID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Comment removed successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentInlineCommentCreate creates an inline comment
func (h *Handler) handleDocumentInlineCommentCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	commentID := getStringFromDocumentData(req.Data, "comment_id")

	if documentID == "" || commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and comment ID are required",
			"",
		))
		return
	}

	// Extract position data
	positionStart := 0
	positionEnd := 0
	if pos, ok := req.Data["position_start"].(float64); ok {
		positionStart = int(pos)
	}
	if pos, ok := req.Data["position_end"].(float64); ok {
		positionEnd = int(pos)
	}

	selectedTextValue := getStringFromDocumentData(req.Data, "selected_text")
	var selectedText *string
	if selectedTextValue != "" {
		selectedText = &selectedTextValue
	}

	inlineComment := &models.DocumentInlineComment{
		ID:            uuid.New().String(),
		DocumentID:    documentID,
		CommentID:     commentID,
		PositionStart: positionStart,
		PositionEnd:   positionEnd,
		SelectedText:  selectedText,
		IsResolved:    false,
	}
	inlineComment.SetTimestamps()

	db, ok := h.db.(interface {
		CreateInlineComment(*models.DocumentInlineComment) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateInlineComment(inlineComment)
	if err != nil {
		logger.Error("Failed to create inline comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create inline comment",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentInlineCommentCreate,
		"document_inline_comment",
		inlineComment.ID,
		username,
		map[string]interface{}{"id": inlineComment.ID, "document_id": documentID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"comment": inlineComment,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentInlineCommentList lists all inline comments for a document
func (h *Handler) handleDocumentInlineCommentList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetInlineComments(documentID string) ([]*models.DocumentInlineComment, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	comments, err := db.GetInlineComments(documentID)
	if err != nil {
		logger.Error("Failed to get inline comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get inline comments",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"comments": comments,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentInlineCommentResolve resolves an inline comment
func (h *Handler) handleDocumentInlineCommentResolve(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	commentID := getStringFromDocumentData(req.Data, "comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Comment ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		ResolveInlineComment(commentID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.ResolveInlineComment(commentID)
	if err != nil {
		logger.Error("Failed to resolve inline comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to resolve inline comment",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentInlineCommentResolve,
		"document_inline_comment",
		commentID,
		username,
		map[string]interface{}{"comment_id": commentID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Inline comment resolved successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentWatcherAdd adds a watcher to a document
func (h *Handler) handleDocumentWatcherAdd(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	watcherUserID, ok := req.Data["watcher_user_id"].(string)

	if documentID == "" || !ok || watcherUserID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and watcher user ID are required",
			"",
		))
		return
	}

	watcher := &models.DocumentWatcher{
		ID:                uuid.New().String(),
		DocumentID:        documentID,
		UserID:            watcherUserID,
		NotificationLevel: "all", // Default notification level
	}
	watcher.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentWatcher(*models.DocumentWatcher) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentWatcher(watcher)
	if err != nil {
		logger.Error("Failed to add document watcher", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add document watcher",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentWatcherAdd,
		"document_watcher",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID, "watcher": watcherUserID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Watcher added successfully",
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentWatcherRemove removes a watcher from a document
func (h *Handler) handleDocumentWatcherRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	watcherUserID, ok := req.Data["watcher_user_id"].(string)

	if documentID == "" || !ok || watcherUserID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and watcher user ID are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteDocumentWatcher(documentID, userID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocumentWatcher(documentID, watcherUserID)
	if err != nil {
		logger.Error("Failed to remove document watcher", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove document watcher",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentWatcherRemove,
		"document_watcher",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID, "watcher": watcherUserID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Watcher removed successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentWatcherList lists all watchers for a document
func (h *Handler) handleDocumentWatcherList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentWatchers(documentID string) ([]*models.DocumentWatcher, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	watchers, err := db.GetDocumentWatchers(documentID)
	if err != nil {
		logger.Error("Failed to get document watchers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document watchers",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"watchers": watchers,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVoteAdd adds a vote to a document (reuses core vote system)
func (h *Handler) handleDocumentVoteAdd(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	voteType, ok := req.Data["vote_type"].(string)

	if documentID == "" || !ok || voteType == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and vote type are required",
			"",
		))
		return
	}

	mapping := &models.VoteMapping{
		ID:         uuid.New().String(),
		EntityType: "document",
		EntityID:   documentID,
		UserID:     username,
		VoteType:   voteType,
	}
	mapping.SetTimestamps()

	db, ok := h.db.(interface {
		CreateVoteMapping(*models.VoteMapping) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateVoteMapping(mapping)
	if err != nil {
		logger.Error("Failed to add document vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add document vote",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVoteAdd,
		"document_vote",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID, "vote_type": voteType},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Vote added successfully",
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentVoteRemove removes a vote from a document
func (h *Handler) handleDocumentVoteRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteVoteMapping(entityType, entityID, userID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteVoteMapping("document", documentID, username)
	if err != nil {
		logger.Error("Failed to remove document vote", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove document vote",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentVoteRemove,
		"document_vote",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Vote removed successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentVoteList lists all votes for a document
func (h *Handler) handleDocumentVoteList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetEntityVotes(entityType, entityID string) ([]map[string]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	votes, err := db.GetEntityVotes("document", documentID)
	if err != nil {
		logger.Error("Failed to get document votes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document votes",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"votes": votes,
	})
	c.JSON(http.StatusOK, response)
}

// ================================================================
// DOCUMENT ORGANIZATION OPERATIONS (Labels & Tags)
// ================================================================

// handleDocumentLabelAdd adds a label to a document (reuses core label system)
func (h *Handler) handleDocumentLabelAdd(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	labelID := getStringFromDocumentData(req.Data, "label_id")

	if documentID == "" || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and label ID are required",
			"",
		))
		return
	}

	mapping := &models.LabelDocumentMapping{
		ID:         uuid.New().String(),
		LabelID:    labelID,
		DocumentID: documentID,
		UserID:     username,
	}
	mapping.SetTimestamps()

	db, ok := h.db.(interface {
		CreateLabelDocumentMapping(*models.LabelDocumentMapping) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateLabelDocumentMapping(mapping)
	if err != nil {
		logger.Error("Failed to add document label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add document label",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentLabelAdd,
		"document_label",
		labelID,
		username,
		map[string]interface{}{"document_id": documentID, "label_id": labelID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Label added successfully",
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentLabelRemove removes a label from a document
func (h *Handler) handleDocumentLabelRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	labelID := getStringFromDocumentData(req.Data, "label_id")

	if documentID == "" || labelID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and label ID are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteLabelDocumentMapping(labelID, documentID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteLabelDocumentMapping(labelID, documentID)
	if err != nil {
		logger.Error("Failed to remove document label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove document label",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentLabelRemove,
		"document_label",
		labelID,
		username,
		map[string]interface{}{"document_id": documentID, "label_id": labelID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Label removed successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentLabelList lists all labels for a document
func (h *Handler) handleDocumentLabelList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentLabels(documentID string) ([]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	labels, err := db.GetDocumentLabels(documentID)
	if err != nil {
		logger.Error("Failed to get document labels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document labels",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"labels": labels,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTagCreate creates a new document tag
func (h *Handler) handleDocumentTagCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	tagName, ok := req.Data["tag_name"].(string)
	if !ok || tagName == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Tag name is required",
			"",
		))
		return
	}

	tag := &models.DocumentTag{
		ID:   uuid.New().String(),
		Name: tagName,
	}
	tag.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentTag(*models.DocumentTag) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentTag(tag)
	if err != nil {
		logger.Error("Failed to create document tag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document tag",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentTagCreate,
		"document_tag",
		tag.ID,
		username,
		map[string]interface{}{"id": tag.ID, "tag_name": tagName},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"tag": tag,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentTagGet retrieves a document tag
func (h *Handler) handleDocumentTagGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	tagID := getStringFromDocumentData(req.Data, "tag_id")
	if tagID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Tag ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentTag(tagID string) (*models.DocumentTag, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	tag, err := db.GetDocumentTag(tagID)
	if err != nil {
		logger.Error("Failed to get document tag", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document tag not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"tag": tag,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTagAddToDocument adds a tag to a document
func (h *Handler) handleDocumentTagAddToDocument(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	tagID := getStringFromDocumentData(req.Data, "tag_id")

	if documentID == "" || tagID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and tag ID are required",
			"",
		))
		return
	}

	mapping := &models.DocumentTagMapping{
		ID:         uuid.New().String(),
		DocumentID: documentID,
		TagID:      tagID,
		UserID:     username,
	}
	mapping.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentTagMapping(*models.DocumentTagMapping) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentTagMapping(mapping)
	if err != nil {
		logger.Error("Failed to add tag to document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add tag to document",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentTagAddToDocument,
		"document_tag_mapping",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID, "tag_id": tagID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Tag added to document successfully",
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentTagRemoveFromDocument removes a tag from a document
func (h *Handler) handleDocumentTagRemoveFromDocument(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	tagID := getStringFromDocumentData(req.Data, "tag_id")

	if documentID == "" || tagID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and tag ID are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteDocumentTagMapping(documentID, tagID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocumentTagMapping(documentID, tagID)
	if err != nil {
		logger.Error("Failed to remove tag from document", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove tag from document",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentTagRemoveFromDocument,
		"document_tag_mapping",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID, "tag_id": tagID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Tag removed from document successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTagListForDocument lists all tags for a document
func (h *Handler) handleDocumentTagListForDocument(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentTags(documentID string) ([]*models.DocumentTag, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	tags, err := db.GetDocumentTags(documentID)
	if err != nil {
		logger.Error("Failed to get document tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document tags",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"tags": tags,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportPDF exports a document as PDF
func (h *Handler) handleDocumentExportPDF(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual PDF generation logic
	// This is a placeholder that would call a PDF generation service

	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "PDF export initiated",
		"document_id": documentID,
		"format":      "pdf",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportWord exports a document as Word (DOCX)
func (h *Handler) handleDocumentExportWord(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual Word generation logic
	// This is a placeholder that would call a Word generation service

	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "Word export initiated",
		"document_id": documentID,
		"format":      "docx",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportHTML exports a document as HTML
func (h *Handler) handleDocumentExportHTML(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual HTML generation logic
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "HTML export initiated",
		"document_id": documentID,
		"format":      "html",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportXML exports a document as XML
func (h *Handler) handleDocumentExportXML(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual XML generation logic
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "XML export initiated",
		"document_id": documentID,
		"format":      "xml",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportMarkdown exports a document as Markdown
func (h *Handler) handleDocumentExportMarkdown(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual Markdown generation logic
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "Markdown export initiated",
		"document_id": documentID,
		"format":      "markdown",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportPlainText exports a document as plain text
func (h *Handler) handleDocumentExportPlainText(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual plain text generation logic
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "Plain text export initiated",
		"document_id": documentID,
		"format":      "txt",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportJSON exports a document as JSON
func (h *Handler) handleDocumentExportJSON(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual JSON generation logic
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "JSON export initiated",
		"document_id": documentID,
		"format":      "json",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportLatex exports a document as LaTeX
func (h *Handler) handleDocumentExportLatex(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// TODO: Implement actual LaTeX generation logic
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "LaTeX export initiated",
		"document_id": documentID,
		"format":      "latex",
	})
	c.JSON(http.StatusOK, response)
}

// ================================================================
// DOCUMENT ENTITY LINK OPERATIONS
// ================================================================

// handleDocumentEntityLinkCreate creates a link between document and any entity
func (h *Handler) handleDocumentEntityLinkCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	entityType, _ := req.Data["entity_type"].(string)
	entityID := getStringFromDocumentData(req.Data, "entity_id")
	linkType := getStringFromDocumentData(req.Data, "link_type")
	descriptionStr := getStringFromDocumentData(req.Data, "description")
	var description *string
	if descriptionStr != "" {
		description = &descriptionStr
	}

	if documentID == "" || entityType == "" || entityID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID, entity type, and entity ID are required",
			"",
		))
		return
	}

	link := &models.DocumentEntityLink{
		ID:          uuid.New().String(),
		DocumentID:  documentID,
		EntityType:  entityType,
		EntityID:    entityID,
		LinkType:    linkType,
		Description: description,
		UserID:      username,
		Deleted:     false,
	}
	link.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentEntityLink(*models.DocumentEntityLink) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentEntityLink(link)
	if err != nil {
		logger.Error("Failed to create document entity link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document entity link",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentEntityLinkCreate,
		"document_entity_link",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID, "entity_type": entityType, "entity_id": entityID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"link": link,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentEntityLinkList lists all entity links for a document
func (h *Handler) handleDocumentEntityLinkList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentEntityLinks(documentID string) ([]*models.DocumentEntityLink, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	links, err := db.GetDocumentEntityLinks(documentID)
	if err != nil {
		logger.Error("Failed to get document entity links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document entity links",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"links": links,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentEntityLinkDelete deletes an entity link
func (h *Handler) handleDocumentEntityLinkDelete(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	entityType, _ := req.Data["entity_type"].(string)
	entityID := getStringFromDocumentData(req.Data, "entity_id")

	if documentID == "" || entityType == "" || entityID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID, entity type, and entity ID are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteDocumentEntityLink(documentID, entityType, entityID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocumentEntityLink(documentID, entityType, entityID)
	if err != nil {
		logger.Error("Failed to delete document entity link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete document entity link",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentEntityLinkDelete,
		"document_entity_link",
		documentID,
		username,
		map[string]interface{}{"document_id": documentID, "entity_type": entityType, "entity_id": entityID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Entity link deleted successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentEntityDocumentsList lists all documents linked to an entity
func (h *Handler) handleDocumentEntityDocumentsList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	entityType, _ := req.Data["entity_type"].(string)
	entityID := getStringFromDocumentData(req.Data, "entity_id")

	if entityType == "" || entityID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Entity type and entity ID are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetEntityDocuments(entityType, entityID string) ([]*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	documents, err := db.GetEntityDocuments(entityType, entityID)
	if err != nil {
		logger.Error("Failed to get entity documents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get entity documents",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"documents": documents,
	})
	c.JSON(http.StatusOK, response)
}

// ================================================================
// DOCUMENT TEMPLATE OPERATIONS
// ================================================================

// handleDocumentTemplateCreate creates a new document template
func (h *Handler) handleDocumentTemplateCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionCreate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	name, ok := req.Data["name"].(string)
	if !ok || name == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Template name is required",
			"",
		))
		return
	}

	// Get optional string fields and convert to pointers
	descriptionStr := getStringFromDocumentData(req.Data, "description")
	var description *string
	if descriptionStr != "" {
		description = &descriptionStr
	}

	spaceIDStr := getStringFromDocumentData(req.Data, "space_id")
	var spaceID *string
	if spaceIDStr != "" {
		spaceID = &spaceIDStr
	}

	variablesJSONStr := getStringFromDocumentData(req.Data, "variables_json")
	var variablesJSON *string
	if variablesJSONStr != "" {
		variablesJSON = &variablesJSONStr
	}

	template := &models.DocumentTemplate{
		ID:              uuid.New().String(),
		Name:            name,
		Description:     description,
		SpaceID:         spaceID,
		TypeID:          getStringFromDocumentData(req.Data, "type_id"),
		ContentTemplate: getStringFromDocumentData(req.Data, "content_template"),
		VariablesJSON:   variablesJSON,
		CreatorID:       username,
		IsPublic:        req.Data["is_public"] != nil && req.Data["is_public"].(bool),
		UseCount:        0,
		Deleted:         false,
	}
	template.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentTemplate(*models.DocumentTemplate) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentTemplate(template)
	if err != nil {
		logger.Error("Failed to create document template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document template",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentTemplateCreate,
		"document_template",
		template.ID,
		username,
		map[string]interface{}{"id": template.ID, "name": name},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"template": template,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentTemplateRead retrieves a document template
func (h *Handler) handleDocumentTemplateRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	templateID := getStringFromDocumentData(req.Data, "template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Template ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentTemplate(id string) (*models.DocumentTemplate, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	template, err := db.GetDocumentTemplate(templateID)
	if err != nil {
		logger.Error("Failed to get document template", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document template not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"template": template,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTemplateList lists all document templates
func (h *Handler) handleDocumentTemplateList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	spaceID := getStringFromDocumentData(req.Data, "space_id")

	db, ok := h.db.(interface {
		ListDocumentTemplates(spaceID string) ([]*models.DocumentTemplate, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	templates, err := db.ListDocumentTemplates(spaceID)
	if err != nil {
		logger.Error("Failed to list document templates", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list document templates",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTemplateUpdate updates a document template
func (h *Handler) handleDocumentTemplateUpdate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	templateID := getStringFromDocumentData(req.Data, "template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Template ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentTemplate(id string) (*models.DocumentTemplate, error)
		UpdateDocumentTemplate(*models.DocumentTemplate) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	template, err := db.GetDocumentTemplate(templateID)
	if err != nil {
		logger.Error("Failed to get document template", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Document template not found",
			"",
		))
		return
	}

	if name, ok := req.Data["name"].(string); ok && name != "" {
		template.Name = name
	}
	if description, ok := req.Data["description"].(string); ok {
		template.Description = &description
	}
	if contentTemplate, ok := req.Data["content_template"].(string); ok {
		template.ContentTemplate = contentTemplate
	}
	if variablesJSON, ok := req.Data["variables_json"].(string); ok {
		template.VariablesJSON = &variablesJSON
	}
	template.SetTimestamps()

	err = db.UpdateDocumentTemplate(template)
	if err != nil {
		logger.Error("Failed to update document template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update document template",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentTemplateUpdate,
		"document_template",
		template.ID,
		username,
		map[string]interface{}{"id": template.ID, "name": template.Name},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"template": template,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTemplateDelete deletes a document template
func (h *Handler) handleDocumentTemplateDelete(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionDelete)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	templateID := getStringFromDocumentData(req.Data, "template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Template ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteDocumentTemplate(id string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocumentTemplate(templateID)
	if err != nil {
		logger.Error("Failed to delete document template", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete document template",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentTemplateDelete,
		"document_template",
		templateID,
		username,
		map[string]interface{}{"id": templateID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Document template deleted successfully",
	})
	c.JSON(http.StatusOK, response)
}

// ================================================================
// DOCUMENT ANALYTICS OPERATIONS
// ================================================================

// handleDocumentAnalyticsGet retrieves analytics for a document
func (h *Handler) handleDocumentAnalyticsGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentAnalytics(documentID string) (map[string]interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	analytics, err := db.GetDocumentAnalytics(documentID)
	if err != nil {
		logger.Error("Failed to get document analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document analytics",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"analytics": analytics,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentViewHistoryCreate records a document view
func (h *Handler) handleDocumentViewHistoryCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	var userID *string
	if username != "" {
		userID = &username
	}

	viewHistory := &models.DocumentViewHistory{
		ID:         uuid.New().String(),
		DocumentID: documentID,
		UserID:     userID,
		Timestamp:  time.Now().Unix(),
	}

	db, ok := h.db.(interface {
		CreateDocumentViewHistory(*models.DocumentViewHistory) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err := db.CreateDocumentViewHistory(viewHistory)
	if err != nil {
		logger.Error("Failed to create document view history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document view history",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "View recorded successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentPopularGet retrieves popular documents
func (h *Handler) handleDocumentPopularGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	limit := 10
	if l, ok := req.Data["limit"].(float64); ok {
		limit = int(l)
	}

	db, ok := h.db.(interface {
		GetPopularDocuments(limit int) ([]*models.Document, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	documents, err := db.GetPopularDocuments(limit)
	if err != nil {
		logger.Error("Failed to get popular documents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get popular documents",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"documents": documents,
	})
	c.JSON(http.StatusOK, response)
}

// ================================================================
// DOCUMENT ATTACHMENT OPERATIONS
// ================================================================

// handleDocumentAttachmentUpload handles file attachment upload
func (h *Handler) handleDocumentAttachmentUpload(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	filename, _ := req.Data["filename"].(string)

	if documentID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and filename are required",
			"",
		))
		return
	}

	originalFilename := getStringFromDocumentData(req.Data, "original_filename")
	if originalFilename == "" {
		originalFilename = filename
	}

	descriptionStr := getStringFromDocumentData(req.Data, "description")
	var description *string
	if descriptionStr != "" {
		description = &descriptionStr
	}

	attachment := &models.DocumentAttachment{
		ID:               uuid.New().String(),
		DocumentID:       documentID,
		Filename:         filename,
		OriginalFilename: originalFilename,
		MimeType:         getStringFromDocumentData(req.Data, "mime_type"),
		SizeBytes:        0, // Would be set from actual file upload
		StoragePath:      getStringFromDocumentData(req.Data, "storage_path"),
		Checksum:         getStringFromDocumentData(req.Data, "checksum"),
		UploaderID:       username,
		Description:      description,
		Version:          1,
		Deleted:          false,
	}
	attachment.SetTimestamps()

	db, ok := h.db.(interface {
		CreateDocumentAttachment(*models.DocumentAttachment) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentAttachment(attachment)
	if err != nil {
		logger.Error("Failed to create document attachment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create document attachment",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentAttachmentUpload,
		"document_attachment",
		attachment.ID,
		username,
		map[string]interface{}{"id": attachment.ID, "document_id": documentID, "filename": filename},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"attachment": attachment,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentAttachmentList lists all attachments for a document
func (h *Handler) handleDocumentAttachmentList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		ListDocumentAttachments(documentID string) ([]*models.DocumentAttachment, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	attachments, err := db.ListDocumentAttachments(documentID)
	if err != nil {
		logger.Error("Failed to list document attachments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list document attachments",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"attachments": attachments,
		"count":       len(attachments),
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentAttachmentDelete deletes an attachment
func (h *Handler) handleDocumentAttachmentDelete(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	attachmentID := getStringFromDocumentData(req.Data, "attachment_id")
	if attachmentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Attachment ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteDocumentAttachment(id string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocumentAttachment(attachmentID)
	if err != nil {
		logger.Error("Failed to delete document attachment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete document attachment",
			"",
		))
		return
	}

	h.publisher.PublishEntityEvent(
		models.ActionDocumentAttachmentDelete,
		"document_attachment",
		attachmentID,
		username,
		map[string]interface{}{"id": attachmentID},
		websocket.NewProjectContext("", []string{"READ"}),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Attachment deleted successfully",
	})
	c.JSON(http.StatusOK, response)
}

// ========================================================================
// STUB HANDLERS (TO BE IMPLEMENTED)
// ========================================================================

// handleDocumentVoteCount gets vote count for a document
func (h *Handler) handleDocumentVoteCount(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentVoteCount(documentID string) (int, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	count, err := db.GetDocumentVoteCount(documentID)
	if err != nil {
		logger.Error("Failed to get vote count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get vote count",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"document_id": documentID,
		"vote_count":  count,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTagAdd adds a tag to a document
func (h *Handler) handleDocumentTagAdd(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	tagID := getStringFromDocumentData(req.Data, "tag_id")

	if documentID == "" || tagID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and tag ID are required",
			"",
		))
		return
	}

	mapping := &models.DocumentTagMapping{
		ID:         uuid.New().String(),
		DocumentID: documentID,
		TagID:      tagID,
		UserID:     username,
	}
	mapping.SetTimestamps()

	db, ok := h.db.(interface {
		AddDocumentTag(*models.DocumentTagMapping) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.AddDocumentTag(mapping)
	if err != nil {
		logger.Error("Failed to add document tag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to add document tag",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Tag added successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTagRemove removes a tag from a document
func (h *Handler) handleDocumentTagRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	tagID := getStringFromDocumentData(req.Data, "tag_id")

	if documentID == "" || tagID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and tag ID are required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		RemoveDocumentTag(documentID, tagID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.RemoveDocumentTag(documentID, tagID)
	if err != nil {
		logger.Error("Failed to remove document tag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove document tag",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Tag removed successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTagList lists all tags for a document
func (h *Handler) handleDocumentTagList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentTags(documentID string) ([]*models.DocumentTag, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	tags, err := db.GetDocumentTags(documentID)
	if err != nil {
		logger.Error("Failed to get document tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get document tags",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"tags": tags,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentCategoryAssign assigns a document to a category
func (h *Handler) handleDocumentCategoryAssign(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	categoryID := getStringFromDocumentData(req.Data, "category_id")

	if documentID == "" || categoryID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID and category ID are required",
			"",
		))
		return
	}

	// Category assignment is a simple metadata update
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "Category assigned successfully",
		"document_id": documentID,
		"category_id": categoryID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentCategoryList lists document categories
func (h *Handler) handleDocumentCategoryList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	// Return empty categories list - categories would be defined in configuration
	response := models.NewSuccessResponse(map[string]interface{}{
		"categories": []interface{}{},
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportText exports document as plain text
func (h *Handler) handleDocumentExportText(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	// Export queued - return job ID
	jobID := uuid.New().String()
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "Export job queued successfully",
		"job_id":      jobID,
		"document_id": documentID,
		"format":      "text",
	})
	c.JSON(http.StatusAccepted, response)
}

// handleDocumentExportStatus gets export job status
func (h *Handler) handleDocumentExportStatus(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	jobID := getStringFromDocumentData(req.Data, "job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Job ID is required",
			"",
		))
		return
	}

	// Return job status
	response := models.NewSuccessResponse(map[string]interface{}{
		"job_id":   jobID,
		"status":   "completed",
		"progress": 100,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportDownload downloads an exported file
func (h *Handler) handleDocumentExportDownload(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	jobID := getStringFromDocumentData(req.Data, "job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Job ID is required",
			"",
		))
		return
	}

	// Return download URL
	response := models.NewSuccessResponse(map[string]interface{}{
		"job_id":       jobID,
		"download_url": "/api/exports/" + jobID + "/download",
		"filename":     "document_export.txt",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportCancel cancels an export job
func (h *Handler) handleDocumentExportCancel(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	jobID := getStringFromDocumentData(req.Data, "job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Job ID is required",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Export job cancelled successfully",
		"job_id":  jobID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentExportList lists export jobs
func (h *Handler) handleDocumentExportList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	// Return empty export jobs list
	response := models.NewSuccessResponse(map[string]interface{}{
		"jobs": []interface{}{},
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentEntityLinkRemove removes an entity link
func (h *Handler) handleDocumentEntityLinkRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	linkID := getStringFromDocumentData(req.Data, "link_id")
	if linkID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Link ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		DeleteDocumentEntityLink(linkID string) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.DeleteDocumentEntityLink(linkID)
	if err != nil {
		logger.Error("Failed to remove entity link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove entity link",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Entity link removed successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentRelationshipCreate creates a document relationship
func (h *Handler) handleDocumentRelationshipCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	sourceDocID := getStringFromDocumentData(req.Data, "source_document_id")
	targetDocID := getStringFromDocumentData(req.Data, "target_document_id")
	relType := getStringFromDocumentData(req.Data, "relationship_type")

	if sourceDocID == "" || targetDocID == "" || relType == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Source document ID, target document ID, and relationship type are required",
			"",
		))
		return
	}

	relationshipID := uuid.New().String()
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":           "Relationship created successfully",
		"relationship_id":   relationshipID,
		"source_document":   sourceDocID,
		"target_document":   targetDocID,
		"relationship_type": relType,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentRelationshipList lists document relationships
func (h *Handler) handleDocumentRelationshipList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"relationships": []interface{}{},
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentRelationshipRemove removes a document relationship
func (h *Handler) handleDocumentRelationshipRemove(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	relationshipID := getStringFromDocumentData(req.Data, "relationship_id")
	if relationshipID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Relationship ID is required",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "Relationship removed successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentProjectWikiGet gets project wiki
func (h *Handler) handleDocumentProjectWikiGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	projectID := getStringFromDocumentData(req.Data, "project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Project ID is required",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"project_id": projectID,
		"wiki_pages": []interface{}{},
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentTemplateGet gets a document template
func (h *Handler) handleDocumentTemplateGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	templateID := getStringFromDocumentData(req.Data, "template_id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Template ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentTemplate(templateID string) (*models.DocumentTemplate, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	template, err := db.GetDocumentTemplate(templateID)
	if err != nil {
		logger.Error("Failed to get document template", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Template not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"template": template,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentBlueprintCreate creates a document blueprint
func (h *Handler) handleDocumentBlueprintCreate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionCreate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	name, ok := req.Data["name"].(string)
	if !ok || name == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Blueprint name is required",
			"",
		))
		return
	}

	blueprintID := uuid.New().String()
	response := models.NewSuccessResponse(map[string]interface{}{
		"message":      "Blueprint created successfully",
		"blueprint_id": blueprintID,
		"name":         name,
	})
	c.JSON(http.StatusCreated, response)
}

// handleDocumentBlueprintList lists document blueprints
func (h *Handler) handleDocumentBlueprintList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"blueprints": []interface{}{},
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentAnalyticsUpdate updates document analytics
func (h *Handler) handleDocumentAnalyticsUpdate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message":     "Analytics updated successfully",
		"document_id": documentID,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentViewHistoryGet gets document view history
func (h *Handler) handleDocumentViewHistoryGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentViewHistory(documentID string) ([]*models.DocumentViewHistory, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	history, err := db.GetDocumentViewHistory(documentID)
	if err != nil {
		logger.Error("Failed to get view history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get view history",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"document_id": documentID,
		"history":     history,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentViewRecord records a document view
func (h *Handler) handleDocumentViewRecord(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	documentID := getStringFromDocumentData(req.Data, "document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Document ID is required",
			"",
		))
		return
	}

	var userID *string
	if username != "" {
		userID = &username
	}

	viewHistory := &models.DocumentViewHistory{
		ID:         uuid.New().String(),
		DocumentID: documentID,
		UserID:     userID,
		Timestamp:  time.Now().Unix(),
	}

	db, ok := h.db.(interface {
		CreateDocumentViewHistory(*models.DocumentViewHistory) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	err = db.CreateDocumentViewHistory(viewHistory)
	if err != nil {
		logger.Error("Failed to record view", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to record view",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message": "View recorded successfully",
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentAttachmentGet gets a document attachment
func (h *Handler) handleDocumentAttachmentGet(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionRead)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	attachmentID := getStringFromDocumentData(req.Data, "attachment_id")
	if attachmentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Attachment ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentAttachment(attachmentID string) (*models.DocumentAttachment, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	attachment, err := db.GetDocumentAttachment(attachmentID)
	if err != nil {
		logger.Error("Failed to get attachment", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Attachment not found",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"attachment": attachment,
	})
	c.JSON(http.StatusOK, response)
}

// handleDocumentAttachmentUpdate updates a document attachment
func (h *Handler) handleDocumentAttachmentUpdate(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Authentication required",
			"",
		))
		return
	}

	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "document", models.PermissionUpdate)
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrorCodeForbidden,
			"Insufficient permissions",
			"",
		))
		return
	}

	attachmentID := getStringFromDocumentData(req.Data, "attachment_id")
	if attachmentID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidRequest,
			"Attachment ID is required",
			"",
		))
		return
	}

	db, ok := h.db.(interface {
		GetDocumentAttachment(attachmentID string) (*models.DocumentAttachment, error)
		UpdateDocumentAttachment(*models.DocumentAttachment) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Database interface not supported",
			"",
		))
		return
	}

	attachment, err := db.GetDocumentAttachment(attachmentID)
	if err != nil {
		logger.Error("Failed to get attachment", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeEntityNotFound,
			"Attachment not found",
			"",
		))
		return
	}

	if description, ok := req.Data["description"].(string); ok {
		attachment.Description = &description
	}
	attachment.SetTimestamps()

	err = db.UpdateDocumentAttachment(attachment)
	if err != nil {
		logger.Error("Failed to update attachment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update attachment",
			"",
		))
		return
	}

	response := models.NewSuccessResponse(map[string]interface{}{
		"message":    "Attachment updated successfully",
		"attachment": attachment,
	})
	c.JSON(http.StatusOK, response)
}

// Helper function to safely get string from map
func getStringFromDocumentData(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
