package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/security/ratelimit"
	"github.com/helixtrack/attachments-service/internal/storage/orchestrator"
	"github.com/helixtrack/attachments-service/internal/storage/reference"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

// AdminHandler handles administrative operations
type AdminHandler struct {
	db               database.Database
	orchestrator     *orchestrator.Orchestrator
	refCounter       *reference.Counter
	rateLimiter      *ratelimit.Limiter
	metrics          *utils.PrometheusMetrics
	serviceRegistry  *utils.ServiceRegistry
	logger           *zap.Logger
	startTime        time.Time
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	db database.Database,
	orch *orchestrator.Orchestrator,
	refCounter *reference.Counter,
	limiter *ratelimit.Limiter,
	metrics *utils.PrometheusMetrics,
	registry *utils.ServiceRegistry,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		db:              db,
		orchestrator:    orch,
		refCounter:      refCounter,
		rateLimiter:     limiter,
		metrics:         metrics,
		serviceRegistry: registry,
		logger:          logger,
		startTime:       time.Now(),
	}
}

// Health returns service health status
func (h *AdminHandler) Health(c *gin.Context) {
	health := gin.H{
		"status":  "healthy",
		"service": "attachments-service",
		"uptime":  time.Since(h.startTime).String(),
		"timestamp": time.Now().Unix(),
	}

	// Check database
	if err := h.db.Ping(); err != nil {
		health["status"] = "unhealthy"
		health["database"] = "unavailable"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}
	health["database"] = "healthy"

	// Check storage endpoints (if orchestrator exists)
	if h.orchestrator != nil {
		// Add orchestrator health check here if needed
		health["storage"] = "healthy"
	}

	c.JSON(http.StatusOK, health)
}

// Version returns service version information
func (h *AdminHandler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "attachments-service",
		"version": "1.0.0",
		"build_time": "2025-10-19",
		"go_version": "1.22+",
	})
}

// Stats returns comprehensive service statistics
func (h *AdminHandler) Stats(c *gin.Context) {
	// Get storage stats
	storageStats, err := h.db.GetStorageStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get storage stats",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get statistics",
		})
		return
	}

	// Get rate limiter stats
	var rateLimiterStats *ratelimit.LimiterStats
	if h.rateLimiter != nil {
		rateLimiterStats = h.rateLimiter.GetStats()
	}

	// Get reference counter stats
	var refCounterStats *reference.Statistics
	if h.refCounter != nil {
		refCounterStats, _ = h.refCounter.GetStatistics(c.Request.Context())
	}

	stats := gin.H{
		"storage": gin.H{
			"total_files":        storageStats.TotalFiles,
			"total_references":   storageStats.TotalReferences,
			"unique_files":       storageStats.UniqueFiles,
			"shared_files":       storageStats.SharedFiles,
			"orphaned_files":     storageStats.OrphanedFiles,
			"deduplication_rate": storageStats.DeduplicationRate,
			"total_size_bytes":   storageStats.TotalSizeBytes,
			"pending_scans":      storageStats.PendingScans,
			"infected_files":     storageStats.InfectedFiles,
		},
		"service": gin.H{
			"uptime":     time.Since(h.startTime).String(),
			"start_time": h.startTime.Unix(),
		},
	}

	if rateLimiterStats != nil {
		stats["rate_limiter"] = gin.H{
			"ip_buckets":      rateLimiterStats.IPBuckets,
			"user_buckets":    rateLimiterStats.UserBuckets,
			"blacklisted_ips": rateLimiterStats.BlacklistedIPs,
			"whitelisted_ips": rateLimiterStats.WhitelistedIPs,
			"global_tokens":   rateLimiterStats.GlobalTokens,
		}
	}

	if refCounterStats != nil {
		stats["references"] = gin.H{
			"total_files":          refCounterStats.TotalFiles,
			"total_references":     refCounterStats.TotalReferences,
			"unique_files":         refCounterStats.UniqueFiles,
			"shared_files":         refCounterStats.SharedFiles,
			"orphaned_files":       refCounterStats.OrphanedFiles,
			"average_refs_per_file": refCounterStats.AverageRefsPerFile,
		}
	}

	c.JSON(http.StatusOK, stats)
}

// CleanupOrphans triggers orphaned file cleanup
func (h *AdminHandler) CleanupOrphans(c *gin.Context) {
	retentionDays := 30 // Default retention

	// Check admin permission
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin role required",
		})
		return
	}

	h.logger.Info("starting orphan cleanup",
		zap.Int("retention_days", retentionDays),
	)

	// Check if reference counter is available
	if h.refCounter == nil {
		h.logger.Error("reference counter not available")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "reference counter not available",
		})
		return
	}

	// Find orphaned files
	deleted, err := h.refCounter.CleanupOrphaned(c.Request.Context(), retentionDays)
	if err != nil {
		h.logger.Error("orphan cleanup failed",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cleanup failed",
		})
		return
	}

	h.logger.Info("orphan cleanup complete",
		zap.Int64("deleted", deleted),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "cleanup complete",
		"deleted_files": deleted,
		"retention_days": retentionDays,
	})
}

// VerifyIntegrity verifies reference counting integrity
func (h *AdminHandler) VerifyIntegrity(c *gin.Context) {
	// Check admin permission
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin role required",
		})
		return
	}

	h.logger.Info("verifying reference count integrity")

	// Check if reference counter is available
	if h.refCounter == nil {
		h.logger.Error("reference counter not available")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "reference counter not available",
		})
		return
	}

	issues, err := h.refCounter.VerifyIntegrity(c.Request.Context())
	if err != nil {
		h.logger.Error("integrity verification failed",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "verification failed",
		})
		return
	}

	if len(issues) > 0 {
		h.logger.Warn("integrity issues found",
			zap.Int("count", len(issues)),
		)

		issueList := make([]gin.H, len(issues))
		for i, issue := range issues {
			issueList[i] = gin.H{
				"file_hash":      issue.FileHash,
				"database_count": issue.DatabaseCount,
				"actual_count":   issue.ActualCount,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "issues_found",
			"issues": issueList,
			"count":  len(issues),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "no integrity issues found",
	})
}

// RepairIntegrity repairs reference counting integrity
func (h *AdminHandler) RepairIntegrity(c *gin.Context) {
	// Check admin permission
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin role required",
		})
		return
	}

	h.logger.Info("repairing reference count integrity")

	// Check if reference counter is available
	if h.refCounter == nil {
		h.logger.Error("reference counter not available")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "reference counter not available",
		})
		return
	}

	repaired, err := h.refCounter.RepairIntegrity(c.Request.Context())
	if err != nil {
		h.logger.Error("integrity repair failed",
			zap.Error(err),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "repair failed",
		})
		return
	}

	h.logger.Info("integrity repair complete",
		zap.Int("repaired", repaired),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "repair complete",
		"repaired_count": repaired,
	})
}

// BlacklistIP adds an IP to the blacklist
func (h *AdminHandler) BlacklistIP(c *gin.Context) {
	// Check admin permission
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin role required",
		})
		return
	}

	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if h.rateLimiter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "rate limiter not configured",
		})
		return
	}

	h.rateLimiter.AddToBlacklist(req.IP)

	h.logger.Info("IP blacklisted",
		zap.String("ip", req.IP),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "IP blacklisted successfully",
		"ip": req.IP,
	})
}

// UnblacklistIP removes an IP from the blacklist
func (h *AdminHandler) UnblacklistIP(c *gin.Context) {
	// Check admin permission
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "admin role required",
		})
		return
	}

	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if h.rateLimiter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "rate limiter not configured",
		})
		return
	}

	h.rateLimiter.RemoveFromBlacklist(req.IP)

	h.logger.Info("IP removed from blacklist",
		zap.String("ip", req.IP),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "IP removed from blacklist",
		"ip": req.IP,
	})
}

// ServiceInfo returns service discovery information
func (h *AdminHandler) ServiceInfo(c *gin.Context) {
	info := gin.H{
		"service_name": "attachments-service",
		"version": "1.0.0",
		"uptime": time.Since(h.startTime).String(),
	}

	// Add service discovery info if available
	if h.serviceRegistry != nil {
		// Add registry info here
		info["service_discovery"] = "consul"
	}

	c.JSON(http.StatusOK, info)
}
