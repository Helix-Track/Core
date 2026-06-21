package models

import "time"

type UserRole struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
	Version     int    `json:"version" db:"version"`
}

func NewUserRole(id, name, description string) *UserRole {
	now := time.Now().Unix()
	return &UserRole{ID: id, Name: name, Description: description,
		Created: now, Modified: now, Deleted: false, Version: 1}
}
