#!/bin/bash

# HelixTrack Core - DrawIO to PNG Export Script
# This script exports all .drawio files to high-resolution PNG images

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================="
echo "HelixTrack Core - Diagram Export to PNG"
echo "========================================="
echo ""

# Check for DrawIO CLI
if command -v drawio &> /dev/null; then
    echo "✓ DrawIO CLI found"
    EXPORT_METHOD="cli"
elif command -v docker &> /dev/null; then
    echo "✓ Docker found, will use Docker-based export"
    EXPORT_METHOD="docker"
else
    echo "✗ Neither DrawIO CLI nor Docker found"
    echo ""
    echo "Please install one of the following:"
    echo ""
    echo "Option 1: DrawIO Desktop (Recommended)"
    echo "  Download from: https://github.com/jgraph/drawio-desktop/releases"
    echo "  After installation, use File → Export as → PNG"
    echo "  Settings: Scale 3x, Transparent background, Border width: 10"
    echo ""
    echo "Option 2: Docker (Automated)"
    echo "  Install Docker from: https://docs.docker.com/get-docker/"
    echo "  Then run this script again"
    echo ""
    echo "Option 3: Online Export"
    echo "  1. Open https://app.diagrams.net/"
    echo "  2. Open each .drawio file"
    echo "  3. File → Export as → PNG"
    echo "  4. Use these settings:"
    echo "     - Zoom: 300%"
    echo "     - Transparent Background: Yes"
    echo "     - Border Width: 10"
    echo ""
    exit 1
fi

echo ""
echo "Export Method: $EXPORT_METHOD"
echo ""

# List of diagrams to export
DIAGRAMS=(
    "01-system-architecture"
    "02-database-schema-overview"
    "03-api-request-flow"
    "04-auth-permissions-flow"
    "05-microservices-interaction"
)

# Export function using DrawIO CLI
export_cli() {
    local input="$1"
    local output="$2"
    echo "  Exporting: $input → $output"
    drawio --export --format png --scale 3 --transparent --border 10 \
        --output "$output" "$input"
}

# Export function using Docker
export_docker() {
    local input="$1"
    local output="$2"
    echo "  Exporting: $input → $output"
    docker run -it --rm \
        -v "$SCRIPT_DIR:/data" \
        rlespinasse/drawio-export:latest \
        --format png --scale 3 --transparent --border 10 \
        --output "/data/$(basename "$output")" \
        "/data/$(basename "$input")"
}

# Export all diagrams
echo "Exporting diagrams..."
echo ""

for diagram in "${DIAGRAMS[@]}"; do
    INPUT_FILE="${diagram}.drawio"
    OUTPUT_FILE="${diagram}.png"

    if [ ! -f "$INPUT_FILE" ]; then
        echo "  ✗ Skipping $INPUT_FILE (not found)"
        continue
    fi

    if [ "$EXPORT_METHOD" = "cli" ]; then
        export_cli "$INPUT_FILE" "$OUTPUT_FILE"
    elif [ "$EXPORT_METHOD" = "docker" ]; then
        export_docker "$INPUT_FILE" "$OUTPUT_FILE"
    fi

    if [ -f "$OUTPUT_FILE" ]; then
        SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
        echo "  ✓ Created: $OUTPUT_FILE ($SIZE)"
    else
        echo "  ✗ Failed to create: $OUTPUT_FILE"
    fi
    echo ""
done

echo "========================================="
echo "Export Complete!"
echo "========================================="
echo ""
echo "Generated PNG files:"
ls -lh *.png 2>/dev/null || echo "  (No PNG files found - export may have failed)"
echo ""

# Verify all exports
echo "Verification:"
MISSING=0
for diagram in "${DIAGRAMS[@]}"; do
    if [ -f "${diagram}.png" ]; then
        echo "  ✓ ${diagram}.png"
    else
        echo "  ✗ ${diagram}.png (MISSING)"
        MISSING=$((MISSING + 1))
    fi
done

echo ""
if [ $MISSING -eq 0 ]; then
    echo "✓ All diagrams exported successfully!"
    exit 0
else
    echo "✗ $MISSING diagram(s) failed to export"
    echo ""
    echo "Manual export instructions:"
    echo "  1. Open DrawIO desktop: https://github.com/jgraph/drawio-desktop/releases"
    echo "  2. Open each .drawio file"
    echo "  3. File → Export as → PNG"
    echo "  4. Settings: Scale 300%, Transparent background, Border 10"
    echo "  5. Save as the corresponding .png filename"
    exit 1
fi
