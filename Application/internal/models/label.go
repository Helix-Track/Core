package models

import "time"

type Label struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Color     string `json:"color" db:"color"`
	ProjectID string `json:"projectId" db:"project_id"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
	Version   int    `json:"version" db:"version"`
}

func NewLabel(id, name, color, projectID string) *Label {
	now := time.Now().Unix()
	return &Label{ID: id, Name: name, Color: color, ProjectID: projectID,
		Created: now, Modified: now, Deleted: false, Version: 1}
}
