#!/bin/bash
# Run AI QA tests for HelixTrack Core with both databases

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "HelixTrack Core - AI QA Test Suite"
echo "========================================"
echo ""

# Create test results directory
mkdir -p test-results

# Stop any existing test containers
echo "Stopping existing test containers..."
docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true

echo ""
echo "Starting test environment..."
echo ""

# Start test databases
docker-compose -f docker-compose.test.yml up -d postgres-test

echo "Waiting for PostgreSQL test database to be ready..."
sleep 5

# Run SQLite tests
echo ""
echo "=========================================="
echo "Running tests with SQLite"
echo "=========================================="
echo ""

docker-compose -f docker-compose.test.yml --profile sqlite-test up --build -d helixtrack-core-sqlite-test

# Wait for service
sleep 10

# Run tests against SQLite
echo "Executing AI QA tests against SQLite..."
docker run --rm \
    --network helixtrack-test-network \
    -v "$(pwd)/tests/ai-qa:/tests" \
    -v "$(pwd)/test-results:/test-results" \
    -e SQLITE_API_URL=http://helixtrack-core-sqlite:8080 \
    -e POSTGRES_API_URL=http://helixtrack-core-sqlite:8080 \
    -e TEST_TIMEOUT=300 \
    -e VERBOSE=true \
    helixtrack-ai-qa-runner:latest \
    python3 /tests/run_all_tests.py

SQLITE_EXIT_CODE=$?

# Stop SQLite containers
docker-compose -f docker-compose.test.yml --profile sqlite-test down

# Run PostgreSQL tests
echo ""
echo "=========================================="
echo "Running tests with PostgreSQL"
echo "=========================================="
echo ""

docker-compose -f docker-compose.test.yml --profile postgres-test up --build -d helixtrack-core-postgres-test

# Wait for service
sleep 15

# Run tests against PostgreSQL
echo "Executing AI QA tests against PostgreSQL..."
docker run --rm \
    --network helixtrack-test-network \
    -v "$(pwd)/tests/ai-qa:/tests" \
    -v "$(pwd)/test-results:/test-results" \
    -e SQLITE_API_URL=http://helixtrack-core-postgres:8081 \
    -e POSTGRES_API_URL=http://helixtrack-core-postgres:8081 \
    -e TEST_TIMEOUT=300 \
    -e VERBOSE=true \
    helixtrack-ai-qa-runner:latest \
    python3 /tests/run_all_tests.py

POSTGRES_EXIT_CODE=$?

# Cleanup
echo ""
echo "Cleaning up test environment..."
docker-compose -f docker-compose.test.yml down -v

# Summary
echo ""
echo "========================================"
echo "Test Results Summary"
echo "========================================"
echo ""
echo "SQLite tests:     $([ $SQLITE_EXIT_CODE -eq 0 ] && echo '✓ PASSED' || echo '✗ FAILED')"
echo "PostgreSQL tests: $([ $POSTGRES_EXIT_CODE -eq 0 ] && echo '✓ PASSED' || echo '✗ FAILED')"
echo ""

if [ -f "test-results/ai-qa-report.json" ]; then
    echo "Detailed report: test-results/ai-qa-report.json"
    echo ""
fi

# Exit with error if any tests failed
if [ $SQLITE_EXIT_CODE -ne 0 ] || [ $POSTGRES_EXIT_CODE -ne 0 ]; then
    echo "✗ Some tests failed"
    exit 1
fi

echo "✓ All tests passed!"
exit 0
