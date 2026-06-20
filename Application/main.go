package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
	"helixtrack.ru/core/internal/server"
)

const (
	defaultConfigPath = "Configurations/default.json"
	version           = "1.0.0"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", defaultConfigPath, "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	spaceRoot := flag.String("space-root", "", "Path to space data root directory for per-space data isolation")
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("Helix Track Core v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Set version
	cfg.Version = version

	// ----- Space root resolution (Phase 1 multi-space data isolation) -----
	effectiveSpaceRoot := *spaceRoot
	if effectiveSpaceRoot == "" {
		effectiveSpaceRoot = cfg.SpaceRoot
	}
	if effectiveSpaceRoot == "" {
		var wdErr error
		effectiveSpaceRoot, wdErr = os.Getwd()
		if wdErr != nil {
			logger.Fatal("Failed to determine working directory", zap.Error(wdErr))
		}
	}
	logger.Info("Space root resolved", zap.String("space_root", effectiveSpaceRoot))

	// Load or create space config (onboarding wizard for new spaces)
	spaceCfg, cfgErr := config.LoadSpaceConfig(effectiveSpaceRoot)
	if cfgErr != nil {
		logger.Fatal("Failed to load space config", zap.Error(cfgErr))
	}
	cfg.SpaceConfig = spaceCfg

	// When space-root is explicitly provided (flag or config), use per-space database
	if *spaceRoot != "" || cfg.SpaceRoot != "" {
		cfg.Database.SQLitePath = filepath.Join(effectiveSpaceRoot, "data", "helixtrack.db")
		dataDir := filepath.Dir(cfg.Database.SQLitePath)
		if mkErr := os.MkdirAll(dataDir, 0755); mkErr != nil {
			logger.Fatal("Failed to create space data directory", zap.Error(mkErr))
		}
		logger.Info("Per-space database path set", zap.String("path", cfg.Database.SQLitePath))
	}

	// Initialize logger
	if err := logger.Initialize(cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Infof("Starting Helix Track Core v%s", version)
	logger.Infof("Configuration loaded from: %s", *configPath)

	// ----- First-start database initialization (space has no existing DB) -----
	dbPath := cfg.Database.SQLitePath
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		logger.Info("No existing database -- running DDL migrations", zap.String("path", dbPath))
		ddlDir := filepath.Join(effectiveSpaceRoot, "../Database/DDL")
		// Fallback to relative path for built binary
		if _, ddlStatErr := os.Stat(ddlDir); os.IsNotExist(ddlStatErr) {
			ddlDir = "Database/DDL"
		}
		migrator := database.NewMigrator(ddlDir)
		if migErr := migrator.RunMigrations(cfg.Database); migErr != nil {
			logger.Fatal("DDL migration failed", zap.Error(migErr))
		}
		logger.Info("DDL migrations completed successfully")
	}

	// Create server
	srv, err := server.NewServer(cfg)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info("Server exited successfully")
}
