#!/bin/bash
# Run all API tests

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================="
echo "Running all API tests"
echo "========================================="
echo ""

# Public endpoints (no authentication required)
echo "=== Public Endpoints ==="
bash "${SCRIPT_DIR}/test-version.sh"
bash "${SCRIPT_DIR}/test-jwt-capable.sh"
bash "${SCRIPT_DIR}/test-db-capable.sh"
bash "${SCRIPT_DIR}/test-health.sh"

# Authentication endpoint
echo "=== Authentication ==="
bash "${SCRIPT_DIR}/test-authenticate.sh"

# Protected endpoints (require JWT)
echo "=== Protected Endpoints ==="
echo "Note: These require a valid JWT token"
echo "Set JWT_TOKEN environment variable before running:"
echo "  export JWT_TOKEN=your-token-here"
echo ""

if [ -n "$JWT_TOKEN" ]; then
  bash "${SCRIPT_DIR}/test-create.sh"
else
  echo "Skipping protected endpoints (JWT_TOKEN not set)"
fi

echo ""
echo "========================================="
echo "All tests completed"
echo "========================================="
