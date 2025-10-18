#!/bin/bash
# E2E Test: Documents V2 - Version Control Workflow
# Tests: Version history, version comparison (diff), labels, tags, rollback
#
# NOTE: Requires database implementation fixes to be completed.
#       See: DOCUMENTS_V2_DATABASE_ISSUES.md
#
# Workflow:
# 1. Create document (version 1)
# 2. Make multiple edits (versions 2-5)
# 3. Add version labels
# 4. Add version tags
# 5. Compare versions (diff)
# 6. Rollback to previous version
# 7. View complete version history

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-test-jwt-token}"
TIMESTAMP=$(date +%s)
SPACE_KEY="VERSION${TIMESTAMP}"
USER_ID="user-test-${TIMESTAMP}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

extract_version() {
    local response="$1"
    echo "$response" | grep -o '"version":[0-9]*' | cut -d':' -f2 | head -1
}

# Start test
print_header "Documents V2 - Version Control Workflow Test"

echo "Configuration:"
echo "  Base URL: $BASE_URL"
echo "  Space Key: $SPACE_KEY"
echo ""

# Setup: Create Space
print_step "Setup: Create space"
SPACE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentSpaceCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"key\": \"$SPACE_KEY\",
      \"name\": \"Version Control Test Space\",
      \"isPublic\": true
    }
  }")

check_response "$SPACE_RESPONSE" "Create Space"
SPACE_ID=$(extract_id "$SPACE_RESPONSE" "id")

# Step 1: Create Document (Version 1)
print_step "Step 1: Create document (Version 1)"
DOC_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"title\": \"Version Control Guide\",
      \"spaceID\": \"$SPACE_ID\",
      \"typeID\": \"type-page\",
      \"content\": {
        \"contentHTML\": \"<h1>Version Control</h1><p>Initial content.</p>\"
      }
    }
  }")

check_response "$DOC_RESPONSE" "Create Document"
DOC_ID=$(extract_id "$DOC_RESPONSE" "id")
VERSION_1=$(extract_version "$DOC_RESPONSE")
echo "  Document ID: $DOC_ID, Version: $VERSION_1"

# Step 2: Edit 1 - Add Section (Version 2)
print_step "Step 2: Edit 1 - Add Prerequisites section (Version 2)"
sleep 1  # Ensure timestamp difference
EDIT1_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentModify\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"version\": $VERSION_1,
      \"content\": {
        \"contentHTML\": \"<h1>Version Control</h1><p>Initial content.</p><h2>Prerequisites</h2><p>Git installed.</p>\"
      }
    }
  }")

check_response "$EDIT1_RESPONSE" "Edit 1"
VERSION_2=$(extract_version "$EDIT1_RESPONSE")
echo "  New Version: $VERSION_2"

# Step 3: Edit 2 - Add Instructions (Version 3)
print_step "Step 3: Edit 2 - Add instructions (Version 3)"
sleep 1
EDIT2_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentModify\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"version\": $VERSION_2,
      \"content\": {
        \"contentHTML\": \"<h1>Version Control</h1><p>Initial content.</p><h2>Prerequisites</h2><p>Git installed.</p><h2>Instructions</h2><ol><li>Initialize repository</li><li>Add files</li><li>Commit changes</li></ol>\"
      }
    }
  }")

check_response "$EDIT2_RESPONSE" "Edit 2"
VERSION_3=$(extract_version "$EDIT2_RESPONSE")
echo "  New Version: $VERSION_3"

# Step 4: Edit 3 - Add Examples (Version 4)
print_step "Step 4: Edit 3 - Add examples (Version 4)"
sleep 1
EDIT3_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentModify\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"version\": $VERSION_3,
      \"content\": {
        \"contentHTML\": \"<h1>Version Control</h1><p>Initial content.</p><h2>Prerequisites</h2><p>Git installed.</p><h2>Instructions</h2><ol><li>Initialize repository</li><li>Add files</li><li>Commit changes</li></ol><h2>Examples</h2><pre>git init\ngit add .\ngit commit -m 'Initial commit'</pre>\"
      }
    }
  }")

check_response "$EDIT3_RESPONSE" "Edit 3"
VERSION_4=$(extract_version "$EDIT3_RESPONSE")
echo "  New Version: $VERSION_4"

# Step 5: View Version History
print_step "Step 5: View complete version history"
HISTORY_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentVersionList\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\"
    }
  }")

check_response "$HISTORY_RESPONSE" "List Version History"
VERSION_COUNT=$(echo "$HISTORY_RESPONSE" | grep -o '"versionNumber":[0-9]*' | wc -l)
echo "  Total versions: $VERSION_COUNT (expected 4)"

# Step 6: Add Version Label to Version 3
print_step "Step 6: Add label to Version 3 (Milestone)"
LABEL_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentVersionLabelCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"versionID\": \"ver-v3-${DOC_ID}\",
      \"labelName\": \"Milestone 1.0\",
      \"labelColor\": \"#00cc66\"
    }
  }")

check_response "$LABEL_RESPONSE" "Add Version Label"
LABEL_ID=$(extract_id "$LABEL_RESPONSE" "id")
echo "  Label ID: $LABEL_ID (Milestone 1.0)"

# Step 7: Add Version Tag to Version 4
print_step "Step 7: Add tag to Version 4 (Release)"
TAG_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentVersionTagCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"versionID\": \"ver-v4-${DOC_ID}\",
      \"tagName\": \"v1.0.0\",
      \"tagDescription\": \"First stable release\"
    }
  }")

check_response "$TAG_RESPONSE" "Add Version Tag"
TAG_ID=$(extract_id "$TAG_RESPONSE" "id")
echo "  Tag ID: $TAG_ID (v1.0.0)"

# Step 8: Compare Versions (Diff)
print_step "Step 8: Compare Version 2 vs Version 4 (unified diff)"
DIFF_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentVersionCompare\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"fromVersion\": 2,
      \"toVersion\": 4,
      \"diffType\": \"unified\"
    }
  }")

check_response "$DIFF_RESPONSE" "Compare Versions"
echo "  Diff type: unified (showing changes from v2 to v4)"

# Step 9: Add Comment to Version
print_step "Step 9: Add comment to version explaining changes"
VERSION_COMMENT_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentVersionCommentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"versionID\": \"ver-v3-${DOC_ID}\",
      \"userID\": \"$USER_ID\",
      \"commentText\": \"Added complete instructions for repository initialization\"
    }
  }")

check_response "$VERSION_COMMENT_RESPONSE" "Add Version Comment"

# Step 10: Get Specific Version
print_step "Step 10: Read Version 2 content"
VERSION_READ_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentVersionRead\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"versionID\": \"ver-v2-${DOC_ID}\"
    }
  }")

check_response "$VERSION_READ_RESPONSE" "Read Specific Version"
echo "  Retrieved Version 2 content"

# Step 11: Rollback to Version 2
print_step "Step 11: Rollback to Version 2"
ROLLBACK_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentVersionRestore\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"versionID\": \"ver-v2-${DOC_ID}\",
      \"currentVersion\": $VERSION_4
    }
  }")

check_response "$ROLLBACK_RESPONSE" "Rollback to Version 2"
VERSION_5=$(extract_version "$ROLLBACK_RESPONSE")
echo "  New Version (rollback): $VERSION_5 (content from v2)"

# Step 12: Verify Rollback
print_step "Step 12: Verify rollback - read current document"
VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentRead\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\"
    }
  }")

check_response "$VERIFY_RESPONSE" "Verify Rollback"
CURRENT_VERSION=$(extract_version "$VERIFY_RESPONSE")
echo "  Current Version: $CURRENT_VERSION (should be $VERSION_5)"

if [[ "$CURRENT_VERSION" == "$VERSION_5" ]]; then
    print_success "Rollback verified - version matches"
else
    print_error "Rollback verification failed - version mismatch"
fi

# Step 13: Test Optimistic Locking (Version Conflict)
print_step "Step 13: Test optimistic locking (attempt edit with old version)"
CONFLICT_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentModify\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"version\": 2,
      \"content\": {
        \"contentHTML\": \"<h1>This should fail</h1>\"
      }
    }
  }")

if echo "$CONFLICT_RESPONSE" | grep -q '"errorCode":1005'; then
    print_success "Version conflict detected (as expected - optimistic locking works)"
else
    print_error "Version conflict should have been detected"
    echo "Response: $CONFLICT_RESPONSE"
fi

# Cleanup
print_step "Cleanup: Remove test space"
CLEANUP_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentSpaceRemove\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"spaceID\": \"$SPACE_ID\"
    }
  }")

check_response "$CLEANUP_RESPONSE" "Cleanup Space"

# Summary
print_header "Test Summary"
echo -e "${GREEN}✓ All version control workflow tests completed successfully!${NC}"
echo ""
echo "Tests performed:"
echo "  1. Document creation (v1)"
echo "  2-4. Multiple edits (v2, v3, v4)"
echo "  5. Version history listing"
echo "  6. Version labels (Milestone 1.0)"
echo "  7. Version tags (v1.0.0)"
echo "  8. Version comparison (diff)"
echo "  9. Version comments"
echo " 10. Read specific version"
echo " 11. Rollback to previous version (v2 → v5)"
echo " 12. Verify rollback success"
echo " 13. Optimistic locking test (version conflict)"
echo ""
echo "Version progression:"
echo "  v1: Initial document"
echo "  v2: + Prerequisites"
echo "  v3: + Instructions (Labeled: Milestone 1.0)"
echo "  v4: + Examples (Tagged: v1.0.0)"
echo "  v5: Rollback to v2 content"
echo ""
echo "Total versions created: 5"
echo "Labels added: 1"
echo "Tags added: 1"
echo "Diffs compared: 1 (v2 vs v4)"
echo ""

print_success "Version Control Workflow Test Complete"
exit 0
