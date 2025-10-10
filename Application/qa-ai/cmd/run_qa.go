package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"helixtrack.ru/core/qa-ai/config"
	"helixtrack.ru/core/qa-ai/orchestrator"
	"helixtrack.ru/core/qa-ai/reports"
)

func main() {
	// Parse command line flags
	suite := flag.String("suite", "", "Run specific test suite")
	_ = flag.String("profile", "", "Use specific user profile") // Reserved for future use
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	reportFormat := flag.String("report", "html", "Report format (html, json, markdown)")
	flag.Parse()

	// Load configuration
	cfg := config.DefaultQAConfig()
	if *verbose {
		cfg.VerboseLogging = true
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("HelixTrack QA-AI System Starting...")
	log.Printf("Configuration loaded: %+v", cfg)

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(cfg)

	// Initialize orchestrator
	if err := orch.Initialize(); err != nil {
		log.Fatalf("Failed to initialize orchestrator: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Run tests
	var err error
	if *suite != "" {
		log.Printf("Running specific suite: %s", *suite)
		err = orch.RunTestSuite(ctx, *suite)
	} else {
		log.Printf("Running all test suites")
		err = orch.RunAllTests(ctx)
	}

	if err != nil {
		log.Fatalf("Test execution failed: %v", err)
	}

	// Generate report
	if cfg.GenerateReport {
		log.Printf("Generating test report...")
		reporter := reports.NewReporter(cfg.ReportPath)

		switch *reportFormat {
		case "html":
			if err := reporter.GenerateHTMLReport(orch); err != nil {
				log.Printf("Failed to generate HTML report: %v", err)
			}
		case "json":
			if err := reporter.GenerateJSONReport(orch); err != nil {
				log.Printf("Failed to generate JSON report: %v", err)
			}
		case "markdown":
			if err := reporter.GenerateMarkdownReport(orch); err != nil {
				log.Printf("Failed to generate Markdown report: %v", err)
			}
		default:
			log.Printf("Unknown report format: %s", *reportFormat)
		}
	}

	// Exit with appropriate code
	successRate := orch.GetSuccessRate()
	if successRate < 100.0 {
		log.Printf("Tests completed with failures (Success Rate: %.2f%%)", successRate)
		os.Exit(1)
	}

	log.Printf("All tests passed successfully!")
	os.Exit(0)
}
