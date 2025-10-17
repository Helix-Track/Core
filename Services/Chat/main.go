package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"helixtrack.ru/chat/configs"
	"helixtrack.ru/chat/internal/database"
	"helixtrack.ru/chat/internal/logger"
	"helixtrack.ru/chat/internal/server"
	"helixtrack.ru/chat/internal/services"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "configs/dev.json", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("HelixTrack Chat Service\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Load configuration
	config, err := configs.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.InitLogger(&config.Logger); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting HelixTrack Chat Service",
		zap.String("version", version),
		zap.String("config", *configPath),
	)

	// Initialize database
	db, err := database.NewPostgresDB(&config.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Initialize Core service client
	coreService := services.NewCoreService(config.JWT.Issuer) // Use JWT issuer as base URL for now
	logger.Info("Core service client initialized")

	// Create server
	srv := server.NewServer(config, db, coreService)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	// Print routes
	logger.Info("Registered routes:")
	srv.PrintRoutes()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-errChan:
		logger.Error("Server error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	}

	// Graceful shutdown
	logger.Info("Initiating graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}
