package models

// ProjectCategory represents a category for organizing projects
type ProjectCategory struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"`
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Common project category constants
const (
	CategorySoftware     = "Software Development"
	CategoryInfrastructure = "Infrastructure"
	CategoryMarketing    = "Marketing"
	CategorySales        = "Sales"
	CategorySupport      = "Customer Support"
	CategoryResearch     = "Research & Development"
	CategoryInternal     = "Internal"
	CategoryExternal     = "External"
)

// IsValid checks if the category has valid data
func (pc *ProjectCategory) IsValid() bool {
	return pc.Title != ""
}

// GetDisplayName returns a user-friendly display name
func (pc *ProjectCategory) GetDisplayName() string {
	if pc.Title != "" {
		return pc.Title
	}
	return "Uncategorized"
}
