#!/bin/bash
# E2E Test: Documents V2 - Collaboration Workflow
# Tests: Comments, inline comments, mentions, reactions, watchers
#
# NOTE: Requires database implementation fixes to be completed.
#       See: DOCUMENTS_V2_DATABASE_ISSUES.md
#
# Workflow:
# 1. Create space and document
# 2. Add comments (threaded)
# 3. Add inline comments
# 4. Create mentions (@username)
# 5. Add reactions (emoji)
# 6. Add watchers
# 7. List all collaboration items
# 8. Cleanup

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-test-jwt-token}"
TIMESTAMP=$(date +%s)
SPACE_KEY="COLLAB${TIMESTAMP}"
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
print_header "Documents V2 - Collaboration Workflow Test"

echo "Configuration:"
echo "  Base URL: $BASE_URL"
echo "  Space Key: $SPACE_KEY"
echo "  User ID: $USER_ID"
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
      \"name\": \"Collaboration Test Space\",
      \"isPublic\": true
    }
  }")

check_response "$SPACE_RESPONSE" "Create Space"
SPACE_ID=$(extract_id "$SPACE_RESPONSE" "id")

# Setup: Create Document
print_step "Setup: Create document"
DOC_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"title\": \"Team Collaboration Document\",
      \"spaceID\": \"$SPACE_ID\",
      \"typeID\": \"type-page\",
      \"content\": {
        \"contentHTML\": \"<h1>Team Collaboration</h1><p>This document demonstrates collaboration features.</p><p>We can comment on this section.</p>\"
      }
    }
  }")

check_response "$DOC_RESPONSE" "Create Document"
DOC_ID=$(extract_id "$DOC_RESPONSE" "id")
echo "  Document ID: $DOC_ID"

# Step 1: Add Comment
print_step "Step 1: Add top-level comment"
COMMENT_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCommentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"userID\": \"$USER_ID\",
      \"commentText\": \"Great documentation! This will help the team.\"
    }
  }")

check_response "$COMMENT_RESPONSE" "Add Comment"
COMMENT_ID=$(extract_id "$COMMENT_RESPONSE" "id")
echo "  Comment ID: $COMMENT_ID"

# Step 2: Reply to Comment (Threaded)
print_step "Step 2: Reply to comment (threaded discussion)"
REPLY_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCommentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"parentCommentID\": \"$COMMENT_ID\",
      \"userID\": \"user-reply-${TIMESTAMP}\",
      \"commentText\": \"Thanks! I'll add more examples in the next update.\"
    }
  }")

check_response "$REPLY_RESPONSE" "Reply to Comment"
REPLY_ID=$(extract_id "$REPLY_RESPONSE" "id")
echo "  Reply ID: $REPLY_ID"

# Step 3: Add Inline Comment
print_step "Step 3: Add inline comment on specific text"
INLINE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentInlineCommentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"commentID\": \"$COMMENT_ID\",
      \"positionStart\": 25,
      \"positionEnd\": 50,
      \"selectedText\": \"collaboration features\"
    }
  }")

check_response "$INLINE_RESPONSE" "Add Inline Comment"
INLINE_ID=$(extract_id "$INLINE_RESPONSE" "id")
echo "  Inline Comment ID: $INLINE_ID"

# Step 4: Create Mention
print_step "Step 4: Mention user in comment"
MENTION_COMMENT_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCommentCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"userID\": \"$USER_ID\",
      \"commentText\": \"@john.doe Can you review this section?\"
    }
  }")

check_response "$MENTION_COMMENT_RESPONSE" "Add Comment with Mention"
MENTION_COMMENT_ID=$(extract_id "$MENTION_COMMENT_RESPONSE" "id")

MENTION_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentMentionCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"commentID\": \"$MENTION_COMMENT_ID\",
      \"mentionedUserID\": \"user-john.doe\",
      \"mentionText\": \"@john.doe\",
      \"mentionerID\": \"$USER_ID\"
    }
  }")

check_response "$MENTION_RESPONSE" "Create Mention"
MENTION_ID=$(extract_id "$MENTION_RESPONSE" "id")
echo "  Mention ID: $MENTION_ID"

# Step 5: Add Reactions
print_step "Step 5: Add emoji reactions"

# Thumbs up
REACTION1_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentReactionCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"userID\": \"$USER_ID\",
      \"reactionType\": \"thumbsup\",
      \"reactionEmoji\": \"👍\"
    }
  }")

check_response "$REACTION1_RESPONSE" "Add Reaction (thumbsup)"
REACTION1_ID=$(extract_id "$REACTION1_RESPONSE" "id")

# Heart
REACTION2_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentReactionCreate\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"userID\": \"user-other-${TIMESTAMP}\",
      \"reactionType\": \"heart\",
      \"reactionEmoji\": \"❤️\"
    }
  }")

check_response "$REACTION2_RESPONSE" "Add Reaction (heart)"
echo "  Reaction IDs: $REACTION1_ID, $(extract_id "$REACTION2_RESPONSE" "id")"

# Step 6: Add Watchers
print_step "Step 6: Add watchers for notifications"

# Watcher with 'all' notifications
WATCHER1_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentWatcherAdd\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"userID\": \"$USER_ID\",
      \"notificationLevel\": \"all\"
    }
  }")

check_response "$WATCHER1_RESPONSE" "Add Watcher (all notifications)"
WATCHER1_ID=$(extract_id "$WATCHER1_RESPONSE" "id")

# Watcher with 'mentions' only
WATCHER2_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentWatcherAdd\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"userID\": \"user-watcher-${TIMESTAMP}\",
      \"notificationLevel\": \"mentions\"
    }
  }")

check_response "$WATCHER2_RESPONSE" "Add Watcher (mentions only)"
echo "  Watcher IDs: $WATCHER1_ID, $(extract_id "$WATCHER2_RESPONSE" "id")"

# Step 7: List Comments
print_step "Step 7: List all comments"
COMMENTS_LIST_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentCommentList\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"includeReplies\": true
    }
  }")

check_response "$COMMENTS_LIST_RESPONSE" "List Comments"
COMMENT_COUNT=$(echo "$COMMENTS_LIST_RESPONSE" | grep -o '"id":"comment-' | wc -l)
echo "  Comments found: $COMMENT_COUNT (expected 3)"

# Step 8: List Reactions
print_step "Step 8: List all reactions"
REACTIONS_LIST_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentReactionList\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\",
      \"groupByType\": true
    }
  }")

check_response "$REACTIONS_LIST_RESPONSE" "List Reactions"
echo "  Reactions: thumbsup (1), heart (1)"

# Step 9: List Watchers
print_step "Step 9: List all watchers"
WATCHERS_LIST_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentWatcherList\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"documentID\": \"$DOC_ID\"
    }
  }")

check_response "$WATCHERS_LIST_RESPONSE" "List Watchers"
WATCHER_COUNT=$(echo "$WATCHERS_LIST_RESPONSE" | grep -o '"id":"watcher-' | wc -l)
echo "  Watchers found: $WATCHER_COUNT (expected 2)"

# Step 10: Resolve Inline Comment
print_step "Step 10: Resolve inline comment"
RESOLVE_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentInlineCommentResolve\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"inlineCommentID\": \"$INLINE_ID\",
      \"isResolved\": true
    }
  }")

check_response "$RESOLVE_RESPONSE" "Resolve Inline Comment"

# Step 11: List User's Mentions
print_step "Step 11: List mentions for john.doe"
MENTIONS_LIST_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentMentionListByUser\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"userID\": \"user-john.doe\",
      \"includeRead\": true
    }
  }")

check_response "$MENTIONS_LIST_RESPONSE" "List User Mentions"

# Step 12: Update Watcher Notification Level
print_step "Step 12: Update watcher notification level"
UPDATE_WATCHER_RESPONSE=$(curl -s -X POST "$BASE_URL/do" \
  -H "Content-Type: application/json" \
  -d "{
    \"action\": \"documentWatcherModify\",
    \"jwt\": \"$JWT_TOKEN\",
    \"data\": {
      \"watcherID\": \"$WATCHER1_ID\",
      \"notificationLevel\": \"mentions\"
    }
  }")

check_response "$UPDATE_WATCHER_RESPONSE" "Update Watcher Level"
echo "  Changed from 'all' to 'mentions'"

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
echo -e "${GREEN}✓ All collaboration workflow tests completed successfully!${NC}"
echo ""
echo "Tests performed:"
echo "  1. Comment creation (threaded)"
echo "  2. Reply to comments"
echo "  3. Inline comments with text selection"
echo "  4. User mentions (@username)"
echo "  5. Emoji reactions (👍, ❤️)"
echo "  6. Watcher subscriptions (all/mentions)"
echo "  7. List comments"
echo "  8. List reactions (grouped)"
echo "  9. List watchers"
echo " 10. Resolve inline comments"
echo " 11. List user mentions"
echo " 12. Update watcher notification level"
echo ""
echo "Collaboration items created:"
echo "  Comments: $COMMENT_COUNT"
echo "  Reactions: 2 (thumbsup, heart)"
echo "  Watchers: $WATCHER_COUNT"
echo "  Mentions: 1 (@john.doe)"
echo "  Inline Comments: 1 (resolved)"
echo ""

print_success "Collaboration Workflow Test Complete"
exit 0
