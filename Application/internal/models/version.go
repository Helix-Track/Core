package models

// Version represents a product version/release
type Version struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	ProjectID   string `json:"projectId" db:"project_id" binding:"required"`
	StartDate   *int64 `json:"startDate,omitempty" db:"start_date"`   // Unix timestamp, pointer for nullable
	ReleaseDate *int64 `json:"releaseDate,omitempty" db:"release_date"` // Unix timestamp, pointer for nullable
	Released    bool   `json:"released" db:"released"`
	Archived    bool   `json:"archived" db:"archived"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// TicketVersionMapping represents the relationship between a ticket and a version
type TicketVersionMapping struct {
	ID        string `json:"id" db:"id"`
	TicketID  string `json:"ticketId" db:"ticket_id" binding:"required"`
	VersionID string `json:"versionId" db:"version_id" binding:"required"`
	Created   int64  `json:"created" db:"created"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

// IsReleased checks if the version has been released
func (v *Version) IsReleased() bool {
	return v.Released
}

// IsArchived checks if the version is archived
func (v *Version) IsArchived() bool {
	return v.Archived
}

// IsActive checks if the version is active (not archived and not deleted)
func (v *Version) IsActive() bool {
	return !v.Archived && !v.Deleted
}
