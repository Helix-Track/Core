package models

import "time"

type Permission struct {
	ID        string `json:"id" db:"id"`
	RoleID    string `json:"roleId" db:"role_id"`
	Resource  string `json:"resource" db:"resource"`
	Action    string `json:"action" db:"action"`
	Created   int64  `json:"created" db:"created"`
	Modified  int64  `json:"modified" db:"modified"`
	Deleted   bool   `json:"deleted" db:"deleted"`
	Version   int    `json:"version" db:"version"`
}

func NewPermission(id, roleID, resource, action string) *Permission {
	now := time.Now().Unix()
	return &Permission{ID: id, RoleID: roleID, Resource: resource, Action: action,
		Created: now, Modified: now, Deleted: false, Version: 1}
}
