package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/config"
	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/security/ratelimit"
	"github.com/helixtrack/attachments-service/internal/security/scanner"
	"github.com/helixtrack/attachments-service/internal/security/validation"
	"github.com/helixtrack/attachments-service/internal/storage/deduplication"
	"github.com/helixtrack/attachments-service/internal/storage/orchestrator"
	"github.com/helixtrack/attachments-service/internal/storage/reference"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

// FileHandlerDeps contains dependencies for file handlers
type FileHandlerDeps struct {
	DB                  database.Database
	DeduplicationEngine *deduplication.Engine
	RefCounter          *reference.Counter
	SecurityScanner     *scanner.Scanner
	StorageOrch         *orchestrator.Orchestrator
	Config              *config.Config
	Logger              *zap.Logger
}

// MetadataHandlerDeps contains dependencies for metadata handlers
type MetadataHandlerDeps struct {
	DB         database.Database
	RefCounter *reference.Counter
	Logger     *zap.Logger
}

// AdminHandlerDeps contains dependencies for admin handlers
type AdminHandlerDeps struct {
	DB             database.Database
	StorageOrch    *orchestrator.Orchestrator
	RefCounter     *reference.Counter
	RateLimiter    *ratelimit.Limiter
	ServiceRegistry *utils.ServiceRegistry
	Logger         *zap.Logger
}

// RegisterFileHandlers registers file upload and download handlers
func RegisterFileHandlers(router *gin.RouterGroup, deps *FileHandlerDeps) {
	// Initialize metrics
	metrics := utils.NewPrometheusMetrics()

	// Initialize validator
	validator := validation.NewValidator(&validation.ValidationConfig{
		MaxFilenameLength:    255,
		AllowedFilenameChars: "a-zA-Z0-9._ -",
		ForbiddenFilenames:   []string{"CON", "PRN", "AUX", "NUL"},
		MaxEntityTypeLength:  50,
		MaxEntityIDLength:    100,
		AllowedEntityTypes:   []string{"project", "ticket", "epic", "sprint", "board"},
		MaxUserIDLength:      100,
		MinUserIDLength:      3,
		MaxDescriptionLength: 1000,
		MaxTagLength:         50,
		MaxTagsPerFile:       20,
		AllowAbsolutePaths:   false,
		AllowPathTraversal:   false,
	})

	// Upload handler configuration
	uploadConfig := &UploadConfig{
		MaxFileSize: int64(deps.Config.Security.MaxFileSizeMB * 1024 * 1024),
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
		EnableVirusScan: deps.Config.Security.VirusScanning.Enabled,
	}

	// Download handler configuration
	downloadConfig := &DownloadConfig{
		EnableRangeRequests: true,
		EnableCaching:       true,
		CacheMaxAge:         3600,
		BufferSize:          32 * 1024,
	}

	// Initialize upload handler
	uploadHandler := NewUploadHandler(
		deps.DeduplicationEngine,
		deps.SecurityScanner,
		validator,
		metrics,
		deps.Logger,
		uploadConfig,
	)

	// Initialize download handler
	downloadHandler := NewDownloadHandler(
		deps.DeduplicationEngine,
		metrics,
		deps.Logger,
		downloadConfig,
	)

	// Register routes
	router.POST("/upload", uploadHandler.Handle)
	router.POST("/upload/multiple", uploadHandler.HandleMultiple)
	router.GET("/download/:reference_id", downloadHandler.Handle)
	router.GET("/view/:reference_id", downloadHandler.HandleInline)
	router.GET("/info/:reference_id", downloadHandler.HandleMetadata)
}

// RegisterMetadataHandlers registers metadata management handlers
func RegisterMetadataHandlers(router *gin.RouterGroup, deps *MetadataHandlerDeps) {
	// Initialize metrics
	metrics := utils.NewPrometheusMetrics()

	// Initialize metadata handler
	metadataHandler := NewMetadataHandler(
		deps.DB,
		&deduplication.Engine{},
		metrics,
		deps.Logger,
	)

	// Register routes
	router.GET("/attachments/:entity_type/:entity_id", metadataHandler.ListByEntity)
	router.DELETE("/attachments/:reference_id", metadataHandler.Delete)
	router.PATCH("/attachments/:reference_id", metadataHandler.Update)
	router.GET("/attachments/hash/:file_hash", metadataHandler.GetByHash)
	router.GET("/attachments/search", metadataHandler.Search)
	router.GET("/attachments/stats", metadataHandler.GetStats)
}

// RegisterAdminHandlers registers admin-only handlers
func RegisterAdminHandlers(router *gin.RouterGroup, deps *AdminHandlerDeps) {
	// Initialize metrics
	metrics := utils.NewPrometheusMetrics()

	// Initialize admin handler
	adminHandler := NewAdminHandler(
		deps.DB,
		deps.StorageOrch,
		deps.RefCounter,
		deps.RateLimiter,
		metrics,
		deps.ServiceRegistry,
		deps.Logger,
	)

	// Register routes
	router.GET("/health", adminHandler.Health)
	router.GET("/version", adminHandler.Version)
	router.GET("/stats", adminHandler.Stats)
	router.POST("/storage/verify", adminHandler.VerifyIntegrity)
	router.POST("/storage/repair", adminHandler.RepairIntegrity)
	router.POST("/cleanup/orphans", adminHandler.CleanupOrphans)
	router.POST("/blacklist/:ip", adminHandler.BlacklistIP)
	router.DELETE("/blacklist/:ip", adminHandler.UnblacklistIP)
	router.GET("/info", adminHandler.ServiceInfo)
}
