package testcases

import (
	"fmt"
	"time"
)

// PM Workflow Test Cases for AI QA Framework
// These test cases cover real-world project management scenarios

// ProjectOnboardingTestCase tests complete project setup from scratch
var ProjectOnboardingTestCase = TestCase{
	ID:          "PM-001",
	Name:        "Complete Project Onboarding",
	Category:    "ProjectManagement",
	Priority:    "High",
	Description: "Test complete project setup workflow: org → project → team → workflow → backlog",
	Steps: []TestStep{
		{
			ID:          "PM-001-01",
			Action:      "CreateOrganization",
			Description: "Create organization for the project",
			Request: map[string]interface{}{
				"action": "create",
				"object": "organization",
				"data": map[string]interface{}{
					"name":        "Tech Startup Inc",
					"description": "Innovative software solutions",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.organization.name == 'Tech Startup Inc'",
				},
			},
		},
		{
			ID:          "PM-001-02",
			Action:      "CreateProject",
			Description: "Create new software project",
			Request: map[string]interface{}{
				"action": "create",
				"object": "project",
				"data": map[string]interface{}{
					"name":        "Mobile Banking App",
					"key":         "MBA",
					"description": "Secure mobile banking application",
					"type":        "software",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.project.key == 'MBA'",
				},
			},
		},
		{
			ID:          "PM-001-03",
			Action:      "AddTeamMembers",
			Description: "Add team members to project",
			Request: map[string]interface{}{
				"action": "create",
				"object": "team",
				"data": map[string]interface{}{
					"name":    "MBA Development Team",
					"members": []string{"john_dev", "sarah_qa", "mike_pm", "lisa_designer"},
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"len(response.data.team.members) == 4",
				},
			},
		},
		{
			ID:          "PM-001-04",
			Action:      "ConfigureWorkflow",
			Description: "Setup custom workflow for the project",
			Request: map[string]interface{}{
				"action": "create",
				"object": "workflow",
				"data": map[string]interface{}{
					"name":        "Agile Development Workflow",
					"description": "Standard agile workflow",
					"steps":       []string{"Backlog", "To Do", "In Progress", "Code Review", "QA Testing", "Done"},
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"len(response.data.workflow.steps) == 6",
				},
			},
		},
		{
			ID:          "PM-001-05",
			Action:      "CreateInitialBacklog",
			Description: "Create initial product backlog",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":    "User Authentication Feature",
					"type":     "story",
					"priority": "high",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
	},
	Tags: []string{"project", "onboarding", "workflow", "team"},
}

// SprintPlanningTestCase tests sprint planning and execution
var SprintPlanningTestCase = TestCase{
	ID:          "PM-002",
	Name:        "Sprint Planning and Execution",
	Category:    "ProjectManagement",
	Priority:    "High",
	Description: "Test complete sprint workflow: create sprint → add tickets → estimate → start → execute → complete",
	Steps: []TestStep{
		{
			ID:          "PM-002-01",
			Action:      "CreateSprint",
			Description: "Create new 2-week sprint",
			Request: map[string]interface{}{
				"action": "create",
				"object": "cycle",
				"data": map[string]interface{}{
					"name":       "Sprint 1 - Authentication",
					"start_date": time.Now().Unix(),
					"end_date":   time.Now().Add(14 * 24 * time.Hour).Unix(),
					"goal":       "Implement user authentication and registration",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.cycle.name contains 'Sprint 1'",
				},
			},
		},
		{
			ID:          "PM-002-02",
			Action:      "AddTicketsToSprint",
			Description: "Add estimated tickets to sprint",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":             "Implement login API",
					"type":              "story",
					"priority":          "high",
					"original_estimate": 8,
					"assignee":          "john_dev",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.ticket.original_estimate == 8",
				},
			},
		},
		{
			ID:          "PM-002-03",
			Action:      "StartSprint",
			Description: "Activate the sprint",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "cycle",
				"data": map[string]interface{}{
					"id":     "${sprint_id}",
					"status": "active",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-002-04",
			Action:      "WorkOnTicket",
			Description: "Move ticket to in progress and log time",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":         "${ticket_id}",
					"status":     "in_progress",
					"time_spent": 4,
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-002-05",
			Action:      "CompleteTicket",
			Description: "Mark ticket as done",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":         "${ticket_id}",
					"status":     "done",
					"resolution": "completed",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-002-06",
			Action:      "CompleteSprint",
			Description: "Close the sprint",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "cycle",
				"data": map[string]interface{}{
					"id":     "${sprint_id}",
					"status": "completed",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
	},
	Tags: []string{"sprint", "agile", "planning", "execution"},
}

// BugTriageTestCase tests bug reporting and resolution workflow
var BugTriageTestCase = TestCase{
	ID:          "PM-003",
	Name:        "Bug Triage and Resolution",
	Category:    "ProjectManagement",
	Priority:    "Critical",
	Description: "Test bug lifecycle: report → triage → assign → fix → verify → close",
	Steps: []TestStep{
		{
			ID:          "PM-003-01",
			Action:      "ReportBug",
			Description: "User reports a critical bug",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":       "Payment processing fails on checkout",
					"description": "Users cannot complete payment on checkout page. Error: 'Transaction timeout'",
					"type":        "bug",
					"severity":    "critical",
					"reporter":    "sarah_qa",
					"priority":    "highest",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.ticket.type == 'bug'",
					"response.data.ticket.priority == 'highest'",
				},
			},
		},
		{
			ID:          "PM-003-02",
			Action:      "TriageBug",
			Description: "PM triages and categorizes the bug",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":       "${bug_id}",
					"priority": "critical",
					"labels":   []string{"payment", "production", "urgent"},
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-003-03",
			Action:      "AssignBug",
			Description: "Assign bug to senior developer",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":       "${bug_id}",
					"assignee": "john_dev",
					"status":   "assigned",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-003-04",
			Action:      "InvestigateBug",
			Description: "Developer adds investigation comments",
			Request: map[string]interface{}{
				"action": "create",
				"object": "comment",
				"data": map[string]interface{}{
					"ticket_id": "${bug_id}",
					"text":      "Found the issue - payment gateway timeout set to 5 seconds, needs to be 30 seconds. Deploying fix now.",
					"author":    "john_dev",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-003-05",
			Action:      "FixBug",
			Description: "Mark bug as fixed and ready for QA",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":     "${bug_id}",
					"status": "ready_for_qa",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-003-06",
			Action:      "VerifyFix",
			Description: "QA verifies the fix",
			Request: map[string]interface{}{
				"action": "create",
				"object": "comment",
				"data": map[string]interface{}{
					"ticket_id": "${bug_id}",
					"text":      "Verified fix in production. Payments are processing successfully. Closing bug.",
					"author":    "sarah_qa",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-003-07",
			Action:      "CloseBug",
			Description: "Close bug as resolved",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":         "${bug_id}",
					"status":     "closed",
					"resolution": "fixed",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.ticket.resolution == 'fixed'",
				},
			},
		},
	},
	Tags: []string{"bug", "triage", "critical", "production"},
}

// FeatureDevelopmentTestCase tests feature from request to release
var FeatureDevelopmentTestCase = TestCase{
	ID:          "PM-004",
	Name:        "Feature Development Lifecycle",
	Category:    "ProjectManagement",
	Priority:    "High",
	Description: "Test complete feature development: request → break down → implement → test → release",
	Steps: []TestStep{
		{
			ID:          "PM-004-01",
			Action:      "CreateFeatureRequest",
			Description: "Product manager creates feature request",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":       "Add biometric authentication",
					"description": "Users want to login using fingerprint/face recognition for better security",
					"type":        "feature",
					"priority":    "medium",
					"reporter":    "mike_pm",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.ticket.type == 'feature'",
				},
			},
		},
		{
			ID:          "PM-004-02",
			Action:      "BreakDownIntoTasks",
			Description: "Break feature into implementation tasks",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":             "Research biometric APIs",
					"type":              "task",
					"parent_id":         "${feature_id}",
					"assignee":          "john_dev",
					"original_estimate": 4,
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-004-03",
			Action:      "ImplementFeature",
			Description: "Complete implementation tasks",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":     "${task_id}",
					"status": "done",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-004-04",
			Action:      "TestFeature",
			Description: "QA tests the new feature",
			Request: map[string]interface{}{
				"action": "create",
				"object": "comment",
				"data": map[string]interface{}{
					"ticket_id": "${feature_id}",
					"text":      "Tested on iOS and Android. Biometric auth working perfectly!",
					"author":    "sarah_qa",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-004-05",
			Action:      "CompleteFeature",
			Description: "Mark feature as complete",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":         "${feature_id}",
					"status":     "done",
					"resolution": "completed",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.ticket.resolution == 'completed'",
				},
			},
		},
	},
	Tags: []string{"feature", "development", "lifecycle"},
}

// ReleaseManagementTestCase tests version management and release
var ReleaseManagementTestCase = TestCase{
	ID:          "PM-005",
	Name:        "Release Management Workflow",
	Category:    "ProjectManagement",
	Priority:    "High",
	Description: "Test release management: create version → assign tickets → track progress → release",
	Steps: []TestStep{
		{
			ID:          "PM-005-01",
			Action:      "CreateVersion",
			Description: "Create new version for upcoming release",
			Request: map[string]interface{}{
				"action": "versionCreate",
				"data": map[string]interface{}{
					"name":         "v2.0.0",
					"description":  "Major release with new features",
					"release_date": time.Now().Add(30 * 24 * time.Hour).Unix(),
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.version.name == 'v2.0.0'",
				},
			},
		},
		{
			ID:          "PM-005-02",
			Action:      "AssignTicketsToVersion",
			Description: "Assign tickets to the release version",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":       "Performance improvements",
					"type":        "improvement",
					"fix_version": "${version_id}",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-005-03",
			Action:      "TrackProgress",
			Description: "Monitor release progress",
			Request: map[string]interface{}{
				"action": "versionRead",
				"data": map[string]interface{}{
					"id": "${version_id}",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-005-04",
			Action:      "ReleaseVersion",
			Description: "Release the version",
			Request: map[string]interface{}{
				"action": "versionRelease",
				"data": map[string]interface{}{
					"id":           "${version_id}",
					"release_date": time.Now().Unix(),
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
	},
	Tags: []string{"release", "version", "deployment"},
}

// TeamCollaborationTestCase tests collaboration features
var TeamCollaborationTestCase = TestCase{
	ID:          "PM-006",
	Name:        "Team Collaboration Workflow",
	Category:    "ProjectManagement",
	Priority:    "Medium",
	Description: "Test team collaboration: watchers, comments, mentions, notifications",
	Steps: []TestStep{
		{
			ID:          "PM-006-01",
			Action:      "CreateTaskForTeam",
			Description: "Create task requiring team collaboration",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":       "Design new onboarding flow",
					"description": "Redesign user onboarding for better conversion",
					"type":        "task",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-006-02",
			Action:      "AddWatchers",
			Description: "Add team members as watchers",
			Request: map[string]interface{}{
				"action": "watcherAdd",
				"data": map[string]interface{}{
					"ticket_id": "${ticket_id}",
					"username":  "mike_pm",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-006-03",
			Action:      "AddCollaborativeComments",
			Description: "Team members discuss the task",
			Request: map[string]interface{}{
				"action": "create",
				"object": "comment",
				"data": map[string]interface{}{
					"ticket_id": "${ticket_id}",
					"text":      "@lisa_designer Can you create mockups by tomorrow? @john_dev will implement after review.",
					"author":    "mike_pm",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-006-04",
			Action:      "ListWatchers",
			Description: "Get all watchers for the ticket",
			Request: map[string]interface{}{
				"action": "watcherList",
				"data": map[string]interface{}{
					"ticket_id": "${ticket_id}",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
					"len(response.data.watchers) > 0",
				},
			},
		},
	},
	Tags: []string{"collaboration", "watchers", "comments", "team"},
}

// CrossTeamDependenciesTestCase tests ticket linking and dependencies
var CrossTeamDependenciesTestCase = TestCase{
	ID:          "PM-007",
	Name:        "Cross-Team Dependencies",
	Category:    "ProjectManagement",
	Priority:    "High",
	Description: "Test cross-team coordination: dependencies, blockers, ticket linking",
	Steps: []TestStep{
		{
			ID:          "PM-007-01",
			Action:      "CreateBackendTask",
			Description: "Backend team creates API task",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":    "Create user profile API endpoint",
					"type":     "task",
					"assignee": "backend_dev",
					"team":     "Backend Team",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-007-02",
			Action:      "CreateFrontendTask",
			Description: "Frontend team creates UI task (blocked by backend)",
			Request: map[string]interface{}{
				"action": "create",
				"object": "ticket",
				"data": map[string]interface{}{
					"title":      "Implement user profile UI",
					"type":       "task",
					"assignee":   "frontend_dev",
					"team":       "Frontend Team",
					"blocked_by": "${backend_ticket_id}",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-007-03",
			Action:      "CompleteBackendTask",
			Description: "Backend team completes their task",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":     "${backend_ticket_id}",
					"status": "done",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-007-04",
			Action:      "UnblockFrontendTask",
			Description: "Frontend team can now work on their task",
			Request: map[string]interface{}{
				"action": "modify",
				"object": "ticket",
				"data": map[string]interface{}{
					"id":     "${frontend_ticket_id}",
					"status": "in_progress",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
	},
	Tags: []string{"dependencies", "cross-team", "coordination", "blocking"},
}

// FilterManagementTestCase tests saved filters and searches
var FilterManagementTestCase = TestCase{
	ID:          "PM-008",
	Name:        "Filter and Search Management",
	Category:    "ProjectManagement",
	Priority:    "Medium",
	Description: "Test filter management: create, save, share, use filters",
	Steps: []TestStep{
		{
			ID:          "PM-008-01",
			Action:      "CreatePersonalFilter",
			Description: "User creates personal filter for their tickets",
			Request: map[string]interface{}{
				"action": "filterSave",
				"data": map[string]interface{}{
					"name":        "My Open Tickets",
					"description": "All my open tickets across all projects",
					"jql":         "assignee = currentUser() AND status != closed",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
					"response.data.filter.name == 'My Open Tickets'",
				},
			},
		},
		{
			ID:          "PM-008-02",
			Action:      "CreateTeamFilter",
			Description: "Create filter for critical team bugs",
			Request: map[string]interface{}{
				"action": "filterSave",
				"data": map[string]interface{}{
					"name":        "Critical Bugs",
					"description": "All critical priority bugs",
					"jql":         "type = bug AND priority = critical AND status != closed",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 201,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-008-03",
			Action:      "ShareFilter",
			Description: "Share filter with team",
			Request: map[string]interface{}{
				"action": "filterShare",
				"data": map[string]interface{}{
					"id":         "${filter_id}",
					"share_type": "team",
					"share_with": "development_team",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
		{
			ID:          "PM-008-04",
			Action:      "UseFilter",
			Description: "Load and apply the filter",
			Request: map[string]interface{}{
				"action": "filterLoad",
				"data": map[string]interface{}{
					"id": "${filter_id}",
				},
			},
			ExpectedResult: ExpectedResult{
				StatusCode: 200,
				Assertions: []string{
					"response.errorCode == -1",
				},
			},
		},
	},
	Tags: []string{"filter", "search", "jql", "sharing"},
}

// TestCase represents a complete test scenario
type TestCase struct {
	ID          string
	Name        string
	Category    string
	Priority    string
	Description string
	Steps       []TestStep
	Tags        []string
}

// TestStep represents a single step in a test case
type TestStep struct {
	ID             string
	Action         string
	Description    string
	Request        map[string]interface{}
	ExpectedResult ExpectedResult
}

// ExpectedResult represents expected outcome of a test step
type ExpectedResult struct {
	StatusCode int
	Assertions []string
}

// GetAllPMWorkflowTestCases returns all PM workflow test cases
func GetAllPMWorkflowTestCases() []TestCase {
	return []TestCase{
		ProjectOnboardingTestCase,
		SprintPlanningTestCase,
		BugTriageTestCase,
		FeatureDevelopmentTestCase,
		ReleaseManagementTestCase,
		TeamCollaborationTestCase,
		CrossTeamDependenciesTestCase,
		FilterManagementTestCase,
	}
}
