#!/bin/bash

# Start localization service for testing with HTTP (no TLS/HTTP3)
# This script starts the service separately from docker compose for testing

cd "$(dirname "$0")/.."

# Build the service first
echo "Building localization service..."
go build -o localization-service-test ./cmd/main.go

# Kill any existing test service
pkill -f localization-service-test

# Wait a moment for process to die
sleep 2

# Start service in test mode with HTTP
echo "Starting localization service for testing (HTTP only)..."
./localization-service-test -config=configs/test.json &
TEST_PID=$!

echo "Test service started with PID: $TEST_PID"
echo "Service URL: http://localhost:8085"
echo "Health check: http://localhost:8085/health"

# Wait for service to be ready
echo "Waiting for service to be ready..."
sleep 5

# Check if service is responding
if curl -s http://localhost:8085/health > /dev/null; then
    echo "✓ Service is ready for testing"
else
    echo "✗ Service failed to start properly"
    kill $TEST_PID
    exit 1
fi

# Store PID for later cleanup
echo $TEST_PID > .test-service-pid
echo "Use 'kill \$(cat .test-service-pid)' or 'pkill -f localization-service-test' to stop"