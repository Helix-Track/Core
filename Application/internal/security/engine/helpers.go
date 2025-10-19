package engine

import (
	"context"

	"go.uber.org/zap"
	"helixtrack.ru/core/internal/logger"
)

// HelperMethods provides convenience methods for common security operations
type HelperMethods struct {
	engine Engine
}

// NewHelperMethods creates a new helper instance
func NewHelperMethods(engine Engine) *HelperMethods {
	return &HelperMethods{engine: engine}
}

// CanUserCreate checks if a user can create a resource
func (h *HelperMethods) CanUserCreate(ctx context.Context, username, resource string, context map[string]string) (bool, error) {
	req := AccessRequest{
		Username: username,
		Resource: resource,
		Action:   ActionCreate,
		Context:  context,
	}

	response, err := h.engine.CheckAccess(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Allowed, nil
}

// CanUserRead checks if a user can read a resource
func (h *HelperMethods) CanUserRead(ctx context.Context, username, resource, resourceID string, context map[string]string) (bool, error) {
	req := AccessRequest{
		Username:   username,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     ActionRead,
		Context:    context,
	}

	response, err := h.engine.CheckAccess(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Allowed, nil
}

// CanUserUpdate checks if a user can update a resource
func (h *HelperMethods) CanUserUpdate(ctx context.Context, username, resource, resourceID string, context map[string]string) (bool, error) {
	req := AccessRequest{
		Username:   username,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     ActionUpdate,
		Context:    context,
	}

	response, err := h.engine.CheckAccess(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Allowed, nil
}

// CanUserDelete checks if a user can delete a resource
func (h *HelperMethods) CanUserDelete(ctx context.Context, username, resource, resourceID string, context map[string]string) (bool, error) {
	req := AccessRequest{
		Username:   username,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     ActionDelete,
		Context:    context,
	}

	response, err := h.engine.CheckAccess(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Allowed, nil
}

// CanUserList checks if a user can list resources
func (h *HelperMethods) CanUserList(ctx context.Context, username, resource string, context map[string]string) (bool, error) {
	req := AccessRequest{
		Username: username,
		Resource: resource,
		Action:   ActionList,
		Context:  context,
	}

	response, err := h.engine.CheckAccess(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Allowed, nil
}

// RequirePermission checks permission and returns error details if denied
func (h *HelperMethods) RequirePermission(ctx context.Context, req AccessRequest) (AccessResponse, error) {
	response, err := h.engine.CheckAccess(ctx, req)
	if err != nil {
		logger.Error("Permission check failed",
			zap.Error(err),
			zap.String("username", req.Username),
			zap.String("resource", req.Resource),
		)
		return AccessResponse{
			Allowed: false,
			Reason:  "Permission check failed",
		}, err
	}

	if !response.Allowed {
		logger.Warn("Permission denied",
			zap.String("username", req.Username),
			zap.String("resource", req.Resource),
			zap.String("action", string(req.Action)),
			zap.String("reason", response.Reason),
		)
	}

	return response, nil
}

// GetUserPermissions returns all permissions for a user on a resource
func (h *HelperMethods) GetUserPermissions(ctx context.Context, username, resourceType, resourceID string) (PermissionSet, error) {
	return h.engine.GetEffectivePermissions(ctx, username, resourceType, resourceID)
}

// FilterBySecurityLevel filters a list of entity IDs based on user's security clearance
func (h *HelperMethods) FilterBySecurityLevel(ctx context.Context, username string, entityIDs []string) ([]string, error) {
	allowedIDs := make([]string, 0, len(entityIDs))

	for _, entityID := range entityIDs {
		hasAccess, err := h.engine.ValidateSecurityLevel(ctx, username, entityID)
		if err != nil {
			logger.Error("Security level check failed",
				zap.Error(err),
				zap.String("username", username),
				zap.String("entity_id", entityID),
			)
			continue
		}

		if hasAccess {
			allowedIDs = append(allowedIDs, entityID)
		}
	}

	return allowedIDs, nil
}

// CheckMultiplePermissions checks multiple permissions at once
func (h *HelperMethods) CheckMultiplePermissions(ctx context.Context, username, resource string, actions []Action) (map[Action]bool, error) {
	results := make(map[Action]bool)

	for _, action := range actions {
		req := AccessRequest{
			Username: username,
			Resource: resource,
			Action:   action,
		}

		response, err := h.engine.CheckAccess(ctx, req)
		if err != nil {
			logger.Error("Permission check failed",
				zap.Error(err),
				zap.String("username", username),
				zap.String("action", string(action)),
			)
			results[action] = false
			continue
		}

		results[action] = response.Allowed
	}

	return results, nil
}

// GetAccessSummary returns a summary of user's access to a resource
func (h *HelperMethods) GetAccessSummary(ctx context.Context, username, resource, resourceID string) (AccessSummary, error) {
	summary := AccessSummary{
		Username:   username,
		Resource:   resource,
		ResourceID: resourceID,
	}

	// Check all standard actions
	actions := []Action{ActionRead, ActionCreate, ActionUpdate, ActionDelete, ActionList}
	permissions, err := h.CheckMultiplePermissions(ctx, username, resource, actions)
	if err != nil {
		return summary, err
	}

	summary.CanCreate = permissions[ActionCreate]
	summary.CanRead = permissions[ActionRead]
	summary.CanUpdate = permissions[ActionUpdate]
	summary.CanDelete = permissions[ActionDelete]
	summary.CanList = permissions[ActionList]

	// Get effective permissions
	permSet, err := h.engine.GetEffectivePermissions(ctx, username, resource, resourceID)
	if err != nil {
		return summary, err
	}

	summary.EffectivePermissions = permSet
	summary.Roles = permSet.Roles

	return summary, nil
}

// InvalidateUserCache invalidates all cached permissions for a user
// Should be called when user's roles or teams change
func (h *HelperMethods) InvalidateUserCache(username string) {
	h.engine.InvalidateCache(username)
	logger.Info("Security cache invalidated for user", zap.String("username", username))
}

// AccessSummary represents a summary of user's access to a resource
type AccessSummary struct {
	Username             string
	Resource             string
	ResourceID           string
	CanCreate            bool
	CanRead              bool
	CanUpdate            bool
	CanDelete            bool
	CanList              bool
	EffectivePermissions PermissionSet
	Roles                []Role
}
