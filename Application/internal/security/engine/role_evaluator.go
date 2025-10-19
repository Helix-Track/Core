package engine

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
)

// RoleEvaluator evaluates role-based permissions
type RoleEvaluator struct {
	db database.Database
}

// NewRoleEvaluator creates a new role evaluator
func NewRoleEvaluator(db database.Database) *RoleEvaluator {
	return &RoleEvaluator{db: db}
}

// HasRole checks if a user has a specific role in a project
func (re *RoleEvaluator) HasRole(ctx context.Context, username, projectID, requiredRole string) (bool, error) {
	logger.Debug("Role evaluator: checking role",
		zap.String("username", username),
		zap.String("project_id", projectID),
		zap.String("required_role", requiredRole),
	)

	// Query for role assignment
	query := `
		SELECT COUNT(*)
		FROM project_role pr
		INNER JOIN project_role_user_mapping prum ON pr.id = prum.project_role_id
		INNER JOIN user u ON prum.user_id = u.id
		WHERE u.username = ?
		AND pr.title = ?
		AND (pr.project_id = ? OR pr.project_id IS NULL)
		AND pr.deleted = 0
		AND prum.deleted = 0
	`

	var count int
	err := re.db.QueryRow(ctx, query, username, requiredRole, projectID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check role", zap.Error(err))
		return false, fmt.Errorf("failed to check role: %w", err)
	}

	return count > 0, nil
}

// CheckProjectAccess checks if user's role in a project permits the action
func (re *RoleEvaluator) CheckProjectAccess(ctx context.Context, username, projectID string, action Action) (bool, error) {
	logger.Debug("Role evaluator: checking project access",
		zap.String("username", username),
		zap.String("project_id", projectID),
		zap.String("action", string(action)),
	)

	// Get user's roles for the project
	roles, err := re.GetProjectRoles(ctx, username, projectID)
	if err != nil {
		return false, err
	}

	if len(roles) == 0 {
		logger.Debug("User has no roles in project",
			zap.String("username", username),
			zap.String("project_id", projectID),
		)
		return false, nil
	}

	// Check if any role grants permission for the action
	requiredPermission := re.actionToPermission(action)
	for _, role := range roles {
		if re.rolePermissionLevel(role.Title) >= requiredPermission {
			logger.Debug("Role grants access",
				zap.String("username", username),
				zap.String("role", role.Title),
				zap.String("action", string(action)),
			)
			return true, nil
		}
	}

	logger.Debug("No role grants access",
		zap.String("username", username),
		zap.String("project_id", projectID),
		zap.String("action", string(action)),
	)
	return false, nil
}

// GetUserRoles returns all roles assigned to a user across all projects
func (re *RoleEvaluator) GetUserRoles(ctx context.Context, username string) ([]Role, error) {
	query := `
		SELECT pr.id, pr.title, pr.project_id
		FROM project_role pr
		INNER JOIN project_role_user_mapping prum ON pr.id = prum.project_role_id
		INNER JOIN user u ON prum.user_id = u.id
		WHERE u.username = ?
		AND pr.deleted = 0
		AND prum.deleted = 0
		ORDER BY pr.title
	`

	rows, err := re.db.Query(ctx, query, username)
	if err != nil {
		logger.Error("Failed to query user roles", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]Role, 0)
	for rows.Next() {
		var role Role
		var projectID *string

		if err := rows.Scan(&role.ID, &role.Title, &projectID); err != nil {
			logger.Error("Failed to scan role", zap.Error(err))
			continue
		}

		role.ProjectID = projectID
		role.Permissions = re.getRolePermissions(role.Title)
		roles = append(roles, role)
	}

	logger.Debug("Retrieved user roles",
		zap.String("username", username),
		zap.Int("role_count", len(roles)),
	)

	return roles, nil
}

// GetProjectRoles returns all roles assigned to a user in a specific project
func (re *RoleEvaluator) GetProjectRoles(ctx context.Context, username, projectID string) ([]Role, error) {
	query := `
		SELECT pr.id, pr.title, pr.project_id
		FROM project_role pr
		INNER JOIN project_role_user_mapping prum ON pr.id = prum.project_role_id
		INNER JOIN user u ON prum.user_id = u.id
		WHERE u.username = ?
		AND (prum.project_id = ? OR pr.project_id IS NULL)
		AND pr.deleted = 0
		AND prum.deleted = 0
		ORDER BY pr.title
	`

	rows, err := re.db.Query(ctx, query, username, projectID)
	if err != nil {
		logger.Error("Failed to query project roles",
			zap.Error(err),
			zap.String("username", username),
			zap.String("project_id", projectID),
		)
		return nil, fmt.Errorf("failed to query project roles: %w", err)
	}
	defer rows.Close()

	roles := make([]Role, 0)
	for rows.Next() {
		var role Role
		var projID *string

		if err := rows.Scan(&role.ID, &role.Title, &projID); err != nil {
			logger.Error("Failed to scan role", zap.Error(err))
			continue
		}

		role.ProjectID = projID
		role.Permissions = re.getRolePermissions(role.Title)
		roles = append(roles, role)
	}

	return roles, nil
}

// actionToPermission maps an action to a permission level
func (re *RoleEvaluator) actionToPermission(action Action) int {
	switch action {
	case ActionRead, ActionList:
		return 1 // READ
	case ActionCreate:
		return 2 // CREATE
	case ActionUpdate, ActionExecute:
		return 3 // UPDATE
	case ActionDelete:
		return 5 // DELETE
	default:
		return 5 // Default to highest permission
	}
}

// rolePermissionLevel returns the permission level granted by a role
func (re *RoleEvaluator) rolePermissionLevel(roleTitle string) int {
	// Role hierarchy (from lowest to highest permission)
	// In production, this should be configurable or stored in database
	roleHierarchy := map[string]int{
		"Viewer":                1, // READ only
		"Contributor":           2, // CREATE
		"Tester":                2, // CREATE
		"Developer":             3, // UPDATE
		"Project Lead":          3, // UPDATE
		"Project Administrator": 5, // DELETE (all permissions)
	}

	level, ok := roleHierarchy[roleTitle]
	if !ok {
		// Unknown role, grant minimal permission
		return 1
	}

	return level
}

// getRolePermissions returns the permission set for a role
func (re *RoleEvaluator) getRolePermissions(roleTitle string) PermissionSet {
	level := re.rolePermissionLevel(roleTitle)

	permissions := PermissionSet{
		CanRead:   level >= 1,
		CanCreate: level >= 2,
		CanUpdate: level >= 3,
		CanDelete: level >= 5,
		CanList:   level >= 1,
		Level:     level,
	}

	return permissions
}

// IsGlobalRole checks if a role is global (not project-specific)
func (re *RoleEvaluator) IsGlobalRole(ctx context.Context, roleID string) (bool, error) {
	query := `SELECT project_id FROM project_role WHERE id = ? AND deleted = 0`

	var projectID *string
	err := re.db.QueryRow(ctx, query, roleID).Scan(&projectID)
	if err != nil {
		return false, fmt.Errorf("failed to check role type: %w", err)
	}

	return projectID == nil || *projectID == "", nil
}
