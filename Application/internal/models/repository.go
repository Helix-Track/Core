package models

// Repository represents a source code repository (Git, SVN, Mercurial, etc.)
type Repository struct {
	ID               string `json:"id" db:"id"`
	Repository       string `json:"repository" db:"repository" binding:"required"` // Git URL or repository path
	Description      string `json:"description,omitempty" db:"description"`
	RepositoryTypeID string `json:"repositoryTypeId" db:"repository_type_id" binding:"required"`
	Created          int64  `json:"created" db:"created"`
	Modified         int64  `json:"modified" db:"modified"`
	Deleted          bool   `json:"deleted" db:"deleted"`
}

// RepositoryType represents the type of repository (git, svn, mercurial, etc.)
type RepositoryType struct {
	ID          string `json:"id" db:"id"`
	Title       string `json:"title" db:"title" binding:"required"` // git, svn, mercurial
	Description string `json:"description,omitempty" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
}

// Default repository type IDs
const (
	RepositoryTypeGit       = "repo-type-git"
	RepositoryTypeSVN       = "repo-type-svn"
	RepositoryTypeMercurial = "repo-type-mercurial"
	RepositoryTypeCVS       = "repo-type-cvs"
	RepositoryTypePerforce  = "repo-type-perforce"
)

// RepositoryProjectMapping maps repositories to projects
type RepositoryProjectMapping struct {
	ID           string `json:"id" db:"id"`
	RepositoryID string `json:"repositoryId" db:"repository_id" binding:"required"`
	ProjectID    string `json:"projectId" db:"project_id" binding:"required"`
	Created      int64  `json:"created" db:"created"`
	Modified     int64  `json:"modified" db:"modified"`
	Deleted      bool   `json:"deleted" db:"deleted"`
}

// RepositoryCommitTicketMapping maps repository commits to tickets
type RepositoryCommitTicketMapping struct {
	ID           string `json:"id" db:"id"`
	RepositoryID string `json:"repositoryId" db:"repository_id" binding:"required"`
	TicketID     string `json:"ticketId" db:"ticket_id" binding:"required"`
	CommitHash   string `json:"commitHash" db:"commit_hash" binding:"required"`
	Created      int64  `json:"created" db:"created"`
	Modified     int64  `json:"modified" db:"modified"`
	Deleted      bool   `json:"deleted" db:"deleted"`
}
