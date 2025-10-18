# Documents V2 - Architecture Refactor for Core Entity Reuse

**Date**: 2025-10-18
**Purpose**: Refactor Documents V2 to reuse existing core entities instead of duplicating them

---

## Core Entities Available for Reuse

### From Definition.V1.sql

**1. Comment System** (lines 1003-1017)
```sql
CREATE TABLE comment (
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL
);
```
**Mapping Tables:**
- `comment_ticket_mapping`
- `asset_comment_mapping`

**2. Label System** (lines 782-799)
```sql
CREATE TABLE label (
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);
```
**Mapping Tables:**
- `label_project_mapping`
- `label_team_mapping`
- `label_ticket_mapping`
- `label_asset_mapping`

**3. Label Categories** (lines 804-819)
```sql
CREATE TABLE label_category (
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL
);
```

### From Definition.V3.sql

**4. Vote System** (line 506)
```sql
CREATE TABLE ticket_vote_mapping
```

**5. Mention System** (line 639)
```sql
CREATE TABLE comment_mention_mapping
```

---

## Refactoring Strategy

### Remove These Document-Specific Tables

**From Documents V2 Schema:**
- ❌ `document_comment` → Use core `comment` + mapping
- ❌ `document_comment_thread` → Use core `comment` with parent_id
- ❌ `document_mention` → Use core `comment_mention_mapping`
- ❌ `document_label` → Use core `label`
- ❌ `document_tag` → Keep as separate (tags != labels)
- ❌ `document_label_mapping` → Use new `label_document_mapping`
- ❌ `document_reaction` → Replace with generic vote system

### Add These Generic Mapping Tables

**To Core System (V5):**
- ✅ `comment_document_mapping` - Link comments to documents
- ✅ `label_document_mapping` - Link labels to documents
- ✅ `vote_mapping` - Generic votes (replaces ticket_vote_mapping)

### Keep These Document-Specific Tables

**Unique to Documents:**
- ✅ `document_space` - Document-specific
- ✅ `document_type` - Document-specific
- ✅ `document` - Document-specific
- ✅ `document_content` - Document-specific
- ✅ `document_version` - Document-specific
- ✅ `document_version_label` - Version-specific
- ✅ `document_version_tag` - Version-specific
- ✅ `document_version_comment` - Version-specific
- ✅ `document_version_mention` - Version-specific
- ✅ `document_version_diff` - Document-specific
- ✅ `document_inline_comment` - Document-specific (has position data)
- ✅ `document_tag` - Tags are different from labels
- ✅ `document_tag_mapping` - Document-specific
- ✅ `document_watcher` - Document-specific
- ✅ `document_entity_link` - Document-specific
- ✅ `document_relationship` - Document-specific
- ✅ `document_template` - Document-specific
- ✅ `document_blueprint` - Document-specific
- ✅ `document_view_history` - Document-specific
- ✅ `document_analytics` - Document-specific
- ✅ `document_attachment` - Document-specific

---

## Updated Table Count

### Original Documents V2
- **Total**: 27 tables
- **Reusable**: 4 tables (comment, label, mention, vote)
- **After Refactor**: 23 document-specific tables

### New Mapping Tables (Add to Core V5)
- `comment_document_mapping`
- `label_document_mapping`
- `vote_mapping` (generic, replaces ticket_vote_mapping)

---

## Benefits of Refactoring

### 1. Unified Data Model
- Comments work the same everywhere (tickets, documents, assets)
- Labels work the same everywhere
- Votes/reactions unified

### 2. Simplified Queries
- Single query to get all comments by user (across all entities)
- Single query to get all entities with a label
- Consistent API patterns

### 3. Reduced Code Duplication
- Reuse existing comment handlers
- Reuse existing label handlers
- Reuse existing vote handlers

### 4. Better Integration
- Documents integrate seamlessly with core
- Cross-entity features work automatically
- Permissions handled consistently

---

## Implementation Changes Needed

### 1. Update Documents V2 Schema

**Remove:**
```sql
DROP TABLE IF EXISTS document_comment;
DROP TABLE IF EXISTS document_comment_thread;
DROP TABLE IF EXISTS document_mention;
DROP TABLE IF EXISTS document_label;
DROP TABLE IF EXISTS document_label_mapping;
DROP TABLE IF EXISTS document_reaction;
```

**Keep Enhanced comment Table for Inline Comments:**
```sql
-- Keep this - it has document-specific position data
CREATE TABLE document_inline_comment (
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id     TEXT    NOT NULL,
    comment_id      TEXT    NOT NULL,  -- Links to core comment table
    position_start  INTEGER NOT NULL,
    position_end    INTEGER NOT NULL,
    selected_text   TEXT,
    is_resolved     BOOLEAN NOT NULL DEFAULT 0,
    created         INTEGER NOT NULL
);
```

### 2. Add to Core V5 Schema

**Generic Comment Mapping:**
```sql
CREATE TABLE comment_document_mapping (
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id  TEXT    NOT NULL,
    document_id TEXT    NOT NULL,
    user_id     TEXT    NOT NULL,  -- User who added comment
    is_resolved BOOLEAN NOT NULL DEFAULT 0,
    created     INTEGER NOT NULL,
    UNIQUE(comment_id, document_id),
    FOREIGN KEY (comment_id) REFERENCES comment(id)
);
```

**Generic Label Mapping:**
```sql
CREATE TABLE label_document_mapping (
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id    TEXT    NOT NULL,
    document_id TEXT    NOT NULL,
    user_id     TEXT    NOT NULL,
    created     INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(label_id, document_id),
    FOREIGN KEY (label_id) REFERENCES label(id)
);
```

**Generic Vote System (Replaces ticket_vote_mapping):**
```sql
CREATE TABLE vote_mapping (
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    entity_type TEXT    NOT NULL,  -- "ticket", "document", "comment", etc.
    entity_id   TEXT    NOT NULL,
    user_id     TEXT    NOT NULL,
    vote_type   TEXT    NOT NULL DEFAULT 'upvote',  -- "upvote", "downvote", "like", "love", etc.
    emoji       TEXT,                               -- Optional emoji
    created     INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE(entity_type, entity_id, user_id, vote_type)
);
```

### 3. Enhance Core Comment Table

**Option 1: Add Optional Fields (Non-Breaking)**
```sql
ALTER TABLE comment ADD COLUMN parent_id TEXT;  -- For threading
ALTER TABLE comment ADD COLUMN version INTEGER DEFAULT 1;  -- For edit history
ALTER TABLE comment ADD COLUMN user_id TEXT;  -- Who wrote the comment

CREATE INDEX comments_get_by_parent_id ON comment (parent_id);
CREATE INDEX comments_get_by_user_id ON comment (user_id);
```

**Option 2: Use Mapping Table for Metadata**
```sql
CREATE TABLE comment_metadata (
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id  TEXT    NOT NULL UNIQUE,
    parent_id   TEXT,
    user_id     TEXT    NOT NULL,
    version     INTEGER NOT NULL DEFAULT 1,
    is_edited   BOOLEAN NOT NULL DEFAULT 0,
    created     INTEGER NOT NULL,
    FOREIGN KEY (comment_id) REFERENCES comment(id),
    FOREIGN KEY (parent_id) REFERENCES comment(id)
);
```

---

## Updated API Actions

### Actions That Now Use Core Entities

**Comments (use core comment + mapping):**
- `documentCommentAdd` → Creates comment + comment_document_mapping
- `documentCommentReply` → Creates comment with parent_id
- `documentCommentEdit` → Updates core comment
- `documentCommentRemove` → Soft deletes comment
- `documentCommentList` → Queries via comment_document_mapping

**Labels (use core label + mapping):**
- `documentLabelAdd` → Links to existing label or creates new
- `documentLabelRemove` → Deletes label_document_mapping
- `documentLabelList` → Queries via label_document_mapping

**Votes/Reactions (use core vote_mapping):**
- `documentReact` → Creates vote_mapping entry
- `documentGetReactions` → Queries vote_mapping

**Mentions (use core comment_mention_mapping):**
- `documentMention` → Creates mention via comment

---

## Go Model Changes

### Models to Remove

- ❌ `document_collaboration.go` - Most structs removed
  - Remove `DocumentComment` → use core `Comment`
  - Remove `DocumentMention` → use core mention system
  - Remove `DocumentReaction` → use core `Vote`
  - Remove `DocumentLabel` → use core `Label`

### Models to Keep

- ✅ `DocumentInlineComment` - Has position-specific data
- ✅ `DocumentTag` - Tags are separate from labels
- ✅ `DocumentWatcher` - Document-specific

### New Models to Add

```go
// Comment-document mapping
type CommentDocumentMapping struct {
    ID         string `json:"id"`
    CommentID  string `json:"comment_id"`
    DocumentID string `json:"document_id"`
    UserID     string `json:"user_id"`
    IsResolved bool   `json:"is_resolved"`
    Created    int64  `json:"created"`
}

// Label-document mapping
type LabelDocumentMapping struct {
    ID         string `json:"id"`
    LabelID    string `json:"label_id"`
    DocumentID string `json:"document_id"`
    UserID     string `json:"user_id"`
    Created    int64  `json:"created"`
    Deleted    bool   `json:"deleted"`
}

// Generic vote mapping
type VoteMapping struct {
    ID         string  `json:"id"`
    EntityType string  `json:"entity_type"`
    EntityID   string  `json:"entity_id"`
    UserID     string  `json:"user_id"`
    VoteType   string  `json:"vote_type"`
    Emoji      *string `json:"emoji,omitempty"`
    Created    int64   `json:"created"`
    Deleted    bool    `json:"deleted"`
}
```

---

## Migration Impact

### Existing ticket_vote_mapping Migration

**Need to migrate to generic vote_mapping:**
```sql
INSERT INTO vote_mapping (id, entity_type, entity_id, user_id, vote_type, emoji, created, deleted)
SELECT id, 'ticket' AS entity_type, ticket_id AS entity_id, user_id, 'upvote' AS vote_type, NULL, created, deleted
FROM ticket_vote_mapping;

-- Then drop old table
DROP TABLE ticket_vote_mapping;
```

---

## Revised Table Count

### Documents Extension V2 (Refactored)
- **Document-Specific Tables**: 23
- **Removed (reusing core)**: 4
- **Total Saved**: 4 tables

### Core V5 (Enhanced)
- **Original V5 Tables**: 8
- **New Mapping Tables**: 3
- **Enhanced Core Tables**: 1 (comment)
- **Total V5**: 12 tables

### Overall System
- **Before**: 27 document tables + core tables
- **After**: 23 document tables + 12 V5 tables (including 3 new generic mappings)
- **Net**: Cleaner, more integrated architecture

---

## Recommendation

✅ **Proceed with Refactoring**

**Reasons:**
1. Better long-term architecture
2. Code reuse and consistency
3. Easier maintenance
4. Seamless cross-entity features
5. Reduced duplication

**Implementation Order:**
1. Update core V5 schema with new mapping tables
2. Refactor Documents V2 schema to remove duplicates
3. Update Go models to use core entities
4. Implement handlers using core + mappings
5. Update tests to reflect new architecture

---

**Status**: Architecture refactor recommended
**Impact**: Medium (affects schema and models)
**Benefit**: High (better integration, less duplication)
**Risk**: Low (well-defined refactoring)
