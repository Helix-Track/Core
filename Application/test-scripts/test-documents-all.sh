#!/bin/bash
# E2E Test: Documents V2 - Master Test Suite
# Runs all Documents V2 workflow tests
#
# NOTE: Requires database implementation fixes to be completed.
#       See: DOCUMENTS_V2_DATABASE_ISSUES.md
#
# Test Suites:
# 1. Basic Workflow - Space, document CRUD, hierarchy
# 2. Collaboration - Comments, mentions, reactions, watchers
# 3. Version Control - History, diff, rollback, optimistic locking
# 4. Templates & Blueprints - Templates, variables, wizard creation

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-test-jwt-token}"
TIMESTAMP=$(date +%s)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
START_TIME=$(date +%s)

# Helper functions
print_header() {
    echo -e "\n${CYAN}================================================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}================================================================${NC}\n"
}

print_test_header() {
    echo -e "\n${BLUE}>>> $1${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

run_test() {
    local test_script="$1"
    local test_name="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    print_test_header "Running: $test_name"

    if bash "$test_script"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        print_success "$test_name passed"
        return 0
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        print_error "$test_name failed"
        return 1
    fi
}

# Main execution
print_header "HelixTrack Documents V2 - Complete Test Suite"

echo "Test Configuration:"
echo "  Base URL: $BASE_URL"
echo "  JWT Token: ${JWT_TOKEN:0:20}..."
echo "  Timestamp: $TIMESTAMP"
echo "  Date: $(date)"
echo ""

print_warning "NOTE: These tests require database implementation fixes"
print_warning "See: DOCUMENTS_V2_DATABASE_ISSUES.md"
echo ""

read -p "Continue with tests? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Tests cancelled by user"
    exit 1
fi

# Check if server is running
print_test_header "Pre-flight check: Testing server connectivity"
if curl -s -X POST "$BASE_URL/do" \
    -H "Content-Type: application/json" \
    -d '{"action": "version"}' | grep -q '"errorCode":-1'; then
    print_success "Server is running and responding"
else
    print_error "Server is not responding at $BASE_URL"
    print_error "Please start the server first: ./htCore"
    exit 1
fi

# Run test suites
print_header "Test Suite 1: Basic Workflow"
run_test "./test-documents-workflow-basic.sh" "Basic Workflow (Space, Document CRUD, Hierarchy)"

print_header "Test Suite 2: Collaboration Features"
run_test "./test-documents-workflow-collaboration.sh" "Collaboration (Comments, Mentions, Reactions, Watchers)"

print_header "Test Suite 3: Version Control"
run_test "./test-documents-workflow-versioning.sh" "Version Control (History, Diff, Rollback, Locking)"

print_header "Test Suite 4: Templates & Blueprints"
run_test "./test-documents-workflow-templates.sh" "Templates & Blueprints (Variables, Wizard Creation)"

# Calculate results
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
PASS_RATE=$(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")

# Print summary
print_header "Test Suite Summary"

echo "Results:"
echo "  Total Tests: $TOTAL_TESTS"
echo -e "  Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "  Failed: ${RED}$FAILED_TESTS${NC}"
echo "  Pass Rate: ${PASS_RATE}%"
echo "  Duration: ${DURATION}s"
echo ""

echo "Test Coverage:"
echo "  ✓ Space Management (create, list, archive, delete)"
echo "  ✓ Document CRUD (create, read, modify, delete)"
echo "  ✓ Document Hierarchy (parent-child relationships)"
echo "  ✓ Comments (create, reply, list, threaded)"
echo "  ✓ Inline Comments (text selection, resolve)"
echo "  ✓ Mentions (@username notifications)"
echo "  ✓ Reactions (emoji reactions)"
echo "  ✓ Watchers (subscription, notification levels)"
echo "  ✓ Version History (list, read specific version)"
echo "  ✓ Version Comparison (diff: unified, split, html)"
echo "  ✓ Version Labels & Tags (milestone markers)"
echo "  ✓ Version Rollback (restore previous version)"
echo "  ✓ Optimistic Locking (version conflict detection)"
echo "  ✓ Templates (create, use, variable substitution)"
echo "  ✓ Blueprints (wizard creation, multi-step)"
echo "  ✓ Template Usage Tracking (use count)"
echo ""

echo "API Actions Tested:"
echo "  - documentSpaceCreate, documentSpaceList, documentSpaceArchive, documentSpaceRemove"
echo "  - documentCreate, documentRead, documentModify, documentRemove, documentList"
echo "  - documentCreateFromTemplate, documentCreateFromBlueprint"
echo "  - documentCommentCreate, documentCommentList"
echo "  - documentInlineCommentCreate, documentInlineCommentResolve"
echo "  - documentMentionCreate, documentMentionListByUser"
echo "  - documentReactionCreate, documentReactionList"
echo "  - documentWatcherAdd, documentWatcherList, documentWatcherModify"
echo "  - documentVersionList, documentVersionRead, documentVersionCompare, documentVersionRestore"
echo "  - documentVersionLabelCreate, documentVersionTagCreate, documentVersionCommentCreate"
echo "  - documentTemplateCreate, documentTemplateRead, documentTemplateList, documentTemplateModify"
echo "  - documentBlueprintCreate, documentBlueprintList"
echo ""
echo "  Total: 30+ API actions tested"
echo ""

# Exit with appropriate code
if [[ $FAILED_TESTS -eq 0 ]]; then
    print_success "ALL TESTS PASSED! 🎉"
    echo ""
    echo "Documents V2 implementation verified:"
    echo "  ✓ All basic workflows functioning"
    echo "  ✓ Collaboration features working"
    echo "  ✓ Version control operational"
    echo "  ✓ Templates & blueprints functional"
    echo ""
    exit 0
else
    print_error "SOME TESTS FAILED"
    echo ""
    echo "Failed tests: $FAILED_TESTS/$TOTAL_TESTS"
    echo ""
    print_warning "Check individual test output above for details"
    print_warning "See DOCUMENTS_V2_DATABASE_ISSUES.md for known issues"
    echo ""
    exit 1
fi
