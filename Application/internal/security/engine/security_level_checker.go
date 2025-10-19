package engine

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
)

// SecurityLevelChecker validates access based on security levels
type SecurityLevelChecker struct {
	db database.Database
}

// NewSecurityLevelChecker creates a new security level checker
func NewSecurityLevelChecker(db database.Database) *SecurityLevelChecker {
	return &SecurityLevelChecker{db: db}
}

// CheckAccess checks if a user has access to an entity based on its security level
func (slc *SecurityLevelChecker) CheckAccess(ctx context.Context, username, entityID, entityType string) (bool, error) {
	logger.Debug("Security level checker: checking access",
		zap.String("username", username),
		zap.String("entity_id", entityID),
		zap.String("entity_type", entityType),
	)

	// Get the entity's security level ID
	securityLevelID, err := slc.getEntitySecurityLevel(ctx, entityID, entityType)
	if err != nil {
		logger.Error("Failed to get entity security level", zap.Error(err))
		return false, fmt.Errorf("failed to get entity security level: %w", err)
	}

	// If no security level assigned, allow access (unrestricted)
	if securityLevelID == "" {
		logger.Debug("No security level assigned, allowing access",
			zap.String("entity_id", entityID),
		)
		return true, nil
	}

	// Check if user has access to this security level
	hasAccess, err := slc.checkSecurityLevelAccess(ctx, username, securityLevelID)
	if err != nil {
		return false, err
	}

	if hasAccess {
		logger.Debug("Security level access granted",
			zap.String("username", username),
			zap.String("security_level_id", securityLevelID),
		)
		return true, nil
	}

	logger.Debug("Security level access denied",
		zap.String("username", username),
		zap.String("security_level_id", securityLevelID),
	)
	return false, nil
}

// getEntitySecurityLevel retrieves the security level ID for an entity
func (slc *SecurityLevelChecker) getEntitySecurityLevel(ctx context.Context, entityID, entityType string) (string, error) {
	// Map entity type to table and column
	// Currently only tickets support security levels, but this can be extended
	var query string
	switch entityType {
	case "ticket":
		query = `SELECT security_level_id FROM ticket WHERE id = ? AND deleted = 0`
	case "project":
		// Projects might have security levels in the future
		query = `SELECT security_level_id FROM project WHERE id = ? AND deleted = 0`
	default:
		// Unknown entity type, no security level
		return "", nil
	}

	var securityLevelID *string
	err := slc.db.QueryRow(ctx, query, entityID).Scan(&securityLevelID)
	if err != nil {
		logger.Error("Failed to query entity security level",
			zap.Error(err),
			zap.String("entity_id", entityID),
			zap.String("entity_type", entityType),
		)
		return "", fmt.Errorf("failed to query entity security level: %w", err)
	}

	if securityLevelID == nil || *securityLevelID == "" {
		return "", nil
	}

	return *securityLevelID, nil
}

// checkSecurityLevelAccess checks if a user has access to a security level
func (slc *SecurityLevelChecker) checkSecurityLevelAccess(ctx context.Context, username, securityLevelID string) (bool, error) {
	// Check for direct user grant
	hasDirectAccess, err := slc.checkDirectSecurityAccess(ctx, username, securityLevelID)
	if err != nil {
		return false, err
	}

	if hasDirectAccess {
		logger.Debug("Direct security level access granted",
			zap.String("username", username),
			zap.String("security_level_id", securityLevelID),
		)
		return true, nil
	}

	// Check for team-based access
	hasTeamAccess, err := slc.checkTeamSecurityAccess(ctx, username, securityLevelID)
	if err != nil {
		return false, err
	}

	if hasTeamAccess {
		logger.Debug("Team-based security level access granted",
			zap.String("username", username),
			zap.String("security_level_id", securityLevelID),
		)
		return true, nil
	}

	// Check for role-based access
	hasRoleAccess, err := slc.checkRoleSecurityAccess(ctx, username, securityLevelID)
	if err != nil {
		return false, err
	}

	if hasRoleAccess {
		logger.Debug("Role-based security level access granted",
			zap.String("username", username),
			zap.String("security_level_id", securityLevelID),
		)
		return true, nil
	}

	return false, nil
}

// checkDirectSecurityAccess checks for direct user-level security grants
func (slc *SecurityLevelChecker) checkDirectSecurityAccess(ctx context.Context, username, securityLevelID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM security_level_permission_mapping slpm
		INNER JOIN user u ON slpm.user_id = u.id
		WHERE slpm.security_level_id = ?
		AND u.username = ?
		AND slpm.deleted = 0
	`

	var count int
	err := slc.db.QueryRow(ctx, query, securityLevelID, username).Scan(&count)
	if err != nil {
		logger.Error("Failed to check direct security access", zap.Error(err))
		return false, fmt.Errorf("failed to check direct security access: %w", err)
	}

	return count > 0, nil
}

// checkTeamSecurityAccess checks for team-based security grants
func (slc *SecurityLevelChecker) checkTeamSecurityAccess(ctx context.Context, username, securityLevelID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM security_level_permission_mapping slpm
		INNER JOIN team_user tu ON slpm.team_id = tu.team_id
		INNER JOIN user u ON tu.user_id = u.id
		WHERE slpm.security_level_id = ?
		AND u.username = ?
		AND slpm.deleted = 0
		AND tu.deleted = 0
	`

	var count int
	err := slc.db.QueryRow(ctx, query, securityLevelID, username).Scan(&count)
	if err != nil {
		logger.Error("Failed to check team security access", zap.Error(err))
		return false, fmt.Errorf("failed to check team security access: %w", err)
	}

	return count > 0, nil
}

// checkRoleSecurityAccess checks for role-based security grants
func (slc *SecurityLevelChecker) checkRoleSecurityAccess(ctx context.Context, username, securityLevelID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM security_level_permission_mapping slpm
		INNER JOIN project_role_user_mapping prum ON slpm.project_role_id = prum.project_role_id
		INNER JOIN user u ON prum.user_id = u.id
		WHERE slpm.security_level_id = ?
		AND u.username = ?
		AND slpm.deleted = 0
		AND prum.deleted = 0
	`

	var count int
	err := slc.db.QueryRow(ctx, query, securityLevelID, username).Scan(&count)
	if err != nil {
		logger.Error("Failed to check role security access", zap.Error(err))
		return false, fmt.Errorf("failed to check role security access: %w", err)
	}

	return count > 0, nil
}

// GetSecurityLevel retrieves the security level details for an entity
func (slc *SecurityLevelChecker) GetSecurityLevel(ctx context.Context, securityLevelID string) (*SecurityLevel, error) {
	query := `
		SELECT id, title, description, project_id, level
		FROM security_level
		WHERE id = ? AND deleted = 0
	`

	var sl SecurityLevel
	err := slc.db.QueryRow(ctx, query, securityLevelID).Scan(
		&sl.ID,
		&sl.Title,
		&sl.Description,
		&sl.ProjectID,
		&sl.Level,
	)

	if err != nil {
		logger.Error("Failed to get security level", zap.Error(err), zap.String("id", securityLevelID))
		return nil, fmt.Errorf("failed to get security level: %w", err)
	}

	return &sl, nil
}

// SecurityLevel represents a security level entity (local definition for this package)
type SecurityLevel struct {
	ID          string
	Title       string
	Description string
	ProjectID   string
	Level       int
}
