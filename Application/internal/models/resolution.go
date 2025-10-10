package models

// Resolution represents how a ticket was resolved
type Resolution struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Default resolution IDs
const (
	ResolutionIDFixed            = "resolution-fixed"
	ResolutionIDWontFix          = "resolution-wont-fix"
	ResolutionIDDuplicate        = "resolution-duplicate"
	ResolutionIDIncomplete       = "resolution-incomplete"
	ResolutionIDCannotReproduce  = "resolution-cannot-reproduce"
	ResolutionIDDone             = "resolution-done"
)

// GetDisplayName returns a user-friendly display name
func (r *Resolution) GetDisplayName() string {
	if r.Title != "" {
		return r.Title
	}
	return "Unknown Resolution"
}
