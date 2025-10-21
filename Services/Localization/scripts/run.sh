#!/bin/bash

# HelixTrack Localization Service - Run Script

set -e

echo "🚀 Starting HelixTrack Localization Service..."

# Navigate to project root
cd "$(dirname "$0")/.."

# Check if binary exists
if [ ! -f "bin/localization-service" ]; then
    echo "Binary not found. Building first..."
    ./scripts/build.sh
fi

# Use config file if specified, otherwise use default
CONFIG_FILE="${1:-configs/default.json}"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Config file not found: $CONFIG_FILE"
    exit 1
fi

echo "Using configuration: $CONFIG_FILE"

# Run the service
./bin/localization-service --config="$CONFIG_FILE"
