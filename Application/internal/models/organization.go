package models

type Organization struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

type OrganizationAccountMapping struct {
	ID             string `json:"id" db:"id"`
	OrganizationID string `json:"organizationId" db:"organization_id" binding:"required"`
	AccountID      string `json:"accountId" db:"account_id" binding:"required"`
	Created        int64  `json:"created" db:"created"`
	Modified       int64  `json:"modified" db:"modified"`
	Deleted        bool   `json:"deleted" db:"deleted"`
}
