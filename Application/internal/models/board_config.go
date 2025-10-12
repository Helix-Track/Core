package models

// BoardColumn represents a column on a board (for Kanban/Scrum)
type BoardColumn struct {
	ID       string  `json:"id" db:"id"`
	BoardID  string  `json:"boardId" db:"board_id" binding:"required"`
	Title    string  `json:"title" db:"title" binding:"required"`
	StatusID *string `json:"statusId,omitempty" db:"status_id"` // Maps to ticket_status
	Position int     `json:"position" db:"position" binding:"required"`
	MaxItems *int    `json:"maxItems,omitempty" db:"max_items"` // WIP limit
	Created  int64   `json:"created" db:"created"`
	Modified int64   `json:"modified" db:"modified"`
	Deleted  bool    `json:"deleted" db:"deleted"`
}

// BoardSwimlane represents a swimlane on a board
type BoardSwimlane struct {
	ID       string  `json:"id" db:"id"`
	BoardID  string  `json:"boardId" db:"board_id" binding:"required"`
	Title    string  `json:"title" db:"title" binding:"required"`
	Query    *string `json:"query,omitempty" db:"query"` // JQL-like query for swimlane
	Position int     `json:"position" db:"position" binding:"required"`
	Created  int64   `json:"created" db:"created"`
	Modified int64   `json:"modified" db:"modified"`
	Deleted  bool    `json:"deleted" db:"deleted"`
}

// BoardQuickFilter represents a quick filter on a board
type BoardQuickFilter struct {
	ID       string  `json:"id" db:"id"`
	BoardID  string  `json:"boardId" db:"board_id" binding:"required"`
	Title    string  `json:"title" db:"title" binding:"required"`
	Query    *string `json:"query,omitempty" db:"query"` // JQL-like query for quick filter
	Position int     `json:"position" db:"position" binding:"required"`
	Created  int64   `json:"created" db:"created"`
	Deleted  bool    `json:"deleted" db:"deleted"`
}

// HasWIPLimit checks if the column has a WIP (Work In Progress) limit
func (bc *BoardColumn) HasWIPLimit() bool {
	return bc.MaxItems != nil && *bc.MaxItems > 0
}

// IsWIPLimitExceeded checks if the current item count exceeds the WIP limit
func (bc *BoardColumn) IsWIPLimitExceeded(currentCount int) bool {
	if !bc.HasWIPLimit() {
		return false
	}
	return currentCount > *bc.MaxItems
}

// HasQuery checks if the swimlane has a query defined
func (bs *BoardSwimlane) HasQuery() bool {
	return bs.Query != nil && *bs.Query != ""
}

// HasQuery checks if the quick filter has a query defined
func (bq *BoardQuickFilter) HasQuery() bool {
	return bq.Query != nil && *bq.Query != ""
}
