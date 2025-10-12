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
			identifier TEXT UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			workflow_id TEXT,
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

	// Create ticket table with V3 columns
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
			ticket_key TEXT,
			status_id TEXT,
			is_epic BOOLEAN DEFAULT 0,
			epic_id TEXT,
			epic_color TEXT,
			epic_name TEXT,
			is_subtask BOOLEAN DEFAULT 0,
			parent_ticket_id TEXT,
			security_level_id TEXT,
			vote_count INTEGER DEFAULT 0,
			UNIQUE (ticket_number, project_id)
		);
		CREATE INDEX IF NOT EXISTS tickets_get_by_project_id ON ticket (project_id);
		CREATE INDEX IF NOT EXISTS tickets_get_by_title ON ticket (title);
		CREATE INDEX IF NOT EXISTS tickets_get_by_deleted ON ticket (deleted);
		CREATE INDEX IF NOT EXISTS tickets_get_by_is_epic ON ticket (is_epic);
		CREATE INDEX IF NOT EXISTS tickets_get_by_epic_id ON ticket (epic_id);
		CREATE INDEX IF NOT EXISTS tickets_get_by_is_subtask ON ticket (is_subtask);
		CREATE INDEX IF NOT EXISTS tickets_get_by_parent_ticket_id ON ticket (parent_ticket_id);
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

	// Create project_role table
	projectRoleSQL := `
		CREATE TABLE IF NOT EXISTS project_role (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS project_role_by_project_id ON project_role (project_id);
		CREATE INDEX IF NOT EXISTS project_role_by_title ON project_role (title);
		CREATE INDEX IF NOT EXISTS project_role_by_deleted ON project_role (deleted);
	`
	if _, err := db.Exec(ctx, projectRoleSQL); err != nil {
		return err
	}

	// Create project_role_user_mapping table
	projectRoleUserMappingSQL := `
		CREATE TABLE IF NOT EXISTS project_role_user_mapping (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			project_role_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS project_role_user_mapping_by_role ON project_role_user_mapping (project_role_id);
		CREATE INDEX IF NOT EXISTS project_role_user_mapping_by_project ON project_role_user_mapping (project_id);
		CREATE INDEX IF NOT EXISTS project_role_user_mapping_by_user ON project_role_user_mapping (user_id);
		CREATE INDEX IF NOT EXISTS project_role_user_mapping_by_deleted ON project_role_user_mapping (deleted);
	`
	if _, err := db.Exec(ctx, projectRoleUserMappingSQL); err != nil {
		return err
	}

	// Create work_log table
	workLogSQL := `
		CREATE TABLE IF NOT EXISTS work_log (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			ticket_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			time_spent INTEGER NOT NULL,
			work_date INTEGER NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS work_log_by_ticket_id ON work_log (ticket_id);
		CREATE INDEX IF NOT EXISTS work_log_by_user_id ON work_log (user_id);
		CREATE INDEX IF NOT EXISTS work_log_by_work_date ON work_log (work_date);
		CREATE INDEX IF NOT EXISTS work_log_by_deleted ON work_log (deleted);
	`
	if _, err := db.Exec(ctx, workLogSQL); err != nil {
		return err
	}

	// Create user table
	userSQL := `
		CREATE TABLE IF NOT EXISTS user (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS user_by_username ON user (username);
		CREATE INDEX IF NOT EXISTS user_by_email ON user (email);
		CREATE INDEX IF NOT EXISTS user_by_deleted ON user (deleted);
	`
	if _, err := db.Exec(ctx, userSQL); err != nil {
		return err
	}

	// Create team table
	teamSQL := `
		CREATE TABLE IF NOT EXISTS team (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			project_id TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS team_by_project_id ON team (project_id);
		CREATE INDEX IF NOT EXISTS team_by_title ON team (title);
		CREATE INDEX IF NOT EXISTS team_by_deleted ON team (deleted);
	`
	if _, err := db.Exec(ctx, teamSQL); err != nil {
		return err
	}

	// Create team_user table
	teamUserSQL := `
		CREATE TABLE IF NOT EXISTS team_user (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			team_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS team_user_by_team ON team_user (team_id);
		CREATE INDEX IF NOT EXISTS team_user_by_user ON team_user (user_id);
		CREATE INDEX IF NOT EXISTS team_user_by_deleted ON team_user (deleted);
	`
	if _, err := db.Exec(ctx, teamUserSQL); err != nil {
		return err
	}

	// Create security_level table
	securityLevelSQL := `
		CREATE TABLE IF NOT EXISTS security_level (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			project_id TEXT NOT NULL,
			level INTEGER NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS security_level_by_project_id ON security_level (project_id);
		CREATE INDEX IF NOT EXISTS security_level_by_level ON security_level (level);
		CREATE INDEX IF NOT EXISTS security_level_by_deleted ON security_level (deleted);
	`
	if _, err := db.Exec(ctx, securityLevelSQL); err != nil {
		return err
	}

	// Create security_level_permission_mapping table
	securityLevelPermissionMappingSQL := `
		CREATE TABLE IF NOT EXISTS security_level_permission_mapping (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			security_level_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_role_id TEXT,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS security_level_permission_by_level ON security_level_permission_mapping (security_level_id);
		CREATE INDEX IF NOT EXISTS security_level_permission_by_user ON security_level_permission_mapping (user_id);
		CREATE INDEX IF NOT EXISTS security_level_permission_by_team ON security_level_permission_mapping (team_id);
		CREATE INDEX IF NOT EXISTS security_level_permission_by_role ON security_level_permission_mapping (project_role_id);
		CREATE INDEX IF NOT EXISTS security_level_permission_by_deleted ON security_level_permission_mapping (deleted);
	`
	if _, err := db.Exec(ctx, securityLevelPermissionMappingSQL); err != nil {
		return err
	}

	// Create dashboard table
	dashboardSQL := `
		CREATE TABLE IF NOT EXISTS dashboard (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			is_public BOOLEAN DEFAULT 0,
			is_favorite BOOLEAN DEFAULT 0,
			layout TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS dashboard_by_owner_id ON dashboard (owner_id);
		CREATE INDEX IF NOT EXISTS dashboard_by_is_public ON dashboard (is_public);
		CREATE INDEX IF NOT EXISTS dashboard_by_is_favorite ON dashboard (is_favorite);
		CREATE INDEX IF NOT EXISTS dashboard_by_deleted ON dashboard (deleted);
	`
	if _, err := db.Exec(ctx, dashboardSQL); err != nil {
		return err
	}

	// Create dashboard_widget table
	dashboardWidgetSQL := `
		CREATE TABLE IF NOT EXISTS dashboard_widget (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			dashboard_id TEXT NOT NULL,
			widget_type TEXT NOT NULL,
			title TEXT,
			position_x INTEGER,
			position_y INTEGER,
			width INTEGER,
			height INTEGER,
			configuration TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS dashboard_widget_by_dashboard_id ON dashboard_widget (dashboard_id);
		CREATE INDEX IF NOT EXISTS dashboard_widget_by_widget_type ON dashboard_widget (widget_type);
		CREATE INDEX IF NOT EXISTS dashboard_widget_by_deleted ON dashboard_widget (deleted);
	`
	if _, err := db.Exec(ctx, dashboardWidgetSQL); err != nil {
		return err
	}

	// Create dashboard_share_mapping table
	dashboardShareMappingSQL := `
		CREATE TABLE IF NOT EXISTS dashboard_share_mapping (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			dashboard_id TEXT NOT NULL,
			user_id TEXT,
			team_id TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS dashboard_share_mapping_by_dashboard_id ON dashboard_share_mapping (dashboard_id);
		CREATE INDEX IF NOT EXISTS dashboard_share_mapping_by_user_id ON dashboard_share_mapping (user_id);
		CREATE INDEX IF NOT EXISTS dashboard_share_mapping_by_team_id ON dashboard_share_mapping (team_id);
		CREATE INDEX IF NOT EXISTS dashboard_share_mapping_by_project_id ON dashboard_share_mapping (project_id);
		CREATE INDEX IF NOT EXISTS dashboard_share_mapping_by_deleted ON dashboard_share_mapping (deleted);
	`
	if _, err := db.Exec(ctx, dashboardShareMappingSQL); err != nil {
		return err
	}

	// Create board table
	boardSQL := `
		CREATE TABLE IF NOT EXISTS board (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS board_by_title ON board (title);
		CREATE INDEX IF NOT EXISTS board_by_type ON board (type);
		CREATE INDEX IF NOT EXISTS board_by_deleted ON board (deleted);
	`
	if _, err := db.Exec(ctx, boardSQL); err != nil {
		return err
	}

	// Create board_column table
	boardColumnSQL := `
		CREATE TABLE IF NOT EXISTS board_column (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			board_id TEXT NOT NULL,
			title TEXT NOT NULL,
			status_id TEXT,
			position INTEGER NOT NULL,
			max_items INTEGER,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS board_column_by_board_id ON board_column (board_id);
		CREATE INDEX IF NOT EXISTS board_column_by_position ON board_column (position);
		CREATE INDEX IF NOT EXISTS board_column_by_deleted ON board_column (deleted);
	`
	if _, err := db.Exec(ctx, boardColumnSQL); err != nil {
		return err
	}

	// Create board_swimlane table
	boardSwimlaneSQL := `
		CREATE TABLE IF NOT EXISTS board_swimlane (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			board_id TEXT NOT NULL,
			title TEXT NOT NULL,
			query TEXT,
			position INTEGER NOT NULL,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS board_swimlane_by_board_id ON board_swimlane (board_id);
		CREATE INDEX IF NOT EXISTS board_swimlane_by_position ON board_swimlane (position);
		CREATE INDEX IF NOT EXISTS board_swimlane_by_deleted ON board_swimlane (deleted);
	`
	if _, err := db.Exec(ctx, boardSwimlaneSQL); err != nil {
		return err
	}

	// Create board_quick_filter table
	boardQuickFilterSQL := `
		CREATE TABLE IF NOT EXISTS board_quick_filter (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			board_id TEXT NOT NULL,
			title TEXT NOT NULL,
			query TEXT,
			position INTEGER NOT NULL,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS board_quick_filter_by_board_id ON board_quick_filter (board_id);
		CREATE INDEX IF NOT EXISTS board_quick_filter_by_position ON board_quick_filter (position);
		CREATE INDEX IF NOT EXISTS board_quick_filter_by_deleted ON board_quick_filter (deleted);
	`
	if _, err := db.Exec(ctx, boardQuickFilterSQL); err != nil {
		return err
	}

	// Create ticket_vote_mapping table (Phase 3)
	ticketVoteMappingSQL := `
		CREATE TABLE IF NOT EXISTS ticket_vote_mapping (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			ticket_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS ticket_vote_mapping_by_ticket_id ON ticket_vote_mapping (ticket_id);
		CREATE INDEX IF NOT EXISTS ticket_vote_mapping_by_user_id ON ticket_vote_mapping (user_id);
		CREATE INDEX IF NOT EXISTS ticket_vote_mapping_by_deleted ON ticket_vote_mapping (deleted);
	`
	if _, err := db.Exec(ctx, ticketVoteMappingSQL); err != nil {
		return err
	}

	// Create project_category table (Phase 3)
	projectCategorySQL := `
		CREATE TABLE IF NOT EXISTS project_category (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS project_category_by_title ON project_category (title);
		CREATE INDEX IF NOT EXISTS project_category_by_deleted ON project_category (deleted);
	`
	if _, err := db.Exec(ctx, projectCategorySQL); err != nil {
		return err
	}

	// Update project table to add project_category_id column if needed
	alterProjectSQL := `ALTER TABLE project ADD COLUMN project_category_id TEXT`
	db.Exec(ctx, alterProjectSQL) // Ignore error if column already exists

	// Create notification_scheme table (Phase 3)
	notificationSchemeSQL := `
		CREATE TABLE IF NOT EXISTS notification_scheme (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			title TEXT NOT NULL,
			description TEXT,
			project_id TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS notification_scheme_by_project_id ON notification_scheme (project_id);
		CREATE INDEX IF NOT EXISTS notification_scheme_by_title ON notification_scheme (title);
		CREATE INDEX IF NOT EXISTS notification_scheme_by_deleted ON notification_scheme (deleted);
	`
	if _, err := db.Exec(ctx, notificationSchemeSQL); err != nil {
		return err
	}

	// Create notification_event table (Phase 3)
	notificationEventSQL := `
		CREATE TABLE IF NOT EXISTS notification_event (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			event_type TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS notification_event_by_event_type ON notification_event (event_type);
		CREATE INDEX IF NOT EXISTS notification_event_by_deleted ON notification_event (deleted);
	`
	if _, err := db.Exec(ctx, notificationEventSQL); err != nil {
		return err
	}

	// Create notification_rule table (Phase 3)
	notificationRuleSQL := `
		CREATE TABLE IF NOT EXISTS notification_rule (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			notification_scheme_id TEXT NOT NULL,
			notification_event_id TEXT NOT NULL,
			recipient_type TEXT NOT NULL,
			recipient_id TEXT,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS notification_rule_by_scheme_id ON notification_rule (notification_scheme_id);
		CREATE INDEX IF NOT EXISTS notification_rule_by_event_id ON notification_rule (notification_event_id);
		CREATE INDEX IF NOT EXISTS notification_rule_by_recipient_type ON notification_rule (recipient_type);
		CREATE INDEX IF NOT EXISTS notification_rule_by_deleted ON notification_rule (deleted);
	`
	if _, err := db.Exec(ctx, notificationRuleSQL); err != nil {
		return err
	}

	// Create audit table (Phase 3 - enhanced for activity stream)
	auditSQL := `
		CREATE TABLE IF NOT EXISTS audit (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			action TEXT NOT NULL,
			user_id TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			details TEXT,
			is_public BOOLEAN DEFAULT 1,
			activity_type TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS audit_by_user_id ON audit (user_id);
		CREATE INDEX IF NOT EXISTS audit_by_entity_id ON audit (entity_id);
		CREATE INDEX IF NOT EXISTS audit_by_entity_type ON audit (entity_type);
		CREATE INDEX IF NOT EXISTS audit_by_is_public ON audit (is_public);
		CREATE INDEX IF NOT EXISTS audit_by_activity_type ON audit (activity_type);
		CREATE INDEX IF NOT EXISTS audit_by_created ON audit (created);
		CREATE INDEX IF NOT EXISTS audit_by_deleted ON audit (deleted);
	`
	if _, err := db.Exec(ctx, auditSQL); err != nil {
		return err
	}

	// Create comment_mention_mapping table (Phase 3)
	commentMentionMappingSQL := `
		CREATE TABLE IF NOT EXISTS comment_mention_mapping (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			comment_id TEXT NOT NULL,
			mentioned_user_id TEXT NOT NULL,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS comment_mention_mapping_by_comment_id ON comment_mention_mapping (comment_id);
		CREATE INDEX IF NOT EXISTS comment_mention_mapping_by_user_id ON comment_mention_mapping (mentioned_user_id);
		CREATE INDEX IF NOT EXISTS comment_mention_mapping_by_deleted ON comment_mention_mapping (deleted);
	`
	if _, err := db.Exec(ctx, commentMentionMappingSQL); err != nil {
		return err
	}

	// Create users table (if needed for mention tests)
	usersSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT NOT NULL PRIMARY KEY UNIQUE,
			username TEXT NOT NULL UNIQUE,
			email TEXT,
			created INTEGER NOT NULL,
			deleted BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS users_by_username ON users (username);
		CREATE INDEX IF NOT EXISTS users_by_deleted ON users (deleted);
	`
	if _, err := db.Exec(ctx, usersSQL); err != nil {
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
