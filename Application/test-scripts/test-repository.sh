#!/bin/bash
# Repository Management API Test Script
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"

echo "Testing Repository Management API"
echo "=================================================="

# Create Repository
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"repositoryCreate\", \"jwt\": \"$JWT_TOKEN\", \"data\": {\"name\": \"core-repo\", \"url\": \"https://github.com/org/repo.git\", \"type\": \"git\"}}" | jq '.'

# List Repositories
curl -s -X POST "$BASE_URL/do" -H "Content-Type: application/json" \
  -d "{\"action\": \"repositoryList\", \"jwt\": \"$JWT_TOKEN\", \"data\": {}}" | jq '.'

echo "Repository tests completed"
