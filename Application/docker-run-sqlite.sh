#!/bin/bash
# Run HelixTrack Core with SQLite database

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "HelixTrack Core - SQLite Configuration"
echo "========================================"
echo ""

# Check if .env.sqlite exists
if [ ! -f ".env.sqlite" ]; then
    echo "Error: .env.sqlite not found"
    echo "Please copy .env.example to .env.sqlite and configure"
    exit 1
fi

# Create necessary directories
mkdir -p Database logs

# Build and start services
echo "Building and starting services..."
docker-compose -f docker-compose.yml up --build -d

echo ""
echo "Services started successfully!"
echo ""
echo "Core API:     http://localhost:8080"
echo "Metrics:      http://localhost:9090"
echo "Mock Auth:    http://localhost:8081"
echo "Mock Perm:    http://localhost:8082"
echo ""
echo "To view logs: docker-compose -f docker-compose.yml logs -f"
echo "To stop:      docker-compose -f docker-compose.yml down"
echo ""

# Wait for service to be healthy
echo "Waiting for service to be healthy..."
timeout=60
counter=0

while [ $counter -lt $timeout ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✓ Service is healthy!"
        break
    fi
    sleep 1
    counter=$((counter + 1))
    echo -n "."
done

if [ $counter -eq $timeout ]; then
    echo "✗ Service did not become healthy within ${timeout}s"
    echo "Check logs: docker-compose -f docker-compose.yml logs"
    exit 1
fi

echo ""
echo "✓ HelixTrack Core (SQLite) is ready!"
echo ""
