#!/bin/bash

# HelixTrack Core - Export DrawIO Diagrams to PNG
# This script exports all .drawio diagram files to high-resolution PNG images

echo "========================================"
echo "HelixTrack Core - Diagram Export Script"
echo "========================================"
echo ""

# Check if drawio is installed
if ! command -v drawio &> /dev/null && ! command -v draw.io &> /dev/null; then
    echo "ERROR: DrawIO CLI not found!"
    echo ""
    echo "Please install DrawIO desktop application with CLI support:"
    echo ""
    echo "Ubuntu/Debian:"
    echo "  wget https://github.com/jgraph/drawio-desktop/releases/download/v24.0.0/drawio-amd64-24.0.0.deb"
    echo "  sudo dpkg -i drawio-amd64-24.0.0.deb"
    echo "  sudo ln -s /opt/drawio/drawio /usr/local/bin/drawio"
    echo ""
    echo "Arch Linux:"
    echo "  yay -S drawio-desktop-bin"
    echo ""
    echo "macOS:"
    echo "  brew install --cask drawio"
    echo ""
    echo "Or download from: https://github.com/jgraph/drawio-desktop/releases"
    echo ""
    exit 1
fi

DRAWIO_CMD=$(command -v drawio || command -v draw.io)
echo "Using DrawIO: $DRAWIO_CMD"
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Export settings
EXPORT_FORMAT="png"
EXPORT_SCALE="2"  # 2x resolution for high quality
EXPORT_BORDER="10"  # 10px border

# Count total diagrams
TOTAL_DIAGRAMS=$(ls -1 *.drawio 2>/dev/null | wc -l)

if [ "$TOTAL_DIAGRAMS" -eq 0 ]; then
    echo "No .drawio files found in $SCRIPT_DIR"
    exit 1
fi

echo "Found $TOTAL_DIAGRAMS diagram(s) to export"
echo ""

# Export each diagram
COUNT=0
for DIAGRAM in *.drawio; do
    COUNT=$((COUNT + 1))
    PNG_FILE="${DIAGRAM%.drawio}.png"

    echo "[$COUNT/$TOTAL_DIAGRAMS] Exporting: $DIAGRAM"
    echo "           Output: $PNG_FILE"

    # Export using DrawIO CLI
    $DRAWIO_CMD --export --format "$EXPORT_FORMAT" --scale "$EXPORT_SCALE" --border "$EXPORT_BORDER" --output "$PNG_FILE" "$DIAGRAM" 2>&1 | grep -v "Gtk-Message"

    if [ -f "$PNG_FILE" ]; then
        FILE_SIZE=$(du -h "$PNG_FILE" | cut -f1)
        echo "           Success! Size: $FILE_SIZE"
    else
        echo "           ERROR: Export failed!"
    fi
    echo ""
done

echo "========================================"
echo "Export complete!"
echo "Exported $COUNT diagram(s) to PNG format"
echo "Location: $SCRIPT_DIR"
echo "========================================"

# Also copy to Website if it exists
WEBSITE_DIAGRAMS_DIR="$SCRIPT_DIR/../../../Website/docs/assets/diagrams"
if [ -d "$WEBSITE_DIAGRAMS_DIR" ]; then
    echo ""
    echo "Copying PNG files to Website..."
    cp -v *.png "$WEBSITE_DIAGRAMS_DIR/"
    echo "Website diagrams updated!"
fi

echo ""
echo "Done!"
