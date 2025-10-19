package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/attachments-service/internal/config"
	"github.com/helixtrack/attachments-service/internal/database"
	"github.com/helixtrack/attachments-service/internal/handlers"
	"github.com/helixtrack/attachments-service/internal/middleware"
	"github.com/helixtrack/attachments-service/internal/security/ratelimit"
	"github.com/helixtrack/attachments-service/internal/security/scanner"
	"github.com/helixtrack/attachments-service/internal/storage/adapters"
	"github.com/helixtrack/attachments-service/internal/storage/deduplication"
	"github.com/helixtrack/attachments-service/internal/storage/orchestrator"
	"github.com/helixtrack/attachments-service/internal/storage/reference"
	"github.com/helixtrack/attachments-service/internal/utils"
	"go.uber.org/zap"
)

const (
	serviceName    = "attachments-service"
	serviceVersion = "1.0.0"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "../configs/default.json", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger, err := utils.NewLogger("info", "json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting HelixTrack Attachments Service",
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
	)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration",
			zap.String("config_path", *configPath),
			zap.Error(err),
		)
	}

	logger.Info("Configuration loaded successfully",
		zap.String("config_path", *configPath),
	)

	// Find available port (auto port selection)
	port, err := findAvailablePort(cfg.Service.Port, cfg.Service.PortRange)
	if err != nil {
		logger.Fatal("Failed to find available port", zap.Error(err))
	}

	cfg.Service.Port = port
	logger.Info("Port selected",
		zap.Int("port", port),
	)

	// Initialize database connection
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database connection established",
		zap.String("driver", cfg.Database.Driver),
		zap.String("host", cfg.Database.Host),
	)

	// Run database migrations
	if err := db.Migrate(); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	logger.Info("Database migrations completed successfully")

	// Initialize security scanner
	scanConfig := &scanner.ScanConfig{
		AllowedMimeTypes:          cfg.Security.AllowedMimeTypes,
		AllowedExtensions:         []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".txt"},
		MaxFileSize:               cfg.Security.MaxFileSizeMB * 1024 * 1024,
		MaxImageWidth:             cfg.Security.ImageValidation.MaxWidthPixels,
		MaxImageHeight:            cfg.Security.ImageValidation.MaxHeightPixels,
		MaxImagePixels:            int64(cfg.Security.ImageValidation.MaxWidthPixels) * int64(cfg.Security.ImageValidation.MaxHeightPixels),
		EnableImageBombProtection: true,
		EnableClamAV:              cfg.Security.VirusScanning.Enabled,
		ClamAVSocket:              cfg.Security.VirusScanning.ClamdSocket,
		ClamAVTimeout:             time.Duration(cfg.Security.VirusScanning.TimeoutSeconds) * time.Second,
		EnableMagicBytes:          true,
		StrictMagicBytes:          true,
		EnableContentAnalysis:     true,
		MaxScanBytes:              cfg.Security.VirusScanning.MaxScanSizeMB * 1024 * 1024,
	}
	securityScanner := scanner.NewScanner(scanConfig, logger)

	logger.Info("Security scanner initialized",
		zap.Bool("virus_scanning", cfg.Security.VirusScanning.Enabled),
	)

	// Initialize storage orchestrator
	orchConfig := &orchestrator.OrchestratorConfig{
		EnableFailover:           true,
		FailoverTimeout:          30 * time.Second,
		MaxRetries:               3,
		EnableMirroring:          len(cfg.Storage.Endpoints) > 1,
		MirrorAsync:              cfg.Storage.ReplicationMode == "asynchronous" || cfg.Storage.ReplicationMode == "hybrid",
		RequireAllMirrorsSuccess: cfg.Storage.ReplicationMode == "synchronous",
		HealthCheckInterval:      30 * time.Second,
		HealthCheckTimeout:       5 * time.Second,
		UnhealthyThreshold:       3,
		HealthyThreshold:         2,
		CircuitBreakerThreshold:  5,
		CircuitBreakerTimeout:    60 * time.Second,
	}
	storageOrch := orchestrator.NewOrchestrator(db, orchConfig, logger)

	// Initialize storage adapters from endpoints
	ctx := context.Background()
	for _, endpoint := range cfg.Storage.Endpoints {
		if !endpoint.Enabled {
			logger.Info("Skipping disabled storage endpoint",
				zap.String("id", endpoint.ID),
			)
			continue
		}

		// Create adapter based on type
		var adapter adapters.StorageAdapter
		var err error

		switch endpoint.Type {
		case "local":
			basePath, ok := endpoint.AdapterConfig["base_path"].(string)
			if !ok || basePath == "" {
				logger.Warn("Local adapter requires 'base_path' in adapter_config",
					zap.String("id", endpoint.ID),
				)
				continue
			}
			adapter, err = adapters.NewLocalAdapter(basePath, logger)

		case "s3":
			s3Cfg, err := parseS3Config(endpoint.AdapterConfig)
			if err != nil {
				logger.Warn("Failed to parse S3 config",
					zap.String("id", endpoint.ID),
					zap.Error(err),
				)
				continue
			}
			adapter, err = adapters.NewS3Adapter(ctx, s3Cfg, logger)

		case "minio":
			minioCfg, err := parseMinIOConfig(endpoint.AdapterConfig)
			if err != nil {
				logger.Warn("Failed to parse MinIO config",
					zap.String("id", endpoint.ID),
					zap.Error(err),
				)
				continue
			}
			adapter, err = adapters.NewMinIOAdapter(ctx, minioCfg, logger)

		default:
			logger.Warn("Unknown storage adapter type",
				zap.String("id", endpoint.ID),
				zap.String("type", endpoint.Type),
			)
			continue
		}

		if err != nil {
			logger.Warn("Failed to create storage adapter",
				zap.String("id", endpoint.ID),
				zap.String("type", endpoint.Type),
				zap.Error(err),
			)
			continue
		}

		// Register adapter with orchestrator
		if err := storageOrch.RegisterEndpoint(endpoint.ID, adapter, endpoint.Role); err != nil {
			logger.Warn("Failed to register storage endpoint",
				zap.String("id", endpoint.ID),
				zap.Error(err),
			)
			continue
		}

		logger.Info("Storage endpoint registered successfully",
			zap.String("id", endpoint.ID),
			zap.String("type", endpoint.Type),
			zap.String("role", endpoint.Role),
		)
	}

	logger.Info("Storage orchestrator initialized",
		zap.Int("endpoints", len(cfg.Storage.Endpoints)),
	)

	// Start storage health monitor
	go storageOrch.StartHealthMonitor(context.Background(), 30*time.Second)

	// Initialize deduplication engine
	deduplicationEngine := deduplication.NewEngine(db, storageOrch, logger)

	// Initialize reference counter
	refCounter := reference.NewCounter(db, logger)

	// Initialize rate limiter
	limiterConfig := &ratelimit.LimiterConfig{
		EnableIPRateLimit:       true,
		IPRequestsPerSecond:     cfg.Security.RateLimiting.PerIPRequestsPerMinute / 60,
		IPBurstSize:             cfg.Security.RateLimiting.PerIPRequestsPerMinute / 10,
		EnableUserRateLimit:     true,
		UserRequestsPerSecond:   cfg.Security.RateLimiting.PerUserRequestsPerMinute / 60,
		UserBurstSize:           cfg.Security.RateLimiting.PerUserRequestsPerMinute / 10,
		EnableGlobalRateLimit:   true,
		GlobalRequestsPerSecond: cfg.Security.RateLimiting.GlobalRequestsPerMinute / 60,
		GlobalBurstSize:         cfg.Security.RateLimiting.GlobalRequestsPerMinute / 10,
	}
	rateLimiter := ratelimit.NewLimiter(limiterConfig, logger)

	// Service discovery registration (declare early for use in handlers)
	var serviceRegistry *utils.ServiceRegistry

	// Initialize HTTP server with Gin
	if cfg.Service.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestSize(cfg.Security.MaxFileSizeMB * 1024 * 1024))
	router.Use(middleware.RateLimiter(rateLimiter))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		healthStatus := getHealthStatus(db, storageOrch, securityScanner)
		if healthStatus["status"] == "healthy" {
			c.JSON(http.StatusOK, healthStatus)
		} else {
			c.JSON(http.StatusServiceUnavailable, healthStatus)
		}
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Authentication middleware (JWT validation)
		v1.Use(middleware.JWTAuth(cfg.Security.JWTSecret, logger))

		// File operations
		handlers.RegisterFileHandlers(v1, &handlers.FileHandlerDeps{
			DB:                  db,
			DeduplicationEngine: deduplicationEngine,
			RefCounter:          refCounter,
			SecurityScanner:     securityScanner,
			StorageOrch:         storageOrch,
			Config:              cfg,
			Logger:              logger,
		})

		// Metadata operations
		handlers.RegisterMetadataHandlers(v1, &handlers.MetadataHandlerDeps{
			DB:         db,
			RefCounter: refCounter,
			Logger:     logger,
		})

		// Admin operations (restricted)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			handlers.RegisterAdminHandlers(admin, &handlers.AdminHandlerDeps{
				DB:              db,
				StorageOrch:     storageOrch,
				RefCounter:      refCounter,
				RateLimiter:     rateLimiter,
				ServiceRegistry: serviceRegistry,
				Logger:          logger,
			})
		}
	}

	// Metrics endpoint (Prometheus)
	router.GET("/metrics", gin.WrapH(utils.PrometheusHandler()))

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Service.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Service.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Service.MaxHeaderBytes,
	}

	// Service discovery registration (if enabled)
	if cfg.Service.Discovery.Enabled {
		serviceRegistry, err = utils.NewServiceRegistry(
			cfg.Service.Discovery.Provider,
			cfg.Service.Discovery.ConsulAddress,
			serviceName,
			port,
			logger,
		)
		if err != nil {
			logger.Warn("Failed to initialize service registry",
				zap.Error(err),
			)
		} else {
			if err := serviceRegistry.Register(); err != nil {
				logger.Warn("Failed to register service",
					zap.Error(err),
				)
			} else {
				logger.Info("Service registered successfully",
					zap.String("provider", cfg.Service.Discovery.Provider),
				)
			}
		}
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server",
			zap.Int("port", port),
			zap.String("environment", cfg.Service.Environment),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Deregister from service discovery
	if serviceRegistry != nil {
		if err := serviceRegistry.Deregister(); err != nil {
			logger.Error("Failed to deregister service", zap.Error(err))
		} else {
			logger.Info("Service deregistered successfully")
		}
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited successfully")
}

// findAvailablePort attempts to find an available port in the given range
func findAvailablePort(preferredPort int, portRange []int) (int, error) {
	// Try preferred port first
	if isPortAvailable(preferredPort) {
		return preferredPort, nil
	}

	// Try port range if provided
	if len(portRange) >= 2 {
		for port := portRange[0]; port <= portRange[1]; port++ {
			if isPortAvailable(port) {
				return port, nil
			}
		}
	}

	return 0, fmt.Errorf("no available port found in range %d-%d", portRange[0], portRange[1])
}

// isPortAvailable checks if a port is available for binding
func isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// getHealthStatus returns the health status of the service and its dependencies
func getHealthStatus(db database.Database, storageOrch *orchestrator.Orchestrator, scanner *scanner.Scanner) map[string]interface{} {
	status := map[string]interface{}{
		"service": serviceName,
		"version": serviceVersion,
		"status":  "healthy",
		"checks":  make(map[string]interface{}),
	}

	checks := status["checks"].(map[string]interface{})

	// Check database
	dbStart := time.Now()
	if err := db.Ping(); err != nil {
		checks["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		status["status"] = "unhealthy"
	} else {
		checks["database"] = map[string]interface{}{
			"status":     "healthy",
			"latency_ms": time.Since(dbStart).Milliseconds(),
		}
	}

	// Check storage endpoints
	endpointHealths := storageOrch.GetEndpointHealth()
	primaryHealthy := false
	for _, health := range endpointHealths {
		if health.Role == "primary" && health.Status == "healthy" {
			primaryHealthy = true
		}
		checks[fmt.Sprintf("storage_%s", health.EndpointID)] = map[string]interface{}{
			"status":     health.Status,
			"latency_ms": health.LatencyMs,
		}
	}

	if !primaryHealthy {
		status["status"] = "degraded"
	}

	// Check virus scanner
	if scanner.IsEnabled() {
		if err := scanner.Ping(context.Background()); err != nil {
			checks["virus_scanner"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			// Virus scanner failure is not critical
			if status["status"] == "healthy" {
				status["status"] = "degraded"
			}
		} else {
			checks["virus_scanner"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		checks["virus_scanner"] = map[string]interface{}{
			"status": "disabled",
		}
	}

	return status
}

// parseS3Config parses S3 adapter configuration from map
func parseS3Config(cfg map[string]interface{}) (*adapters.S3Config, error) {
	s3Cfg := &adapters.S3Config{}

	// Required fields
	if bucket, ok := cfg["bucket"].(string); ok {
		s3Cfg.Bucket = bucket
	} else {
		return nil, fmt.Errorf("bucket is required for S3 adapter")
	}

	// Optional but recommended fields
	if region, ok := cfg["region"].(string); ok {
		s3Cfg.Region = region
	}

	if accessKey, ok := cfg["access_key_id"].(string); ok {
		s3Cfg.AccessKeyID = accessKey
	}

	if secretKey, ok := cfg["secret_access_key"].(string); ok {
		s3Cfg.SecretAccessKey = secretKey
	}

	if sessionToken, ok := cfg["session_token"].(string); ok {
		s3Cfg.SessionToken = sessionToken
	}

	if endpoint, ok := cfg["endpoint"].(string); ok {
		s3Cfg.Endpoint = endpoint
	}

	if prefix, ok := cfg["prefix"].(string); ok {
		s3Cfg.Prefix = prefix
	}

	if usePathStyle, ok := cfg["use_path_style"].(bool); ok {
		s3Cfg.UsePathStyle = usePathStyle
	}

	if disableSSL, ok := cfg["disable_ssl"].(bool); ok {
		s3Cfg.DisableSSL = disableSSL
	}

	return s3Cfg, nil
}

// parseMinIOConfig parses MinIO adapter configuration from map
func parseMinIOConfig(cfg map[string]interface{}) (*adapters.MinIOConfig, error) {
	minioCfg := &adapters.MinIOConfig{}

	// Required fields
	if endpoint, ok := cfg["endpoint"].(string); ok {
		minioCfg.Endpoint = endpoint
	} else {
		return nil, fmt.Errorf("endpoint is required for MinIO adapter")
	}

	if bucket, ok := cfg["bucket"].(string); ok {
		minioCfg.Bucket = bucket
	} else {
		return nil, fmt.Errorf("bucket is required for MinIO adapter")
	}

	if accessKey, ok := cfg["access_key_id"].(string); ok {
		minioCfg.AccessKeyID = accessKey
	} else {
		return nil, fmt.Errorf("access_key_id is required for MinIO adapter")
	}

	if secretKey, ok := cfg["secret_access_key"].(string); ok {
		minioCfg.SecretAccessKey = secretKey
	} else {
		return nil, fmt.Errorf("secret_access_key is required for MinIO adapter")
	}

	// Optional fields
	if useSSL, ok := cfg["use_ssl"].(bool); ok {
		minioCfg.UseSSL = useSSL
	}

	if prefix, ok := cfg["prefix"].(string); ok {
		minioCfg.Prefix = prefix
	}

	if storageClass, ok := cfg["storage_class"].(string); ok {
		minioCfg.StorageClass = storageClass
	}

	return minioCfg, nil
}
