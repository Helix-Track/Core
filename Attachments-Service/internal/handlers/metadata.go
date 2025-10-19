package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/storage/deduplication"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

// MetadataHandler handles metadata operations
type MetadataHandler struct {
	db                  database.Database
	deduplicationEngine deduplication.DeduplicationEngine
	metrics             utils.MetricsRecorder
	logger              *zap.Logger
}

// NewMetadataHandler creates a new metadata handler
func NewMetadataHandler(
	db database.Database,
	engine deduplication.DeduplicationEngine,
	metrics utils.MetricsRecorder,
	logger *zap.Logger,
) *MetadataHandler {
	return &MetadataHandler{
		db:                  db,
		deduplicationEngine: engine,
		metrics:             metrics,
		logger:              logger,
	}
}

// ListByEntity lists all attachments for an entity
func (h *MetadataHandler) ListByEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityID := c.Param("entity_id")

	if entityType == "" || entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "entity_type and entity_id are required",
		})
		return
	}

	// Get references for entity
	references, err := h.db.ListReferencesByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		h.logger.Error("failed to list references",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list attachments",
		})
		return
	}

	// Build response
	attachments := make([]gin.H, 0, len(references))
	for _, ref := range references {
		// Get file info
		file, err := h.db.GetFile(c.Request.Context(), ref.FileHash)
		if err != nil {
			h.logger.Warn("file not found for reference",
				zap.String("reference_id", ref.ID),
				zap.String("file_hash", ref.FileHash),
			)
			continue
		}

		attachments = append(attachments, gin.H{
			"reference_id": ref.ID,
			"filename":     ref.Filename,
			"size_bytes":   file.SizeBytes,
			"mime_type":    file.MimeType,
			"created_at":   ref.Created,
			"uploaded_by":  ref.UploaderID,
			"tags":         ref.Tags,
			"description":  ref.Description,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_type":  entityType,
		"entity_id":    entityID,
		"attachments":  attachments,
		"total_count":  len(attachments),
	})
}

// Delete deletes an attachment reference
func (h *MetadataHandler) Delete(c *gin.Context) {
	referenceID := c.Param("reference_id")
	if referenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "reference_id is required",
		})
		return
	}

	// Check ownership (optional: add permission check here)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	h.logger.Info("deleting attachment",
		zap.String("reference_id", referenceID),
		zap.String("user_id", userID.(string)),
	)

	// Delete reference
	err := h.deduplicationEngine.DeleteReference(c.Request.Context(), referenceID)
	if err != nil {
		h.logger.Error("failed to delete reference",
			zap.String("reference_id", referenceID),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete attachment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attachment deleted successfully",
		"reference_id": referenceID,
	})
}

// Update updates attachment metadata (tags, description)
func (h *MetadataHandler) Update(c *gin.Context) {
	referenceID := c.Param("reference_id")
	if referenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "reference_id is required",
		})
		return
	}

	var req struct {
		Description *string  `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Get existing reference
	reference, err := h.db.GetReference(c.Request.Context(), referenceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "reference not found",
		})
		return
	}

	// Update fields
	if req.Description != nil {
		reference.Description = req.Description
	}

	if req.Tags != nil {
		reference.Tags = req.Tags
	}

	// Save changes
	err = h.db.UpdateReference(c.Request.Context(), reference)
	if err != nil {
		h.logger.Error("failed to update reference",
			zap.String("reference_id", referenceID),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update attachment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attachment updated successfully",
		"reference_id": referenceID,
	})
}

// GetStats returns attachment statistics
func (h *MetadataHandler) GetStats(c *gin.Context) {
	stats, err := h.deduplicationEngine.GetDeduplicationStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get stats",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_files":        stats.TotalFiles,
		"total_references":   stats.TotalReferences,
		"unique_files":       stats.UniqueFiles,
		"shared_files":       stats.SharedFiles,
		"deduplication_rate": stats.DeduplicationRate,
		"saved_files":        stats.SavedFiles,
	})
}

// Search searches attachments by various criteria
func (h *MetadataHandler) Search(c *gin.Context) {
	// Get query parameters
	filename := c.Query("filename")
	mimeType := c.Query("mime_type")
	uploaderID := c.Query("uploader_id")
	tag := c.Query("tag")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Build filter
	filter := &database.ReferenceFilter{
		UploaderID: uploaderID,
		Limit:      limit,
		Offset:     offset,
	}

	// Add tag filter if specified
	if tag != "" {
		filter.Tags = []string{tag}
	}

	// Search references
	references, total, err := h.db.ListReferences(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("search failed",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "search failed",
		})
		return
	}

	// Build results
	results := make([]gin.H, 0, len(references))
	for _, ref := range references {
		// Get file info
		file, err := h.db.GetFile(c.Request.Context(), ref.FileHash)
		if err != nil {
			continue
		}

		// Apply filename filter if specified
		if filename != "" && !strings.Contains(strings.ToLower(ref.Filename), strings.ToLower(filename)) {
			continue
		}

		// Apply MIME type filter if specified
		if mimeType != "" && file.MimeType != mimeType {
			continue
		}

		results = append(results, gin.H{
			"reference_id": ref.ID,
			"filename":     ref.Filename,
			"size_bytes":   file.SizeBytes,
			"mime_type":    file.MimeType,
			"entity_type":  ref.EntityType,
			"entity_id":    ref.EntityID,
			"created_at":   ref.Created,
			"uploaded_by":  ref.UploaderID,
			"tags":         ref.Tags,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"results":      results,
		"total_count":  total,
		"limit":        limit,
		"offset":       offset,
	})
}

// GetByHash gets all references for a specific file hash
func (h *MetadataHandler) GetByHash(c *gin.Context) {
	fileHash := c.Param("file_hash")
	if fileHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_hash is required",
		})
		return
	}

	// Get file
	file, err := h.db.GetFile(c.Request.Context(), fileHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}

	// Get all references
	references, err := h.db.ListReferencesByHash(c.Request.Context(), fileHash)
	if err != nil {
		h.logger.Error("failed to list references",
			zap.String("file_hash", fileHash),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list references",
		})
		return
	}

	// Build reference list
	refList := make([]gin.H, 0, len(references))
	for _, ref := range references {
		refList = append(refList, gin.H{
			"reference_id": ref.ID,
			"filename":     ref.Filename,
			"entity_type":  ref.EntityType,
			"entity_id":    ref.EntityID,
			"uploaded_by":  ref.UploaderID,
			"created_at":   ref.Created,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"file_hash":  fileHash,
		"size_bytes": file.SizeBytes,
		"mime_type":  file.MimeType,
		"ref_count":  file.RefCount,
		"references": refList,
	})
}
