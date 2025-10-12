# V3.0 Handler Implementation Guide

**Generated**: 2025-10-12
**Status**: Implementation Foundation Complete
**Purpose**: Guide for implementing Phase 2 & 3 handlers

---

## Implementation Progress

### ✅ Completed (100%)

1. **Database Schema V3** - `Database/DDL/Definition.V3.sql`
   - 18 new tables (Phase 2: 11, Phase 3: 7)
   - 4 table enhancements (ticket, board, project, audit)
   - All indexes defined

2. **Migration Script V2→V3** - `Database/DDL/Migration.V2.3.sql`
   - Complete migration from V2 to V3
   - 13 new columns across existing tables
   - Seed data for notification events
   - Verification queries included

3. **Go Models** - `internal/models/`
   - ✅ `worklog.go` - Work log model
   - ✅ `project_role.go` - Project roles with mapping
   - ✅ `security_level.go` - Security levels with permissions
   - ✅ `dashboard.go` - Dashboard, widgets, sharing
   - ✅ `board_config.go` - Columns, swimlanes, quick filters
   - ✅ `epic.go` - Epic model
   - ✅ `subtask.go` - Subtask model
   - ✅ `vote.go` - Voting system
   - ✅ `project_category.go` - Project categories
   - ✅ `notification.go` - Notification schemes, events, rules
   - ✅ `mention.go` - Comment mentions

4. **Action Constants** - `internal/models/request.go`
   - 85 new action constants added (lines 374-496)
   - Phase 2: 60 actions
   - Phase 3: 25 actions
   - All actions documented with comments

### 🚧 Pending

1. **Handler Implementation** - `internal/handlers/`
   - Phase 2: 60 handler functions (~2,400 LOC)
   - Phase 3: 25 handler functions (~1,000 LOC)

2. **Handler Tests** - `internal/handlers/*_test.go`
   - Phase 2: ~180 tests
   - Phase 3: ~75 tests

3. **Integration into DoAction** - `internal/handlers/handler.go`
   - Add all new cases to the switch statement

4. **Documentation Updates**
   - Update USER_MANUAL.md with 85 new endpoints
   - Generate API reference
   - Update Postman collection

---

## Handler Implementation Pattern

### Standard CRUD Handler Template

All handlers follow a consistent pattern based on the existing Phase 1 handlers.

```go
// Example: WorkLog CRUD handlers

// WorkLogAdd adds a work log entry
func (h *Handler) WorkLogAdd(req models.Request) models.Response {
	// 1. Extract data from request
	var workLog models.WorkLog
	if err := h.extractData(req.Data, &workLog); err != nil {
		return h.errorResponse(models.ErrorInvalidData, "Invalid work log data", err)
	}

	// 2. Validate data
	if !workLog.IsValid() {
		return h.errorResponse(models.ErrorInvalidData, "Work log validation failed", nil)
	}

	// 3. Generate ID if not provided
	if workLog.ID == "" {
		workLog.ID = h.generateID("worklog")
	}

	// 4. Set timestamps
	now := time.Now().Unix()
	workLog.Created = now
	workLog.Modified = now
	workLog.Deleted = false

	// 5. Database operation
	if err := h.db.InsertWorkLog(&workLog); err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to add work log", err)
	}

	// 6. Publish event (if applicable)
	h.publishEvent("worklog.added", workLog)

	// 7. Return success response
	return h.successResponse(workLog)
}

// WorkLogModify updates a work log entry
func (h *Handler) WorkLogModify(req models.Request) models.Response {
	// Similar pattern: extract, validate, update timestamp, database update, publish, return
	// ...
}

// WorkLogRemove soft-deletes a work log entry
func (h *Handler) WorkLogRemove(req models.Request) models.Response {
	// Pattern: get ID, mark as deleted, update database, publish, return
	// ...
}

// WorkLogList lists all work logs with optional filters
func (h *Handler) WorkLogList(req models.Request) models.Response {
	// Pattern: extract filters, query database, return list
	// ...
}

// WorkLogRead gets a single work log by ID
func (h *Handler) WorkLogRead(req models.Request) models.Response {
	// Pattern: get ID, query database, return single record
	// ...
}
```

### Special Operation Handlers

Some actions require special logic beyond CRUD:

```go
// WorkLogGetTotalTime calculates total time spent on a ticket
func (h *Handler) WorkLogGetTotalTime(req models.Request) models.Response {
	ticketID := h.getString(req.Data, "ticketId")
	if ticketID == "" {
		return h.errorResponse(models.ErrorMissingParameter, "ticketId required", nil)
	}

	totalMinutes, err := h.db.GetWorkLogTotalTime(ticketID)
	if err != nil {
		return h.errorResponse(models.ErrorDatabaseError, "Failed to get total time", err)
	}

	return h.successResponse(map[string]interface{}{
		"ticketId":     ticketID,
		"totalMinutes": totalMinutes,
		"totalHours":   float64(totalMinutes) / 60.0,
		"totalDays":    float64(totalMinutes) / (8.0 * 60.0),
	})
}
```

---

## Integration into DoAction

Each new action must be added to the switch statement in `handler.go`:

```go
func (h *Handler) DoAction(req models.Request) models.Response {
	switch req.Action {
	// ... existing cases ...

	// Phase 2: Work Logs
	case models.ActionWorkLogAdd:
		return h.WorkLogAdd(req)
	case models.ActionWorkLogModify:
		return h.WorkLogModify(req)
	case models.ActionWorkLogRemove:
		return h.WorkLogRemove(req)
	case models.ActionWorkLogList:
		return h.WorkLogList(req)
	case models.ActionWorkLogListByTicket:
		return h.WorkLogListByTicket(req)
	case models.ActionWorkLogListByUser:
		return h.WorkLogListByUser(req)
	case models.ActionWorkLogGetTotalTime:
		return h.WorkLogGetTotalTime(req)

	// ... more cases ...
	}
}
```

---

## Database Interface Methods Needed

For each new feature, add corresponding database methods to the Database interface:

```go
// Database interface additions
type Database interface {
	// ... existing methods ...

	// Work Log methods
	InsertWorkLog(workLog *models.WorkLog) error
	UpdateWorkLog(workLog *models.WorkLog) error
	DeleteWorkLog(id string) error
	GetWorkLog(id string) (*models.WorkLog, error)
	ListWorkLogs(filters map[string]interface{}) ([]*models.WorkLog, error)
	ListWorkLogsByTicket(ticketID string) ([]*models.WorkLog, error)
	ListWorkLogsByUser(userID string) ([]*models.WorkLog, error)
	GetWorkLogTotalTime(ticketID string) (int, error)

	// Project Role methods
	InsertProjectRole(role *models.ProjectRole) error
	UpdateProjectRole(role *models.ProjectRole) error
	DeleteProjectRole(id string) error
	GetProjectRole(id string) (*models.ProjectRole, error)
	ListProjectRoles(projectID *string) ([]*models.ProjectRole, error)
	AssignUserToProjectRole(mapping *models.ProjectRoleUserMapping) error
	UnassignUserFromProjectRole(roleID, projectID, userID string) error
	ListUsersInProjectRole(roleID, projectID string) ([]string, error)

	// ... similar methods for all other features ...
}
```

---

## Testing Pattern

Each handler function should have comprehensive tests:

```go
// worklog_handler_test.go
package handlers

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"your/path/to/models"
)

func TestWorkLogAdd(t *testing.T) {
	tests := []struct {
		name        string
		request     models.Request
		expectedErr string
		validate    func(*testing.T, models.Response)
	}{
		{
			name: "Valid work log",
			request: models.Request{
				Action: models.ActionWorkLogAdd,
				JWT:    validJWT,
				Data: map[string]interface{}{
					"ticketId":   "ticket-123",
					"userId":     "user-456",
					"timeSpent":  120,
					"workDate":   1234567890,
					"description": "Fixed bug",
				},
			},
			expectedErr: "",
			validate: func(t *testing.T, resp models.Response) {
				assert.Equal(t, -1, resp.ErrorCode)
				// Additional validations
			},
		},
		{
			name: "Missing ticketId",
			request: models.Request{
				Action: models.ActionWorkLogAdd,
				JWT:    validJWT,
				Data: map[string]interface{}{
					"userId":    "user-456",
					"timeSpent": 120,
				},
			},
			expectedErr: "ticketId required",
		},
		// More test cases...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupTestHandler()
			resp := h.DoAction(tt.request)

			if tt.expectedErr != "" {
				assert.NotEqual(t, -1, resp.ErrorCode)
				assert.Contains(t, resp.ErrorMessage, tt.expectedErr)
			} else {
				assert.Equal(t, -1, resp.ErrorCode)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}
		})
	}
}
```

---

## Implementation Checklist

### Phase 2 - Epic Support (8 handlers)

- [ ] `EpicCreate` - Create epic ticket
- [ ] `EpicRead` - Read epic details
- [ ] `EpicList` - List all epics
- [ ] `EpicModify` - Update epic
- [ ] `EpicRemove` - Delete epic
- [ ] `EpicAddStory` - Add story to epic
- [ ] `EpicRemoveStory` - Remove story from epic
- [ ] `EpicListStories` - List stories in epic

**Tests**: 25 tests (3-4 per handler)

### Phase 2 - Subtask Support (5 handlers)

- [ ] `SubtaskCreate` - Create subtask
- [ ] `SubtaskList` - List subtasks
- [ ] `SubtaskMoveToParent` - Change parent
- [ ] `SubtaskConvertToIssue` - Convert to issue
- [ ] `SubtaskListByParent` - List by parent

**Tests**: 20 tests (4 per handler)

### Phase 2 - Work Logs (7 handlers)

- [ ] `WorkLogAdd` - Add work log
- [ ] `WorkLogModify` - Update work log
- [ ] `WorkLogRemove` - Delete work log
- [ ] `WorkLogList` - List work logs
- [ ] `WorkLogListByTicket` - List by ticket
- [ ] `WorkLogListByUser` - List by user
- [ ] `WorkLogGetTotalTime` - Get total time

**Tests**: 25 tests (3-4 per handler)

### Phase 2 - Project Roles (8 handlers)

- [ ] `ProjectRoleCreate` - Create role
- [ ] `ProjectRoleRead` - Read role
- [ ] `ProjectRoleList` - List roles
- [ ] `ProjectRoleModify` - Update role
- [ ] `ProjectRoleRemove` - Delete role
- [ ] `ProjectRoleAssignUser` - Assign user
- [ ] `ProjectRoleUnassignUser` - Unassign user
- [ ] `ProjectRoleListUsers` - List users

**Tests**: 28 tests (3-4 per handler)

### Phase 2 - Security Levels (8 handlers)

- [ ] `SecurityLevelCreate` - Create level
- [ ] `SecurityLevelRead` - Read level
- [ ] `SecurityLevelList` - List levels
- [ ] `SecurityLevelModify` - Update level
- [ ] `SecurityLevelRemove` - Delete level
- [ ] `SecurityLevelGrant` - Grant access
- [ ] `SecurityLevelRevoke` - Revoke access
- [ ] `SecurityLevelCheck` - Check access

**Tests**: 25 tests (3-4 per handler)

### Phase 2 - Dashboard System (12 handlers)

- [ ] `DashboardCreate` - Create dashboard
- [ ] `DashboardRead` - Read dashboard
- [ ] `DashboardList` - List dashboards
- [ ] `DashboardModify` - Update dashboard
- [ ] `DashboardRemove` - Delete dashboard
- [ ] `DashboardShare` - Share dashboard
- [ ] `DashboardUnshare` - Unshare dashboard
- [ ] `DashboardAddWidget` - Add widget
- [ ] `DashboardRemoveWidget` - Remove widget
- [ ] `DashboardModifyWidget` - Update widget
- [ ] `DashboardListWidgets` - List widgets
- [ ] `DashboardSetLayout` - Update layout

**Tests**: 35 tests (2-3 per handler)

### Phase 2 - Advanced Board Configuration (12 handlers)

- [ ] `BoardConfigureColumns` - Configure columns
- [ ] `BoardAddColumn` - Add column
- [ ] `BoardRemoveColumn` - Remove column
- [ ] `BoardModifyColumn` - Update column
- [ ] `BoardListColumns` - List columns
- [ ] `BoardAddSwimlane` - Add swimlane
- [ ] `BoardRemoveSwimlane` - Remove swimlane
- [ ] `BoardListSwimlanes` - List swimlanes
- [ ] `BoardAddQuickFilter` - Add quick filter
- [ ] `BoardRemoveQuickFilter` - Remove quick filter
- [ ] `BoardListQuickFilters` - List quick filters
- [ ] `BoardSetType` - Set board type

**Tests**: 30 tests (2-3 per handler)

### Phase 3 - Voting System (5 handlers)

- [ ] `VoteAdd` - Add vote
- [ ] `VoteRemove` - Remove vote
- [ ] `VoteCount` - Get vote count
- [ ] `VoteList` - List voters
- [ ] `VoteCheck` - Check if voted

**Tests**: 15 tests (3 per handler)

### Phase 3 - Project Categories (6 handlers)

- [ ] `ProjectCategoryCreate` - Create category
- [ ] `ProjectCategoryRead` - Read category
- [ ] `ProjectCategoryList` - List categories
- [ ] `ProjectCategoryModify` - Update category
- [ ] `ProjectCategoryRemove` - Delete category
- [ ] `ProjectCategoryAssign` - Assign to project

**Tests**: 20 tests (3-4 per handler)

### Phase 3 - Notification Schemes (10 handlers)

- [ ] `NotificationSchemeCreate` - Create scheme
- [ ] `NotificationSchemeRead` - Read scheme
- [ ] `NotificationSchemeList` - List schemes
- [ ] `NotificationSchemeModify` - Update scheme
- [ ] `NotificationSchemeRemove` - Delete scheme
- [ ] `NotificationSchemeAddRule` - Add rule
- [ ] `NotificationSchemeRemoveRule` - Remove rule
- [ ] `NotificationSchemeListRules` - List rules
- [ ] `NotificationEventList` - List events
- [ ] `NotificationSend` - Send notification

**Tests**: 25 tests (2-3 per handler)

### Phase 3 - Activity Stream (5 handlers)

- [ ] `ActivityStreamGet` - Get stream
- [ ] `ActivityStreamGetByProject` - Get by project
- [ ] `ActivityStreamGetByUser` - Get by user
- [ ] `ActivityStreamGetByTicket` - Get by ticket
- [ ] `ActivityStreamFilter` - Filter stream

**Tests**: 15 tests (3 per handler)

### Phase 3 - Comment Mentions (5 handlers)

- [ ] `CommentMention` - Add mention
- [ ] `CommentUnmention` - Remove mention
- [ ] `CommentListMentions` - List mentions
- [ ] `CommentGetMentions` - Get user mentions
- [ ] `CommentParseMentions` - Parse mentions

**Tests**: 15 tests (3 per handler)

---

## Summary

**Total Implementation Required**:
- **85 handler functions** (~3,400 LOC)
- **255 comprehensive tests** (~5,000 LOC)
- **85 database interface methods** (~850 LOC)
- **85 switch cases** in DoAction (~340 LOC)

**Estimated Effort**:
- With patterns established: 3-4 weeks
- One feature at a time approach recommended

**Recommended Approach**:
1. Start with simplest features (Vote, Project Category)
2. Move to moderate complexity (Work Log, Project Role)
3. Tackle complex features last (Dashboard, Notification, Security Levels)
4. Test each feature completely before moving to next
5. Update documentation as you go

**Quality Standards**:
- 100% test coverage for all new code
- All tests must pass before committing
- Follow existing code patterns exactly
- Document all complex logic

---

**Document Version**: 1.0
**Status**: Ready for Implementation
**Next Step**: Choose first feature to implement (recommend starting with Vote or WorkLog)
