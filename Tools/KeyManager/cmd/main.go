package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/helixtrack/keymanager/internal/generator"
	"github.com/helixtrack/keymanager/internal/storage"
)

const (
	version = "1.0.0"
)

func main() {
	// Command-line flags
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	rotateCmd := flag.NewFlagSet("rotate", flag.ExitOnError)
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	importCmd := flag.NewFlagSet("import", flag.ExitOnError)

	// Generate command flags
	keyType := generateCmd.String("type", "", "Type of key to generate (jwt, db, tls, redis, api)")
	keyName := generateCmd.String("name", "", "Name/identifier for the key")
	service := generateCmd.String("service", "", "Service name (e.g., localization, authentication)")
	length := generateCmd.Int("length", 0, "Key length in bytes (for jwt, db, api keys)")
	outputFile := generateCmd.String("output", "", "Output file path (optional)")

	// Rotate command flags
	rotateKeyName := rotateCmd.String("name", "", "Name of key to rotate")
	rotateService := rotateCmd.String("service", "", "Service name")

	// Export command flags
	exportPath := exportCmd.String("path", "", "Export directory path")
	exportFormat := exportCmd.String("format", "json", "Export format (json, env, yaml)")

	// Import command flags
	importPath := importCmd.String("path", "", "Import file path")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		if err := handleGenerate(*keyType, *keyName, *service, *length, *outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "list":
		listCmd.Parse(os.Args[2:])
		if err := handleList(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "rotate":
		rotateCmd.Parse(os.Args[2:])
		if err := handleRotate(*rotateKeyName, *rotateService); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "export":
		exportCmd.Parse(os.Args[2:])
		if err := handleExport(*exportPath, *exportFormat); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "import":
		importCmd.Parse(os.Args[2:])
		if err := handleImport(*importPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "version":
		fmt.Printf("HelixTrack Key Manager v%s\n", version)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("HelixTrack Key Manager - Secure key generation and management for all Core services")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  keymanager <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate    Generate a new key")
	fmt.Println("  list        List all managed keys")
	fmt.Println("  rotate      Rotate an existing key")
	fmt.Println("  export      Export keys to file")
	fmt.Println("  import      Import keys from file")
	fmt.Println("  version     Show version information")
	fmt.Println()
	fmt.Println("Generate Options:")
	fmt.Println("  -type string       Type of key (jwt, db, tls, redis, api)")
	fmt.Println("  -name string       Key name/identifier")
	fmt.Println("  -service string    Service name")
	fmt.Println("  -length int        Key length in bytes (default: type-specific)")
	fmt.Println("  -output string     Output file path (optional)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate JWT secret for authentication service")
	fmt.Println("  keymanager generate -type jwt -name auth-jwt-secret -service authentication")
	fmt.Println()
	fmt.Println("  # Generate database encryption key for localization service")
	fmt.Println("  keymanager generate -type db -name loc-db-key -service localization -length 32")
	fmt.Println()
	fmt.Println("  # Generate TLS certificates")
	fmt.Println("  keymanager generate -type tls -name loc-tls -service localization")
	fmt.Println()
	fmt.Println("  # List all keys")
	fmt.Println("  keymanager list")
	fmt.Println()
	fmt.Println("  # Rotate a key")
	fmt.Println("  keymanager rotate -name auth-jwt-secret -service authentication")
	fmt.Println()
	fmt.Println("  # Export keys to JSON")
	fmt.Println("  keymanager export -path ./keys -format json")
}

func handleGenerate(keyType, name, service string, length int, outputFile string) error {
	if keyType == "" {
		return fmt.Errorf("key type is required (-type)")
	}
	if name == "" {
		return fmt.Errorf("key name is required (-name)")
	}
	if service == "" {
		return fmt.Errorf("service name is required (-service)")
	}

	// Initialize key generator
	gen := generator.New()

	var key *generator.Key
	var err error

	switch keyType {
	case "jwt":
		if length == 0 {
			length = 64 // Default JWT secret length
		}
		key, err = gen.GenerateJWTSecret(name, service, length)

	case "db":
		if length == 0 {
			length = 32 // Default database encryption key length (256-bit)
		}
		key, err = gen.GenerateDatabaseKey(name, service, length)

	case "tls":
		key, err = gen.GenerateTLSCertificate(name, service)

	case "redis":
		if length == 0 {
			length = 32
		}
		key, err = gen.GenerateRedisPassword(name, service, length)

	case "api":
		if length == 0 {
			length = 32
		}
		key, err = gen.GenerateAPIKey(name, service, length)

	default:
		return fmt.Errorf("unsupported key type: %s (supported: jwt, db, tls, redis, api)", keyType)
	}

	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Store key
	store, err := storage.New()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	if err := store.SaveKey(key); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	// Output key information
	fmt.Println("✓ Key generated successfully!")
	fmt.Printf("  Type:    %s\n", key.Type)
	fmt.Printf("  Name:    %s\n", key.Name)
	fmt.Printf("  Service: %s\n", key.Service)
	fmt.Printf("  ID:      %s\n", key.ID)

	if keyType == "tls" {
		fmt.Printf("  Cert:    %s\n", key.Metadata["cert_path"])
		fmt.Printf("  Key:     %s\n", key.Metadata["key_path"])
	} else {
		fmt.Printf("  Value:   %s\n", key.Value)
	}

	// Write to output file if specified
	if outputFile != "" {
		if err := store.ExportKeyToFile(key, outputFile); err != nil {
			return fmt.Errorf("failed to export key to file: %w", err)
		}
		fmt.Printf("  Exported to: %s\n", outputFile)
	}

	return nil
}

func handleList() error {
	store, err := storage.New()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	keys, err := store.ListKeys()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	if len(keys) == 0 {
		fmt.Println("No keys found")
		return nil
	}

	fmt.Printf("Found %d key(s):\n\n", len(keys))
	fmt.Printf("%-20s %-15s %-20s %-30s %s\n", "NAME", "TYPE", "SERVICE", "ID", "CREATED")
	fmt.Println("---------------------------------------------------------------------------------------------------")

	for _, key := range keys {
		fmt.Printf("%-20s %-15s %-20s %-30s %s\n",
			truncate(key.Name, 20),
			key.Type,
			truncate(key.Service, 20),
			truncate(key.ID, 30),
			key.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	return nil
}

func handleRotate(name, service string) error {
	if name == "" {
		return fmt.Errorf("key name is required (-name)")
	}
	if service == "" {
		return fmt.Errorf("service name is required (-service)")
	}

	store, err := storage.New()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	key, err := store.GetKey(name, service)
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	gen := generator.New()
	newKey, err := gen.RotateKey(key)
	if err != nil {
		return fmt.Errorf("failed to rotate key: %w", err)
	}

	if err := store.SaveKey(newKey); err != nil {
		return fmt.Errorf("failed to save rotated key: %w", err)
	}

	fmt.Println("✓ Key rotated successfully!")
	fmt.Printf("  Name:       %s\n", newKey.Name)
	fmt.Printf("  Service:    %s\n", newKey.Service)
	fmt.Printf("  Old ID:     %s\n", key.ID)
	fmt.Printf("  New ID:     %s\n", newKey.ID)
	fmt.Printf("  New Value:  %s\n", newKey.Value)

	return nil
}

func handleExport(path, format string) error {
	if path == "" {
		return fmt.Errorf("export path is required (-path)")
	}

	store, err := storage.New()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	if err := store.ExportKeys(path, format); err != nil {
		return fmt.Errorf("failed to export keys: %w", err)
	}

	fmt.Printf("✓ Keys exported successfully to: %s\n", path)
	return nil
}

func handleImport(path string) error {
	if path == "" {
		return fmt.Errorf("import path is required (-path)")
	}

	store, err := storage.New()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	count, err := store.ImportKeys(path)
	if err != nil {
		return fmt.Errorf("failed to import keys: %w", err)
	}

	fmt.Printf("✓ Imported %d key(s) successfully\n", count)
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
