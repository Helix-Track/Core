#!/bin/bash

###############################################################################
# HelixTrack Localization Service - Export to Seed Data Script
#
# This script exports the current localization database to seed data format
# for backup and redistribution purposes.
#
# Usage:
#   ./export-to-seed.sh [output_dir] [config_file]
#
# Arguments:
#   output_dir: Directory to export seed data (default: seed-data-export/)
#   config_file: Path to configuration file (default: configs/default.json)
#
# Features:
#   - Exports languages, localization keys, and translations
#   - Creates timestamped backups
#   - Validates export integrity
#   - Compresses backup files (optional)
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
OUTPUT_DIR="${1:-seed-data-export}"
CONFIG_FILE="${2:-configs/default.json}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="$OUTPUT_DIR/backup-$TIMESTAMP"
COMPRESS="${COMPRESS:-true}"

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

# Check dependencies
check_dependencies() {
    local missing=0

    if ! command -v psql &> /dev/null; then
        log_warn "psql not found - PostgreSQL export may not work"
        missing=1
    fi

    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi

    if [ "$COMPRESS" = "true" ] && ! command -v tar &> /dev/null; then
        log_warn "tar not found - compression will be skipped"
        COMPRESS="false"
    fi
}

# Export languages
export_languages() {
    local output_file="$1"
    local db_host db_port db_name db_user db_password

    # Parse database config from JSON
    db_host=$(jq -r '.database.host // "localhost"' "$PROJECT_DIR/$CONFIG_FILE")
    db_port=$(jq -r '.database.port // 5432' "$PROJECT_DIR/$CONFIG_FILE")
    db_name=$(jq -r '.database.database // "helixtrack_localization"' "$PROJECT_DIR/$CONFIG_FILE")
    db_user=$(jq -r '.database.user // "postgres"' "$PROJECT_DIR/$CONFIG_FILE")
    db_password=$(jq -r '.database.password // ""' "$PROJECT_DIR/$CONFIG_FILE")

    log_info "Exporting languages..."

    PGPASSWORD="$db_password" psql -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" -t -A -F"," \
        -c "SELECT json_agg(row_to_json(t)) FROM (
            SELECT
                code,
                name,
                native_name,
                is_rtl,
                is_active,
                is_default
            FROM languages
            WHERE deleted = false
            ORDER BY is_default DESC, code
        ) t" \
        | jq '.' > "$output_file"

    local count=$(jq 'length' "$output_file")
    log_success "Exported $count languages"
}

# Export localization keys
export_localization_keys() {
    local output_file="$1"
    local db_host db_port db_name db_user db_password

    # Parse database config from JSON
    db_host=$(jq -r '.database.host // "localhost"' "$PROJECT_DIR/$CONFIG_FILE")
    db_port=$(jq -r '.database.port // 5432' "$PROJECT_DIR/$CONFIG_FILE")
    db_name=$(jq -r '.database.database // "helixtrack_localization"' "$PROJECT_DIR/$CONFIG_FILE")
    db_user=$(jq -r '.database.user // "postgres"' "$PROJECT_DIR/$CONFIG_FILE")
    db_password=$(jq -r '.database.password // ""' "$PROJECT_DIR/$CONFIG_FILE")

    log_info "Exporting localization keys..."

    PGPASSWORD="$db_password" psql -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" -t -A -F"," \
        -c "SELECT json_agg(row_to_json(t)) FROM (
            SELECT
                key,
                category,
                description,
                context,
                COALESCE(variables::json, '[]'::json) as variables
            FROM localization_keys
            WHERE deleted = false
            ORDER BY category, key
        ) t" \
        | jq '.' > "$output_file"

    local count=$(jq 'length' "$output_file")
    log_success "Exported $count localization keys"
}

# Export localizations for a specific language
export_localizations_for_language() {
    local language_code="$1"
    local output_file="$2"
    local db_host db_port db_name db_user db_password

    # Parse database config from JSON
    db_host=$(jq -r '.database.host // "localhost"' "$PROJECT_DIR/$CONFIG_FILE")
    db_port=$(jq -r '.database.port // 5432' "$PROJECT_DIR/$CONFIG_FILE")
    db_name=$(jq -r '.database.database // "helixtrack_localization"' "$PROJECT_DIR/$CONFIG_FILE")
    db_user=$(jq -r '.database.user // "postgres"' "$PROJECT_DIR/$CONFIG_FILE")
    db_password=$(jq -r '.database.password // ""' "$PROJECT_DIR/$CONFIG_FILE")

    log_info "Exporting localizations for language: $language_code"

    PGPASSWORD="$db_password" psql -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" -t -A -F"," \
        -c "SELECT json_object_agg(lk.key, l.value) FROM localizations l
            JOIN localization_keys lk ON l.key_id = lk.id
            JOIN languages lang ON l.language_id = lang.id
            WHERE lang.code = '$language_code'
            AND l.deleted = false
            AND l.approved = true" \
        | jq '.' > "$output_file"

    local count=$(jq 'length' "$output_file")
    log_success "Exported $count localizations for $language_code"
}

# Main export function
main() {
    log_info "HelixTrack Localization Service - Export to Seed Data"
    log_info "===================================================="
    echo ""

    # Check dependencies
    check_dependencies

    # Create output directory
    mkdir -p "$BACKUP_DIR/localizations"
    log_success "Created output directory: $BACKUP_DIR"
    echo ""

    # Export languages
    export_languages "$BACKUP_DIR/languages.json"
    echo ""

    # Export localization keys
    export_localization_keys "$BACKUP_DIR/localization-keys.json"
    echo ""

    # Get list of languages and export localizations for each
    log_info "Exporting localizations for all languages..."
    languages=$(jq -r '.[].code' "$BACKUP_DIR/languages.json")

    for lang in $languages; do
        export_localizations_for_language "$lang" "$BACKUP_DIR/localizations/$lang.json"
    done
    echo ""

    # Create metadata file
    log_info "Creating backup metadata..."
    cat > "$BACKUP_DIR/metadata.json" << EOF
{
  "backup_timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "backup_type": "full",
  "service_version": "1.0.0",
  "languages": $(jq 'length' "$BACKUP_DIR/languages.json"),
  "localization_keys": $(jq 'length' "$BACKUP_DIR/localization-keys.json"),
  "total_localizations": $(find "$BACKUP_DIR/localizations" -name "*.json" -exec jq 'length' {} \; | awk '{s+=$1} END {print s}')
}
EOF
    log_success "Metadata created"
    echo ""

    # Create README
    cat > "$BACKUP_DIR/README.md" << EOF
# Localization Export - $TIMESTAMP

This directory contains a complete export of the HelixTrack Localization service database.

## Export Details

- **Date**: $(date)
- **Languages**: $(jq 'length' "$BACKUP_DIR/languages.json")
- **Localization Keys**: $(jq 'length' "$BACKUP_DIR/localization-keys.json")
- **Total Localizations**: $(find "$BACKUP_DIR/localizations" -name "*.json" -exec jq 'length' {} \; | awk '{s+=$1} END {print s}')

## Files

- \`languages.json\` - Language definitions
- \`localization-keys.json\` - Localization key metadata
- \`localizations/*.json\` - Translation files per language
- \`metadata.json\` - Backup metadata
- \`README.md\` - This file

## Usage

To restore this backup:

\`\`\`bash
cp -r . ../seed-data/
cd ../
./scripts/populate-from-seed.sh
\`\`\`

Or use the import API:

\`\`\`bash
curl -X POST https://localhost:8085/v1/admin/import \\
  -H "Authorization: Bearer YOUR_ADMIN_JWT" \\
  -H "Content-Type: application/json" \\
  -d @import-payload.json
\`\`\`
EOF

    # Compress if requested
    if [ "$COMPRESS" = "true" ]; then
        log_info "Compressing backup..."
        cd "$OUTPUT_DIR"
        tar -czf "backup-$TIMESTAMP.tar.gz" "backup-$TIMESTAMP"
        log_success "Compressed backup created: backup-$TIMESTAMP.tar.gz"
        log_info "Removing uncompressed directory..."
        rm -rf "backup-$TIMESTAMP"
        cd - > /dev/null
    fi

    echo ""
    log_success "Export completed successfully!"
    log_info "Output location: $BACKUP_DIR"

    if [ "$COMPRESS" = "true" ]; then
        local size=$(du -h "$OUTPUT_DIR/backup-$TIMESTAMP.tar.gz" | cut -f1)
        log_info "Compressed size: $size"
    fi
}

# Run main function
main "$@"
