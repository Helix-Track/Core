package models

// TicketRelationship represents a relationship between two tickets
type TicketRelationship struct {
	ID                        string `json:"id" db:"id"`
	TicketID                  string `json:"ticketId" db:"ticket_id" binding:"required"`
	ChildTicketID             string `json:"childTicketId" db:"child_ticket_id" binding:"required"`
	TicketRelationshipTypeID  string `json:"ticketRelationshipTypeId" db:"ticket_relationship_type_id" binding:"required"`
	Created                   int64  `json:"created" db:"created"`
	Modified                  int64  `json:"modified" db:"modified"`
	Deleted                   bool   `json:"deleted" db:"deleted"`
}

// TicketRelationshipType represents the type of relationship between tickets
type TicketRelationshipType struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Default relationship type IDs
const (
	RelationshipTypeBlocks       = "rel-blocks"
	RelationshipTypeBlockedBy    = "rel-blocked-by"
	RelationshipTypeRelatesTo    = "rel-relates-to"
	RelationshipTypeDuplicates   = "rel-duplicates"
	RelationshipTypeDuplicatedBy = "rel-duplicated-by"
	RelationshipTypeCauses       = "rel-causes"
	RelationshipTypeCausedBy     = "rel-caused-by"
	RelationshipTypeClones       = "rel-clones"
	RelationshipTypeClonedBy     = "rel-cloned-by"
)
