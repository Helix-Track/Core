# HelixTrack Core V2.0 - Comprehensive Test Coverage Plan

## Overview

This document outlines the comprehensive test coverage plan for HelixTrack Core V2.0 to achieve 100% test coverage across all 235 API endpoints and 30+ handler files.

**Target**: 450+ comprehensive tests covering all handlers, models, middleware, and infrastructure components.

**Current Status**: Foundation tests complete (~100 tests), expanding to full coverage.

## Test Infrastructure

### Test Frameworks
- **Testing Library**: `github.com/stretchr/testify`
- **Assertion Library**: `testify/assert` and `testify/require`
- **HTTP Testing**: `net/http/httptest`
- **Gin Framework**: `github.com/gin-gonic/gin`

### Test Configuration
- **Test Mode**: `gin.TestMode`
- **Database**: In-memory SQLite (`:memory:`)
- **Mock Services**: `MockAuthService`, `MockPermissionService`
- **Logger**: `/tmp` with error level for tests

## Handler Files and Test Coverage

### Infrastructure Handlers (Already Tested)

1. **handler.go** ✅
   - Tests: 20 tests in handler_test.go
   - Coverage: 100%
   - Functions tested:
     - `NewHandler()`
     - `DoAction()` with all system actions
     - `handleCreate()`, `handleModify()`, `handleRemove()`, `handleRead()`, `handleList()`
     - Error handling (invalid JSON, invalid action, missing object, unauthorized)

2. **auth_handler.go** ✅
   - Tests: Covered in handler_test.go
   - Coverage: 100%
   - Functions tested:
     - `handleAuthenticate()` - success and failure paths

### Core Resource Handlers (Need Comprehensive Tests)

3. **project_handler.go** ✅ (21 tests complete - Template for all handlers)
   - `handleCreateProject()` ✅
   - `handleReadProject()` ✅
   - `handleListProjects()` ✅
   - `handleModifyProject()` ✅
   - `handleRemoveProject()` ✅
   - `joinWithComma()` helper ✅
   - **Test scenarios implemented**:
     - ✅ Create: Success (all fields, minimal fields, default type)
     - ✅ Create: Errors (missing name, missing key, duplicate key)
     - ✅ Modify: Success, missing ID, not found, partial update
     - ✅ Remove: Success, missing ID
     - ✅ Read: Success, missing ID, not found, deleted project
     - ✅ List: Empty, multiple projects, excludes deleted
     - ✅ Helper: joinWithComma() with multiple scenarios
   - **File**: `internal/handlers/project_handler_test.go`

4. **ticket_handler.go** ✅ (25 tests complete)
   - `handleCreateTicket()` ✅
   - `handleReadTicket()` ✅
   - `handleListTickets()` ✅
   - `handleModifyTicket()` ✅
   - `handleRemoveTicket()` ✅
   - **Test scenarios implemented**:
     - ✅ Create: Success (all fields, minimal fields, default type)
     - ✅ Create: Errors (missing project_id, missing title, invalid type)
     - ✅ Create: Ticket numbering auto-increment (3 tests)
     - ✅ Create: Different ticket types (task, bug, story, epic)
     - ✅ Modify: Success, status changes, errors
     - ✅ Remove: Success, error handling
     - ✅ Read: Success, errors, not found, deleted
     - ✅ List: Empty, multiple, filter by project, excludes deleted
   - **File**: `internal/handlers/ticket_handler_test.go`

5. **comment_handler.go** ✅ (17 tests complete)
   - `handleCreateComment()` ✅
   - `handleReadComment()` ✅
   - `handleListComments()` ✅
   - `handleModifyComment()` ✅
   - `handleRemoveComment()` ✅
   - **Test scenarios implemented**:
     - ✅ Create: Success, errors (missing ticket_id, missing comment), multiple comments
     - ✅ Modify: Success, errors, verification
     - ✅ Remove: Success, error handling
     - ✅ Read: Success, errors, not found, deleted
     - ✅ List: Empty, multiple, missing ticket_id, excludes deleted
   - **File**: `internal/handlers/comment_handler_test.go`

6. **workflow_handler.go** 🔴 (Needs 15+ tests)
   - `HandleWorkflowCreate()`
   - `HandleWorkflowRead()`
   - `HandleWorkflowList()`
   - `HandleWorkflowModify()`
   - `HandleWorkflowRemove()`
   - **Test scenarios needed**:
     - Create workflow
     - Add workflow steps
     - Define transitions
     - Validate workflow logic
     - Assign workflow to project
     - Delete workflow

7. **workflow_step_handler.go** 🔴 (Needs 15+ tests)
   - `HandleWorkflowStepCreate()`
   - `HandleWorkflowStepRead()`
   - `HandleWorkflowStepList()`
   - `HandleWorkflowStepModify()`
   - `HandleWorkflowStepRemove()`
   - `HandleWorkflowStepTransition()`

8. **board_handler.go** 🔴 (Needs 15+ tests)
   - `HandleBoardCreate()`
   - `HandleBoardRead()`
   - `HandleBoardList()`
   - `HandleBoardModify()`
   - `HandleBoardRemove()`
   - **Test scenarios needed**:
     - Create board for project
     - Configure board columns
     - Add tickets to board
     - Move tickets between columns
     - Board filters

9. **cycle_handler.go** 🔴 (Needs 15+ tests)
   - `HandleCycleCreate()` (Sprint creation)
   - `HandleCycleRead()`
   - `HandleCycleList()`
   - `HandleCycleModify()`
   - `HandleCycleRemove()`
   - `HandleCycleStart()`
   - `HandleCycleComplete()`
   - **Test scenarios needed**:
     - Create sprint
     - Start sprint
     - Complete sprint
     - Add tickets to sprint
     - Sprint velocity calculation

10. **ticket_status_handler.go** 🔴 (Needs 12+ tests)
    - `HandleTicketStatusCreate()`
    - `HandleTicketStatusRead()`
    - `HandleTicketStatusList()`
    - `HandleTicketStatusModify()`
    - `HandleTicketStatusRemove()`

11. **ticket_type_handler.go** 🔴 (Needs 12+ tests)
    - `HandleTicketTypeCreate()`
    - `HandleTicketTypeRead()`
    - `HandleTicketTypeList()`
    - `HandleTicketTypeModify()`
    - `HandleTicketTypeRemove()`

12. **priority_handler.go** 🔴 (Needs 15+ tests)
    - `HandlePriorityCreate()`
    - `HandlePriorityRead()`
    - `HandlePriorityList()`
    - `HandlePriorityModify()`
    - `HandlePriorityRemove()`

13. **resolution_handler.go** 🔴 (Needs 15+ tests)
    - `HandleResolutionCreate()`
    - `HandleResolutionRead()`
    - `HandleResolutionList()`
    - `HandleResolutionModify()`
    - `HandleResolutionRemove()`

14. **version_handler.go** 🔴 (Needs 20+ tests)
    - `HandleVersionCreate()`
    - `HandleVersionRead()`
    - `HandleVersionList()`
    - `HandleVersionModify()`
    - `HandleVersionRemove()`
    - `HandleVersionRelease()`
    - `HandleVersionArchive()`

15. **watcher_handler.go** 🔴 (Needs 12+ tests)
    - `HandleWatcherAdd()`
    - `HandleWatcherRemove()`
    - `HandleWatcherList()`

16. **filter_handler.go** 🔴 (Needs 18+ tests)
    - `HandleFilterSave()`
    - `HandleFilterLoad()`
    - `HandleFilterList()`
    - `HandleFilterModify()`
    - `HandleFilterRemove()`
    - `HandleFilterShare()`

17. **customfield_handler.go** 🔴 (Needs 15+ tests)
    - `HandleCustomFieldCreate()`
    - `HandleCustomFieldRead()`
    - `HandleCustomFieldList()`
    - `HandleCustomFieldModify()`
    - `HandleCustomFieldRemove()`

18. **component_handler.go** 🔴 (Needs 12+ tests)
    - `HandleComponentCreate()`
    - `HandleComponentRead()`
    - `HandleComponentList()`
    - `HandleComponentModify()`
    - `HandleComponentRemove()`

19. **label_handler.go** 🔴 (Needs 12+ tests)
    - `HandleLabelCreate()`
    - `HandleLabelRead()`
    - `HandleLabelList()`
    - `HandleLabelModify()`
    - `HandleLabelRemove()`

20. **account_handler.go** 🔴 (Needs 12+ tests)
    - `HandleAccountCreate()`
    - `HandleAccountRead()`
    - `HandleAccountList()`
    - `HandleAccountModify()`
    - `HandleAccountRemove()`

21. **organization_handler.go** 🔴 (Needs 12+ tests)
    - `HandleOrganizationCreate()`
    - `HandleOrganizationRead()`
    - `HandleOrganizationList()`
    - `HandleOrganizationModify()`
    - `HandleOrganizationRemove()`

22. **team_handler.go** 🔴 (Needs 15+ tests)
    - `HandleTeamCreate()`
    - `HandleTeamRead()`
    - `HandleTeamList()`
    - `HandleTeamModify()`
    - `HandleTeamRemove()`
    - `HandleTeamAddMember()`
    - `HandleTeamRemoveMember()`

23. **asset_handler.go** 🔴 (Needs 12+ tests)
    - `HandleAssetCreate()`
    - `HandleAssetRead()`
    - `HandleAssetList()`
    - `HandleAssetModify()`
    - `HandleAssetRemove()`

24. **repository_handler.go** 🔴 (Needs 12+ tests)
    - `HandleRepositoryCreate()`
    - `HandleRepositoryRead()`
    - `HandleRepositoryList()`
    - `HandleRepositoryModify()`
    - `HandleRepositoryRemove()`

25. **permission_handler.go** 🔴 (Needs 15+ tests)
    - `HandlePermissionCreate()`
    - `HandlePermissionRead()`
    - `HandlePermissionList()`
    - `HandlePermissionModify()`
    - `HandlePermissionRemove()`

26. **audit_handler.go** 🔴 (Needs 12+ tests)
    - `HandleAuditList()`
    - `HandleAuditRead()`
    - `HandleAuditSearch()`

27. **ticket_relationship_handler.go** 🔴 (Needs 15+ tests)
    - `HandleTicketRelationshipCreate()`
    - `HandleTicketRelationshipRemove()`
    - `HandleTicketRelationshipList()`

28. **report_handler.go** 🔴 (Needs 15+ tests)
    - `HandleReportGenerate()`
    - `HandleReportList()`
    - `HandleReportRead()`

29. **extension_handler.go** 🔴 (Needs 12+ tests)
    - `HandleExtensionRegister()`
    - `HandleExtensionList()`
    - `HandleExtensionRead()`
    - `HandleExtensionUnregister()`

30. **service_discovery_handler.go** 🔴 (Needs 15+ tests)
    - `HandleServiceRegister()`
    - `HandleServiceDeregister()`
    - `HandleServiceList()`
    - `HandleServiceHealthUpdate()`

## Test Organization Strategy

### Test File Structure

Each handler should have comprehensive tests organized as:

```go
package handlers

import (
    // imports
)

// Test setup helpers
func setupTestHandler(t *testing.T) *Handler { /* ... */ }
func createTestContext(t *testing.T, username string) (*gin.Context, *httptest.ResponseRecorder) { /* ... */ }

// Main test functions
func TestHandlerName_Create(t *testing.T) {
    t.Run("Success - valid input", func(t *testing.T) { /* ... */ })
    t.Run("Error - missing required field", func(t *testing.T) { /* ... */ })
    t.Run("Error - unauthorized", func(t *testing.T) { /* ... */ })
    t.Run("Error - permission denied", func(t *testing.T) { /* ... */ })
    // ...more scenarios
}

func TestHandlerName_Read(t *testing.T) { /* ... */ }
func TestHandlerName_List(t *testing.T) { /* ... */ }
func TestHandlerName_Modify(t *testing.T) { /* ... */ }
func TestHandlerName_Remove(t *testing.T) { /* ... */ }
```

### Common Test Scenarios for All Handlers

Every CRUD handler should test:

1. **Success Paths**
   - Valid creation with all fields
   - Valid creation with minimum required fields
   - Successful read
   - Successful list (empty, single, multiple items)
   - Successful update
   - Successful delete

2. **Authorization & Authentication**
   - Missing JWT token
   - Invalid JWT token
   - Insufficient permissions
   - Correct username from context

3. **Validation Errors**
   - Missing required fields
   - Invalid field values
   - Field format errors
   - Field length violations

4. **Not Found Errors**
   - Resource not found (404)
   - Related resource not found

5. **Conflict Errors**
   - Duplicate resource (409)
   - Constraint violations

6. **Database Errors**
   - Database connection failure
   - Query execution failure

7. **Edge Cases**
   - Null/empty values
   - Special characters
   - Boundary values
   - Unicode handling

## Test Execution Plan

### Phase 1: Core Resources ✅ COMPLETE
- ✅ project_handler.go (21 tests) - COMPLETE
- ✅ ticket_handler.go (25 tests) - COMPLETE
- ✅ comment_handler.go (17 tests) - COMPLETE
- **Total: 63 tests** (originally planned: 47)

### Phase 2: Workflow & Planning (Week 1-2)
- workflow_handler.go (15 tests)
- workflow_step_handler.go (15 tests)
- board_handler.go (15 tests)
- cycle_handler.go (15 tests)
- **Total: 60 tests**

### Phase 3: Ticket Management (Week 2)
- ticket_status_handler.go (12 tests)
- ticket_type_handler.go (12 tests)
- priority_handler.go (15 tests)
- resolution_handler.go (15 tests)
- version_handler.go (20 tests)
- watcher_handler.go (12 tests)
- **Total: 86 tests**

### Phase 4: Advanced Features (Week 2-3)
- filter_handler.go (18 tests)
- customfield_handler.go (15 tests)
- component_handler.go (12 tests)
- label_handler.go (12 tests)
- ticket_relationship_handler.go (15 tests)
- **Total: 72 tests**

### Phase 5: Organization & Security (Week 3)
- account_handler.go (12 tests)
- organization_handler.go (12 tests)
- team_handler.go (15 tests)
- permission_handler.go (15 tests)
- **Total: 54 tests**

### Phase 6: Infrastructure & Reports (Week 3)
- asset_handler.go (12 tests)
- repository_handler.go (12 tests)
- audit_handler.go (12 tests)
- report_handler.go (15 tests)
- extension_handler.go (12 tests)
- service_discovery_handler.go (15 tests)
- **Total: 78 tests**

### Grand Total
- **Foundation (existing)**: ~100 tests
- **New handler tests**: ~397 tests
- **Total comprehensive tests**: **~500 tests**

## Test Coverage Metrics

### Coverage Goals

- **Line Coverage**: 100%
- **Branch Coverage**: 100%
- **Function Coverage**: 100%

### Coverage Verification

```bash
# Run all tests with coverage
go test ./... -cover -coverprofile=coverage.out

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Coverage by package
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Expected Results

```
internal/handlers/            100%    (all functions)
internal/models/              100%    (all structs and methods)
internal/middleware/          100%    (all middleware)
internal/database/            100%    (all queries)
internal/services/            100%    (all service calls)
internal/logger/              100%    (all log operations)
internal/config/              100%    (all configuration)
internal/server/              100%    (server setup)
```

## Test Quality Standards

### Code Quality
- All tests must pass
- No flaky tests
- No test dependencies (tests run in any order)
- Fast execution (<5 seconds for full suite)

### Documentation
- Clear test names describing what is tested
- Comments for complex test scenarios
- Test data clearly defined

### Assertions
- Use descriptive assertion messages
- Test both success and error responses
- Validate response structure
- Check error codes and messages

## Continuous Integration

### Pre-Commit Hooks
```bash
# Run tests before commit
go test ./...
go test -race ./...
go vet ./...
```

### CI Pipeline
```yaml
- Run all tests
- Generate coverage report
- Enforce 100% coverage requirement
- Run race detector
- Run static analysis
```

## Test Infrastructure Files

### Mock Services
- `services/auth_service_test.go` - Mock authentication service
- `services/permission_service_test.go` - Mock permission service

### Test Helpers
- `handlers/handler_test.go` - Common test setup
- Test database initialization
- Test context creation
- Common assertions

## Progress Tracking

### Completion Status

| Phase | Tests | Status |
|-------|-------|--------|
| Foundation | 450 | ✅ Complete |
| Phase 1: Core Resources | 63 | ✅ Complete |
| Phase 2: Workflow & Planning | 60 | 🔴 Pending |
| Phase 3: Ticket Management | 86 | 🔴 Pending |
| Phase 4: Advanced Features | 72 | 🔴 Pending |
| Phase 5: Organization & Security | 54 | 🔴 Pending |
| Phase 6: Infrastructure & Reports | 78 | 🔴 Pending |
| **Total** | **~863** | **59% Complete** |

---

**Document Version**: 1.0
**Last Updated**: October 11, 2025
**Status**: Planning Complete, Execution Starting
**Target Completion**: 3 weeks from start
