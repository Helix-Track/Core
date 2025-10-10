#!/bin/bash
# Asset Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Asset Management API"
echo "=================================================="

# Create Asset
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"assetCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"filename\": \"screenshot.png\", \"mimeType\": \"image/png\", \"size\": 102400}}" | jq '.'

# List Assets
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"assetList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Asset tests completed"
