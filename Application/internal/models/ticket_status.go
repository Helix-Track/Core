package models

import "time"

type TicketStatus struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Category    string `json:"category" db:"category"`
	Color       string `json:"color" db:"color"`
	IsDefault   bool   `json:"isDefault" db:"is_default"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
	Version     int    `json:"version" db:"version"`
}

func NewTicketStatus(id, name, description, category, color string, isDefault bool) *TicketStatus {
	now := time.Now().Unix()
	return &TicketStatus{
		ID: id, Name: name, Description: description, Category: category, Color: color,
		IsDefault: isDefault, Created: now, Modified: now, Deleted: false, Version: 1,
	}
}
