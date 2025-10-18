package database

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/config"
	"helixtrack.ru/core/internal/models"
)

// setupDocumentTestDB creates a test database with document schema
func setupDocumentTestDB(t *testing.T) *db {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_documents.db")

	cfg := config.DatabaseConfig{
		Type:       "sqlite",
		SQLitePath: dbPath,
	}

	database, err := NewDatabase(cfg)
	require.NoError(t, err)

	// Cast to concrete type for testing
	db, ok := database.(*db)
	require.True(t, ok, "Database should be *db type")

	// Create document tables schema
	ctx := context.Background()
	schemas := []string{
		// document table
		`CREATE TABLE document (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			space_id TEXT NOT NULL,
			parent_id TEXT,
			type_id TEXT NOT NULL,
			project_id TEXT,
			creator_id TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			position INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 0,
			is_archived INTEGER DEFAULT 0,
			publish_date INTEGER,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_content table
		`CREATE TABLE document_content (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			version INTEGER NOT NULL,
			content_type TEXT NOT NULL,
			content TEXT NOT NULL,
			content_hash TEXT,
			size_bytes INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_space table
		`CREATE TABLE document_space (
			id TEXT PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			is_public INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_type table
		`CREATE TABLE document_type (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_version table
		`CREATE TABLE document_version (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			version_number INTEGER NOT NULL,
			editor_id TEXT NOT NULL,
			change_summary TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_version_label table
		`CREATE TABLE document_version_label (
			id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL,
			label_name TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_version_tag table
		`CREATE TABLE document_version_tag (
			id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL,
			tag_name TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_version_comment table
		`CREATE TABLE document_version_comment (
			id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			comment_text TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_version_mention table
		`CREATE TABLE document_version_mention (
			id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL,
			mentioned_user_id TEXT NOT NULL,
			mentioner_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_version_diff table
		`CREATE TABLE document_version_diff (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			from_version INTEGER NOT NULL,
			to_version INTEGER NOT NULL,
			diff_type TEXT NOT NULL,
			diff_content TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// comment_document_mapping table
		`CREATE TABLE comment_document_mapping (
			id TEXT PRIMARY KEY,
			comment_id TEXT NOT NULL,
			document_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			is_resolved INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_inline_comment table
		`CREATE TABLE document_inline_comment (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			comment_text TEXT NOT NULL,
			selection_start INTEGER NOT NULL,
			selection_end INTEGER NOT NULL,
			is_resolved INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_watcher table
		`CREATE TABLE document_watcher (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			notification_level TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// label_document_mapping table
		`CREATE TABLE label_document_mapping (
			id TEXT PRIMARY KEY,
			label_id TEXT NOT NULL,
			document_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_tag table
		`CREATE TABLE document_tag (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_tag_mapping table
		`CREATE TABLE document_tag_mapping (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// vote_mapping table
		`CREATE TABLE vote_mapping (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			vote_type TEXT NOT NULL,
			emoji TEXT,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_entity_link table
		`CREATE TABLE document_entity_link (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			link_type TEXT NOT NULL,
			description TEXT,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_relationship table
		`CREATE TABLE document_relationship (
			id TEXT PRIMARY KEY,
			source_document_id TEXT NOT NULL,
			target_document_id TEXT NOT NULL,
			relationship_type TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_template table
		`CREATE TABLE document_template (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			space_id TEXT,
			type_id TEXT NOT NULL,
			content_template TEXT NOT NULL,
			variables_json TEXT,
			creator_id TEXT NOT NULL,
			is_public INTEGER DEFAULT 0,
			use_count INTEGER DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_blueprint table
		`CREATE TABLE document_blueprint (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			template_id TEXT NOT NULL,
			wizard_steps TEXT,
			is_enabled INTEGER DEFAULT 1,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
		// document_view_history table
		`CREATE TABLE document_view_history (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT,
			user_agent TEXT,
			session_id TEXT,
			view_duration INTEGER,
			timestamp INTEGER NOT NULL
		)`,
		// document_analytics table
		`CREATE TABLE document_analytics (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			total_views INTEGER DEFAULT 0,
			unique_viewers INTEGER DEFAULT 0,
			total_edits INTEGER DEFAULT 0,
			unique_editors INTEGER DEFAULT 0,
			total_comments INTEGER DEFAULT 0,
			total_reactions INTEGER DEFAULT 0,
			total_watchers INTEGER DEFAULT 0,
			avg_view_duration INTEGER,
			last_viewed INTEGER,
			last_edited INTEGER,
			popularity_score REAL DEFAULT 0,
			updated INTEGER NOT NULL
		)`,
		// document_attachment table
		`CREATE TABLE document_attachment (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			original_filename TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			storage_path TEXT NOT NULL,
			checksum TEXT NOT NULL,
			uploader_id TEXT NOT NULL,
			description TEXT,
			version INTEGER NOT NULL DEFAULT 1,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER DEFAULT 0
		)`,
	}

	for _, schema := range schemas {
		_, err := db.Exec(ctx, schema)
		require.NoError(t, err, "Failed to create schema")
	}

	return db
}

// ========================================================================
// CORE DOCUMENT OPERATIONS TESTS
// ========================================================================

func TestCreateDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	doc := &models.Document{
		ID:        "doc-123",
		Title:     "Test Document",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()

	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	assert.NoError(t, err)
	assert.Greater(t, doc.Created, int64(0))
	assert.Greater(t, doc.Modified, int64(0))
}

func TestCreateDocument_InvalidDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	doc := &models.Document{
		ID: "", // Invalid: empty ID
	}

	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid document")
}

func TestGetDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create document
	original := &models.Document{
		ID:        "doc-456",
		Title:     "Test Get Document",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}

	original.SetTimestamps()
	err := db.CreateDocument(original)
	require.NoError(t, err)

	// Retrieve document
	retrieved, err := db.GetDocument("doc-456")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "doc-456", retrieved.ID)
	assert.Equal(t, "Test Get Document", retrieved.Title)
	assert.Equal(t, "space-1", retrieved.SpaceID)
	assert.Equal(t, 1, retrieved.Version)
}

func TestGetDocument_NotFound(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	retrieved, err := db.GetDocument("nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Contains(t, err.Error(), "document not found")
}

func TestListDocuments(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create test documents
	docs := []*models.Document{
		{
			ID:        "doc-1",
			Title:     "Document 1",
			SpaceID:   "space-1",
			TypeID:    "type-page",
			CreatorID: "user-1",
			Version:   1,
		},
		{
			ID:        "doc-2",
			Title:     "Document 2",
			SpaceID:   "space-1",
			TypeID:    "type-page",
			CreatorID: "user-1",
			Version:   1,
		},
		{
			ID:        "doc-3",
			Title:     "Document 3",
			SpaceID:   "space-2",
			TypeID:    "type-page",
			CreatorID: "user-1",
			Version:   1,
		},
	}

	for _, doc := range docs {
		doc.SetTimestamps()
		err := db.CreateDocument(doc)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Test list all
	allDocs, err := db.ListDocuments(nil, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, allDocs, 3)

	// Test filter by space
	filters := map[string]interface{}{
		"space_id": "space-1",
	}
	filteredDocs, err := db.ListDocuments(filters, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, filteredDocs, 2)

	// Test pagination
	paginatedDocs, err := db.ListDocuments(nil, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, paginatedDocs, 2)

	// Test offset
	offsetDocs, err := db.ListDocuments(nil, 2, 1)
	assert.NoError(t, err)
	assert.Len(t, offsetDocs, 2)
}

func TestUpdateDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create document
	doc := &models.Document{
		ID:        "doc-789",
		Title:     "Original Title",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)

	// Update document
	doc.Title = "Updated Title"
	err = db.UpdateDocument(doc)
	assert.NoError(t, err)
	assert.Equal(t, 2, doc.Version) // Version should increment

	// Verify update
	updated, err := db.GetDocument("doc-789")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, 2, updated.Version)
}

func TestUpdateDocument_VersionConflict(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create document
	doc := &models.Document{
		ID:        "doc-conflict",
		Title:     "Original Title",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)

	// Simulate concurrent modification
	doc1 := &models.Document{
		ID:        "doc-conflict",
		Title:     "Update 1",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1, // Same version
	}
	doc2 := &models.Document{
		ID:        "doc-conflict",
		Title:     "Update 2",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1, // Same version
	}

	// First update should succeed
	doc1.SetTimestamps()
	err = db.UpdateDocument(doc1)
	assert.NoError(t, err)

	// Second update should fail due to version conflict
	doc2.SetTimestamps()
	err = db.UpdateDocument(doc2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version conflict")
}

func TestDeleteDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create document
	doc := &models.Document{
		ID:        "doc-delete",
		Title:     "To Delete",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)

	// Delete document
	err = db.DeleteDocument("doc-delete")
	assert.NoError(t, err)

	// Verify document is soft-deleted
	deleted, err := db.GetDocument("doc-delete")
	assert.Error(t, err)
	assert.Nil(t, deleted)
}

func TestDeleteDocument_NotFound(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	err := db.DeleteDocument("nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
}

func TestRestoreDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create and delete document
	doc := &models.Document{
		ID:        "doc-restore",
		Title:     "To Restore",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)
	err = db.DeleteDocument("doc-restore")
	require.NoError(t, err)

	// Restore document
	err = db.RestoreDocument("doc-restore")
	assert.NoError(t, err)

	// Verify document is restored
	restored, err := db.GetDocument("doc-restore")
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, "doc-restore", restored.ID)
}

func TestArchiveDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create document
	doc := &models.Document{
		ID:        "doc-archive",
		Title:     "To Archive",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)

	// Archive document
	err = db.ArchiveDocument("doc-archive")
	assert.NoError(t, err)

	// Verify document is archived
	archived, err := db.GetDocument("doc-archive")
	assert.NoError(t, err)
	assert.True(t, archived.IsArchived)
}

func TestUnarchiveDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create and archive document
	doc := &models.Document{
		ID:        "doc-unarchive",
		Title:     "To Unarchive",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)
	err = db.ArchiveDocument("doc-unarchive")
	require.NoError(t, err)

	// Unarchive document
	err = db.UnarchiveDocument("doc-unarchive")
	assert.NoError(t, err)

	// Verify document is unarchived
	unarchived, err := db.GetDocument("doc-unarchive")
	assert.NoError(t, err)
	assert.False(t, unarchived.IsArchived)
}

func TestPublishDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create document
	doc := &models.Document{
		ID:        "doc-publish",
		Title:     "To Publish",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)

	// Publish document
	err = db.PublishDocument("doc-publish")
	assert.NoError(t, err)

	// Verify document is published
	published, err := db.GetDocument("doc-publish")
	assert.NoError(t, err)
	assert.True(t, published.IsPublished)
}

func TestUnpublishDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create and publish document
	doc := &models.Document{
		ID:        "doc-unpublish",
		Title:     "To Unpublish",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)
	err = db.PublishDocument("doc-unpublish")
	require.NoError(t, err)

	// Unpublish document
	err = db.UnpublishDocument("doc-unpublish")
	assert.NoError(t, err)

	// Verify document is unpublished
	unpublished, err := db.GetDocument("doc-unpublish")
	assert.NoError(t, err)
	assert.False(t, unpublished.IsPublished)
}

func TestMoveDocument(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create document
	doc := &models.Document{
		ID:        "doc-move",
		Title:     "To Move",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	doc.SetTimestamps()
	err := db.CreateDocument(doc)
	require.NoError(t, err)

	// Move document
	err = db.MoveDocument("doc-move", "space-2")
	assert.NoError(t, err)

	// Verify document moved
	moved, err := db.GetDocument("doc-move")
	assert.NoError(t, err)
	assert.Equal(t, "space-2", moved.SpaceID)
}

func TestSetDocumentParent(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create documents
	parent := &models.Document{
		ID:        "doc-parent",
		Title:     "Parent",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	child := &models.Document{
		ID:        "doc-child",
		Title:     "Child",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}
	parent.SetTimestamps()
	err := db.CreateDocument(parent)
	require.NoError(t, err)
	child.SetTimestamps()
	err = db.CreateDocument(child)
	require.NoError(t, err)

	// Set parent
	err = db.SetDocumentParent("doc-child", "doc-parent")
	assert.NoError(t, err)

	// Verify parent set
	childDoc, err := db.GetDocument("doc-child")
	assert.NoError(t, err)
	assert.NotNil(t, childDoc.ParentID)
	assert.Equal(t, "doc-parent", *childDoc.ParentID)
}

func TestGetDocumentChildren(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create parent
	parent := &models.Document{
		ID:        "doc-parent-2",
		Title:     "Parent",
		SpaceID:   "space-1",
		TypeID:    "type-page",
		CreatorID: "user-1",
		Version:   1,
	}

	parent.SetTimestamps()
	err := db.CreateDocument(parent)
	require.NoError(t, err)

	// Create children
	parentID := "doc-parent-2"
	for i := 1; i <= 3; i++ {
		child := &models.Document{
			ID:        fmt.Sprintf("doc-child-%d", i),
			Title:     fmt.Sprintf("Child %d", i),
			SpaceID:   "space-1",
			ParentID:  &parentID,
			TypeID:    "type-page",
			CreatorID: "user-1",
			Version:   1,
		}

		child.SetTimestamps()
		err = db.CreateDocument(child)
		require.NoError(t, err)
	}

	// Get children
	children, err := db.GetDocumentChildren("doc-parent-2")
	assert.NoError(t, err)
	assert.Len(t, children, 3)
}

// ========================================================================
// DOCUMENT CONTENT OPERATIONS TESTS
// ========================================================================

func TestCreateDocumentContent(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	content := &models.DocumentContent{
		ID:          "content-123",
		DocumentID:  "doc-123",
		Version:     1,
		ContentType: "html",
		Content:     stringPtr("<p>Test content</p>"),
		ContentHash: stringPtr("hash123"),
		SizeBytes:   19,
	}


	content.SetTimestamps()

	content.SetTimestamps()
	err := db.CreateDocumentContent(content)
	assert.NoError(t, err)
	assert.Greater(t, content.Created, int64(0))
	assert.Greater(t, content.Modified, int64(0))
}

func TestGetDocumentContent(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create content
	original := &models.DocumentContent{
		ID:          "content-456",
		DocumentID:  "doc-456",
		Version:     1,
		ContentType: "html",
		Content:     stringPtr("<p>Test content</p>"),
		ContentHash: stringPtr("hash456"),
		SizeBytes:   19,
	}

	original.SetTimestamps()
	err := db.CreateDocumentContent(original)
	require.NoError(t, err)

	// Retrieve content
	retrieved, err := db.GetDocumentContent("doc-456", 1)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "content-456", retrieved.ID)
	assert.Equal(t, "doc-456", retrieved.DocumentID)
	assert.Equal(t, 1, retrieved.Version)
}

func TestGetLatestDocumentContent(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create multiple versions
	for i := 1; i <= 3; i++ {
		contentStr := fmt.Sprintf("<p>Version %d</p>", i)
		hashStr := fmt.Sprintf("hash-v%d", i)
		content := &models.DocumentContent{
			ID:          fmt.Sprintf("content-v%d", i),
			DocumentID:  "doc-latest",
			Version:     i,
			ContentType: "html",
			Content:     &contentStr,
			ContentHash: &hashStr,
			SizeBytes:   20,
		}

		content.SetTimestamps()
		err := db.CreateDocumentContent(content)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
	}

	// Get latest
	latest, err := db.GetLatestDocumentContent("doc-latest")
	assert.NoError(t, err)
	assert.Equal(t, 3, latest.Version)
}

// ========================================================================
// DOCUMENT SPACE OPERATIONS TESTS
// ========================================================================

func TestCreateDocumentSpace(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	space := &models.DocumentSpace{
		ID:       "space-123",
		Key:      "TEST",
		Name:     "Test Space",
		OwnerID:  "user-1",
		IsPublic: true,
	}


	space.SetTimestamps()

	space.SetTimestamps()
	err := db.CreateDocumentSpace(space)
	assert.NoError(t, err)
	assert.Greater(t, space.Created, int64(0))
	assert.Greater(t, space.Modified, int64(0))
}

func TestGetDocumentSpace(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create space
	original := &models.DocumentSpace{
		ID:       "space-456",
		Key:      "DEV",
		Name:     "Development",
		OwnerID:  "user-1",
		IsPublic: false,
	}

	original.SetTimestamps()
	err := db.CreateDocumentSpace(original)
	require.NoError(t, err)

	// Retrieve space
	retrieved, err := db.GetDocumentSpace("space-456")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "space-456", retrieved.ID)
	assert.Equal(t, "DEV", retrieved.Key)
	assert.Equal(t, "Development", retrieved.Name)
}

func TestListDocumentSpaces(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create spaces
	spaces := []*models.DocumentSpace{
		{
			ID:       "space-1",
			Key:      "SPC1",
			Name:     "Space 1",
			OwnerID:  "user-1",
			IsPublic: true,
		},
		{
			ID:       "space-2",
			Key:      "SPC2",
			Name:     "Space 2",
			OwnerID:  "user-1",
			IsPublic: false,
		},
	}

	for _, space := range spaces {
		space.SetTimestamps()
		err := db.CreateDocumentSpace(space)
		require.NoError(t, err)
	}

	// List all
	allSpaces, err := db.ListDocumentSpaces(nil)
	assert.NoError(t, err)
	assert.Len(t, allSpaces, 2)
}

// ========================================================================
// DOCUMENT COLLABORATION OPERATIONS TESTS
// ========================================================================

func TestCreateDocumentWatcher(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	watcher := &models.DocumentWatcher{
		ID:                "watcher-123",
		DocumentID:        "doc-123",
		UserID:            "user-123",
		NotificationLevel: "all",
	}


	watcher.SetTimestamps()

	watcher.SetTimestamps()
	err := db.CreateDocumentWatcher(watcher)
	assert.NoError(t, err)
	assert.Greater(t, watcher.Created, int64(0))
}

func TestGetDocumentWatchers(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create watchers
	watchers := []*models.DocumentWatcher{
		{
			ID:                "watcher-1",
			DocumentID:        "doc-watchers",
			UserID:            "user-1",
			NotificationLevel: "all",
		},
		{
			ID:                "watcher-2",
			DocumentID:        "doc-watchers",
			UserID:            "user-2",
			NotificationLevel: "mentions",
		},
	}

	for _, watcher := range watchers {
		watcher.SetTimestamps()
		err := db.CreateDocumentWatcher(watcher)
		require.NoError(t, err)
	}

	// Get watchers
	retrieved, err := db.GetDocumentWatchers("doc-watchers")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestDeleteDocumentWatcher(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create watcher
	watcher := &models.DocumentWatcher{
		ID:                "watcher-delete",
		DocumentID:        "doc-delete-watcher",
		UserID:            "user-delete",
		NotificationLevel: "all",
	}

	watcher.SetTimestamps()
	err := db.CreateDocumentWatcher(watcher)
	require.NoError(t, err)

	// Delete watcher
	err = db.DeleteDocumentWatcher("doc-delete-watcher", "user-delete")
	assert.NoError(t, err)
}

// ========================================================================
// DOCUMENT TAG OPERATIONS TESTS
// ========================================================================

func TestCreateDocumentTag(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	tag := &models.DocumentTag{
		ID:   "tag-123",
		Name: "important",
	}


	tag.SetTimestamps()

	tag.SetTimestamps()
	err := db.CreateDocumentTag(tag)
	assert.NoError(t, err)
	assert.Greater(t, tag.Created, int64(0))
}

func TestGetDocumentTag(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create tag
	original := &models.DocumentTag{
		ID:   "tag-456",
		Name: "urgent",
	}

	original.SetTimestamps()
	err := db.CreateDocumentTag(original)
	require.NoError(t, err)

	// Retrieve tag
	retrieved, err := db.GetDocumentTag("tag-456")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "tag-456", retrieved.ID)
	assert.Equal(t, "urgent", retrieved.Name)
}

func TestGetOrCreateDocumentTag(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// First call should create
	tag1, err := db.GetOrCreateDocumentTag("new-tag")
	assert.NoError(t, err)
	assert.NotNil(t, tag1)
	assert.Equal(t, "new-tag", tag1.Name)

	// Second call should return existing
	tag2, err := db.GetOrCreateDocumentTag("new-tag")
	assert.NoError(t, err)
	assert.NotNil(t, tag2)
	assert.Equal(t, tag1.ID, tag2.ID)
}

// ========================================================================
// DOCUMENT TEMPLATE OPERATIONS TESTS
// ========================================================================

func TestCreateDocumentTemplate(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	spaceID := "space-1"
	template := &models.DocumentTemplate{
		ID:              "template-123",
		Name:            "Meeting Notes",
		SpaceID:         &spaceID,
		TypeID:          "type-page",
		CreatorID:       "user-1",
		ContentTemplate: "# Meeting Notes\n\nDate: {{date}}",
		IsPublic:        true,
	}
	template.SetTimestamps()

	template.SetTimestamps()
	err := db.CreateDocumentTemplate(template)
	assert.NoError(t, err)
	assert.Greater(t, template.Created, int64(0))
}

func TestGetDocumentTemplate(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create template
	spaceID := "space-1"
	original := &models.DocumentTemplate{
		ID:              "template-456",
		Name:            "Project Plan",
		SpaceID:         &spaceID,
		TypeID:          "type-page",
		CreatorID:       "user-1",
		ContentTemplate: "# Project: {{project_name}}",
		IsPublic:        false,
	}
	original.SetTimestamps()
	err := db.CreateDocumentTemplate(original)
	require.NoError(t, err)

	// Retrieve template
	retrieved, err := db.GetDocumentTemplate("template-456")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "template-456", retrieved.ID)
	assert.Equal(t, "Project Plan", retrieved.Name)
}

func TestIncrementTemplateUseCount(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create template
	spaceID := "space-1"
	template := &models.DocumentTemplate{
		ID:              "template-count",
		Name:            "Use Count Test",
		SpaceID:         &spaceID,
		TypeID:          "type-page",
		CreatorID:       "user-1",
		ContentTemplate: "Test",
		UseCount:        0,
	}

	template.SetTimestamps()
	err := db.CreateDocumentTemplate(template)
	require.NoError(t, err)

	// Increment use count
	err = db.IncrementTemplateUseCount("template-count")
	assert.NoError(t, err)

	// Verify count incremented
	retrieved, err := db.GetDocumentTemplate("template-count")
	assert.NoError(t, err)
	assert.Equal(t, 1, retrieved.UseCount)
}

// ========================================================================
// DOCUMENT ANALYTICS OPERATIONS TESTS
// ========================================================================

func TestCreateDocumentViewHistory(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	userID := "user-123"
	duration := 30
	view := &models.DocumentViewHistory{
		ID:           "view-123",
		DocumentID:   "doc-123",
		UserID:       &userID,
		ViewDuration: &duration,
	}

	view.SetTimestamps()

	view.SetTimestamps()
	err := db.CreateDocumentViewHistory(view)
	assert.NoError(t, err)
	assert.Greater(t, view.Timestamp, int64(0))
}

func TestGetDocumentViewHistory(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create view history
	for i := 1; i <= 5; i++ {
		userID := fmt.Sprintf("user-%d", i)
		duration := i * 10
		view := &models.DocumentViewHistory{
			ID:           fmt.Sprintf("view-%d", i),
			DocumentID:   "doc-views",
			UserID:       &userID,
			ViewDuration: &duration,
		}

		view.SetTimestamps()
		err := db.CreateDocumentViewHistory(view)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
	}

	// Get history
	history, err := db.GetDocumentViewHistory("doc-views", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, history, 5)
}

func TestCreateDocumentAnalytics(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	analytics := &models.DocumentAnalytics{
		ID:               "analytics-123",
		DocumentID:       "doc-123",
		TotalViews:       100,
		UniqueViewers:    50,
		TotalEdits:       10,
		TotalComments:    20,
		TotalReactions:   15,
		TotalWatchers:    5,
		PopularityScore:  27.5,
	}


	analytics.SetTimestamps()

	analytics.SetTimestamps()
	err := db.CreateDocumentAnalytics(analytics)
	assert.NoError(t, err)
	assert.Greater(t, analytics.Updated, int64(0))
}

func TestGetDocumentAnalytics(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create analytics
	original := &models.DocumentAnalytics{
		ID:              "analytics-456",
		DocumentID:      "doc-456",
		TotalViews:      200,
		UniqueViewers:   100,
		PopularityScore: 50.0,
	}

	original.SetTimestamps()
	err := db.CreateDocumentAnalytics(original)
	require.NoError(t, err)

	// Retrieve analytics
	retrieved, err := db.GetDocumentAnalytics("doc-456")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "doc-456", retrieved.DocumentID)
	assert.Equal(t, 200, retrieved.TotalViews)
}

// ========================================================================
// DOCUMENT ATTACHMENT OPERATIONS TESTS
// ========================================================================

func TestCreateDocumentAttachment(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	attachment := &models.DocumentAttachment{
		ID:               "attach-123",
		DocumentID:       "doc-123",
		Filename:         "test.pdf",
		OriginalFilename: "test.pdf",
		MimeType:         "application/pdf",
		SizeBytes:        1024,
		StoragePath:      "/storage/test.pdf",
		Checksum:         "checksum123",
		UploaderID:       "user-123",
		Version:          1,
	}


	attachment.SetTimestamps()

	attachment.SetTimestamps()
	err := db.CreateDocumentAttachment(attachment)
	assert.NoError(t, err)
	assert.Greater(t, attachment.Created, int64(0))
}

func TestGetDocumentAttachment(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create attachment
	original := &models.DocumentAttachment{
		ID:               "attach-456",
		DocumentID:       "doc-456",
		Filename:         "image.png",
		OriginalFilename: "image.png",
		MimeType:         "image/png",
		SizeBytes:        2048,
		StoragePath:      "/storage/image.png",
		Checksum:         "checksum456",
		UploaderID:       "user-456",
		Version:          1,
	}

	original.SetTimestamps()
	err := db.CreateDocumentAttachment(original)
	require.NoError(t, err)

	// Retrieve attachment
	retrieved, err := db.GetDocumentAttachment("attach-456")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "attach-456", retrieved.ID)
	assert.Equal(t, "image.png", retrieved.Filename)
}

func TestListDocumentAttachments(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create attachments
	for i := 1; i <= 3; i++ {
		attachment := &models.DocumentAttachment{
			ID:               fmt.Sprintf("attach-%d", i),
			DocumentID:       "doc-attachments",
			Filename:         fmt.Sprintf("file%d.pdf", i),
			OriginalFilename: fmt.Sprintf("file%d.pdf", i),
			MimeType:         "application/pdf",
			SizeBytes:        1024 * i,
			StoragePath:      fmt.Sprintf("/storage/file%d.pdf", i),
			Checksum:         fmt.Sprintf("checksum%d", i),
			UploaderID:       "user-1",
			Version:          1,
		}

		attachment.SetTimestamps()
		err := db.CreateDocumentAttachment(attachment)
		require.NoError(t, err)
	}

	// List attachments
	attachments, err := db.ListDocumentAttachments("doc-attachments")
	assert.NoError(t, err)
	assert.Len(t, attachments, 3)
}

func TestDeleteDocumentAttachment(t *testing.T) {
	db := setupDocumentTestDB(t)
	defer db.Close()

	// Create attachment
	attachment := &models.DocumentAttachment{
		ID:               "attach-delete",
		DocumentID:       "doc-delete",
		Filename:         "delete.pdf",
		OriginalFilename: "delete.pdf",
		MimeType:         "application/pdf",
		SizeBytes:        1024,
		StoragePath:      "/storage/delete.pdf",
		Checksum:         "checksumdelete",
		UploaderID:       "user-1",
		Version:          1,
	}

	attachment.SetTimestamps()
	err := db.CreateDocumentAttachment(attachment)
	require.NoError(t, err)

	// Delete attachment
	err = db.DeleteDocumentAttachment("attach-delete")
	assert.NoError(t, err)
}

// ========================================================================
// HELPER FUNCTIONS
// ========================================================================

// Helper functions are in database_documents_impl.go
