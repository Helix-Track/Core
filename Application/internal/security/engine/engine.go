package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/database"
	"helixtrack.ru/core/internal/logger"
)

// Engine is the main security engine interface
type Engine interface {
	// CheckAccess checks if a user has permission to perform an action on a resource
	CheckAccess(ctx context.Context, req AccessRequest) (AccessResponse, error)

	// ValidateSecurityLevel checks if a user has access to a specific security level
	ValidateSecurityLevel(ctx context.Context, username, entityID string) (bool, error)

	// EvaluateRole checks if a user has a specific role in a project
	EvaluateRole(ctx context.Context, username, projectID, requiredRole string) (bool, error)

	// GetEffectivePermissions returns all effective permissions for a user on a resource
	GetEffectivePermissions(ctx context.Context, username, resourceType, resourceID string) (PermissionSet, error)

	// GetSecurityContext retrieves or builds the security context for a user
	GetSecurityContext(ctx context.Context, username string) (*SecurityContext, error)

	// InvalidateCache invalidates cached permissions for a user
	InvalidateCache(username string)

	// InvalidateAllCache clears the entire permission cache
	InvalidateAllCache()

	// AuditAccessAttempt logs an access attempt
	AuditAccessAttempt(ctx context.Context, req AccessRequest, response AccessResponse) error
}

// SecurityEngine is the concrete implementation of the Engine interface
type SecurityEngine struct {
	db                database.Database
	permissionResolver *PermissionResolver
	roleEvaluator     *RoleEvaluator
	securityChecker   *SecurityLevelChecker
	cache             *PermissionCache
	auditLogger       *AuditLogger
	mu                sync.RWMutex
	config            Config
}

// Config holds configuration for the security engine
type Config struct {
	EnableCaching     bool
	CacheTTL          time.Duration
	CacheMaxSize      int
	EnableAuditing    bool
	AuditAllAttempts  bool          // Audit both allowed and denied attempts
	AuditRetention    time.Duration // How long to keep audit logs
}

// DefaultConfig returns the default security engine configuration
func DefaultConfig() Config {
	return Config{
		EnableCaching:    true,
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     10000,
		EnableAuditing:   true,
		AuditAllAttempts: true,
		AuditRetention:   90 * 24 * time.Hour, // 90 days
	}
}

// NewSecurityEngine creates a new security engine instance
func NewSecurityEngine(db database.Database, config Config) *SecurityEngine {
	engine := &SecurityEngine{
		db:     db,
		config: config,
	}

	// Initialize components
	engine.permissionResolver = NewPermissionResolver(db)
	engine.roleEvaluator = NewRoleEvaluator(db)
	engine.securityChecker = NewSecurityLevelChecker(db)

	if config.EnableCaching {
		engine.cache = NewPermissionCache(config.CacheMaxSize, config.CacheTTL)
	}

	if config.EnableAuditing {
		engine.auditLogger = NewAuditLogger(db, config.AuditRetention)
	}

	return engine
}

// CheckAccess is the main authorization method
func (e *SecurityEngine) CheckAccess(ctx context.Context, req AccessRequest) (AccessResponse, error) {
	logger.Debug("Security engine: checking access",
		zap.String("username", req.Username),
		zap.String("resource", req.Resource),
		zap.String("resource_id", req.ResourceID),
		zap.String("action", string(req.Action)),
	)

	// Check cache first
	if e.config.EnableCaching && e.cache != nil {
		if cached, found := e.cache.Get(req); found {
			logger.Debug("Security engine: cache hit",
				zap.String("username", req.Username),
				zap.String("resource", req.Resource),
			)
			return cached, nil
		}
	}

	// Perform access check
	response := AccessResponse{
		Allowed: false,
		Reason:  "Access denied",
	}

	// Step 1: Check basic resource permissions
	hasPermission, err := e.permissionResolver.HasPermission(ctx, req.Username, req.Resource, req.Action)
	if err != nil {
		logger.Error("Security engine: permission check failed",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		return response, fmt.Errorf("permission check failed: %w", err)
	}

	if !hasPermission {
		response.Reason = fmt.Sprintf("User does not have %s permission on resource %s", req.Action, req.Resource)

		// Audit denied access
		if e.config.EnableAuditing {
			e.AuditAccessAttempt(ctx, req, response)
		}

		// Cache negative result (with shorter TTL)
		if e.config.EnableCaching && e.cache != nil {
			e.cache.SetWithTTL(req, response, 1*time.Minute)
		}

		return response, nil
	}

	// Step 2: Check security level if resource ID is provided
	if req.ResourceID != "" {
		hasSecurityAccess, err := e.securityChecker.CheckAccess(ctx, req.Username, req.ResourceID, req.Resource)
		if err != nil {
			logger.Error("Security engine: security level check failed",
				zap.Error(err),
				zap.String("username", req.Username),
				zap.String("resource_id", req.ResourceID),
			)
			return response, fmt.Errorf("security level check failed: %w", err)
		}

		if !hasSecurityAccess {
			response.Reason = "Insufficient security clearance for this resource"

			// Audit denied access
			if e.config.EnableAuditing {
				e.AuditAccessAttempt(ctx, req, response)
			}

			// Cache negative result
			if e.config.EnableCaching && e.cache != nil {
				e.cache.SetWithTTL(req, response, 1*time.Minute)
			}

			return response, nil
		}
	}

	// Step 3: Check project-specific role permissions if project context provided
	if projectID, ok := req.Context["project_id"]; ok && projectID != "" {
		hasRoleAccess, err := e.roleEvaluator.CheckProjectAccess(ctx, req.Username, projectID, req.Action)
		if err != nil {
			logger.Error("Security engine: role check failed",
				zap.Error(err),
				zap.String("username", req.Username),
				zap.String("project_id", projectID),
			)
			return response, fmt.Errorf("role check failed: %w", err)
		}

		if !hasRoleAccess {
			response.Reason = "User's project role does not permit this action"

			// Audit denied access
			if e.config.EnableAuditing {
				e.AuditAccessAttempt(ctx, req, response)
			}

			// Cache negative result
			if e.config.EnableCaching && e.cache != nil {
				e.cache.SetWithTTL(req, response, 1*time.Minute)
			}

			return response, nil
		}
	}

	// All checks passed - grant access
	response.Allowed = true
	response.Reason = "Access granted"

	// Audit allowed access
	if e.config.EnableAuditing && e.config.AuditAllAttempts {
		e.AuditAccessAttempt(ctx, req, response)
	}

	// Cache positive result
	if e.config.EnableCaching && e.cache != nil {
		e.cache.Set(req, response)
	}

	logger.Debug("Security engine: access granted",
		zap.String("username", req.Username),
		zap.String("resource", req.Resource),
		zap.String("resource_id", req.ResourceID),
	)

	return response, nil
}

// ValidateSecurityLevel checks if a user has access to a specific security level
func (e *SecurityEngine) ValidateSecurityLevel(ctx context.Context, username, entityID string) (bool, error) {
	// Delegate to security level checker
	// For now, we check tickets - this can be extended to other entity types
	return e.securityChecker.CheckAccess(ctx, username, entityID, "ticket")
}

// EvaluateRole checks if a user has a specific role in a project
func (e *SecurityEngine) EvaluateRole(ctx context.Context, username, projectID, requiredRole string) (bool, error) {
	return e.roleEvaluator.HasRole(ctx, username, projectID, requiredRole)
}

// GetEffectivePermissions returns all effective permissions for a user on a resource
func (e *SecurityEngine) GetEffectivePermissions(ctx context.Context, username, resourceType, resourceID string) (PermissionSet, error) {
	return e.permissionResolver.GetEffectivePermissions(ctx, username, resourceType, resourceID)
}

// GetSecurityContext retrieves or builds the security context for a user
func (e *SecurityEngine) GetSecurityContext(ctx context.Context, username string) (*SecurityContext, error) {
	// Check cache first
	if e.config.EnableCaching && e.cache != nil {
		if secCtx, found := e.cache.GetContext(username); found {
			logger.Debug("Security engine: context cache hit", zap.String("username", username))
			return secCtx, nil
		}
	}

	// Build security context
	secCtx := &SecurityContext{
		Username:             username,
		Roles:                make([]Role, 0),
		Teams:                make([]string, 0),
		EffectivePermissions: make(map[string]PermissionSet),
		CachedAt:             time.Now(),
		ExpiresAt:            time.Now().Add(e.config.CacheTTL),
	}

	// Get user's roles
	roles, err := e.roleEvaluator.GetUserRoles(ctx, username)
	if err != nil {
		logger.Error("Failed to get user roles", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	secCtx.Roles = roles

	// Get user's teams
	teams, err := e.permissionResolver.GetUserTeams(ctx, username)
	if err != nil {
		logger.Error("Failed to get user teams", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}
	secCtx.Teams = teams

	// Cache the context
	if e.config.EnableCaching && e.cache != nil {
		e.cache.SetContext(username, secCtx)
	}

	return secCtx, nil
}

// InvalidateCache invalidates cached permissions for a specific user
func (e *SecurityEngine) InvalidateCache(username string) {
	if e.config.EnableCaching && e.cache != nil {
		e.cache.InvalidateUser(username)
		logger.Info("Security engine: cache invalidated for user", zap.String("username", username))
	}
}

// InvalidateAllCache clears the entire permission cache
func (e *SecurityEngine) InvalidateAllCache() {
	if e.config.EnableCaching && e.cache != nil {
		e.cache.Clear()
		logger.Info("Security engine: all cache cleared")
	}
}

// AuditAccessAttempt logs an access attempt
func (e *SecurityEngine) AuditAccessAttempt(ctx context.Context, req AccessRequest, response AccessResponse) error {
	if !e.config.EnableAuditing || e.auditLogger == nil {
		return nil
	}

	entry := AuditEntry{
		Timestamp:  time.Now(),
		Username:   req.Username,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Action:     req.Action,
		Allowed:    response.Allowed,
		Reason:     response.Reason,
		Context:    req.Context,
	}

	return e.auditLogger.Log(ctx, entry)
}
