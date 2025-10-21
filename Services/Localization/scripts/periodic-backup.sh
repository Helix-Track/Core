#!/bin/bash

###############################################################################
# HelixTrack Localization Service - Periodic Backup Script
#
# This script runs periodic backups of the localization database.
# Designed to be run as a cron job.
#
# Backup Schedule (recommended):
#   - Hourly: Incremental backups (only recent changes)
#   - Daily: Full backups
#   - Weekly: Full backups with compression and archival
#
# Crontab Examples:
#   # Hourly incremental backup
#   0 * * * * /path/to/periodic-backup.sh hourly
#
#   # Daily full backup (2 AM)
#   0 2 * * * /path/to/periodic-backup.sh daily
#
#   # Weekly full backup (Sunday 3 AM)
#   0 3 * * 0 /path/to/periodic-backup.sh weekly
#
# Usage:
#   ./periodic-backup.sh [backup_type]
#
# Arguments:
#   backup_type: hourly|daily|weekly (default: daily)
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
BACKUP_TYPE="${1:-daily}"
BACKUP_BASE_DIR="${BACKUP_DIR:-$PROJECT_DIR/backups}"
TIMESTAMP=$(date +%Y-%m-%d-%H-%M-%S)
DATE_ONLY=$(date +%Y-%m-%d)
WEEK_NUMBER=$(date +%Y-W%V)

# Retention periods (in days)
RETENTION_HOURLY=3
RETENTION_DAILY=30
RETENTION_WEEKLY=365

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

# Create backup directories
setup_backup_dirs() {
    mkdir -p "$BACKUP_BASE_DIR/hourly"
    mkdir -p "$BACKUP_BASE_DIR/daily"
    mkdir -p "$BACKUP_BASE_DIR/weekly"
    mkdir -p "$BACKUP_BASE_DIR/on-demand"
}

# Perform hourly incremental backup
backup_hourly() {
    local output_dir="$BACKUP_BASE_DIR/hourly/$TIMESTAMP-incremental"

    log_info "Starting hourly incremental backup..."

    # Export only changes from the last hour
    # Note: This is a placeholder - actual implementation would query database for recent changes
    COMPRESS=false "$SCRIPT_DIR/export-to-seed.sh" "$output_dir" > /dev/null 2>&1

    # Add incremental marker
    echo "incremental" > "$output_dir/backup-type.txt"
    echo "since: $(date -d '1 hour ago' -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$output_dir/backup-type.txt"

    log_success "Hourly backup completed: $output_dir"
}

# Perform daily full backup
backup_daily() {
    local output_dir="$BACKUP_BASE_DIR/daily/$DATE_ONLY-full"

    log_info "Starting daily full backup..."

    # Full export
    COMPRESS=true "$SCRIPT_DIR/export-to-seed.sh" "$output_dir" > /dev/null 2>&1

    log_success "Daily backup completed: $output_dir"
}

# Perform weekly full backup with archival
backup_weekly() {
    local output_dir="$BACKUP_BASE_DIR/weekly/$WEEK_NUMBER-full"

    log_info "Starting weekly full backup..."

    # Full export with compression
    COMPRESS=true "$SCRIPT_DIR/export-to-seed.sh" "$output_dir" > /dev/null 2>&1

    # Create version tag
    cd "$PROJECT_DIR"
    if [ -d ".git" ]; then
        git rev-parse --short HEAD > "$output_dir/git-commit.txt" 2>/dev/null || true
    fi

    log_success "Weekly backup completed: $output_dir"
}

# Clean up old backups based on retention policy
cleanup_old_backups() {
    log_info "Cleaning up old backups..."

    # Cleanup hourly backups (older than 3 days)
    find "$BACKUP_BASE_DIR/hourly" -type d -mtime +$RETENTION_HOURLY -exec rm -rf {} + 2>/dev/null || true
    find "$BACKUP_BASE_DIR/hourly" -type f -mtime +$RETENTION_HOURLY -delete 2>/dev/null || true

    # Cleanup daily backups (older than 30 days)
    find "$BACKUP_BASE_DIR/daily" -type d -mtime +$RETENTION_DAILY -exec rm -rf {} + 2>/dev/null || true
    find "$BACKUP_BASE_DIR/daily" -type f -mtime +$RETENTION_DAILY -delete 2>/dev/null || true

    # Cleanup weekly backups (older than 365 days)
    find "$BACKUP_BASE_DIR/weekly" -type d -mtime +$RETENTION_WEEKLY -exec rm -rf {} + 2>/dev/null || true
    find "$BACKUP_BASE_DIR/weekly" -type f -mtime +$RETENTION_WEEKLY -delete 2>/dev/null || true

    log_success "Cleanup completed"
}

# Calculate backup statistics
calculate_stats() {
    local total_size=$(du -sh "$BACKUP_BASE_DIR" 2>/dev/null | cut -f1 || echo "0")
    local hourly_count=$(find "$BACKUP_BASE_DIR/hourly" -mindepth 1 -maxdepth 1 2>/dev/null | wc -l)
    local daily_count=$(find "$BACKUP_BASE_DIR/daily" -mindepth 1 -maxdepth 1 2>/dev/null | wc -l)
    local weekly_count=$(find "$BACKUP_BASE_DIR/weekly" -mindepth 1 -maxdepth 1 2>/dev/null | wc -l)

    log_info "Backup statistics:"
    log_info "  Total size: $total_size"
    log_info "  Hourly backups: $hourly_count"
    log_info "  Daily backups: $daily_count"
    log_info "  Weekly backups: $weekly_count"
}

# Send notification (optional)
send_notification() {
    local status="$1"
    local message="$2"

    # Placeholder for notification system
    # Could integrate with email, Slack, etc.

    if [ -n "$NOTIFICATION_WEBHOOK" ]; then
        curl -X POST "$NOTIFICATION_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{\"status\": \"$status\", \"message\": \"$message\", \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"}" \
            > /dev/null 2>&1 || true
    fi
}

# Main function
main() {
    log_info "HelixTrack Localization Service - Periodic Backup"
    log_info "Backup Type: $BACKUP_TYPE"
    log_info "Timestamp: $TIMESTAMP"
    echo ""

    # Setup directories
    setup_backup_dirs

    # Perform backup based on type
    case "$BACKUP_TYPE" in
        hourly)
            backup_hourly
            ;;
        daily)
            backup_daily
            ;;
        weekly)
            backup_weekly
            ;;
        *)
            log_error "Invalid backup type: $BACKUP_TYPE"
            log_info "Valid types: hourly, daily, weekly"
            exit 1
            ;;
    esac

    # Cleanup old backups
    cleanup_old_backups

    # Show statistics
    echo ""
    calculate_stats

    # Send success notification
    send_notification "success" "Localization backup ($BACKUP_TYPE) completed successfully"

    echo ""
    log_success "Backup process completed!"
}

# Error handler
trap 'log_error "Backup failed!"; send_notification "error" "Localization backup ($BACKUP_TYPE) failed"; exit 1' ERR

# Run main function
main "$@"
