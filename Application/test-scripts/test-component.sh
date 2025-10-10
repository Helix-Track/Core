#!/bin/bash
# Component Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Component Management API"
echo "=================================================="

# Create Component
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"componentCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"API\", \"description\": \"REST API component\"}}" | jq '.'

# List Components
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"componentList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Component tests completed"
