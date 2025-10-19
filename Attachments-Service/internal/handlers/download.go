package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/storage/deduplication"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

// DownloadHandler handles file download requests
type DownloadHandler struct {
	deduplicationEngine deduplication.DeduplicationEngine
	metrics             utils.MetricsRecorder
	logger              *zap.Logger
	config              *DownloadConfig
}

// DownloadConfig contains download handler configuration
type DownloadConfig struct {
	EnableRangeRequests bool
	EnableCaching       bool
	CacheMaxAge         int // seconds
	BufferSize          int // bytes for streaming
}

// DefaultDownloadConfig returns default download configuration
func DefaultDownloadConfig() *DownloadConfig {
	return &DownloadConfig{
		EnableRangeRequests: true,
		EnableCaching:       true,
		CacheMaxAge:         3600, // 1 hour
		BufferSize:          32 * 1024, // 32 KB
	}
}

// NewDownloadHandler creates a new download handler
func NewDownloadHandler(
	engine deduplication.DeduplicationEngine,
	metrics utils.MetricsRecorder,
	logger *zap.Logger,
	config *DownloadConfig,
) *DownloadHandler {
	if config == nil {
		config = DefaultDownloadConfig()
	}

	return &DownloadHandler{
		deduplicationEngine: engine,
		metrics:             metrics,
		logger:              logger,
		config:              config,
	}
}

// Handle handles file download requests
func (h *DownloadHandler) Handle(c *gin.Context) {
	startTime := time.Now()

	// Get reference ID from URL parameter
	referenceID := c.Param("reference_id")
	if referenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "reference_id is required",
		})
		return
	}

	h.logger.Info("processing download request",
		zap.String("reference_id", referenceID),
		zap.String("ip", c.ClientIP()),
	)

	// Download file
	reader, reference, file, err := h.deduplicationEngine.DownloadFile(c.Request.Context(), referenceID)
	if err != nil {
		h.logger.Warn("download failed",
			zap.String("reference_id", referenceID),
			zap.Error(err),
		)

		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}
	defer reader.Close()

	// Set response headers
	h.setDownloadHeaders(c, reference.Filename, file.MimeType, file.SizeBytes)

	// Handle range requests if enabled
	if h.config.EnableRangeRequests {
		rangeHeader := c.GetHeader("Range")
		if rangeHeader != "" {
			h.handleRangeRequest(c, reader, file.SizeBytes, rangeHeader)
			return
		}
	}

	// Stream file to response
	written, err := h.streamFile(c.Writer, reader)
	if err != nil {
		h.logger.Error("failed to stream file",
			zap.String("reference_id", referenceID),
			zap.Error(err),
		)
		return
	}

	// Record metrics
	duration := time.Since(startTime)
	if h.metrics != nil {
		h.metrics.RecordDownload("success", written, duration, false)
	}

	h.logger.Info("download successful",
		zap.String("reference_id", referenceID),
		zap.String("filename", reference.Filename),
		zap.Int64("bytes_sent", written),
		zap.Duration("duration", duration),
	)
}

// HandleInline handles inline file viewing (in browser)
func (h *DownloadHandler) HandleInline(c *gin.Context) {
	startTime := time.Now()

	referenceID := c.Param("reference_id")
	if referenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "reference_id is required",
		})
		return
	}

	// Download file
	reader, reference, file, err := h.deduplicationEngine.DownloadFile(c.Request.Context(), referenceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}
	defer reader.Close()

	// Set inline headers
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", file.SizeBytes))
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", reference.Filename))

	// Cache headers
	if h.config.EnableCaching {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", h.config.CacheMaxAge))
		c.Header("ETag", file.Hash)
	}

	// Stream file
	written, err := h.streamFile(c.Writer, reader)
	if err != nil {
		h.logger.Error("failed to stream file",
			zap.Error(err),
		)
		return
	}

	duration := time.Since(startTime)
	if h.metrics != nil {
		h.metrics.RecordDownload("success", written, duration, false)
	}

	h.logger.Info("inline view successful",
		zap.String("reference_id", referenceID),
		zap.Int64("bytes_sent", written),
		zap.Duration("duration", duration),
	)
}

// setDownloadHeaders sets standard download headers
func (h *DownloadHandler) setDownloadHeaders(c *gin.Context, filename, mimeType string, size int64) {
	// Content headers
	c.Header("Content-Type", mimeType)
	c.Header("Content-Length", fmt.Sprintf("%d", size))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Cache headers
	if h.config.EnableCaching {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", h.config.CacheMaxAge))
	} else {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	// Security headers
	c.Header("X-Content-Type-Options", "nosniff")

	// Accept ranges header
	if h.config.EnableRangeRequests {
		c.Header("Accept-Ranges", "bytes")
	}
}

// streamFile streams file content to the response writer
func (h *DownloadHandler) streamFile(w io.Writer, reader io.Reader) (int64, error) {
	buffer := make([]byte, h.config.BufferSize)
	return io.CopyBuffer(w, reader, buffer)
}

// handleRangeRequest handles HTTP range requests for partial content
func (h *DownloadHandler) handleRangeRequest(c *gin.Context, reader io.Reader, size int64, rangeHeader string) {
	// Parse range header (e.g., "bytes=0-1023")
	start, end, err := parseRangeHeader(rangeHeader, size)
	if err != nil {
		c.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{
			"error": "invalid range",
		})
		return
	}

	// Seek to start position
	if seeker, ok := reader.(io.Seeker); ok {
		_, err := seeker.Seek(start, io.SeekStart)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to seek",
			})
			return
		}
	} else {
		// If reader is not seekable, discard bytes until start
		_, err := io.CopyN(io.Discard, reader, start)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to skip bytes",
			})
			return
		}
	}

	// Calculate content length for range
	contentLength := end - start + 1

	// Set range response headers
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
	c.Header("Content-Length", fmt.Sprintf("%d", contentLength))
	c.Status(http.StatusPartialContent)

	// Stream the range
	limitedReader := io.LimitReader(reader, contentLength)
	buffer := make([]byte, h.config.BufferSize)
	io.CopyBuffer(c.Writer, limitedReader, buffer)
}

// parseRangeHeader parses HTTP Range header
func parseRangeHeader(rangeHeader string, fileSize int64) (start, end int64, err error) {
	// Expected format: "bytes=start-end" or "bytes=start-" or "bytes=-end"
	if len(rangeHeader) < 7 || rangeHeader[:6] != "bytes=" {
		return 0, 0, fmt.Errorf("invalid range header format")
	}

	rangeSpec := rangeHeader[6:]

	// Handle different range formats
	if rangeSpec[0] == '-' {
		// Suffix range: bytes=-500 (last 500 bytes)
		suffixLength, err := strconv.ParseInt(rangeSpec[1:], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		start = fileSize - suffixLength
		end = fileSize - 1
	} else {
		// Parse start-end range
		var startStr, endStr string
		dashIndex := -1
		for i, c := range rangeSpec {
			if c == '-' {
				dashIndex = i
				break
			}
		}

		if dashIndex == -1 {
			return 0, 0, fmt.Errorf("invalid range format")
		}

		startStr = rangeSpec[:dashIndex]
		endStr = rangeSpec[dashIndex+1:]

		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}

		if endStr == "" {
			// bytes=500- (from 500 to end)
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	// Validate range
	if start < 0 || end >= fileSize || start > end {
		return 0, 0, fmt.Errorf("range out of bounds")
	}

	return start, end, nil
}

// HandleMetadata handles requests for file metadata only
func (h *DownloadHandler) HandleMetadata(c *gin.Context) {
	referenceID := c.Param("reference_id")
	if referenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "reference_id is required",
		})
		return
	}

	// Get file metadata without downloading
	_, reference, file, err := h.deduplicationEngine.DownloadFile(c.Request.Context(), referenceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reference_id": reference.ID,
		"filename":     reference.Filename,
		"file_hash":    file.Hash,
		"size_bytes":   file.SizeBytes,
		"mime_type":    file.MimeType,
		"extension":    file.Extension,
		"created_at":   reference.Created,
		"uploaded_by":  reference.UploaderID,
		"entity_type":  reference.EntityType,
		"entity_id":    reference.EntityID,
		"tags":         reference.Tags,
		"description":  reference.Description,
	})
}
