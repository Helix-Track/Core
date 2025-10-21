#!/bin/bash

# HelixTrack Localization Service - Build Script

set -e

echo "🔨 Building HelixTrack Localization Service..."

# Navigate to project root
cd "$(dirname "$0")/.."

# Clean previous builds
echo "Cleaning previous builds..."
rm -f localization-service
rm -rf bin/

# Create bin directory
mkdir -p bin

# Build for current platform
echo "Building for current platform..."
go build -o bin/localization-service cmd/main.go

# Build for multiple platforms (optional)
if [ "$1" == "--all" ]; then
    echo "Building for all platforms..."

    # Linux AMD64
    GOOS=linux GOARCH=amd64 go build -o bin/localization-service-linux-amd64 cmd/main.go

    # Linux ARM64
    GOOS=linux GOARCH=arm64 go build -o bin/localization-service-linux-arm64 cmd/main.go

    # macOS AMD64
    GOOS=darwin GOARCH=amd64 go build -o bin/localization-service-darwin-amd64 cmd/main.go

    # macOS ARM64 (Apple Silicon)
    GOOS=darwin GOARCH=arm64 go build -o bin/localization-service-darwin-arm64 cmd/main.go

    # Windows AMD64
    GOOS=windows GOARCH=amd64 go build -o bin/localization-service-windows-amd64.exe cmd/main.go

    echo "✅ Multi-platform build complete"
else
    echo "✅ Build complete: bin/localization-service"
fi

echo ""
echo "To build for all platforms, run: $0 --all"
