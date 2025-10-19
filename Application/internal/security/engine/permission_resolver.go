package engine

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
)

// PermissionResolver resolves user permissions from multiple sources
type PermissionResolver struct {
	db database.Database
}

// NewPermissionResolver creates a new permission resolver
func NewPermissionResolver(db database.Database) *PermissionResolver {
	return &PermissionResolver{db: db}
}

// HasPermission checks if a user has a specific permission on a resource
func (pr *PermissionResolver) HasPermission(ctx context.Context, username, resource string, action Action) (bool, error) {
	logger.Debug("Permission resolver: checking permission",
		zap.String("username", username),
		zap.String("resource", resource),
		zap.String("action", string(action)),
	)

	// Map action to permission level
	requiredPermission := pr.actionToPermission(action)

	// Query external permission service or database
	// For now, we check if user exists and has basic access
	// This integrates with the existing permission service architecture

	// Check if user has direct permission grants
	hasDirectPermission, err := pr.checkDirectPermission(ctx, username, resource, requiredPermission)
	if err != nil {
		return false, err
	}

	if hasDirectPermission {
		return true, nil
	}

	// Check if user has permission through team membership
	hasTeamPermission, err := pr.checkTeamPermission(ctx, username, resource, requiredPermission)
	if err != nil {
		return false, err
	}

	if hasTeamPermission {
		return true, nil
	}

	// Check if user has permission through project role
	hasRolePermission, err := pr.checkRolePermission(ctx, username, resource, requiredPermission)
	if err != nil {
		return false, err
	}

	return hasRolePermission, nil
}

// GetEffectivePermissions returns all effective permissions for a user on a resource
func (pr *PermissionResolver) GetEffectivePermissions(ctx context.Context, username, resourceType, resourceID string) (PermissionSet, error) {
	permissions := PermissionSet{
		CanCreate: false,
		CanRead:   false,
		CanUpdate: false,
		CanDelete: false,
		CanList:   false,
		Level:     0,
		Roles:     make([]Role, 0),
	}

	// Check each action
	canCreate, _ := pr.HasPermission(ctx, username, resourceType, ActionCreate)
	canRead, _ := pr.HasPermission(ctx, username, resourceType, ActionRead)
	canUpdate, _ := pr.HasPermission(ctx, username, resourceType, ActionUpdate)
	canDelete, _ := pr.HasPermission(ctx, username, resourceType, ActionDelete)
	canList, _ := pr.HasPermission(ctx, username, resourceType, ActionList)

	permissions.CanCreate = canCreate
	permissions.CanRead = canRead
	permissions.CanUpdate = canUpdate
	permissions.CanDelete = canDelete
	permissions.CanList = canList

	return permissions, nil
}

// GetUserTeams returns all teams a user belongs to
func (pr *PermissionResolver) GetUserTeams(ctx context.Context, username string) ([]string, error) {
	query := `
		SELECT DISTINCT t.id
		FROM team t
		INNER JOIN team_user tu ON t.id = tu.team_id
		INNER JOIN user u ON tu.user_id = u.id
		WHERE u.username = ? AND t.deleted = 0 AND tu.deleted = 0
	`

	rows, err := pr.db.Query(ctx, query, username)
	if err != nil {
		logger.Error("Failed to query user teams", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("failed to query user teams: %w", err)
	}
	defer rows.Close()

	teams := make([]string, 0)
	for rows.Next() {
		var teamID string
		if err := rows.Scan(&teamID); err != nil {
			logger.Error("Failed to scan team ID", zap.Error(err))
			continue
		}
		teams = append(teams, teamID)
	}

	return teams, nil
}

// actionToPermission maps an action to a permission level
func (pr *PermissionResolver) actionToPermission(action Action) int {
	// Permission levels from models package:
	// READ = 1, CREATE = 2, UPDATE = 3, DELETE = 5
	switch action {
	case ActionRead, ActionList:
		return 1 // READ
	case ActionCreate:
		return 2 // CREATE
	case ActionUpdate:
		return 3 // UPDATE
	case ActionDelete:
		return 5 // DELETE
	case ActionExecute:
		return 3 // UPDATE (execute requires modification rights)
	default:
		return 5 // Default to highest permission for unknown actions
	}
}

// checkDirectPermission checks if user has direct permission on the resource
func (pr *PermissionResolver) checkDirectPermission(ctx context.Context, username, resource string, requiredPermission int) (bool, error) {
	// This is a simplified implementation
	// In production, this would query the permission service or permission_grant table

	// For now, we assume all authenticated users have READ access
	// and need explicit grants for CREATE, UPDATE, DELETE
	if requiredPermission == 1 {
		// Everyone can read (basic access)
		return true, nil
	}

	// Check if user has explicit permission grant
	// This would integrate with the external permission service
	// For now, return false to require team/role-based permissions
	return false, nil
}

// checkTeamPermission checks if user has permission through team membership
func (pr *PermissionResolver) checkTeamPermission(ctx context.Context, username, resource string, requiredPermission int) (bool, error) {
	// Get user's teams
	teams, err := pr.GetUserTeams(ctx, username)
	if err != nil {
		return false, err
	}

	if len(teams) == 0 {
		return false, nil
	}

	// For now, team membership grants all permissions except DELETE
	// In production, this would check team-specific permission grants
	if requiredPermission < 5 {
		return true, nil
	}

	return false, nil
}

// checkRolePermission checks if user has permission through project role
func (pr *PermissionResolver) checkRolePermission(ctx context.Context, username, resource string, requiredPermission int) (bool, error) {
	// Query user's project roles
	query := `
		SELECT DISTINCT pr.title
		FROM project_role pr
		INNER JOIN project_role_user_mapping prum ON pr.id = prum.project_role_id
		INNER JOIN user u ON prum.user_id = u.id
		WHERE u.username = ? AND pr.deleted = 0 AND prum.deleted = 0
	`

	rows, err := pr.db.Query(ctx, query, username)
	if err != nil {
		logger.Error("Failed to query user roles", zap.Error(err), zap.String("username", username))
		return false, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var roleTitle string
		if err := rows.Scan(&roleTitle); err != nil {
			logger.Error("Failed to scan role title", zap.Error(err))
			continue
		}
		roles = append(roles, roleTitle)
	}

	// Check if any role grants the required permission
	for _, roleTitle := range roles {
		if pr.roleGrantsPermission(roleTitle, requiredPermission) {
			return true, nil
		}
	}

	return false, nil
}

// roleGrantsPermission checks if a role grants the required permission
func (pr *PermissionResolver) roleGrantsPermission(roleTitle string, requiredPermission int) bool {
	// Role permission mapping
	// In production, this would be configurable
	rolePermissions := map[string]int{
		"Project Administrator": 5, // DELETE (all permissions)
		"Project Lead":          3, // UPDATE
		"Developer":             3, // UPDATE
		"Tester":                2, // CREATE
		"Viewer":                1, // READ
		"Contributor":           2, // CREATE
	}

	grantedPermission, ok := rolePermissions[roleTitle]
	if !ok {
		return false
	}

	return grantedPermission >= requiredPermission
}
