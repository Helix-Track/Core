package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helixtrack/localization-service/internal/cache"
	"github.com/helixtrack/localization-service/internal/config"
	"github.com/helixtrack/localization-service/internal/database"
	"github.com/helixtrack/localization-service/internal/handlers"
	"github.com/helixtrack/localization-service/internal/middleware"
	"github.com/helixtrack/localization-service/internal/utils"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"
)

const (
	serviceName    = "localization-service"
	serviceVersion = "1.0.0"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "configs/default.json", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger, err := utils.NewLogger("info", "json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting HelixTrack Localization Service",
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

	// Find available port
	port, err := utils.FindAvailablePort(cfg.Service.Port, cfg.Service.PortRange)
	if err != nil {
		logger.Fatal("Failed to find available port", zap.Error(err))
	}

	cfg.Service.Port = port
	logger.Info("Port selected", zap.Int("port", port))

	// Initialize database connection
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database connection established",
		zap.String("driver", cfg.Database.Driver),
		zap.String("host", cfg.Database.Host),
	)

	// Initialize cache
	var cacheInstance cache.Cache

	if cfg.Cache.Redis.Enabled {
		// Use Redis cache
		redisCache, err := cache.NewRedisCache(
			cfg.Cache.Redis.Addresses,
			cfg.Cache.Redis.Password,
			cfg.Cache.Redis.Database,
			time.Duration(cfg.Cache.Redis.DefaultTTL)*time.Second,
			cfg.Cache.Redis.PoolSize,
			cfg.Cache.Redis.MaxRetries,
			logger,
		)
		if err != nil {
			logger.Warn("Failed to initialize Redis cache, falling back to in-memory",
				zap.Error(err),
			)
			// Fall back to memory cache
			cacheInstance = cache.NewMemoryCache(
				cfg.Cache.InMemory.MaxSizeMB,
				time.Duration(cfg.Cache.InMemory.DefaultTTL)*time.Second,
				time.Duration(cfg.Cache.InMemory.CleanupInterval)*time.Second,
				logger,
			)
		} else {
			cacheInstance = redisCache
			logger.Info("Redis cache initialized")
		}
	} else {
		// Use in-memory cache
		cacheInstance = cache.NewMemoryCache(
			cfg.Cache.InMemory.MaxSizeMB,
			time.Duration(cfg.Cache.InMemory.DefaultTTL)*time.Second,
			time.Duration(cfg.Cache.InMemory.CleanupInterval)*time.Second,
			logger,
		)
		logger.Info("In-memory cache initialized",
			zap.Int("max_size_mb", cfg.Cache.InMemory.MaxSizeMB),
		)
	}
	defer cacheInstance.Close()

	// Initialize HTTP server with Gin
	if cfg.Service.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.CORS())

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(
		cfg.Security.RateLimiting.PerIPRequestsPerMinute,
		cfg.Security.RateLimiting.PerUserRequestsPerMinute,
		cfg.Security.RateLimiting.GlobalRequestsPerMinute,
	)
	defer rateLimiter.Close()

	router.Use(rateLimiter.RateLimit())

	// Initialize handlers
	handler := handlers.NewHandler(db, cacheInstance, logger)
	handler.RegisterRoutes(router, cfg.Security.JWTSecret, cfg.Security.AdminRoles)

	// Create TLS configuration for HTTP/3 QUIC
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		NextProtos: []string{"h3"},  // HTTP/3 protocol identifier
	}

	// Create HTTP/3 server (QUIC)
	srv := &http3.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	// Certificate paths for TLS (required for HTTP/3)
	certFile := cfg.Service.TLSCertFile
	keyFile := cfg.Service.TLSKeyFile

	// Validate certificate files exist
	if certFile == "" || keyFile == "" {
		logger.Fatal("TLS certificate and key files are required for HTTP/3",
			zap.String("cert_file", certFile),
			zap.String("key_file", keyFile),
		)
	}

	// Service discovery registration (if enabled)
	var serviceRegistry *utils.ServiceRegistry
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

	// Start HTTP/3 QUIC server in goroutine
	go func() {
		logger.Info("Starting HTTP/3 QUIC server",
			zap.Int("port", port),
			zap.String("environment", cfg.Service.Environment),
			zap.String("protocol", "HTTP/3 (QUIC)"),
			zap.String("cert_file", certFile),
		)

		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP/3 server", zap.Error(err))
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

	// Graceful shutdown
	// Note: http3.Server.Close() does not accept context like http.Server.Shutdown()
	if err := srv.Close(); err != nil {
		logger.Error("HTTP/3 server forced to shutdown", zap.Error(err))
	}

	logger.Info("HTTP/3 server exited successfully")
}
