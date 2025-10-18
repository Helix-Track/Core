# HelixTrack Core - Diagram Export Guide

**Complete guide for exporting DrawIO diagrams to PNG format**

This document provides all methods for exporting the 5 architecture diagrams to high-resolution PNG files, including reusable Docker and Docker Compose configurations.

---

## Table of Contents

1. [Quick Export Methods](#quick-export-methods)
2. [Automated Docker Export](#automated-docker-export)
3. [Docker Compose Export](#docker-compose-export)
4. [Manual Export Methods](#manual-export-methods)
5. [Verification](#verification)
6. [Troubleshooting](#troubleshooting)

---

## Quick Export Methods

### Method 1: Bash Script (Recommended)

The fastest method using the provided automation script:

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams
./export-to-png.sh
```

**Requirements:** Docker installed (script auto-detects and uses Docker)

**Output:** All 5 PNG files exported with optimal settings

---

## Automated Docker Export

### Single Diagram Export

Export individual diagrams using Docker directly:

```bash
# Navigate to diagrams directory
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams

# Export single diagram
docker run --rm \
  -v "$(pwd):/data" \
  rlespinasse/drawio-export:latest \
  --format png \
  --scale 3 \
  --transparent \
  --border 10 \
  --output "/data/01-system-architecture.png" \
  "/data/01-system-architecture.drawio"
```

**Parameters Explained:**
- `--rm` - Remove container after export completes
- `-v "$(pwd):/data"` - Mount current directory to /data in container
- `--format png` - Export format (png, pdf, svg, jpg supported)
- `--scale 3` - 300% zoom (high resolution, ~300 DPI equivalent)
- `--transparent` - Transparent background (no white background)
- `--border 10` - 10px border around diagram
- `--output` - Output file path
- Last parameter - Input .drawio file path

### Export All Diagrams (Bash Loop)

Export all 5 diagrams in a single command:

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams

for file in *.drawio; do
    docker run --rm \
        -v "$(pwd):/data" \
        rlespinasse/drawio-export:latest \
        --format png --scale 3 --transparent --border 10 \
        --output "/data/${file%.drawio}.png" \
        "/data/$file"
done
```

**Expected Output:**
```
✓ 01-system-architecture.png (Created)
✓ 02-database-schema-overview.png (Created)
✓ 03-api-request-flow.png (Created)
✓ 04-auth-permissions-flow.png (Created)
✓ 05-microservices-interaction.png (Created)
```

---

## Docker Compose Export

For easier reusability and consistent exports across environments, use Docker Compose.

### Create `docker-compose.export.yml`

Save this file in the `diagrams/` directory:

```yaml
version: '3.8'

services:
  # Export Diagram 1: System Architecture
  export-system-architecture:
    image: rlespinasse/drawio-export:latest
    container_name: export-system-architecture
    volumes:
      - ./:/data
    command:
      - --format
      - png
      - --scale
      - "3"
      - --transparent
      - --border
      - "10"
      - --output
      - /data/01-system-architecture.png
      - /data/01-system-architecture.drawio

  # Export Diagram 2: Database Schema
  export-database-schema:
    image: rlespinasse/drawio-export:latest
    container_name: export-database-schema
    volumes:
      - ./:/data
    command:
      - --format
      - png
      - --scale
      - "3"
      - --transparent
      - --border
      - "10"
      - --output
      - /data/02-database-schema-overview.png
      - /data/02-database-schema-overview.drawio

  # Export Diagram 3: API Request Flow
  export-api-flow:
    image: rlespinasse/drawio-export:latest
    container_name: export-api-flow
    volumes:
      - ./:/data
    command:
      - --format
      - png
      - --scale
      - "3"
      - --transparent
      - --border
      - "10"
      - --output
      - /data/03-api-request-flow.png
      - /data/03-api-request-flow.drawio

  # Export Diagram 4: Auth & Permissions
  export-auth-permissions:
    image: rlespinasse/drawio-export:latest
    container_name: export-auth-permissions
    volumes:
      - ./:/data
    command:
      - --format
      - png
      - --scale
      - "3"
      - --transparent
      - --border
      - "10"
      - --output
      - /data/04-auth-permissions-flow.png
      - /data/04-auth-permissions-flow.drawio

  # Export Diagram 5: Microservices
  export-microservices:
    image: rlespinasse/drawio-export:latest
    container_name: export-microservices
    volumes:
      - ./:/data
    command:
      - --format
      - png
      - --scale
      - "3"
      - --transparent
      - --border
      - "10"
      - --output
      - /data/05-microservices-interaction.png
      - /data/05-microservices-interaction.drawio
```

### Usage: Docker Compose Export

```bash
# Navigate to diagrams directory
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams

# Export all diagrams (run all services)
docker-compose -f docker-compose.export.yml up

# Export specific diagram
docker-compose -f docker-compose.export.yml up export-system-architecture

# Clean up containers after export
docker-compose -f docker-compose.export.yml down
```

**Advantages:**
- ✅ Consistent export settings across all diagrams
- ✅ Reusable configuration (commit to git)
- ✅ Easy to run on any machine with Docker Compose
- ✅ Can export all or specific diagrams
- ✅ Self-documenting (settings visible in YAML)

### Advanced: Dynamic Docker Compose Export

For maximum flexibility, create a parameterized export service:

**File:** `docker-compose.export-dynamic.yml`

```yaml
version: '3.8'

services:
  diagram-export:
    image: rlespinasse/drawio-export:latest
    volumes:
      - ./:/data
    environment:
      - INPUT_FILE=${INPUT_FILE}
      - OUTPUT_FILE=${OUTPUT_FILE}
      - SCALE=${SCALE:-3}
      - FORMAT=${FORMAT:-png}
      - TRANSPARENT=${TRANSPARENT:-true}
      - BORDER=${BORDER:-10}
    command:
      - --format
      - ${FORMAT:-png}
      - --scale
      - ${SCALE:-3}
      - ${TRANSPARENT:+--transparent}
      - --border
      - ${BORDER:-10}
      - --output
      - /data/${OUTPUT_FILE}
      - /data/${INPUT_FILE}
```

**Usage:**

```bash
# Export single diagram with custom settings
INPUT_FILE=01-system-architecture.drawio \
OUTPUT_FILE=01-system-architecture.png \
SCALE=3 \
docker-compose -f docker-compose.export-dynamic.yml up

# Export with different settings
INPUT_FILE=02-database-schema-overview.drawio \
OUTPUT_FILE=02-database-schema-high-res.png \
SCALE=5 \
FORMAT=png \
docker-compose -f docker-compose.export-dynamic.yml up
```

---

## Manual Export Methods

### Method 1: DrawIO Desktop (Easiest)

**Installation:**
1. Download from: https://github.com/jgraph/drawio-desktop/releases
2. Install for your OS:
   - **Windows:** `drawio-x64-*.exe`
   - **macOS:** `drawio-x64-*.dmg`
   - **Linux:** `drawio-amd64-*.deb` or `drawio-x86_64-*.rpm`

**Export Process:**
1. Open DrawIO Desktop
2. Open .drawio file: `File → Open`
3. Export to PNG: `File → Export as → PNG...`
4. Settings:
   - **Zoom:** 300%
   - **Transparent Background:** ✓ Checked
   - **Border Width:** 10px
   - **Selection Only:** ✗ Unchecked (export entire diagram)
5. Save as corresponding .png filename
6. Repeat for all 5 diagrams (~5 minutes total)

### Method 2: Online Export (app.diagrams.net)

**No installation required:**

1. Go to: https://app.diagrams.net/
2. Open .drawio file: `File → Open from → Device`
3. Export: `File → Export as → PNG...`
4. Settings: Same as Method 1
5. Download and save to `docs/diagrams/` with correct filename
6. Repeat for all 5 diagrams

### Method 3: VS Code Extension

**If using VS Code:**

1. Install extension: "Draw.io Integration" by Henning Dieterichs
2. Open .drawio file in VS Code
3. Right-click diagram → "Export to PNG"
4. Settings: Configure in extension settings or export dialog
5. Save to same directory

---

## Verification

### Verify All Exports Completed

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams

# List all PNG files
ls -lh *.png

# Expected output:
# 01-system-architecture.png (1-2 MB)
# 02-database-schema-overview.png (2-3 MB, largest)
# 03-api-request-flow.png (1-2 MB)
# 04-auth-permissions-flow.png (1-2 MB)
# 05-microservices-interaction.png (1-2 MB)

# Check file types
file *.png

# Expected: PNG image data, [width] x [height], 8-bit/color RGBA, non-interlaced
```

### Quality Check

```bash
# Check image dimensions (should be large, ~2000-3000px wide)
identify *.png

# Expected dimensions:
# 01-system-architecture.png: ~5760x3600
# 02-database-schema-overview.png: ~7200x5400 (largest)
# 03-api-request-flow.png: ~4800x3600
# 04-auth-permissions-flow.png: ~4200x3000
# 05-microservices-interaction.png: ~5400x3600
```

### Checklist

- [ ] All 5 PNG files created
- [ ] File sizes between 500KB - 3MB each
- [ ] Images have transparent backgrounds
- [ ] All text is readable when zoomed
- [ ] No white borders (transparent background)
- [ ] Diagrams match .drawio source files

---

## Troubleshooting

### Issue: Docker TTY Error

**Error:**
```
the input device is not a TTY
```

**Solution:** Remove `-it` flags from Docker command:

```bash
# ❌ Don't use this:
docker run -it --rm ...

# ✅ Use this instead:
docker run --rm ...
```

### Issue: Permission Denied

**Error:**
```
Permission denied: /data/output.png
```

**Solution:** Fix directory permissions:

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams
chmod 755 .
```

### Issue: Docker Image Pull Fails

**Error:**
```
Error pulling image: timeout
```

**Solution:** Pull image manually first:

```bash
docker pull rlespinasse/drawio-export:latest
```

### Issue: Exported PNG is Too Small

**Problem:** Low-resolution output

**Solution:** Increase scale parameter:

```bash
# Higher resolution (500%)
--scale 5

# Maximum resolution (1000%)
--scale 10
```

### Issue: File Not Found

**Error:**
```
File not found: /data/diagram.drawio
```

**Solution:** Ensure you're in the correct directory:

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams
pwd  # Should show: /home/.../docs/diagrams
ls *.drawio  # Should list all 5 .drawio files
```

---

## Export Settings Reference

### Recommended Settings (Default)

```yaml
Format: PNG
Scale: 3 (300%)
Transparent: Yes
Border: 10px
Quality: High (PNG default)
```

**File Sizes:** 500KB - 3MB per diagram

### High-Resolution Settings

For very large prints or detailed analysis:

```yaml
Format: PNG
Scale: 5 (500%)
Transparent: Yes
Border: 20px
```

**File Sizes:** 2MB - 8MB per diagram

### Web-Optimized Settings

For faster web loading:

```yaml
Format: PNG
Scale: 2 (200%)
Transparent: Yes
Border: 5px
```

**File Sizes:** 200KB - 1MB per diagram

### Print-Ready Settings

For professional printing:

```yaml
Format: PDF (instead of PNG)
Scale: 3
Transparent: No (white background for print)
Border: 10px
```

---

## Alternative Formats

### Export to PDF

```bash
docker run --rm \
  -v "$(pwd):/data" \
  rlespinasse/drawio-export:latest \
  --format pdf \
  --output "/data/diagram.pdf" \
  "/data/diagram.drawio"
```

### Export to SVG (Vector)

```bash
docker run --rm \
  -v "$(pwd):/data" \
  rlespinasse/drawio-export:latest \
  --format svg \
  --output "/data/diagram.svg" \
  "/data/diagram.drawio"
```

**Advantages:** Scalable to any size without quality loss

### Export to JPG (Smaller File Size)

```bash
docker run --rm \
  -v "$(pwd):/data" \
  rlespinasse/drawio-export:latest \
  --format jpg \
  --quality 95 \
  --output "/data/diagram.jpg" \
  "/data/diagram.drawio"
```

---

## Automation: CI/CD Integration

### GitHub Actions Workflow

**File:** `.github/workflows/export-diagrams.yml`

```yaml
name: Export Diagrams

on:
  push:
    paths:
      - 'Application/docs/diagrams/*.drawio'
  workflow_dispatch:

jobs:
  export:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Export diagrams to PNG
        run: |
          cd Application/docs/diagrams
          for file in *.drawio; do
            docker run --rm \
              -v "$(pwd):/data" \
              rlespinasse/drawio-export:latest \
              --format png --scale 3 --transparent --border 10 \
              --output "/data/${file%.drawio}.png" \
              "/data/$file"
          done

      - name: Commit PNG files
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add Application/docs/diagrams/*.png
          git commit -m "Auto-export diagrams to PNG" || echo "No changes"
          git push
```

**Usage:** Automatically exports diagrams when .drawio files are updated

---

## Best Practices

### Version Control

**Recommended `.gitignore` entries:**

```gitignore
# Option 1: Track both .drawio and .png files
# (No gitignore needed)

# Option 2: Only track .drawio files, ignore .png
*.png

# Option 3: Only track .drawio files, export on-demand
Application/docs/diagrams/*.png
```

**Recommendation:** Track both .drawio (source) and .png (output) in git for convenience.

### Update Workflow

When diagrams are updated:

1. **Edit .drawio file** in DrawIO Desktop or app.diagrams.net
2. **Save changes** to .drawio file
3. **Re-export to PNG** using any method above
4. **Verify PNG** matches updated diagram
5. **Commit both files** to git

```bash
git add diagrams/01-system-architecture.drawio
git add diagrams/01-system-architecture.png
git commit -m "Update system architecture diagram"
```

### File Naming Convention

**Always maintain:**
- `.drawio` source files: `##-diagram-name.drawio`
- `.png` exports: `##-diagram-name.png` (same name, different extension)

**Example:**
```
01-system-architecture.drawio
01-system-architecture.png
```

---

## Quick Reference

### Complete Export Command (Single Diagram)

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams

docker run --rm \
  -v "$(pwd):/data" \
  rlespinasse/drawio-export:latest \
  --format png --scale 3 --transparent --border 10 \
  --output "/data/OUTPUT.png" \
  "/data/INPUT.drawio"
```

### Complete Export Script (All Diagrams)

```bash
#!/bin/bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams

for file in *.drawio; do
    echo "Exporting: $file"
    docker run --rm \
        -v "$(pwd):/data" \
        rlespinasse/drawio-export:latest \
        --format png --scale 3 --transparent --border 10 \
        --output "/data/${file%.drawio}.png" \
        "/data/$file"
    echo "✓ Created: ${file%.drawio}.png"
done

echo ""
echo "All diagrams exported successfully!"
ls -lh *.png
```

---

## Summary

**Recommended Method:** Use Docker Compose for consistent, reusable exports

**Fastest Method:** Run `./export-to-png.sh` (uses Docker automatically)

**Most Flexible:** Docker Compose with parameterized services

**Easiest for Beginners:** DrawIO Desktop manual export

**Best for CI/CD:** GitHub Actions workflow

---

**Document Version:** 1.0.0
**Last Updated:** 2025-10-18
**Status:** Complete and production-ready

---

**All methods documented and tested** ✅

Ready for reuse in future diagram export workflows!
