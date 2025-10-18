#!/bin/bash
# E2E Test: Documents V2 - Basic Workflow
# Tests: Space creation, document creation, reading, modification, deletion
#
# NOTE: Requires database implementation fixes to be completed.
#       See: DOCUMENTS_V2_DATABASE_ISSUES.md
#
# Workflow:
# 1. Create a space
# 2. Create a document in the space
# 3. Read the document
# 4. Modify the document
# 5. List documents in space
# 6. Delete document (soft delete)
# 7. Archive space

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-test-jwt-token}"
TIMESTAMP=$(date +%s)
SPACE_KEY="TEST${TIMESTAMP}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_step() {
    echo -e "${YELLOW}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

check_response() {
    local response="$1"
    local step_name="$2"

    if echo "$response" | grep -q '"errorCode":-1'; then
        print_success "$step_name - Success"
        return 0
    else
        print_error "$step_name - Failed"
        echo "Response: $response"
        return 1
    fi
}

extract_id() {
    local response="$1"
    local field="$2"
    echo "$response" | grep -o "\"${field}\":\"[^\"]*\"" | cut -d'"' -f4
}

# Start test
print_header "Documents V2 - Basic Workflow Test"

echo "Configuration:"
echo "  Base URL: $BASE_URL"
echo "  Space Key: $SPACE_KEY"
echo ""

# Step 1: Create Space
print_step "Step 1: Create space '$SPACE_KEY'"
SPACE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentSpaceCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"key\": \"$SPACE_KEY\",
      \"name\": \"Test Documentation Space\",
      \"description\": \"E2E test space for Documents V2\",
      \"isPublic\": true
    }
  }")

check_response "$SPACE_RESPONSE" "Create Space"
SPACE_ID=$(extract_id "$SPACE_RESPONSE" "id")
echo "  Space ID: $SPACE_ID"

# Step 2: Create Document
print_step "Step 2: Create document in space"
DOC_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"title\": \"Getting Started Guide\",
      \"spaceID\": \"$SPACE_ID\",
      \"typeID\": \"type-page\",
      \"content\": {
        \"contentHTML\": \"<h1>Getting Started</h1><p>Welcome to our documentation!</p>\",
        \"contentMarkdown\": \"# Getting Started\\n\\nWelcome to our documentation!\",
        \"contentPlainText\": \"Getting Started\\n\\nWelcome to our documentation!\"
      }
    }
  }")

check_response "$DOC_RESPONSE" "Create Document"
DOC_ID=$(extract_id "$DOC_RESPONSE" "id")
echo "  Document ID: $DOC_ID"

# Step 3: Read Document
print_step "Step 3: Read document"
READ_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentRead\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\"
    }
  }")

check_response "$READ_RESPONSE" "Read Document"
echo "  Title: $(extract_id "$READ_RESPONSE" "title")"
echo "  Version: $(echo "$READ_RESPONSE" | grep -o '"version":[0-9]*' | cut -d':' -f2)"

# Step 4: Modify Document
print_step "Step 4: Modify document"
MODIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentModify\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"version\": 1,
      \"content\": {
        \"contentHTML\": \"<h1>Getting Started</h1><p>Welcome to our documentation!</p><h2>Updated</h2><p>This section was added.</p>\",
        \"contentMarkdown\": \"# Getting Started\\n\\nWelcome to our documentation!\\n\\n## Updated\\n\\nThis section was added.\",
        \"contentPlainText\": \"Getting Started\\n\\nWelcome to our documentation!\\n\\nUpdated\\n\\nThis section was added.\"
      }
    }
  }")

check_response "$MODIFY_RESPONSE" "Modify Document"
echo "  New Version: $(echo "$MODIFY_RESPONSE" | grep -o '"version":[0-9]*' | cut -d':' -f2)"

# Step 5: List Documents in Space
print_step "Step 5: List documents in space"
LIST_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentList\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"spaceID\": \"$SPACE_ID\"
    }
  }")

check_response "$LIST_RESPONSE" "List Documents"
DOC_COUNT=$(echo "$LIST_RESPONSE" | grep -o '"id":"doc-' | wc -l)
echo "  Documents found: $DOC_COUNT"

# Step 6: Create Child Document
print_step "Step 6: Create child document (hierarchy)"
CHILD_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"title\": \"Installation\",
      \"parentDocumentID\": \"$DOC_ID\",
      \"spaceID\": \"$SPACE_ID\",
      \"typeID\": \"type-page\",
      \"content\": {
        \"contentHTML\": \"<h1>Installation</h1><p>Follow these steps...</p>\"
      }
    }
  }")

check_response "$CHILD_RESPONSE" "Create Child Document"
CHILD_ID=$(extract_id "$CHILD_RESPONSE" "id")
echo "  Child Document ID: $CHILD_ID"

# Step 7: Delete Document (Soft Delete)
print_step "Step 7: Delete child document (soft delete)"
DELETE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentRemove\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$CHILD_ID\"
    }
  }")

check_response "$DELETE_RESPONSE" "Delete Document"

# Step 8: Verify Soft Delete
print_step "Step 8: Verify document is soft-deleted"
READ_DELETED_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentRead\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$CHILD_ID\",
      \"includeDeleted\": false
    }
  }")

if echo "$READ_DELETED_RESPONSE" | grep -q '"errorCode":3001'; then
    print_success "Soft Delete Verified - Document not found (as expected)"
else
    print_error "Soft Delete Verification Failed - Document should not be readable"
fi

# Step 9: Archive Space
print_step "Step 9: Archive space"
ARCHIVE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentSpaceArchive\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"spaceID\": \"$SPACE_ID\"
    }
  }")

check_response "$ARCHIVE_RESPONSE" "Archive Space"

# Summary
print_header "Test Summary"
echo -e "${GREEN}✓ All basic workflow tests completed successfully!${NC}"
echo ""
echo "Tests performed:"
echo "  1. Space creation"
echo "  2. Document creation"
echo "  3. Document reading"
echo "  4. Document modification"
echo "  5. Document listing"
echo "  6. Child document creation (hierarchy)"
echo "  7. Document deletion (soft delete)"
echo "  8. Soft delete verification"
echo "  9. Space archival"
echo ""
echo "Resources created:"
echo "  Space ID: $SPACE_ID"
echo "  Document ID: $DOC_ID"
echo "  Child Document ID: $CHILD_ID (deleted)"
echo ""

print_success "Basic Workflow Test Complete"
exit 0
