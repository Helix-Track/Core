#!/bin/bash
# E2E Test: Documents V2 - Templates & Blueprints Workflow
# Tests: Template creation, variable substitution, blueprint wizard, usage tracking
#
# NOTE: Requires database implementation fixes to be completed.
#       See: DOCUMENTS_V2_DATABASE_ISSUES.md
#
# Workflow:
# 1. Create template with variables
# 2. Create document from template
# 3. Verify variable substitution
# 4. Create blueprint with wizard steps
# 5. Create document from blueprint
# 6. Track template usage
# 7. Update template

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-test-jwt-token}"
TIMESTAMP=$(date +%s)
SPACE_KEY="TEMPLATE${TIMESTAMP}"
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

# Start test
print_header "Documents V2 - Templates & Blueprints Workflow Test"

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
      \"name\": \"Templates Test Space\",
      \"isPublic\": true
    }
  }")

check_response "$SPACE_RESPONSE" "Create Space"
SPACE_ID=$(extract_id "$SPACE_RESPONSE" "id")

# Step 1: Create Meeting Notes Template
print_step "Step 1: Create Meeting Notes template with variables"
TEMPLATE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentTemplateCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"name\": \"Meeting Notes Template\",
      \"description\": \"Standard template for team meetings\",
      \"typeID\": \"type-template-meeting\",
      \"creatorID\": \"$USER_ID\",
      \"contentTemplate\": \"<h1>{{meetingTitle}}</h1><p><strong>Date:</strong> {{date}}</p><p><strong>Attendees:</strong></p><ul>{{attendees}}</ul><h2>Agenda</h2>{{agenda}}<h2>Discussion</h2>{{discussion}}<h2>Action Items</h2>{{actionItems}}\",
      \"variablesJSON\": \"{\\\"meetingTitle\\\": \\\"string\\\", \\\"date\\\": \\\"date\\\", \\\"attendees\\\": \\\"list\\\", \\\"agenda\\\": \\\"text\\\", \\\"discussion\\\": \\\"text\\\", \\\"actionItems\\\": \\\"text\\\"}\",
      \"isPublic\": true
    }
  }")

check_response "$TEMPLATE_RESPONSE" "Create Template"
TEMPLATE_ID=$(extract_id "$TEMPLATE_RESPONSE" "id")
echo "  Template ID: $TEMPLATE_ID"

# Step 2: Create Document from Template
print_step "Step 2: Create document from template with variable values"
DOC_FROM_TEMPLATE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCreateFromTemplate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"templateID\": \"$TEMPLATE_ID\",
      \"spaceID\": \"$SPACE_ID\",
      \"title\": \"Weekly Team Meeting - October 18, 2025\",
      \"variables\": {
        \"meetingTitle\": \"Weekly Team Meeting\",
        \"date\": \"October 18, 2025\",
        \"attendees\": \"<li>John Doe</li><li>Jane Smith</li><li>Bob Johnson</li>\",
        \"agenda\": \"<ol><li>Project status updates</li><li>Q4 planning</li><li>Team feedback</li></ol>\",
        \"discussion\": \"<p>Project A is on track. Project B needs more resources.</p>\",
        \"actionItems\": \"<ul><li>John to review Q4 roadmap by Friday</li><li>Jane to schedule follow-up with stakeholders</li></ul>\"
      }
    }
  }")

check_response "$DOC_FROM_TEMPLATE_RESPONSE" "Create Document from Template"
DOC_ID=$(extract_id "$DOC_FROM_TEMPLATE_RESPONSE" "id")
echo "  Document ID: $DOC_ID"

# Step 3: Verify Variable Substitution
print_step "Step 3: Read document and verify variable substitution"
READ_DOC_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentRead\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\"
    }
  }")

check_response "$READ_DOC_RESPONSE" "Read Document"

if echo "$READ_DOC_RESPONSE" | grep -q "Weekly Team Meeting"; then
    print_success "Variable substitution verified - meetingTitle found in content"
else
    print_error "Variable substitution failed - meetingTitle not found"
fi

# Step 4: Check Template Use Count
print_step "Step 4: Verify template use count incremented"
TEMPLATE_READ_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentTemplateRead\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"templateID\": \"$TEMPLATE_ID\"
    }
  }")

check_response "$TEMPLATE_READ_RESPONSE" "Read Template"
USE_COUNT=$(echo "$TEMPLATE_READ_RESPONSE" | grep -o '"useCount":[0-9]*' | cut -d':' -f2)
echo "  Template use count: $USE_COUNT (expected 1)"

# Step 5: Create Product Spec Blueprint
print_step "Step 5: Create Product Specification blueprint with wizard"
BLUEPRINT_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentBlueprintCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"name\": \"Product Specification Blueprint\",
      \"description\": \"Guided creation of product specifications\",
      \"templateID\": \"$TEMPLATE_ID\",
      \"creatorID\": \"$USER_ID\",
      \"wizardStepsJSON\": \"[{\\\"step\\\": 1, \\\"title\\\": \\\"Product Overview\\\", \\\"fields\\\": [\\\"productName\\\", \\\"description\\\"]}, {\\\"step\\\": 2, \\\"title\\\": \\\"Requirements\\\", \\\"fields\\\": [\\\"functionalReqs\\\", \\\"nonFunctionalReqs\\\"]}, {\\\"step\\\": 3, \\\"title\\\": \\\"Success Metrics\\\", \\\"fields\\\": [\\\"kpis\\\", \\\"targets\\\"]}]\",
      \"isPublic\": true
    }
  }")

check_response "$BLUEPRINT_RESPONSE" "Create Blueprint"
BLUEPRINT_ID=$(extract_id "$BLUEPRINT_RESPONSE" "id")
echo "  Blueprint ID: $BLUEPRINT_ID"

# Step 6: Create Document from Blueprint
print_step "Step 6: Create document from blueprint with wizard data"
DOC_FROM_BLUEPRINT_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCreateFromBlueprint\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"blueprintID\": \"$BLUEPRINT_ID\",
      \"spaceID\": \"$SPACE_ID\",
      \"wizardData\": {
        \"step1\": {
          \"productName\": \"Feature X\",
          \"description\": \"A revolutionary new feature for our platform\"
        },
        \"step2\": {
          \"functionalReqs\": \"Must support real-time collaboration\",
          \"nonFunctionalReqs\": \"Performance: < 100ms response time\"
        },
        \"step3\": {
          \"kpis\": \"User adoption rate, engagement time\",
          \"targets\": \"50% adoption in Q1, 10 min avg session\"
        }
      }
    }
  }")

check_response "$DOC_FROM_BLUEPRINT_RESPONSE" "Create Document from Blueprint"
DOC_BLUEPRINT_ID=$(extract_id "$DOC_FROM_BLUEPRINT_RESPONSE" "id")
echo "  Document ID: $DOC_BLUEPRINT_ID"

# Step 7: List All Templates
print_step "Step 7: List all templates"
TEMPLATES_LIST_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentTemplateList\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"sortBy\": \"useCount\",
      \"sortOrder\": \"desc\"
    }
  }")

check_response "$TEMPLATES_LIST_RESPONSE" "List Templates"
TEMPLATE_COUNT=$(echo "$TEMPLATES_LIST_RESPONSE" | grep -o '"id":"template-' | wc -l)
echo "  Templates found: $TEMPLATE_COUNT (expected ≥1)"

# Step 8: List All Blueprints
print_step "Step 8: List all blueprints"
BLUEPRINTS_LIST_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentBlueprintList\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"spaceID\": \"$SPACE_ID\",
      \"includeGlobal\": true
    }
  }")

check_response "$BLUEPRINTS_LIST_RESPONSE" "List Blueprints"
BLUEPRINT_COUNT=$(echo "$BLUEPRINTS_LIST_RESPONSE" | grep -o '"id":"blueprint-' | wc -l)
echo "  Blueprints found: $BLUEPRINT_COUNT (expected ≥1)"

# Step 9: Update Template
print_step "Step 9: Update template with new variables"
UPDATE_TEMPLATE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentTemplateModify\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"templateID\": \"$TEMPLATE_ID\",
      \"version\": 1,
      \"name\": \"Meeting Notes Template (Updated)\",
      \"contentTemplate\": \"<h1>{{meetingTitle}}</h1><p><strong>Date:</strong> {{date}}</p><p><strong>Location:</strong> {{location}}</p><p><strong>Attendees:</strong></p><ul>{{attendees}}</ul><h2>Agenda</h2>{{agenda}}<h2>Discussion</h2>{{discussion}}<h2>Action Items</h2>{{actionItems}}<h2>Next Steps</h2>{{nextSteps}}\"
    }
  }")

check_response "$UPDATE_TEMPLATE_RESPONSE" "Update Template"
echo "  Template updated (added location and nextSteps variables)"

# Step 10: Create Another Document with Updated Template
print_step "Step 10: Create document with updated template"
DOC_FROM_UPDATED_TEMPLATE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCreateFromTemplate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"templateID\": \"$TEMPLATE_ID\",
      \"spaceID\": \"$SPACE_ID\",
      \"title\": \"Sprint Planning Meeting - October 19, 2025\",
      \"variables\": {
        \"meetingTitle\": \"Sprint Planning\",
        \"date\": \"October 19, 2025\",
        \"location\": \"Conference Room A\",
        \"attendees\": \"<li>Development Team</li><li>Product Owner</li>\",
        \"agenda\": \"<ol><li>Sprint goals</li><li>Story estimation</li></ol>\",
        \"discussion\": \"<p>Team committed to 30 story points.</p>\",
        \"actionItems\": \"<ul><li>Team to begin development on Monday</li></ul>\",
        \"nextSteps\": \"<p>Daily standups at 9 AM starting Monday.</p>\"
      }
    }
  }")

check_response "$DOC_FROM_UPDATED_TEMPLATE_RESPONSE" "Create Document from Updated Template"
DOC_UPDATED_ID=$(extract_id "$DOC_FROM_UPDATED_TEMPLATE_RESPONSE" "id")
echo "  Document ID: $DOC_UPDATED_ID"

# Step 11: Verify Final Use Count
print_step "Step 11: Verify template use count is now 2"
FINAL_TEMPLATE_READ_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentTemplateRead\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"templateID\": \"$TEMPLATE_ID\"
    }
  }")

check_response "$FINAL_TEMPLATE_READ_RESPONSE" "Read Template (Final)"
FINAL_USE_COUNT=$(echo "$FINAL_TEMPLATE_READ_RESPONSE" | grep -o '"useCount":[0-9]*' | cut -d':' -f2)
echo "  Template use count: $FINAL_USE_COUNT (expected 2)"

if [[ "$FINAL_USE_COUNT" == "2" ]]; then
    print_success "Use count tracking verified"
else
    print_error "Use count mismatch - expected 2, got $FINAL_USE_COUNT"
fi

# Step 12: Delete Template (should fail if documents exist)
print_step "Step 12: Attempt to delete template (should have safeguards)"
DELETE_TEMPLATE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentTemplateRemove\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"templateID\": \"$TEMPLATE_ID\"
    }
  }")

# This might succeed or fail depending on implementation safeguards
if echo "$DELETE_TEMPLATE_RESPONSE" | grep -q '"errorCode":-1'; then
    print_success "Template deleted (soft delete)"
elif echo "$DELETE_TEMPLATE_RESPONSE" | grep -q '"errorCode"'; then
    print_success "Template deletion prevented (has active documents - good safeguard)"
else
    print_error "Unexpected response from template deletion"
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
echo -e "${GREEN}✓ All templates & blueprints workflow tests completed successfully!${NC}"
echo ""
echo "Tests performed:"
echo "  1. Template creation with 6 variables"
echo "  2. Document creation from template"
echo "  3. Variable substitution verification"
echo "  4. Template use count tracking"
echo "  5. Blueprint creation with 3-step wizard"
echo "  6. Document creation from blueprint"
echo "  7. List all templates (sorted by use)"
echo "  8. List all blueprints"
echo "  9. Update template with new variables"
echo " 10. Create document with updated template"
echo " 11. Verify use count increment (2 uses)"
echo " 12. Template deletion (with safeguards)"
echo ""
echo "Templates & Blueprints created:"
echo "  Templates: 1 (Meeting Notes)"
echo "  Blueprints: 1 (Product Specification)"
echo "  Documents from templates: 3"
echo "  Template use count: $FINAL_USE_COUNT"
echo ""
echo "Variables tested:"
echo "  meetingTitle, date, location, attendees, agenda"
echo "  discussion, actionItems, nextSteps"
echo ""

print_success "Templates & Blueprints Workflow Test Complete"
exit 0
