package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/security/scanner"
	"github.com/helixtrack/attachments-service/internal/security/validation"
	"github.com/helixtrack/attachments-service/internal/storage/deduplication"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	deduplicationEngine deduplication.DeduplicationEngine
	securityScanner     scanner.SecurityScanner
	validator           *validation.Validator
	metrics             utils.MetricsRecorder
	logger              *zap.Logger
	config              *UploadConfig
}

// UploadConfig contains upload handler configuration
type UploadConfig struct {
	MaxFileSize       int64
	AllowedMimeTypes  []string
	AllowedExtensions []string
	RequireAuth       bool
	EnableVirusScan   bool
}

// DefaultUploadConfig returns default upload configuration
func DefaultUploadConfig() *UploadConfig {
	return &UploadConfig{
		MaxFileSize: 100 * 1024 * 1024, // 100 MB
		AllowedMimeTypes: []string{
			"image/jpeg", "image/png", "image/gif",
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"text/plain",
		},
		AllowedExtensions: []string{
			".jpg", ".jpeg", ".png", ".gif",
			".pdf", ".doc", ".docx", ".txt",
		},
		RequireAuth:     true,
		EnableVirusScan: true,
	}
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(
	engine deduplication.DeduplicationEngine,
	secScanner scanner.SecurityScanner,
	validator *validation.Validator,
	metrics utils.MetricsRecorder,
	logger *zap.Logger,
	config *UploadConfig,
) *UploadHandler {
	if config == nil {
		config = DefaultUploadConfig()
	}

	return &UploadHandler{
		deduplicationEngine: engine,
		securityScanner:     secScanner,
		validator:           validator,
		metrics:             metrics,
		logger:              logger,
		config:              config,
	}
}

// UploadRequest represents an upload request
type UploadRequest struct {
	EntityType  string   `form:"entity_type" binding:"required"`
	EntityID    string   `form:"entity_id" binding:"required"`
	Description string   `form:"description"`
	Tags        []string `form:"tags"`
}

// UploadResponse represents an upload response
type UploadResponse struct {
	ReferenceID   string `json:"reference_id"`
	FileHash      string `json:"file_hash"`
	Filename      string `json:"filename"`
	SizeBytes     int64  `json:"size_bytes"`
	MimeType      string `json:"mime_type"`
	Deduplicated  bool   `json:"deduplicated"`
	SavedBytes    int64  `json:"saved_bytes,omitempty"`
	UploadTime    int64  `json:"upload_time"`
}

// Handle handles file upload requests
func (h *UploadHandler) Handle(c *gin.Context) {
	startTime := time.Now()

	// Get user ID from context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists && h.config.RequireAuth {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	uploaderID := ""
	if exists {
		uploaderID = userID.(string)
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(h.config.MaxFileSize)
	if err != nil {
		h.logger.Warn("failed to parse multipart form",
			zap.Error(err),
			zap.String("ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid multipart form",
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Warn("no file in request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file is required",
		})
		return
	}
	defer file.Close()

	// Get upload metadata
	var req UploadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// Validate filename (if validator is available)
	filename := header.Filename
	if h.validator != nil {
		var err error
		filename, err = h.validator.ValidateFilename(header.Filename)
		if err != nil {
			h.logger.Warn("invalid filename",
				zap.String("filename", header.Filename),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid filename",
				"details": err.Error(),
			})
			return
		}
	}

	// Validate entity type (if validator is available)
	if h.validator != nil {
		if err := h.validator.ValidateEntityType(req.EntityType); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid entity type",
				"details": err.Error(),
			})
			return
		}
	}

	// Validate entity ID (if validator is available)
	if h.validator != nil {
		if err := h.validator.ValidateEntityID(req.EntityID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid entity ID",
				"details": err.Error(),
			})
			return
		}
	}

	// Validate description (if validator is available)
	if h.validator != nil && req.Description != "" {
		if err := h.validator.ValidateDescription(req.Description); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid description",
				"details": err.Error(),
			})
			return
		}
	}

	// Validate and sanitize tags (if validator is available)
	if h.validator != nil && len(req.Tags) > 0 {
		req.Tags = h.validator.SanitizeTags(req.Tags)
	}

	h.logger.Info("processing file upload",
		zap.String("filename", filename),
		zap.String("entity_type", req.EntityType),
		zap.String("entity_id", req.EntityID),
		zap.String("uploader_id", uploaderID),
		zap.Int64("size", header.Size),
	)

	// Security scan
	scanResult, err := h.securityScanner.Scan(c.Request.Context(), file, filename)
	if err != nil || !scanResult.Safe {
		h.logger.Warn("security scan failed",
			zap.String("filename", filename),
			zap.Bool("safe", scanResult.Safe),
			zap.Strings("errors", scanResult.Errors),
			zap.Error(err),
		)

		errorMsg := "file failed security scan"
		if scanResult.VirusDetected {
			errorMsg = fmt.Sprintf("virus detected: %s", scanResult.VirusName)
		} else if len(scanResult.Errors) > 0 {
			errorMsg = scanResult.Errors[0]
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": errorMsg,
			"warnings": scanResult.Warnings,
		})
		return
	}

	// Log warnings if any
	if len(scanResult.Warnings) > 0 {
		h.logger.Warn("security scan warnings",
			zap.String("filename", filename),
			zap.Strings("warnings", scanResult.Warnings),
		)
	}

	// Reset file reader position
	file.Seek(0, 0)

	// Prepare upload metadata
	metadata := &deduplication.UploadMetadata{
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		Filename:    filename,
		UploaderID:  uploaderID,
		MimeType:    scanResult.MimeType,
		Extension:   scanResult.Extension,
		Description: req.Description,
		Tags:        req.Tags,
	}

	// Process upload with deduplication
	result, err := h.deduplicationEngine.ProcessUpload(c.Request.Context(), file, metadata)
	if err != nil {
		h.logger.Error("upload processing failed",
			zap.String("filename", filename),
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process upload",
		})
		return
	}

	// Record metrics
	duration := time.Since(startTime)
	if h.metrics != nil {
		h.metrics.RecordUpload("success", scanResult.MimeType, result.SizeBytes, duration)
		if result.Deduplicated {
			h.metrics.RecordDeduplication(true, result.SavedBytes)
		}
	}

	h.logger.Info("upload successful",
		zap.String("reference_id", result.ReferenceID),
		zap.String("file_hash", result.FileHash),
		zap.Int64("size", result.SizeBytes),
		zap.Bool("deduplicated", result.Deduplicated),
		zap.Int64("saved_bytes", result.SavedBytes),
		zap.Duration("duration", duration),
	)

	// Return response
	response := &UploadResponse{
		ReferenceID:  result.ReferenceID,
		FileHash:     result.FileHash,
		Filename:     filename,
		SizeBytes:    result.SizeBytes,
		MimeType:     scanResult.MimeType,
		Deduplicated: result.Deduplicated,
		SavedBytes:   result.SavedBytes,
		UploadTime:   time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// HandleMultiple handles multiple file uploads
func (h *UploadHandler) HandleMultiple(c *gin.Context) {
	startTime := time.Now()

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists && h.config.RequireAuth {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	uploaderID := ""
	if exists {
		uploaderID = userID.(string)
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid multipart form",
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no files provided",
		})
		return
	}

	// Limit number of files
	maxFiles := 10
	if len(files) > maxFiles {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("too many files, maximum %d allowed", maxFiles),
		})
		return
	}

	// Get upload metadata
	var req UploadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request parameters",
		})
		return
	}

	// Process each file
	responses := make([]*UploadResponse, 0, len(files))
	errors := make([]string, 0)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to open", fileHeader.Filename))
			continue
		}

		// Validate filename (if validator is available)
		filename := fileHeader.Filename
		if h.validator != nil {
			var err error
			filename, err = h.validator.ValidateFilename(fileHeader.Filename)
			if err != nil {
				file.Close()
				errors = append(errors, fmt.Sprintf("%s: invalid filename", fileHeader.Filename))
				continue
			}
		}

		// Security scan
		scanResult, err := h.securityScanner.Scan(c.Request.Context(), file, filename)
		if err != nil || !scanResult.Safe {
			file.Close()
			errors = append(errors, fmt.Sprintf("%s: security scan failed", filename))
			continue
		}

		// Reset file position
		file.Seek(0, 0)

		// Prepare metadata
		metadata := &deduplication.UploadMetadata{
			EntityType:  req.EntityType,
			EntityID:    req.EntityID,
			Filename:    filename,
			UploaderID:  uploaderID,
			MimeType:    scanResult.MimeType,
			Extension:   scanResult.Extension,
			Description: req.Description,
			Tags:        req.Tags,
		}

		// Process upload
		result, err := h.deduplicationEngine.ProcessUpload(c.Request.Context(), file, metadata)
		file.Close()

		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: upload failed", filename))
			continue
		}

		// Add to responses
		responses = append(responses, &UploadResponse{
			ReferenceID:  result.ReferenceID,
			FileHash:     result.FileHash,
			Filename:     filename,
			SizeBytes:    result.SizeBytes,
			MimeType:     scanResult.MimeType,
			Deduplicated: result.Deduplicated,
			SavedBytes:   result.SavedBytes,
			UploadTime:   time.Now().Unix(),
		})
	}

	duration := time.Since(startTime)

	h.logger.Info("multiple upload complete",
		zap.Int("total_files", len(files)),
		zap.Int("successful", len(responses)),
		zap.Int("failed", len(errors)),
		zap.Duration("duration", duration),
	)

	c.JSON(http.StatusOK, gin.H{
		"uploads": responses,
		"errors":  errors,
		"summary": gin.H{
			"total":      len(files),
			"successful": len(responses),
			"failed":     len(errors),
		},
	})
}
