package database

import "helixtrack.ru/core/internal/models"

// DocumentDatabase defines all document-related database operations
// This interface extends the main Database interface for document operations
type DocumentDatabase interface {
	// ========================================================================
	// CORE DOCUMENT OPERATIONS
	// ========================================================================

	// CreateDocument creates a new document
	CreateDocument(doc *models.Document) error

	// GetDocument retrieves a document by ID
	GetDocument(id string) (*models.Document, error)

	// ListDocuments lists documents with optional filters
	ListDocuments(filters map[string]interface{}, limit, offset int) ([]*models.Document, error)

	// UpdateDocument updates an existing document (with optimistic locking)
	UpdateDocument(doc *models.Document) error

	// DeleteDocument soft-deletes a document
	DeleteDocument(id string) error

	// RestoreDocument restores a soft-deleted document
	RestoreDocument(id string) error

	// ArchiveDocument archives a document
	ArchiveDocument(id string) error

	// UnarchiveDocument unarchives a document
	UnarchiveDocument(id string) error

	// DuplicateDocument creates a copy of a document
	DuplicateDocument(id string, newTitle string, userID string) (*models.Document, error)

	// MoveDocument moves a document to a different space
	MoveDocument(id string, newSpaceID string) error

	// SetDocumentParent sets the parent document for hierarchy
	SetDocumentParent(id string, parentID string) error

	// GetDocumentChildren gets all child documents
	GetDocumentChildren(id string) ([]*models.Document, error)

	// GetDocumentHierarchy gets the full document tree
	GetDocumentHierarchy(rootID string) ([]*models.Document, error)

	// GetDocumentBreadcrumb gets the breadcrumb trail for a document
	GetDocumentBreadcrumb(id string) ([]*models.Document, error)

	// SearchDocuments performs full-text search on documents
	SearchDocuments(query string, filters map[string]interface{}, limit, offset int) ([]*models.Document, error)

	// GetRelatedDocuments gets related documents
	GetRelatedDocuments(id string, limit int) ([]*models.Document, error)

	// PublishDocument publishes a document
	PublishDocument(id string) error

	// UnpublishDocument unpublishes a document
	UnpublishDocument(id string) error

	// ========================================================================
	// DOCUMENT CONTENT OPERATIONS
	// ========================================================================

	// CreateDocumentContent creates document content
	CreateDocumentContent(content *models.DocumentContent) error

	// GetDocumentContent gets content for a specific version
	GetDocumentContent(documentID string, version int) (*models.DocumentContent, error)

	// GetLatestDocumentContent gets the latest content
	GetLatestDocumentContent(documentID string) (*models.DocumentContent, error)

	// UpdateDocumentContent updates document content
	UpdateDocumentContent(content *models.DocumentContent) error

	// ========================================================================
	// DOCUMENT SPACE OPERATIONS
	// ========================================================================

	// CreateDocumentSpace creates a new document space
	CreateDocumentSpace(space *models.DocumentSpace) error

	// GetDocumentSpace retrieves a space by ID
	GetDocumentSpace(id string) (*models.DocumentSpace, error)

	// ListDocumentSpaces lists all document spaces
	ListDocumentSpaces(filters map[string]interface{}) ([]*models.DocumentSpace, error)

	// UpdateDocumentSpace updates a space
	UpdateDocumentSpace(space *models.DocumentSpace) error

	// DeleteDocumentSpace deletes a space
	DeleteDocumentSpace(id string) error

	// ========================================================================
	// DOCUMENT VERSION OPERATIONS
	// ========================================================================

	// CreateDocumentVersion creates a new version record
	CreateDocumentVersion(version *models.DocumentVersion) error

	// GetDocumentVersion gets a specific version
	GetDocumentVersion(id string) (*models.DocumentVersion, error)

	// ListDocumentVersions lists all versions for a document
	ListDocumentVersions(documentID string) ([]*models.DocumentVersion, error)

	// CompareDocumentVersions compares two versions
	CompareDocumentVersions(documentID string, fromVersion, toVersion int) (*models.DocumentVersionDiff, error)

	// RestoreDocumentVersion restores a document to a specific version
	RestoreDocumentVersion(documentID string, versionNumber int, userID string) error

	// CreateVersionLabel creates a label for a version
	CreateVersionLabel(label *models.DocumentVersionLabel) error

	// GetVersionLabels gets all labels for a version
	GetVersionLabels(versionID string) ([]*models.DocumentVersionLabel, error)

	// CreateVersionTag creates a tag for a version
	CreateVersionTag(tag *models.DocumentVersionTag) error

	// GetVersionTags gets all tags for a version
	GetVersionTags(versionID string) ([]*models.DocumentVersionTag, error)

	// CreateVersionComment creates a comment on a version
	CreateVersionComment(comment *models.DocumentVersionComment) error

	// GetVersionComments gets all comments for a version
	GetVersionComments(versionID string) ([]*models.DocumentVersionComment, error)

	// CreateVersionMention creates a mention in a version
	CreateVersionMention(mention *models.DocumentVersionMention) error

	// GetVersionMentions gets all mentions for a version
	GetVersionMentions(versionID string) ([]*models.DocumentVersionMention, error)

	// GetVersionDiff gets cached diff between versions
	GetVersionDiff(documentID string, fromVersion, toVersion int, diffType string) (*models.DocumentVersionDiff, error)

	// CreateVersionDiff creates a cached diff
	CreateVersionDiff(diff *models.DocumentVersionDiff) error

	// ========================================================================
	// DOCUMENT COLLABORATION OPERATIONS
	// ========================================================================

	// CreateCommentDocumentMapping creates a comment-document link
	CreateCommentDocumentMapping(mapping *models.CommentDocumentMapping) error

	// GetDocumentComments gets all comments for a document
	GetDocumentComments(documentID string) ([]*models.CommentDocumentMapping, error)

	// DeleteCommentDocumentMapping removes a comment from a document
	DeleteCommentDocumentMapping(commentID, documentID string) error

	// CreateInlineComment creates an inline comment
	CreateInlineComment(comment *models.DocumentInlineComment) error

	// GetInlineComments gets all inline comments for a document
	GetInlineComments(documentID string) ([]*models.DocumentInlineComment, error)

	// ResolveInlineComment marks an inline comment as resolved
	ResolveInlineComment(id string) error

	// CreateDocumentWatcher creates a watcher subscription
	CreateDocumentWatcher(watcher *models.DocumentWatcher) error

	// GetDocumentWatchers gets all watchers for a document
	GetDocumentWatchers(documentID string) ([]*models.DocumentWatcher, error)

	// DeleteDocumentWatcher removes a watcher
	DeleteDocumentWatcher(documentID, userID string) error

	// ========================================================================
	// DOCUMENT LABEL/TAG OPERATIONS
	// ========================================================================

	// CreateLabelDocumentMapping creates a label-document link (uses core label)
	CreateLabelDocumentMapping(mapping *models.LabelDocumentMapping) error

	// GetDocumentLabels gets all labels for a document
	GetDocumentLabels(documentID string) ([]*models.LabelDocumentMapping, error)

	// DeleteLabelDocumentMapping removes a label from a document
	DeleteLabelDocumentMapping(labelID, documentID string) error

	// CreateDocumentTag creates a new tag
	CreateDocumentTag(tag *models.DocumentTag) error

	// GetDocumentTag gets a tag by ID
	GetDocumentTag(id string) (*models.DocumentTag, error)

	// GetOrCreateDocumentTag gets existing tag or creates new one
	GetOrCreateDocumentTag(name string) (*models.DocumentTag, error)

	// CreateDocumentTagMapping creates a tag-document link
	CreateDocumentTagMapping(mapping *models.DocumentTagMapping) error

	// GetDocumentTags gets all tags for a document
	GetDocumentTags(documentID string) ([]*models.DocumentTagMapping, error)

	// DeleteDocumentTagMapping removes a tag from a document
	DeleteDocumentTagMapping(tagID, documentID string) error

	// ========================================================================
	// DOCUMENT VOTE/REACTION OPERATIONS
	// ========================================================================

	// CreateVoteMapping creates a vote/reaction (generic system)
	CreateVoteMapping(vote *models.VoteMapping) error

	// GetEntityVotes gets all votes for an entity
	GetEntityVotes(entityType, entityID string) ([]*models.VoteMapping, error)

	// DeleteVoteMapping removes a vote
	DeleteVoteMapping(entityType, entityID, userID, voteType string) error

	// GetVoteCount gets vote count for an entity
	GetVoteCount(entityType, entityID string) (int, error)

	// ========================================================================
	// DOCUMENT ENTITY LINK OPERATIONS
	// ========================================================================

	// CreateDocumentEntityLink creates a link to any entity
	CreateDocumentEntityLink(link *models.DocumentEntityLink) error

	// GetDocumentEntityLinks gets all entity links for a document
	GetDocumentEntityLinks(documentID string) ([]*models.DocumentEntityLink, error)

	// GetEntityDocuments gets all documents linked to an entity
	GetEntityDocuments(entityType, entityID string) ([]*models.DocumentEntityLink, error)

	// DeleteDocumentEntityLink removes an entity link
	DeleteDocumentEntityLink(id string) error

	// CreateDocumentRelationship creates a document-to-document relationship
	CreateDocumentRelationship(rel *models.DocumentRelationship) error

	// GetDocumentRelationships gets all relationships for a document
	GetDocumentRelationships(documentID string) ([]*models.DocumentRelationship, error)

	// DeleteDocumentRelationship removes a relationship
	DeleteDocumentRelationship(id string) error

	// ========================================================================
	// DOCUMENT TEMPLATE OPERATIONS
	// ========================================================================

	// CreateDocumentTemplate creates a new template
	CreateDocumentTemplate(template *models.DocumentTemplate) error

	// GetDocumentTemplate gets a template by ID
	GetDocumentTemplate(id string) (*models.DocumentTemplate, error)

	// ListDocumentTemplates lists all templates
	ListDocumentTemplates(filters map[string]interface{}) ([]*models.DocumentTemplate, error)

	// UpdateDocumentTemplate updates a template
	UpdateDocumentTemplate(template *models.DocumentTemplate) error

	// DeleteDocumentTemplate deletes a template
	DeleteDocumentTemplate(id string) error

	// IncrementTemplateUseCount increments use counter
	IncrementTemplateUseCount(id string) error

	// CreateDocumentBlueprint creates a blueprint
	CreateDocumentBlueprint(blueprint *models.DocumentBlueprint) error

	// GetDocumentBlueprint gets a blueprint by ID
	GetDocumentBlueprint(id string) (*models.DocumentBlueprint, error)

	// ListDocumentBlueprints lists all blueprints
	ListDocumentBlueprints(filters map[string]interface{}) ([]*models.DocumentBlueprint, error)

	// ========================================================================
	// DOCUMENT ANALYTICS OPERATIONS
	// ========================================================================

	// CreateDocumentViewHistory records a document view
	CreateDocumentViewHistory(view *models.DocumentViewHistory) error

	// GetDocumentViewHistory gets view history for a document
	GetDocumentViewHistory(documentID string, limit, offset int) ([]*models.DocumentViewHistory, error)

	// GetDocumentAnalytics gets analytics for a document
	GetDocumentAnalytics(documentID string) (*models.DocumentAnalytics, error)

	// CreateDocumentAnalytics creates analytics record
	CreateDocumentAnalytics(analytics *models.DocumentAnalytics) error

	// UpdateDocumentAnalytics updates analytics record
	UpdateDocumentAnalytics(analytics *models.DocumentAnalytics) error

	// GetPopularDocuments gets most popular documents
	GetPopularDocuments(limit int) ([]*models.DocumentAnalytics, error)

	// ========================================================================
	// DOCUMENT ATTACHMENT OPERATIONS
	// ========================================================================

	// CreateDocumentAttachment creates an attachment record
	CreateDocumentAttachment(attachment *models.DocumentAttachment) error

	// GetDocumentAttachment gets an attachment by ID
	GetDocumentAttachment(id string) (*models.DocumentAttachment, error)

	// ListDocumentAttachments lists all attachments for a document
	ListDocumentAttachments(documentID string) ([]*models.DocumentAttachment, error)

	// UpdateDocumentAttachment updates an attachment
	UpdateDocumentAttachment(attachment *models.DocumentAttachment) error

	// DeleteDocumentAttachment deletes an attachment
	DeleteDocumentAttachment(id string) error
}
