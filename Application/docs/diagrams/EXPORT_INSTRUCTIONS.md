# PNG Export Instructions

The DrawIO diagrams need to be exported to PNG format for use in documentation and websites.

## Automated Export (Recommended)

### Option 1: Using DrawIO Desktop (Easiest)

1. **Install DrawIO Desktop:**
   - Download from: https://github.com/jgraph/drawio-desktop/releases
   - Install for your OS (Windows/Mac/Linux)

2. **Export each diagram:**
   - Open DrawIO Desktop
   - Open each `.drawio` file
   - Go to: **File → Export as → PNG**
   - Settings:
     - **Zoom:** 300%
     - **Transparent Background:** ✓ Checked
     - **Border Width:** 10px
     - **Selection Only:** ✗ Unchecked
   - Save as the corresponding `.png` filename
   - Repeat for all 5 diagrams

3. **Verify:**
   - All 5 PNG files should be in `docs/diagrams/` directory
   - File sizes should be 500KB - 2MB each
   - Open each PNG to verify quality

### Option 2: Using draw.io Online

1. **Go to:** https://app.diagrams.net/
2. **For each diagram:**
   - Open the `.drawio` file
   - **File → Export as → PNG**
   - Settings: Same as Option 1
   - Download and save to `docs/diagrams/` with correct filename

### Option 3: Using Docker (Advanced)

```bash
cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams

# Export each diagram
for file in *.drawio; do
    docker run --rm -v $(pwd):/data \
        rlespinasse/drawio-export:latest \
        --format png --scale 3 --transparent --border 10 \
        --output "/data/${file%.drawio}.png" \
        "/data/$file"
done
```

## Manual Export Checklist

- [ ] 01-system-architecture.png (Expected size: ~1.5MB)
- [ ] 02-database-schema-overview.png (Expected size: ~2MB, largest)
- [ ] 03-api-request-flow.png (Expected size: ~1.2MB)
- [ ] 04-auth-permissions-flow.png (Expected size: ~1MB)
- [ ] 05-microservices-interaction.png (Expected size: ~1.3MB)

## After Export

Once all PNG files are created:

1. **Verify quality:**
   ```bash
   cd /home/milosvasic/Projects/HelixTrack/Core/Application/docs/diagrams
   ls -lh *.png
   file *.png
   ```

2. **Update git:**
   ```bash
   git add *.png
   git commit -m "Add architecture diagram PNG exports"
   ```

3. **Test in documentation:**
   - Open README.md
   - Verify PNG references work
   - Check image quality and readability

## Current Status

**DrawIO Files:** ✅ All 5 created and ready
**PNG Files:** ⏳ Awaiting export (manual step required)

**Next:** Export PNGs using Option 1 (DrawIO Desktop) - takes ~5 minutes total
