package models

// Report represents a saved report definition
type Report struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Query       string `json:"query" db:"query"`              // JSON query definition
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// ReportMetaData represents additional metadata for reports
type ReportMetaData struct {
	ID       string `json:"id" db:"id"`
	ReportID string `json:"reportId" db:"report_id" binding:"required"`
	Property string `json:"property" db:"property" binding:"required"`
	Value    string `json:"value" db:"value"`
	Created  int64  `json:"created" db:"created"`
	Modified int64  `json:"modified" db:"modified"`
	Deleted  bool   `json:"deleted" db:"deleted"`
}

// IsValid validates if the report has required fields
func (r *Report) IsValid() bool {
	return r.Title != ""
}

// HasQuery checks if the report has a query definition
func (r *Report) HasQuery() bool {
	return r.Query != ""
}
