#!/bin/bash

# Build script for HelixTrack Chat Service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Build information
VERSION="${VERSION:-1.0.0}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${GREEN}Building HelixTrack Chat Service${NC}"
echo "Version: $VERSION"
echo "Build Time: $BUILD_TIME"
echo "Git Commit: $GIT_COMMIT"
echo ""

cd "$PROJECT_DIR"

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -f htChat

# Download dependencies
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod download

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
go test ./... -short

# Build the binary
echo -e "${YELLOW}Building binary...${NC}"
go build \
  -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
  -o htChat \
  main.go

# Make binary executable
chmod +x htChat

echo ""
echo -e "${GREEN}Build successful!${NC}"
echo -e "Binary: ${GREEN}./htChat${NC}"
echo ""
echo "Run with: ./htChat --config=configs/dev.json"
