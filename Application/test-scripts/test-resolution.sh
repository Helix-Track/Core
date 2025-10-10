#!/bin/bash
# Resolution Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Resolution Management API"
echo "=================================================="

# Create Resolution
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"resolutionCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"title\": \"Done\", \"description\": \"Task completed successfully\"}}" | jq '.'

# List Resolutions
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"resolutionList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Resolution tests completed"
