package models

import "time"

type Priority struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Level     int    `json:"level" db:"level"`
	Color     string `json:"color" db:"color"`
	IsDefault bool   `json:"isDefault" db:"is_default"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
	Version   int    `json:"version" db:"version"`
}

func NewPriority(id, name, color string, level int, isDefault bool) *Priority {
	now := time.Now().Unix()
	return &Priority{ID: id, Name: name, Level: level, Color: color, IsDefault: isDefault,
		Created: now, Modified: now, Deleted: false, Version: 1}
}
