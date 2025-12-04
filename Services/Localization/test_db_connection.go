package main

import (
	"context"
	"fmt"
	"log"

	"github.com/helixtrack/localization-service/internal/config"
	"github.com/helixtrack/localization-service/internal/database"
	"go.uber.org/zap"
)

func main() {
	// Load test configuration
	cfg, err := config.Load("configs/test.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Connect to database
	db, err := database.New(&cfg.Database, &cfg.Encryption, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("✓ Database connection successful")

	// Test database operations
	ctx := context.Background()

	// Get language
	lang, err := db.GetLanguageByCode(ctx, "en")
	if err != nil {
		log.Fatalf("Failed to get language: %v", err)
	}
	fmt.Printf("✓ Found language: %s (ID: %s)\n", lang.Code, lang.ID)

	// Get localization key
	locKey, err := db.GetLocalizationKeyByKey(ctx, "ui.button.cancel")
	if err != nil {
		log.Fatalf("Failed to get localization key: %v", err)
	}
	fmt.Printf("✓ Found key: %s (ID: %s)\n", locKey.Key, locKey.ID)

	// Get localization
	loc, err := db.GetLocalizationByKeyAndLanguage(ctx, locKey.ID, lang.ID)
	if err != nil {
		log.Fatalf("Failed to get localization: %v", err)
	}
	fmt.Printf("✓ Found localization: %s = %s\n", locKey.Key, loc.Value)
}