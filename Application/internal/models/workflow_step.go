package models

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID         string `json:"id" db:"id"`
	WorkflowID string `json:"workflowId" db:"workflow_id" binding:"required"`
	StatusID   string `json:"statusId" db:"status_id" binding:"required"`
	Position   int    `json:"position" db:"position" binding:"required"`
	Created    int64  `json:"created" db:"created"`
	Modified   int64  `json:"modified" db:"modified"`
	Deleted    bool   `json:"deleted" db:"deleted"`
}
