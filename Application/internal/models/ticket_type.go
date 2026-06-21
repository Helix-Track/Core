package models

import "time"

// TicketType represents a ticket type (Bug, Feature, Task, Epic, Story, SubTask)
type TicketType struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Icon        string `json:"icon" db:"icon"`
	Color       string `json:"color" db:"color"`
	IsDefault   bool   `json:"isDefault" db:"is_default"`
	Created     int64  `json:"created" db:"created"`
	Modified    int64  `json:"modified" db:"modified"`
	Deleted     bool   `json:"deleted" db:"deleted"`
	Version     int    `json:"version" db:"version"`
}

func NewTicketType(id, name, description, icon, color string, isDefault bool) *TicketType {
	now := time.Now().Unix()
	return &TicketType{
		ID: id, Name: name, Description: description, Icon: icon, Color: color,
		IsDefault: isDefault, Created: now, Modified: now, Deleted: false, Version: 1,
	}
}
