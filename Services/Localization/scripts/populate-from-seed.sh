#!/bin/bash

###############################################################################
# HelixTrack Localization Service - Seed Data Population Script
#
# This script manually populates the localization service database with
# seed data from JSON files.
#
# Usage:
#   ./populate-from-seed.sh [config_file]
#
# Arguments:
#   config_file: Path to configuration file (default: configs/default.json)
#
# Environment Variables:
#   SEED_DATA_PATH: Path to seed data directory (default: seed-data/)
#   FORCE_SEED: Force seeding even if database has data (default: false)
###############################################################################

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
CONFIG_FILE="${1:-configs/default.json}"
SEED_DATA_PATH="${SEED_DATA_PATH:-seed-data/}"
FORCE_SEED="${FORCE_SEED:-false}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Main script
main() {
    log_info "HelixTrack Localization Service - Seed Data Population"
    log_info "================================================="
    echo ""

    # Check if seed data directory exists
    if [ ! -d "$PROJECT_DIR/$SEED_DATA_PATH" ]; then
        log_error "Seed data directory not found: $PROJECT_DIR/$SEED_DATA_PATH"
        exit 1
    fi

    # Check for required files
    log_info "Checking seed data files..."

    if [ ! -f "$PROJECT_DIR/$SEED_DATA_PATH/languages.json" ]; then
        log_error "Missing languages.json in seed data directory"
        exit 1
    fi

    if [ ! -f "$PROJECT_DIR/$SEED_DATA_PATH/localization-keys.json" ]; then
        log_error "Missing localization-keys.json in seed data directory"
        exit 1
    fi

    if [ ! -d "$PROJECT_DIR/$SEED_DATA_PATH/localizations" ]; then
        log_warn "Localizations directory not found, will create empty language entries"
    fi

    log_success "All required seed data files found"
    echo ""

    # Count files
    LANG_COUNT=$(jq length "$PROJECT_DIR/$SEED_DATA_PATH/languages.json")
    KEY_COUNT=$(jq length "$PROJECT_DIR/$SEED_DATA_PATH/localization-keys.json")

    log_info "Seed data summary:"
    log_info "  Languages: $LANG_COUNT"
    log_info "  Localization keys: $KEY_COUNT"

    if [ -d "$PROJECT_DIR/$SEED_DATA_PATH/localizations" ]; then
        LOCALIZATION_FILES=$(find "$PROJECT_DIR/$SEED_DATA_PATH/localizations" -name "*.json" | wc -l)
        log_info "  Translation files: $LOCALIZATION_FILES"
    fi
    echo ""

    # Check if database already has data
    log_info "Checking database status..."

    # TODO: Add database check here
    # For now, we'll rely on the Go application to check

    log_info "Starting localization service with seed population..."
    echo ""

    # Run the service with seed population flag
    cd "$PROJECT_DIR"

    # Build if needed
    if [ ! -f "htLocalization" ]; then
        log_info "Building localization service..."
        go build -o htLocalization cmd/main.go
        log_success "Build complete"
        echo ""
    fi

    # Run with environment variables
    export SEED_DATA_PATH="$SEED_DATA_PATH"
    export FORCE_SEED="$FORCE_SEED"

    log_info "Running service (will auto-seed on startup)..."
    ./htLocalization --config="$CONFIG_FILE"
}

# Run main function
main "$@"
