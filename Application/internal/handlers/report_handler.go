package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/middleware"
	"helixtrack.ru/core/internal/models"
)

// handleReportCreate creates a new report
func (h *Handler) handleReportCreate(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "report", models.PermissionCreate)
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

	// Parse report data from request
	title, ok := req.Data["title"].(string)
	if !ok || title == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing title",
			"",
		))
		return
	}

	// Query can be a complex object, convert to JSON string
	var queryStr string
	if query, ok := req.Data["query"]; ok {
		queryBytes, err := json.Marshal(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid query format",
				"",
			))
			return
		}
		queryStr = string(queryBytes)
	}

	report := &models.Report{
		ID:          uuid.New().String(),
		Title:       title,
		Description: getStringFromData(req.Data, "description"),
		Query:       queryStr,
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
		Deleted:     false,
	}

	// Validate report
	if !report.IsValid() {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeInvalidData,
			"Invalid report data",
			"",
		))
		return
	}

	// Insert into database
	query := `
		INSERT INTO report (id, title, description, query, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(c.Request.Context(), query,
		report.ID,
		report.Title,
		report.Description,
		report.Query,
		report.Created,
		report.Modified,
		report.Deleted,
	)

	if err != nil {
		logger.Error("Failed to create report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to create report",
			"",
		))
		return
	}

	logger.Info("Report created",
		zap.String("report_id", report.ID),
		zap.String("title", report.Title),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"report": report,
	})
	c.JSON(http.StatusCreated, response)
}

// handleReportRead reads a single report by ID
func (h *Handler) handleReportRead(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get report ID from request
	reportID, ok := req.Data["id"].(string)
	if !ok || reportID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing report ID",
			"",
		))
		return
	}

	// Query report from database
	query := `
		SELECT id, title, description, query, created, modified, deleted
		FROM report
		WHERE id = ? AND deleted = 0
	`

	var report models.Report
	err := h.db.QueryRow(c.Request.Context(), query, reportID).Scan(
		&report.ID,
		&report.Title,
		&report.Description,
		&report.Query,
		&report.Created,
		&report.Modified,
		&report.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Report not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read report",
			"",
		))
		return
	}

	logger.Info("Report read",
		zap.String("report_id", report.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"report": report,
	})
	c.JSON(http.StatusOK, response)
}

// handleReportList lists all reports
func (h *Handler) handleReportList(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Query all non-deleted reports
	query := `
		SELECT id, title, description, query, created, modified, deleted
		FROM report
		WHERE deleted = 0
		ORDER BY title ASC
	`

	rows, err := h.db.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to list reports", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to list reports",
			"",
		))
		return
	}
	defer rows.Close()

	reports := make([]models.Report, 0)
	for rows.Next() {
		var report models.Report
		err := rows.Scan(
			&report.ID,
			&report.Title,
			&report.Description,
			&report.Query,
			&report.Created,
			&report.Modified,
			&report.Deleted,
		)
		if err != nil {
			logger.Error("Failed to scan report", zap.Error(err))
			continue
		}
		reports = append(reports, report)
	}

	logger.Info("Reports listed",
		zap.Int("count", len(reports)),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"reports": reports,
		"count":   len(reports),
	})
	c.JSON(http.StatusOK, response)
}

// handleReportModify updates an existing report
func (h *Handler) handleReportModify(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "report", models.PermissionUpdate)
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

	// Get report ID
	reportID, ok := req.Data["id"].(string)
	if !ok || reportID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing report ID",
			"",
		))
		return
	}

	// Check if report exists
	checkQuery := `SELECT COUNT(*) FROM report WHERE id = ? AND deleted = 0`
	var count int
	err = h.db.QueryRow(c.Request.Context(), checkQuery, reportID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Report not found",
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
	if query, ok := req.Data["query"]; ok {
		queryBytes, err := json.Marshal(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid query format",
				"",
			))
			return
		}
		updates["query"] = string(queryBytes)
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
	queryStr := "UPDATE report SET "
	args := make([]interface{}, 0)
	first := true

	for key, value := range updates {
		if !first {
			queryStr += ", "
		}
		queryStr += key + " = ?"
		args = append(args, value)
		first = false
	}

	queryStr += " WHERE id = ?"
	args = append(args, reportID)

	_, err = h.db.Exec(c.Request.Context(), queryStr, args...)
	if err != nil {
		logger.Error("Failed to update report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to update report",
			"",
		))
		return
	}

	logger.Info("Report updated",
		zap.String("report_id", reportID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"updated": true,
		"id":      reportID,
	})
	c.JSON(http.StatusOK, response)
}

// handleReportRemove soft-deletes a report
func (h *Handler) handleReportRemove(c *gin.Context, req *models.Request) {
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
	allowed, err := h.permService.CheckPermission(c.Request.Context(), username, "report", models.PermissionDelete)
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

	// Get report ID
	reportID, ok := req.Data["id"].(string)
	if !ok || reportID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing report ID",
			"",
		))
		return
	}

	// Soft delete the report
	query := `UPDATE report SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), reportID)
	if err != nil {
		logger.Error("Failed to delete report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to delete report",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Report not found",
			"",
		))
		return
	}

	logger.Info("Report deleted",
		zap.String("report_id", reportID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      reportID,
	})
	c.JSON(http.StatusOK, response)
}

// handleReportExecute executes a report and returns results
func (h *Handler) handleReportExecute(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get report ID from request
	reportID, ok := req.Data["id"].(string)
	if !ok || reportID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing report ID",
			"",
		))
		return
	}

	// Query report from database
	query := `
		SELECT id, title, description, query, created, modified, deleted
		FROM report
		WHERE id = ? AND deleted = 0
	`

	var report models.Report
	err := h.db.QueryRow(c.Request.Context(), query, reportID).Scan(
		&report.ID,
		&report.Title,
		&report.Description,
		&report.Query,
		&report.Created,
		&report.Modified,
		&report.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Report not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to read report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to read report",
			"",
		))
		return
	}

	// Parse query and execute it
	// For now, we return a placeholder response
	// In a real implementation, this would parse the JSON query and execute it against the database
	logger.Info("Report executed",
		zap.String("report_id", report.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"report": report,
		"results": map[string]interface{}{
			"message": "Report execution not yet implemented - placeholder response",
			"query":   report.Query,
		},
	})
	c.JSON(http.StatusOK, response)
}

// handleReportSetMetadata sets metadata for a report
func (h *Handler) handleReportSetMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Parse metadata from request
	reportID, ok := req.Data["reportId"].(string)
	if !ok || reportID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing reportId",
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

	// Value can be any type, convert to JSON string
	var valueStr string
	if value, ok := req.Data["value"]; ok {
		valueBytes, err := json.Marshal(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrorCodeInvalidData,
				"Invalid value format",
				"",
			))
			return
		}
		valueStr = string(valueBytes)
	}

	metadata := &models.ReportMetaData{
		ID:       uuid.New().String(),
		ReportID: reportID,
		Property: property,
		Value:    valueStr,
		Created:  time.Now().Unix(),
		Modified: time.Now().Unix(),
		Deleted:  false,
	}

	// Insert into database
	query := `
		INSERT INTO report_metadata (id, report_id, property, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(c.Request.Context(), query,
		metadata.ID,
		metadata.ReportID,
		metadata.Property,
		metadata.Value,
		metadata.Created,
		metadata.Modified,
		metadata.Deleted,
	)

	if err != nil {
		logger.Error("Failed to set report metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to set report metadata",
			"",
		))
		return
	}

	logger.Info("Report metadata set",
		zap.String("metadata_id", metadata.ID),
		zap.String("report_id", reportID),
		zap.String("property", property),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
	})
	c.JSON(http.StatusCreated, response)
}

// handleReportGetMetadata gets metadata for a report
func (h *Handler) handleReportGetMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get report ID and property from request
	reportID, ok := req.Data["reportId"].(string)
	if !ok || reportID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing reportId",
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
		SELECT id, report_id, property, value, created, modified, deleted
		FROM report_metadata
		WHERE report_id = ? AND property = ? AND deleted = 0
	`

	var metadata models.ReportMetaData
	err := h.db.QueryRow(c.Request.Context(), query, reportID, property).Scan(
		&metadata.ID,
		&metadata.ReportID,
		&metadata.Property,
		&metadata.Value,
		&metadata.Created,
		&metadata.Modified,
		&metadata.Deleted,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Report metadata not found",
			"",
		))
		return
	}

	if err != nil {
		logger.Error("Failed to get report metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to get report metadata",
			"",
		))
		return
	}

	logger.Info("Report metadata retrieved",
		zap.String("metadata_id", metadata.ID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"metadata": metadata,
	})
	c.JSON(http.StatusOK, response)
}

// handleReportRemoveMetadata removes metadata from a report
func (h *Handler) handleReportRemoveMetadata(c *gin.Context, req *models.Request) {
	username, exists := middleware.GetUsername(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrorCodeUnauthorized,
			"Unauthorized",
			"",
		))
		return
	}

	// Get metadata ID from request
	metadataID, ok := req.Data["id"].(string)
	if !ok || metadataID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrorCodeMissingData,
			"Missing metadata ID",
			"",
		))
		return
	}

	// Soft delete the metadata
	query := `UPDATE report_metadata SET deleted = 1, modified = ? WHERE id = ? AND deleted = 0`
	result, err := h.db.Exec(c.Request.Context(), query, time.Now().Unix(), metadataID)
	if err != nil {
		logger.Error("Failed to remove report metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrorCodeInternalError,
			"Failed to remove report metadata",
			"",
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrorCodeNotFound,
			"Report metadata not found",
			"",
		))
		return
	}

	logger.Info("Report metadata removed",
		zap.String("metadata_id", metadataID),
		zap.String("username", username),
	)

	response := models.NewSuccessResponse(map[string]interface{}{
		"deleted": true,
		"id":      metadataID,
	})
	c.JSON(http.StatusOK, response)
}
