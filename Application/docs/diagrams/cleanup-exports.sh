#!/bin/bash

# HelixTrack Core - PNG Export Cleanup Script
# This script reorganizes PNG files exported by DrawIO Docker tool
# The tool creates directories instead of files, this script fixes that

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================="
echo "Cleaning up PNG export structure"
echo "========================================="
echo ""

# Function to move PNG from directory to file
cleanup_png() {
    local base_name="$1"
    local png_dir="${base_name}.png"

    # Find the actual PNG file inside the directory
    local png_file=$(find "$png_dir" -name "*.png" -type f 2>/dev/null | head -n 1)

    if [ -n "$png_file" ]; then
        echo "Processing: $base_name"
        # Create a temporary file
        cp "$png_file" "${base_name}_temp.png"
        # Remove the directory
        rm -rf "$png_dir"
        # Rename temp file to final name
        mv "${base_name}_temp.png" "${base_name}.png"
        # Fix ownership to current user
        chown $(whoami):$(whoami) "${base_name}.png" 2>/dev/null || true
        echo "  ✓ Created: ${base_name}.png"
    else
        echo "  ✗ No PNG found in: $png_dir"
    fi
}

# Clean up all exported diagrams
cleanup_png "01-system-architecture"
cleanup_png "02-database-schema-overview"
cleanup_png "03-api-request-flow"
cleanup_png "04-auth-permissions-flow"
cleanup_png "05-microservices-interaction"

echo ""
echo "========================================="
echo "Cleanup Complete!"
echo "========================================="
echo ""

# Verify all PNGs exist
echo "Verification:"
for diagram in "01-system-architecture" "02-database-schema-overview" "03-api-request-flow" "04-auth-permissions-flow" "05-microservices-interaction"; do
    if [ -f "${diagram}.png" ]; then
        SIZE=$(du -h "${diagram}.png" | cut -f1)
        echo "  ✓ ${diagram}.png ($SIZE)"
    else
        echo "  ✗ ${diagram}.png (MISSING)"
    fi
done

echo ""
echo "All PNG files ready for use!"
