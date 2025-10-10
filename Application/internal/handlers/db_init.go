package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"helixtrack.ru/core/internal/database"
)

// InitializeProjectTables creates the projects, tickets, and comments tables if they don't exist
func InitializeProjectTables(db database.Database) error {
	ctx := context.Background()

	// Create workflow table (simplified)
	workflowSQL := `
		CREATE TABLE IF NOT EXISTS workflow (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
	`
	if _, err := db.Exec(ctx, workflowSQL); err != nil {
		return err
	}

	// Create project table
	projectSQL := `
		CREATE TABLE IF NOT EXISTS project (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			identifier TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			workflow_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS projects_get_by_identifier ON project (identifier);
		CREATE INDEX IF NOT EXISTS projects_get_by_title ON project (title);
		CREATE INDEX IF NOT EXISTS projects_get_by_deleted ON project (deleted);
	`
	if _, err := db.Exec(ctx, projectSQL); err != nil {
		return err
	}

	// Create ticket_type table
	ticketTypeSQL := `
		CREATE TABLE IF NOT EXISTS ticket_type (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL UNIQUE,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS ticket_types_get_by_title ON ticket_type (title);
	`
	if _, err := db.Exec(ctx, ticketTypeSQL); err != nil {
		return err
	}

	// Create ticket_status table
	ticketStatusSQL := `
		CREATE TABLE IF NOT EXISTS ticket_status (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL UNIQUE,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS ticket_statuses_get_by_title ON ticket_status (title);
	`
	if _, err := db.Exec(ctx, ticketStatusSQL); err != nil {
		return err
	}

	// Create ticket table
	ticketSQL := `
		CREATE TABLE IF NOT EXISTS ticket (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			ticket_number INTEGER NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			title TEXT,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			ticket_type_id TEXT NOT NULL,
			ticket_status_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			user_id TEXT,
			estimation REAL NOT NULL DEFAULT 0,
			story_points INTEGER NOT NULL DEFAULT 0,
			creator TEXT NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0,
			UNIQUE (ticket_number, project_id)
		);
		CREATE INDEX IF NOT EXISTS tickets_get_by_project_id ON ticket (project_id);
		CREATE INDEX IF NOT EXISTS tickets_get_by_title ON ticket (title);
		CREATE INDEX IF NOT EXISTS tickets_get_by_deleted ON ticket (deleted);
	`
	if _, err := db.Exec(ctx, ticketSQL); err != nil {
		return err
	}

	// Create comment table
	commentSQL := `
		CREATE TABLE IF NOT EXISTS comment (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			comment TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS comments_get_by_created ON comment (created);
	`
	if _, err := db.Exec(ctx, commentSQL); err != nil {
		return err
	}

	// Create comment_ticket_mapping table
	commentTicketMappingSQL := `
		CREATE TABLE IF NOT EXISTS comment_ticket_mapping (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			comment_id TEXT NOT NULL,
			ticket_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS comment_ticket_mapping_by_ticket ON comment_ticket_mapping (ticket_id);
		CREATE INDEX IF NOT EXISTS comment_ticket_mapping_by_comment ON comment_ticket_mapping (comment_id);
	`
	if _, err := db.Exec(ctx, commentTicketMappingSQL); err != nil {
		return err
	}

	// Seed default data
	if err := seedDefaultData(db); err != nil {
		return err
	}

	return nil
}

// seedDefaultData inserts default workflow, ticket types, and statuses
func seedDefaultData(db database.Database) error {
	ctx := context.Background()
	now := time.Now().Unix()

	// Check if workflow already exists
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM workflow").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Insert default workflow
		defaultWorkflowID := uuid.New().String()
		_, err = db.Exec(ctx, `
			INSERT INTO workflow (id, title, description, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?)
		`, defaultWorkflowID, "Default Workflow", "Default workflow for projects", now, now, 0)
		if err != nil {
			return err
		}
	}

	// Seed ticket types
	ticketTypes := []struct {
		title       string
		description string
	}{
		{"bug", "Bug or defect"},
		{"feature", "New feature request"},
		{"task", "General task"},
		{"story", "User story"},
		{"epic", "Epic"},
	}

	for _, tt := range ticketTypes {
		// Check if exists
		err := db.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_type WHERE title = ?", tt.title).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			_, err = db.Exec(ctx, `
				INSERT INTO ticket_type (id, title, description, created, modified, deleted)
				VALUES (?, ?, ?, ?, ?, ?)
			`, uuid.New().String(), tt.title, tt.description, now, now, 0)
			if err != nil {
				return err
			}
		}
	}

	// Seed ticket statuses
	ticketStatuses := []struct {
		title       string
		description string
	}{
		{"open", "Ticket is open"},
		{"in_progress", "Work in progress"},
		{"review", "Under review"},
		{"testing", "Being tested"},
		{"done", "Completed"},
		{"closed", "Closed"},
	}

	for _, ts := range ticketStatuses {
		// Check if exists
		err := db.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_status WHERE title = ?", ts.title).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			_, err = db.Exec(ctx, `
				INSERT INTO ticket_status (id, title, description, created, modified, deleted)
				VALUES (?, ?, ?, ?, ?, ?)
			`, uuid.New().String(), ts.title, ts.description, now, now, 0)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
