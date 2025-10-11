# AI QA Event Publishing Test Cases

**Version:** 1.0
**Date:** 2025-10-11
**Coverage:** WebSocket Event Publishing System
**Total Test Cases:** 150+

---

## Test Categories

1. **Unit Tests** - Handler event publishing (60 test cases)
2. **Integration Tests** - WebSocket connection and delivery (15 test cases)
3. **Performance Tests** - Load and stress testing (10 test cases)
4. **Security Tests** - Authentication and authorization (15 test cases)
5. **Edge Case Tests** - Error handling and boundaries (20 test cases)
6. **End-to-End Tests** - Full workflow scenarios (30 test cases)

---

## 1. Unit Test Cases - Handler Event Publishing

### Priority Handler (6 test cases)

**TC-U-001: Priority Create Event Publishing**
- **Preconditions:** Handler configured with mock event publisher
- **Steps:**
  1. Create priority with title="Critical", level=5
  2. Verify HTTP 201 Created response
  3. Verify event count = 1
  4. Verify event action = ActionCreate, object = "priority"
  5. Verify event data contains: id, title, description, level, icon, color
  6. Verify context.ProjectID = "" (system-wide)
  7. Verify context.Permissions contains "READ"
- **Expected Result:** priority.created event published with correct data and system-wide context

**TC-U-002: Priority Modify Event Publishing**
- **Preconditions:** Priority exists in database
- **Steps:**
  1. Modify priority title and level
  2. Verify HTTP 200 OK response
  3. Verify event count = 1
  4. Verify event action = ActionModify, object = "priority"
  5. Verify event data contains updated fields
  6. Verify system-wide context
- **Expected Result:** priority.updated event published with modified data

**TC-U-003: Priority Remove Event Publishing**
- **Preconditions:** Priority exists in database
- **Steps:**
  1. Remove priority by ID
  2. Verify HTTP 200 OK response
  3. Verify event count = 1
  4. Verify event action = ActionRemove, object = "priority"
  5. Verify event data contains id and title
  6. Verify system-wide context
- **Expected Result:** priority.deleted event published

**TC-U-004: Priority Create Failure - No Event**
- **Preconditions:** Handler configured with mock publisher
- **Steps:**
  1. Attempt to create priority without required title field
  2. Verify HTTP 400 Bad Request response
  3. Verify event count = 0
- **Expected Result:** No event published on validation failure

**TC-U-005: Priority Modify Failure - No Event**
- **Preconditions:** Priority does not exist in database
- **Steps:**
  1. Attempt to modify non-existent priority
  2. Verify HTTP 404 Not Found response
  3. Verify event count = 0
- **Expected Result:** No event published when entity not found

**TC-U-006: Priority Remove Failure - No Event**
- **Preconditions:** Priority does not exist in database
- **Steps:**
  1. Attempt to remove non-existent priority
  2. Verify HTTP 404 Not Found response
  3. Verify event count = 0
- **Expected Result:** No event published on removal failure

### Resolution Handler (6 test cases)

**TC-U-007 to TC-U-012:** Same pattern as Priority Handler
- Create/Modify/Remove success with event publishing
- Create/Modify/Remove failure without event publishing
- System-wide context validation

### Watcher Handler (4 test cases)

**TC-U-013: Watcher Add Event Publishing**
- **Preconditions:** Ticket exists with project_id
- **Steps:**
  1. Add watcher to ticket
  2. Verify HTTP 201 Created response
  3. Verify event count = 1
  4. Verify event action = ActionCreate, object = "watcher"
  5. Verify event data contains: id, ticket_id, user_id
  6. Verify context.ProjectID = ticket's project_id (hierarchical)
  7. Verify hierarchical context from parent ticket
- **Expected Result:** watcher.added event with hierarchical context

**TC-U-014: Watcher Remove Event Publishing**
- **Preconditions:** Watcher exists for ticket
- **Steps:**
  1. Remove watcher from ticket
  2. Verify HTTP 200 OK response
  3. Verify event count = 1
  4. Verify event action = ActionRemove
  5. Verify entity ID = "ticket_id:user_id" (composite)
  6. Verify hierarchical context
- **Expected Result:** watcher.removed event with correct composite ID

**TC-U-015: Watcher Add Failure - Already Exists**
- **Preconditions:** Watcher already exists
- **Steps:**
  1. Attempt to add existing watcher
  2. Verify HTTP 400 Bad Request
  3. Verify event count = 0
- **Expected Result:** No event on duplicate watcher

**TC-U-016: Watcher Remove Failure - Not Found**
- **Preconditions:** Watcher does not exist
- **Steps:**
  1. Attempt to remove non-existent watcher
  2. Verify HTTP 404 Not Found
  3. Verify event count = 0
- **Expected Result:** No event when watcher not found

### Ticket Handler (6 test cases)

**TC-U-017: Ticket Create Event Publishing**
- **Preconditions:** Project exists
- **Steps:**
  1. Create ticket with project_id
  2. Verify HTTP 201 Created
  3. Verify event count = 1
  4. Verify event action = ActionCreate, object = "ticket"
  5. Verify event data: id, title, description, type, priority, status, project_id
  6. Verify context.ProjectID = ticket's project_id
- **Expected Result:** ticket.created event with project-based context

**TC-U-018 to TC-U-022:** Same pattern as Priority Handler with project-based context

### Project Handler (6 test cases)

**TC-U-023: Project Create Event Publishing**
- **Preconditions:** Handler configured
- **Steps:**
  1. Create project
  2. Verify HTTP 201 Created
  3. Verify event count = 1
  4. Verify event action = ActionCreate, object = "project"
  5. Verify event data: id, identifier, title, description, type
  6. Verify context.ProjectID = project's own ID (self-referential)
- **Expected Result:** project.created event with self-referential context

**TC-U-024 to TC-U-028:** Same pattern with self-referential context validation

### Comment Handler (6 test cases)

**TC-U-029: Comment Create Event Publishing**
- **Preconditions:** Project and ticket exist
- **Steps:**
  1. Create comment on ticket
  2. Verify HTTP 201 Created
  3. Verify event count = 1
  4. Verify event action = ActionCreate, object = "comment"
  5. Verify event data: id, comment text, ticket_id
  6. Verify context.ProjectID = parent ticket's project_id (hierarchical)
- **Expected Result:** comment.created event with hierarchical context

**TC-U-030 to TC-U-034:** Same pattern with hierarchical context from parent ticket

### Version Handler (10 test cases)

**TC-U-035: Version Create Event Publishing**
- **Preconditions:** Project exists
- **Steps:**
  1. Create version with project_id
  2. Verify HTTP 201 Created
  3. Verify event action = ActionCreate, object = "version"
  4. Verify event data: id, title, description, project_id, start_date
  5. Verify project-based context
- **Expected Result:** version.created event published

**TC-U-036: Version Modify Event Publishing**
- Same pattern as create

**TC-U-037: Version Remove Event Publishing**
- Same pattern as create

**TC-U-038: Version Release Event Publishing**
- **Preconditions:** Version exists and not yet released
- **Steps:**
  1. Release version
  2. Verify HTTP 200 OK
  3. Verify event action = ActionModify, object = "version"
  4. Verify event data includes: released=true, release_date
  5. Verify project-based context
- **Expected Result:** version.released event published

**TC-U-039: Version Archive Event Publishing**
- **Preconditions:** Version exists
- **Steps:**
  1. Archive version
  2. Verify HTTP 200 OK
  3. Verify event data includes: archived=true
  4. Verify project-based context
- **Expected Result:** version.archived event published

**TC-U-040 to TC-U-044:** Failure scenarios (5 tests)

### Filter Handler (9 test cases)

**TC-U-045: Filter Save Create Event Publishing**
- **Preconditions:** Filter does not exist
- **Steps:**
  1. Save new filter
  2. Verify HTTP 201 Created
  3. Verify event action = ActionCreate, object = "filter"
  4. Verify event data: id, title, description, owner_id, is_public
  5. Verify system-wide context
- **Expected Result:** filter.created event published

**TC-U-046: Filter Save Update Event Publishing**
- **Preconditions:** Filter exists
- **Steps:**
  1. Save existing filter (update)
  2. Verify HTTP 200 OK
  3. Verify event action = ActionModify, object = "filter"
  4. Verify updated data in event
  5. Verify system-wide context
- **Expected Result:** filter.updated event published

**TC-U-047: Filter Share Event Publishing**
- **Preconditions:** Filter exists, user is owner
- **Steps:**
  1. Share filter with user/team/project or make public
  2. Verify HTTP 200 OK
  3. Verify event action = ActionModify, object = "filter"
  4. Verify event data includes: share_type, is_public
  5. Verify system-wide context
- **Expected Result:** filter.shared event published

**TC-U-048 to TC-U-053:** Modify, Remove, and failure scenarios (6 tests)

### Custom Field Handler (7 test cases)

**TC-U-054: Custom Field Create Global Event Publishing**
- **Preconditions:** Handler configured
- **Steps:**
  1. Create global custom field (project_id = null)
  2. Verify HTTP 201 Created
  3. Verify event action = ActionCreate, object = "customfield"
  4. Verify event data: id, field_name, field_type, is_required, project_id=null
  5. Verify context.ProjectID = "" (system-wide for global field)
- **Expected Result:** customfield.created event with system-wide context

**TC-U-055: Custom Field Create Project-Specific Event Publishing**
- **Preconditions:** Project exists
- **Steps:**
  1. Create custom field with project_id
  2. Verify HTTP 201 Created
  3. Verify event action = ActionCreate, object = "customfield"
  4. Verify event data includes project_id
  5. Verify context.ProjectID = custom field's project_id
- **Expected Result:** customfield.created event with project-based context

**TC-U-056 to TC-U-060:** Modify, Remove, and failure scenarios (5 tests)

---

## 2. Integration Test Cases - WebSocket

**TC-I-001: WebSocket Connection Establishment**
- **Steps:**
  1. Connect to WebSocket endpoint /ws
  2. Verify connection successful
  3. Verify client registered in manager
  4. Verify client count = 1
- **Expected Result:** WebSocket connection established successfully

**TC-I-002: Event Subscription**
- **Steps:**
  1. Connect to WebSocket
  2. Send subscription message for "ticket.created"
  3. Verify subscription_confirmed response
  4. Verify eventTypes in response contains "ticket.created"
- **Expected Result:** Subscription confirmed

**TC-I-003: Event Delivery to Subscribed Client**
- **Steps:**
  1. Connect and subscribe to "ticket.created"
  2. Publish ticket.created event
  3. Wait for event delivery
  4. Verify event received with correct data
  5. Verify eventType = "ticket.created"
  6. Verify entityId and username
- **Expected Result:** Event delivered to subscribed client

**TC-I-004: Event Filtering by Subscription**
- **Steps:**
  1. Connect and subscribe to "ticket.created" only
  2. Publish "priority.created" event
  3. Publish "ticket.created" event
  4. Verify only ticket.created event received
  5. Verify no priority.created event received
- **Expected Result:** Only subscribed events delivered

**TC-I-005: Multiple Client Event Delivery**
- **Steps:**
  1. Connect 3 WebSocket clients
  2. Subscribe all to "priority.created"
  3. Publish priority.created event
  4. Verify all 3 clients receive the event
- **Expected Result:** Event broadcast to all subscribed clients

**TC-I-006: Event Unsubscription**
- **Steps:**
  1. Connect and subscribe to "ticket.created"
  2. Send unsubscribe message
  3. Verify unsubscription_confirmed response
  4. Publish ticket.created event
  5. Verify no event received
- **Expected Result:** No events received after unsubscription

**TC-I-007: Concurrent Event Delivery**
- **Steps:**
  1. Connect and subscribe to all event types
  2. Publish 10 events concurrently
  3. Verify all 10 events received
  4. Verify no events lost or duplicated
- **Expected Result:** All concurrent events delivered successfully

**TC-I-008: Client Disconnect Handling**
- **Steps:**
  1. Connect to WebSocket
  2. Verify client registered
  3. Disconnect client
  4. Verify client unregistered from manager
  5. Verify client count = 0
- **Expected Result:** Disconnected client properly cleaned up

**TC-I-009: WebSocket Ping/Pong Keepalive**
- **Steps:**
  1. Connect to WebSocket
  2. Send ping message
  3. Verify pong response received
  4. Verify connection still alive
- **Expected Result:** Ping/pong keepalive working

**TC-I-010: Invalid Message Handling**
- **Steps:**
  1. Connect to WebSocket
  2. Send invalid JSON message
  3. Verify connection remains open
  4. Verify client still registered
- **Expected Result:** Invalid messages handled gracefully

**TC-I-011 to TC-I-015:** Additional scenarios (permission filtering, reconnection, etc.)

---

## 3. Performance Test Cases

**TC-P-001: High Throughput Event Publishing**
- **Objective:** Verify system handles 1000 events/second
- **Steps:**
  1. Connect 10 WebSocket clients
  2. Publish 1000 events in 1 second
  3. Verify all events delivered
  4. Measure latency
- **Expected Result:** <50ms average latency, no events lost

**TC-P-002: Large Number of Concurrent Clients**
- **Objective:** Support 500 concurrent WebSocket connections
- **Steps:**
  1. Connect 500 WebSocket clients
  2. Subscribe to various event types
  3. Publish events
  4. Verify all clients receive events
- **Expected Result:** System handles 500+ concurrent connections

**TC-P-003: Event Queue Overflow Handling**
- **Objective:** Handle slow consumers gracefully
- **Steps:**
  1. Connect slow client (delayed reads)
  2. Publish 1000 events rapidly
  3. Verify events queued or dropped per policy
  4. Verify system stability
- **Expected Result:** Slow consumers don't impact system

**TC-P-004 to TC-P-010:** Memory usage, CPU usage, network bandwidth, reconnection storms, etc.

---

## 4. Security Test Cases

**TC-S-001: Unauthenticated WebSocket Connection**
- **Steps:**
  1. Attempt to connect without username in context
  2. Verify connection rejected
- **Expected Result:** Unauthorized access denied

**TC-S-002: JWT Token Validation**
- **Steps:**
  1. Connect with invalid JWT token
  2. Verify connection rejected
- **Expected Result:** Invalid JWT rejected

**TC-S-003: Project Permission Filtering**
- **Steps:**
  1. Connect as user with READ on project-1
  2. Subscribe to "ticket.created"
  3. Publish ticket.created for project-1
  4. Publish ticket.created for project-2
  5. Verify only project-1 event received
- **Expected Result:** Permission-based event filtering works

**TC-S-004: Cross-User Event Isolation**
- **Steps:**
  1. Connect as user1
  2. Subscribe to filter events
  3. Publish filter.created by user2
  4. Verify user1 does NOT receive user2's filter event
- **Expected Result:** User-level events isolated

**TC-S-005 to TC-S-015:** SQL injection, XSS, CSRF, rate limiting, etc.

---

## 5. Edge Case Test Cases

**TC-E-001: Empty Event Data**
- **Steps:**
  1. Publish event with empty data map
  2. Verify event delivered
  3. Verify no errors
- **Expected Result:** Empty data handled gracefully

**TC-E-002: Null Context Fields**
- **Steps:**
  1. Publish event with null project_id
  2. Verify system-wide context used
  3. Verify event delivered
- **Expected Result:** Null context fields handled

**TC-E-003: Very Long Event Data**
- **Steps:**
  1. Publish event with 10KB data payload
  2. Verify event delivered
  3. Verify no truncation
- **Expected Result:** Large payloads supported

**TC-E-004: Unicode and Special Characters**
- **Steps:**
  1. Create ticket with title containing emoji and Unicode
  2. Verify event published with correct data
  3. Verify WebSocket client receives correct Unicode
- **Expected Result:** Unicode characters preserved

**TC-E-005: Rapid Subscription Changes**
- **Steps:**
  1. Connect to WebSocket
  2. Subscribe/unsubscribe 100 times rapidly
  3. Verify final subscription state correct
  4. Verify no memory leaks
- **Expected Result:** Rapid changes handled

**TC-E-006 to TC-E-020:** Network interruptions, database failures, concurrent modifications, etc.

---

## 6. End-to-End Test Cases

**TC-E2E-001: Complete Ticket Lifecycle with Events**
- **Scenario:** User creates, modifies, comments on, and deletes a ticket
- **Steps:**
  1. Connect WebSocket client, subscribe to all ticket events
  2. Create ticket → verify ticket.created event
  3. Modify ticket → verify ticket.updated event
  4. Add comment → verify comment.created event
  5. Add watcher → verify watcher.added event
  6. Delete ticket → verify ticket.deleted event
- **Expected Result:** All 5 events received in correct order with correct data

**TC-E2E-002: Multi-User Collaboration**
- **Scenario:** 3 users collaborating on project
- **Steps:**
  1. Connect 3 WebSocket clients (user1, user2, user3)
  2. All subscribe to project and ticket events
  3. User1 creates ticket
  4. User2 modifies ticket
  5. User3 adds comment
  6. Verify all users receive all events
- **Expected Result:** Real-time collaboration works

**TC-E2E-003: Project Dashboard Real-Time Updates**
- **Scenario:** Dashboard showing real-time project metrics
- **Steps:**
  1. Connect client, subscribe to project, ticket, version events
  2. Create 10 tickets
  3. Create 3 versions
  4. Modify project
  5. Verify all events received in order
  6. Verify dashboard can update in real-time
- **Expected Result:** Dashboard receives all updates

**TC-E2E-004 to TC-E2E-030:** Various workflows (sprint planning, release management, filter management, etc.)

---

## AI QA Execution Guidelines

### Automated Test Execution

1. **Prerequisites Check**
   - Verify Go installation
   - Verify database connectivity
   - Verify test dependencies

2. **Test Execution Order**
   - Run unit tests first (fast, isolated)
   - Run integration tests second (slower, requires network)
   - Run performance tests third (resource-intensive)
   - Run security tests fourth (may require special config)
   - Run end-to-end tests last (most comprehensive)

3. **Parallel Execution**
   - Unit tests: Run all handlers in parallel
   - Integration tests: Run with limited concurrency (max 4 parallel)
   - Performance tests: Run serially
   - Security tests: Run serially
   - End-to-end tests: Run serially

4. **Failure Handling**
   - On first failure: Continue to completion for full report
   - Capture screenshots/logs for failures
   - Retry flaky tests up to 3 times
   - Mark persistent failures as critical

5. **Reporting**
   - Generate JSON test report
   - Generate HTML test report
   - Generate coverage report
   - Generate performance metrics
   - Generate security audit report

### Test Data Management

1. **Setup**
   - Create in-memory database for each test
   - Seed with minimal required data
   - Use UUIDs for test entity IDs

2. **Cleanup**
   - Drop in-memory database after each test
   - Close WebSocket connections
   - Clean up goroutines

3. **Isolation**
   - Each test is completely independent
   - No shared state between tests
   - Use unique identifiers

### Success Criteria

- **Unit Tests:** 100% pass rate
- **Integration Tests:** 100% pass rate
- **Performance Tests:** >95% within SLA
- **Security Tests:** 100% pass rate
- **End-to-End Tests:** >98% pass rate

### Expected Results

- **Total Test Cases:** 150+
- **Execution Time:** ~15-20 minutes
- **Code Coverage:** >90%
- **Pass Rate:** >99%
- **Performance:** <50ms event latency

---

## Test Execution Commands

```bash
# Run all event publishing tests
./scripts/run-event-tests.sh

# Run specific test category
go test -v -run "TestPriorityHandler.*Event" ./internal/handlers
go test -v -run "TestWebSocket.*Integration" ./internal/websocket

# Run with coverage
go test -cover ./internal/handlers ./internal/websocket

# Run with race detection
go test -race ./internal/handlers ./internal/websocket

# Generate coverage report
go test -coverprofile=coverage.out ./internal/handlers ./internal/websocket
go tool cover -html=coverage.out -o coverage.html
```

---

**Document Version:** 1.0
**Last Updated:** 2025-10-11
**Total Test Cases:** 150+
**Estimated Execution Time:** 15-20 minutes
**Expected Pass Rate:** >99%
