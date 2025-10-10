#!/bin/bash
# Cycle/Sprint Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Cycle Management API"
echo "=================================================="

# Create Cycle
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"cycleCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"Sprint 1\", \"type\": \"sprint\", \"startDate\": 1696118400, \"endDate\": 1697328000}}" | jq '.'

# List Cycles
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"cycleList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Cycle tests completed"
