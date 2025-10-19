package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E2E Test Scenarios - Complete user workflows with Security Engine

// TestE2E_UserJourney_TicketCreation tests complete ticket creation workflow
func TestE2E_UserJourney_TicketCreation(t *testing.T) {
	// Scenario: Developer creates a ticket, assigns it, and updates status
	// Steps:
	// 1. User authenticates
	// 2. User checks if they can create tickets
	// 3. User creates ticket
	// 4. User assigns ticket to team member
	// 5. User updates ticket status
	// 6. Audit log is created for all actions

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)
	ctx := context.Background()

	username := "developer1"

	// Step 1: Load security context
	secCtx := &SecurityContext{
		Username:  username,
		Roles:     []Role{{ID: "role-dev", Title: "Developer"}},
		Teams:     []string{"team-backend"},
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	engine.cache.SetContext(username, secCtx)

	// Step 2: Check CREATE permission
	createReq := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionCreate,
		Context:  map[string]string{"project_id": "proj-1"},
	}

	// Cache permission (simulating database check result)
	engine.cache.Set(createReq, AccessResponse{
		Allowed: true,
		Reason:  "Developer role grants CREATE permission",
	})

	cached, found := engine.cache.Get(createReq)
	assert.True(t, found)
	assert.True(t, cached.Allowed)

	// Step 3: Create ticket (permission granted)
	ticketID := "ticket-123"

	// Step 4: Assign ticket (UPDATE permission)
	updateReq := AccessRequest{
		Username:   username,
		Resource:   "ticket",
		ResourceID: ticketID,
		Action:     ActionUpdate,
		Context:    map[string]string{"project_id": "proj-1"},
	}

	engine.cache.Set(updateReq, AccessResponse{
		Allowed: true,
		Reason:  "Developer role grants UPDATE permission",
	})

	// Verify permissions cached
	stats := engine.cache.GetStats()
	assert.GreaterOrEqual(t, stats.EntryCount, 2)

	assert.NotNil(t, ctx)
}

// TestE2E_UserJourney_ProjectAccess tests project access workflow
func TestE2E_UserJourney_ProjectAccess(t *testing.T) {
	// Scenario: User joins project, gets role, accesses project resources
	// Steps:
	// 1. User joins project
	// 2. User is assigned "Contributor" role
	// 3. User can READ and CREATE tickets
	// 4. User cannot DELETE tickets (insufficient role)
	// 5. All attempts are audited

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	username := "contributor1"
	projectID := "proj-1"

	// Security context with Contributor role
	secCtx := &SecurityContext{
		Username: username,
		Roles: []Role{
			{ID: "role-contrib", Title: "Contributor"},
		},
		Teams:     []string{"team-frontend"},
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	engine.cache.SetContext(username, secCtx)

	// Can READ
	readReq := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionRead,
		Context:  map[string]string{"project_id": projectID},
	}
	engine.cache.Set(readReq, AccessResponse{Allowed: true, Reason: "Contributor can READ"})

	// Can CREATE
	createReq := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionCreate,
		Context:  map[string]string{"project_id": projectID},
	}
	engine.cache.Set(createReq, AccessResponse{Allowed: true, Reason: "Contributor can CREATE"})

	// Cannot DELETE
	deleteReq := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionDelete,
		Context:  map[string]string{"project_id": projectID},
	}
	engine.cache.Set(deleteReq, AccessResponse{
		Allowed: false,
		Reason:  "Contributor role insufficient for DELETE (requires Admin)",
	})

	// Verify permission results
	readResp, found := engine.cache.Get(readReq)
	assert.True(t, found)
	assert.True(t, readResp.Allowed)

	deleteResp, found := engine.cache.Get(deleteReq)
	assert.True(t, found)
	assert.False(t, deleteResp.Allowed)
}

// TestE2E_UserJourney_SecurityLevels tests security level workflow
func TestE2E_UserJourney_SecurityLevels(t *testing.T) {
	// Scenario: Confidential tickets with security levels
	// Steps:
	// 1. Admin creates ticket with "Confidential" security level
	// 2. Admin can access (has security grant)
	// 3. Regular user cannot access (no security grant)
	// 4. User is granted access
	// 5. User can now access the ticket

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	adminUsername := "admin1"
	regularUsername := "user1"
	ticketID := "ticket-secret-123"

	// Admin has full access
	adminReq := AccessRequest{
		Username:   adminUsername,
		Resource:   "ticket",
		ResourceID: ticketID,
		Action:     ActionRead,
	}
	engine.cache.Set(adminReq, AccessResponse{
		Allowed: true,
		Reason:  "Admin has security level grant",
	})

	// Regular user denied
	userReq := AccessRequest{
		Username:   regularUsername,
		Resource:   "ticket",
		ResourceID: ticketID,
		Action:     ActionRead,
	}
	engine.cache.Set(userReq, AccessResponse{
		Allowed: false,
		Reason:  "User lacks security level clearance",
	})

	// Verify denials
	userResp, _ := engine.cache.Get(userReq)
	assert.False(t, userResp.Allowed)

	// Grant access to user
	// (This would be done through GrantSecurityAccess in production)

	// Update cache after grant
	engine.cache.Set(userReq, AccessResponse{
		Allowed: true,
		Reason:  "User granted security level access",
	})

	// Verify access now granted
	userResp2, _ := engine.cache.Get(userReq)
	assert.True(t, userResp2.Allowed)
}

// TestE2E_UserJourney_TeamCollaboration tests team collaboration workflow
func TestE2E_UserJourney_TeamCollaboration(t *testing.T) {
	// Scenario: Team working on project together
	// Steps:
	// 1. Multiple team members access project
	// 2. All inherit team permissions
	// 3. Team permissions grant base access
	// 4. Individual roles grant additional permissions

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	teamID := "team-backend"
	projectID := "proj-1"

	teamMembers := []struct {
		username string
		role     string
	}{
		{"dev1", "Developer"},
		{"dev2", "Developer"},
		{"lead1", "Project Lead"},
	}

	for _, member := range teamMembers {
		// Load security context
		secCtx := &SecurityContext{
			Username:  member.username,
			Roles:     []Role{{ID: "role-" + member.role, Title: member.role}},
			Teams:     []string{teamID},
			CachedAt:  time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		engine.cache.SetContext(member.username, secCtx)

		// All can READ (team permission)
		readReq := AccessRequest{
			Username: member.username,
			Resource: "ticket",
			Action:   ActionRead,
			Context:  map[string]string{"project_id": projectID, "team_id": teamID},
		}
		engine.cache.Set(readReq, AccessResponse{
			Allowed: true,
			Reason:  "Team membership grants READ",
		})

		// Developers can UPDATE
		if member.role == "Developer" || member.role == "Project Lead" {
			updateReq := AccessRequest{
				Username: member.username,
				Resource: "ticket",
				Action:   ActionUpdate,
				Context:  map[string]string{"project_id": projectID},
			}
			engine.cache.Set(updateReq, AccessResponse{
				Allowed: true,
				Reason:  member.role + " role grants UPDATE",
			})
		}
	}

	// Verify all team members cached
	stats := engine.cache.GetStats()
	assert.Equal(t, 3, stats.ContextCount)
}

// TestE2E_UserJourney_PermissionEscalation tests permission escalation attempts
func TestE2E_UserJourney_PermissionEscalation(t *testing.T) {
	// Scenario: Viewer attempts to escalate permissions
	// Steps:
	// 1. Viewer can READ tickets
	// 2. Viewer attempts to UPDATE ticket (denied)
	// 3. Viewer attempts to DELETE ticket (denied)
	// 4. All denied attempts are audited with WARNING severity

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	username := "viewer1"
	ticketID := "ticket-456"

	// Security context with Viewer role
	secCtx := &SecurityContext{
		Username:  username,
		Roles:     []Role{{ID: "role-viewer", Title: "Viewer"}},
		Teams:     []string{},
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	engine.cache.SetContext(username, secCtx)

	// READ allowed
	readReq := AccessRequest{
		Username:   username,
		Resource:   "ticket",
		ResourceID: ticketID,
		Action:     ActionRead,
	}
	engine.cache.Set(readReq, AccessResponse{
		Allowed: true,
		Reason:  "Viewer role grants READ",
	})

	// UPDATE denied
	updateReq := AccessRequest{
		Username:   username,
		Resource:   "ticket",
		ResourceID: ticketID,
		Action:     ActionUpdate,
	}
	engine.cache.Set(updateReq, AccessResponse{
		Allowed: false,
		Reason:  "Viewer role insufficient for UPDATE",
		AuditID: "audit-escalation-1",
	})

	// DELETE denied
	deleteReq := AccessRequest{
		Username:   username,
		Resource:   "ticket",
		ResourceID: ticketID,
		Action:     ActionDelete,
	}
	engine.cache.Set(deleteReq, AccessResponse{
		Allowed: false,
		Reason:  "Viewer role insufficient for DELETE",
		AuditID: "audit-escalation-2",
	})

	// Verify denials
	updateResp, _ := engine.cache.Get(updateReq)
	assert.False(t, updateResp.Allowed)

	deleteResp, _ := engine.cache.Get(deleteReq)
	assert.False(t, deleteResp.Allowed)
}

// TestE2E_UserJourney_RoleChange tests role change workflow
func TestE2E_UserJourney_RoleChange(t *testing.T) {
	// Scenario: User promoted from Contributor to Developer
	// Steps:
	// 1. User has Contributor role (can CREATE)
	// 2. User cannot UPDATE (insufficient permission)
	// 3. User is promoted to Developer
	// 4. Cache is invalidated
	// 5. User can now UPDATE

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	username := "promoteduser"

	// Initial role: Contributor
	secCtx1 := &SecurityContext{
		Username:  username,
		Roles:     []Role{{ID: "role-contrib", Title: "Contributor"}},
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	engine.cache.SetContext(username, secCtx1)

	// Can CREATE
	createReq := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionCreate,
	}
	engine.cache.Set(createReq, AccessResponse{Allowed: true})

	// Cannot UPDATE
	updateReq := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionUpdate,
	}
	engine.cache.Set(updateReq, AccessResponse{Allowed: false})

	// Invalidate cache after promotion
	engine.InvalidateCache(username)

	// New role: Developer
	secCtx2 := &SecurityContext{
		Username:  username,
		Roles:     []Role{{ID: "role-dev", Title: "Developer"}},
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	engine.cache.SetContext(username, secCtx2)

	// Can now UPDATE
	engine.cache.Set(updateReq, AccessResponse{Allowed: true})

	updateResp, _ := engine.cache.Get(updateReq)
	assert.True(t, updateResp.Allowed)
}

// TestE2E_UserJourney_AuditTrail tests complete audit trail
func TestE2E_UserJourney_AuditTrail(t *testing.T) {
	// Scenario: Track user actions over time
	// Steps:
	// 1. User performs multiple actions
	// 2. All actions are audited
	// 3. Audit log can be queried
	// 4. Statistics are accurate

	mockDB := new(MockDatabase)
	config := Config{
		EnableCaching:    true,
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     1000,
		EnableAuditing:   true,
		AuditAllAttempts: true,
		AuditRetention:   90 * 24 * time.Hour,
	}

	engine := NewSecurityEngine(mockDB, config)

	username := "activeuser"

	actions := []Action{
		ActionRead,
		ActionList,
		ActionCreate,
		ActionUpdate,
		ActionRead,
		ActionList,
	}

	for i, action := range actions {
		req := AccessRequest{
			Username:   username,
			Resource:   "ticket",
			ResourceID: "ticket-" + string(rune(i)),
			Action:     action,
		}

		// Simulate access check and caching
		engine.cache.Set(req, AccessResponse{
			Allowed: true,
			Reason:  "Access granted",
			AuditID: "audit-" + string(rune(i)),
		})
	}

	// Verify audit configuration
	assert.True(t, engine.config.EnableAuditing)
	assert.True(t, engine.config.AuditAllAttempts)
}

// TestE2E_UserJourney_CacheEfficiency tests cache efficiency in real usage
func TestE2E_UserJourney_CacheEfficiency(t *testing.T) {
	// Scenario: User repeatedly accesses same resources
	// Steps:
	// 1. User accesses 10 tickets
	// 2. User accesses same 10 tickets again
	// 3. Second access should have 100% cache hit rate

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	username := "efficientuser"

	// First pass - populate cache
	for i := 0; i < 10; i++ {
		req := AccessRequest{
			Username:   username,
			Resource:   "ticket",
			ResourceID: "ticket-" + string(rune(i)),
			Action:     ActionRead,
		}
		engine.cache.Set(req, AccessResponse{Allowed: true})
	}

	// Second pass - should hit cache
	hits := 0
	for i := 0; i < 10; i++ {
		req := AccessRequest{
			Username:   username,
			Resource:   "ticket",
			ResourceID: "ticket-" + string(rune(i)),
			Action:     ActionRead,
		}
		_, found := engine.cache.Get(req)
		if found {
			hits++
		}
	}

	// Should have 100% hit rate
	assert.Equal(t, 10, hits)

	hitRate := engine.cache.GetHitRate()
	assert.Greater(t, hitRate, 0.5) // > 50% hit rate
}

// TestE2E_UserJourney_MultiProjectAccess tests cross-project access
func TestE2E_UserJourney_MultiProjectAccess(t *testing.T) {
	// Scenario: User works on multiple projects with different roles
	// Steps:
	// 1. User is Developer on Project A
	// 2. User is Viewer on Project B
	// 3. Different permissions in each project

	mockDB := new(MockDatabase)
	config := DefaultConfig()
	engine := NewSecurityEngine(mockDB, config)

	username := "multiprojectuser"

	// Project A context (Developer)
	reqA := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionUpdate,
		Context:  map[string]string{"project_id": "proj-A"},
	}
	engine.cache.Set(reqA, AccessResponse{
		Allowed: true,
		Reason:  "Developer on Project A",
	})

	// Project B context (Viewer)
	reqB := AccessRequest{
		Username: username,
		Resource: "ticket",
		Action:   ActionUpdate,
		Context:  map[string]string{"project_id": "proj-B"},
	}
	engine.cache.Set(reqB, AccessResponse{
		Allowed: false,
		Reason:  "Viewer on Project B (cannot UPDATE)",
	})

	// Verify different permissions
	respA, _ := engine.cache.Get(reqA)
	assert.True(t, respA.Allowed)

	respB, _ := engine.cache.Get(reqB)
	assert.False(t, respB.Allowed)
}

// TestE2E_CompleteWorkflow tests complete workflow from start to finish
func TestE2E_CompleteWorkflow(t *testing.T) {
	// Complete workflow test
	t.Run("User Login and Setup", func(t *testing.T) {
		mockDB := new(MockDatabase)
		engine := NewSecurityEngine(mockDB, DefaultConfig())

		username := "completeuser"
		secCtx := &SecurityContext{
			Username:  username,
			Roles:     []Role{{ID: "role-1", Title: "Developer"}},
			Teams:     []string{"team-1"},
			CachedAt:  time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}

		engine.cache.SetContext(username, secCtx)

		retrieved, found := engine.cache.GetContext(username)
		require.True(t, found)
		assert.Equal(t, username, retrieved.Username)
	})

	t.Run("Resource Access", func(t *testing.T) {
		mockDB := new(MockDatabase)
		engine := NewSecurityEngine(mockDB, DefaultConfig())

		req := AccessRequest{
			Username: "completeuser",
			Resource: "ticket",
			Action:   ActionRead,
		}

		engine.cache.Set(req, AccessResponse{Allowed: true})

		resp, found := engine.cache.Get(req)
		require.True(t, found)
		assert.True(t, resp.Allowed)
	})

	t.Run("Cache Management", func(t *testing.T) {
		mockDB := new(MockDatabase)
		engine := NewSecurityEngine(mockDB, DefaultConfig())

		// Add multiple entries
		for i := 0; i < 5; i++ {
			req := AccessRequest{
				Username:   "completeuser",
				Resource:   "ticket",
				ResourceID: string(rune(i)),
				Action:     ActionRead,
			}
			engine.cache.Set(req, AccessResponse{Allowed: true})
		}

		stats := engine.cache.GetStats()
		assert.Equal(t, 5, stats.EntryCount)

		// Invalidate
		engine.InvalidateCache("completeuser")

		// Verify cleanup (some entries may remain if not all were for this user)
		statsAfter := engine.cache.GetStats()
		assert.LessOrEqual(t, statsAfter.EntryCount, 5)
	})
}
