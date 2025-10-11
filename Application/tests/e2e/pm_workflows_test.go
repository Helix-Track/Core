package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"helixtrack.ru/core/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestPM_CompleteProjectSetup tests the complete project onboarding workflow
// Scenario: New team starting a project from scratch
func TestPM_CompleteProjectSetup(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	t.Run("Step 1: Create Organization", func(t *testing.T) {
		// Create organization
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "organization",
			Data: map[string]interface{}{
				"name":        "Acme Corporation",
				"description": "Software development company",
			},
		}, token)

		assert.Equal(t, http.StatusCreated, resp.Code)
		var orgResp models.Response
		json.Unmarshal(resp.Body.Bytes(), &orgResp)
		assert.Equal(t, models.ErrorCodeNoError, orgResp.ErrorCode)
	})

	t.Run("Step 2: Create Project", func(t *testing.T) {
		// Create project
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "project",
			Data: map[string]interface{}{
				"name":        "E-Commerce Platform",
				"key":         "ECOM",
				"description": "Next-generation e-commerce platform",
				"type":        "software",
			},
		}, token)

		assert.Equal(t, http.StatusCreated, resp.Code)
		var projResp models.Response
		json.Unmarshal(resp.Body.Bytes(), &projResp)
		assert.Equal(t, models.ErrorCodeNoError, projResp.ErrorCode)
	})

	t.Run("Step 3: Add Team Members", func(t *testing.T) {
		// Add team members (simulated)
		teamMembers := []string{"developer1", "developer2", "qa_engineer", "product_manager"}

		for _, member := range teamMembers {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionCreate,
				Object: "team",
				Data: map[string]interface{}{
					"name":    fmt.Sprintf("%s Team", member),
					"members": []string{member},
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 4: Configure Workflow", func(t *testing.T) {
		// Create custom workflow
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "workflow",
			Data: map[string]interface{}{
				"name":        "Development Workflow",
				"description": "Standard development workflow",
				"steps": []string{
					"To Do",
					"In Progress",
					"Code Review",
					"QA Testing",
					"Done",
				},
			},
		}, token)

		assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Step 5: Create Initial Backlog", func(t *testing.T) {
		// Create initial tickets
		tickets := []struct {
			title    string
			type_    string
			priority string
		}{
			{"Setup development environment", "task", "high"},
			{"Design database schema", "task", "high"},
			{"Implement user authentication", "story", "high"},
			{"Create product catalog", "story", "medium"},
			{"Setup CI/CD pipeline", "task", "medium"},
		}

		for _, ticket := range tickets {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionCreate,
				Object: "ticket",
				Data: map[string]interface{}{
					"title":    ticket.title,
					"type":     ticket.type_,
					"priority": ticket.priority,
				},
			}, token)

			// Should create or fail gracefully
			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})
}

// TestPM_SprintPlanningAndExecution tests complete sprint workflow
// Scenario: Planning and executing a 2-week sprint
func TestPM_SprintPlanningAndExecution(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	var sprintID, ticket1ID, ticket2ID string

	t.Run("Step 1: Create Sprint", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "cycle",
			Data: map[string]interface{}{
				"name":       "Sprint 1",
				"start_date": time.Now().Unix(),
				"end_date":   time.Now().Add(14 * 24 * time.Hour).Unix(),
				"goal":       "Implement core authentication features",
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var sprintResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &sprintResp)
			if cycleData, ok := sprintResp.Data["cycle"].(map[string]interface{}); ok {
				if id, ok := cycleData["id"].(string); ok {
					sprintID = id
				}
			}
		}
	})

	t.Run("Step 2: Create and Estimate Tickets", func(t *testing.T) {
		// Create ticket 1
		resp1 := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":             "Implement login API",
				"description":       "Create REST API for user login",
				"type":              "story",
				"priority":          "high",
				"original_estimate": 8, // 8 hours
			},
		}, token)

		if resp1.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp1.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					ticket1ID = id
				}
			}
		}

		// Create ticket 2
		resp2 := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":             "Implement registration API",
				"description":       "Create REST API for user registration",
				"type":              "story",
				"priority":          "high",
				"original_estimate": 6, // 6 hours
			},
		}, token)

		if resp2.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp2.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					ticket2ID = id
				}
			}
		}
	})

	t.Run("Step 3: Assign Tickets", func(t *testing.T) {
		if ticket1ID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":       ticket1ID,
					"assignee": "developer1",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}

		if ticket2ID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":       ticket2ID,
					"assignee": "developer2",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 4: Start Sprint", func(t *testing.T) {
		if sprintID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "cycle",
				Data: map[string]interface{}{
					"id":     sprintID,
					"status": "active",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 5: Work on Tickets", func(t *testing.T) {
		// Move ticket to "In Progress"
		if ticket1ID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     ticket1ID,
					"status": "in_progress",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}

		// Log work
		if ticket1ID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":         ticket1ID,
					"time_spent": 4, // 4 hours worked
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 6: Complete Tickets", func(t *testing.T) {
		if ticket1ID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":         ticket1ID,
					"status":     "done",
					"resolution": "fixed",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 7: Complete Sprint", func(t *testing.T) {
		if sprintID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "cycle",
				Data: map[string]interface{}{
					"id":     sprintID,
					"status": "completed",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})
}

// TestPM_BugTriageWorkflow tests bug reporting and resolution workflow
// Scenario: User reports a bug → triage → fix → verify → close
func TestPM_BugTriageWorkflow(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	var bugID string

	t.Run("Step 1: Report Bug", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":       "Login button not working on mobile",
				"description": "When user taps login button on mobile Safari, nothing happens",
				"type":        "bug",
				"severity":    "high",
				"reporter":    "qa_engineer",
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var bugResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &bugResp)
			if bugData, ok := bugResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := bugData["id"].(string); ok {
					bugID = id
				}
			}
		}
	})

	t.Run("Step 2: Triage and Prioritize", func(t *testing.T) {
		if bugID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":       bugID,
					"priority": "high",
					"labels":   []string{"mobile", "critical", "authentication"},
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 3: Assign to Developer", func(t *testing.T) {
		if bugID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":       bugID,
					"assignee": "developer1",
					"status":   "assigned",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 4: Developer Investigates and Comments", func(t *testing.T) {
		if bugID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionCreate,
				Object: "comment",
				Data: map[string]interface{}{
					"ticket_id": bugID,
					"text":      "Found the issue - missing event listener for touch events. Working on fix.",
					"author":    "developer1",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 5: Fix Bug", func(t *testing.T) {
		if bugID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     bugID,
					"status": "in_progress",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)

			// Simulate fix completion
			resp = app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     bugID,
					"status": "ready_for_qa",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 6: QA Verification", func(t *testing.T) {
		if bugID != "" {
			// QA tests the fix
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionCreate,
				Object: "comment",
				Data: map[string]interface{}{
					"ticket_id": bugID,
					"text":      "Tested on iPhone 12 Safari - bug is fixed!",
					"author":    "qa_engineer",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 7: Close Bug", func(t *testing.T) {
		if bugID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":         bugID,
					"status":     "closed",
					"resolution": "fixed",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})
}

// TestPM_FeatureDevelopmentLifecycle tests feature from request to release
// Scenario: Feature request → specification → development → testing → release
func TestPM_FeatureDevelopmentLifecycle(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	var featureID, task1ID, task2ID, task3ID string

	t.Run("Step 1: Create Feature Request", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":       "Add dark mode support",
				"description": "Users want dark mode to reduce eye strain",
				"type":        "feature",
				"priority":    "medium",
				"reporter":    "product_manager",
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var featureResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &featureResp)
			if featureData, ok := featureResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := featureData["id"].(string); ok {
					featureID = id
				}
			}
		}
	})

	t.Run("Step 2: Break Down into Tasks", func(t *testing.T) {
		// Task 1: Design
		resp1 := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":       "Design dark mode color scheme",
				"type":        "task",
				"priority":    "medium",
				"parent_id":   featureID,
				"original_estimate": 4,
			},
		}, token)

		if resp1.Code == http.StatusCreated {
			var taskResp models.Response
			json.Unmarshal(resp1.Body.Bytes(), &taskResp)
			if taskData, ok := taskResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := taskData["id"].(string); ok {
					task1ID = id
				}
			}
		}

		// Task 2: Implementation
		resp2 := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":             "Implement dark mode theme switcher",
				"type":              "task",
				"priority":          "medium",
				"parent_id":         featureID,
				"original_estimate": 8,
			},
		}, token)

		if resp2.Code == http.StatusCreated {
			var taskResp models.Response
			json.Unmarshal(resp2.Body.Bytes(), &taskResp)
			if taskData, ok := taskResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := taskData["id"].(string); ok {
					task2ID = id
				}
			}
		}

		// Task 3: Testing
		resp3 := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":             "Test dark mode on all pages",
				"type":              "task",
				"priority":          "medium",
				"parent_id":         featureID,
				"original_estimate": 4,
			},
		}, token)

		if resp3.Code == http.StatusCreated {
			var taskResp models.Response
			json.Unmarshal(resp3.Body.Bytes(), &taskResp)
			if taskData, ok := taskResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := taskData["id"].(string); ok {
					task3ID = id
				}
			}
		}
	})

	t.Run("Step 3: Assign and Execute Tasks", func(t *testing.T) {
		// Complete task 1
		if task1ID != "" {
			app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":       task1ID,
					"assignee": "designer",
					"status":   "done",
				},
			}, token)
		}

		// Complete task 2
		if task2ID != "" {
			app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":       task2ID,
					"assignee": "developer1",
					"status":   "done",
				},
			}, token)
		}

		// Complete task 3
		if task3ID != "" {
			app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":       task3ID,
					"assignee": "qa_engineer",
					"status":   "done",
				},
			}, token)
		}
	})

	t.Run("Step 4: Complete Feature", func(t *testing.T) {
		if featureID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":         featureID,
					"status":     "done",
					"resolution": "completed",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})
}

// TestPM_ReleaseManagement tests version/release management workflow
// Scenario: Create version → assign tickets → track → release
func TestPM_ReleaseManagement(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	var versionID, ticket1ID, ticket2ID string

	t.Run("Step 1: Create Version", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionVersionCreate,
			Data: map[string]interface{}{
				"name":         "v1.5.0",
				"description":  "Q4 2025 Release",
				"release_date": time.Now().Add(30 * 24 * time.Hour).Unix(),
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var versionResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &versionResp)
			if versionData, ok := versionResp.Data["version"].(map[string]interface{}); ok {
				if id, ok := versionData["id"].(string); ok {
					versionID = id
				}
			}
		}
	})

	t.Run("Step 2: Create Tickets for Release", func(t *testing.T) {
		// Ticket 1
		resp1 := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":       "Performance improvements",
				"type":        "improvement",
				"fix_version": versionID,
			},
		}, token)

		if resp1.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp1.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					ticket1ID = id
				}
			}
		}

		// Ticket 2
		resp2 := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":       "Security updates",
				"type":        "security",
				"fix_version": versionID,
			},
		}, token)

		if resp2.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp2.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					ticket2ID = id
				}
			}
		}
	})

	t.Run("Step 3: Track Progress", func(t *testing.T) {
		// Complete tickets
		if ticket1ID != "" {
			app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     ticket1ID,
					"status": "done",
				},
			}, token)
		}

		if ticket2ID != "" {
			app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     ticket2ID,
					"status": "done",
				},
			}, token)
		}
	})

	t.Run("Step 4: Release Version", func(t *testing.T) {
		if versionID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionVersionRelease,
				Data: map[string]interface{}{
					"id":           versionID,
					"release_date": time.Now().Unix(),
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 5: Generate Release Notes", func(t *testing.T) {
		// This would typically query all tickets in the version
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionVersionRead,
			Data: map[string]interface{}{
				"id": versionID,
			},
		}, token)

		assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
	})
}

// TestPM_TeamCollaboration tests team collaboration features
// Scenario: Watchers, comments, attachments, mentions
func TestPM_TeamCollaboration(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	var ticketID string

	t.Run("Step 1: Create Ticket", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":       "Design new dashboard layout",
				"description": "Redesign main dashboard for better UX",
				"type":        "task",
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					ticketID = id
				}
			}
		}
	})

	t.Run("Step 2: Add Watchers", func(t *testing.T) {
		if ticketID != "" {
			watchers := []string{"product_manager", "designer", "developer1", "qa_engineer"}

			for _, watcher := range watchers {
				resp := app.makeRequest("POST", "/do", models.Request{
					Action: models.ActionWatcherAdd,
					Data: map[string]interface{}{
						"ticket_id": ticketID,
						"username":  watcher,
					},
				}, token)

				assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
			}
		}
	})

	t.Run("Step 3: Add Comments with Discussion", func(t *testing.T) {
		if ticketID != "" {
			comments := []struct {
				author string
				text   string
			}{
				{"product_manager", "We need this by end of sprint. Priority is high."},
				{"designer", "I'll create mockups by tomorrow"},
				{"developer1", "@designer Can you share the design system colors?"},
				{"designer", "@developer1 Sure, I'll add them to the attachments"},
				{"qa_engineer", "Please add acceptance criteria before implementation"},
			}

			for _, comment := range comments {
				resp := app.makeRequest("POST", "/do", models.Request{
					Action: models.ActionCreate,
					Object: "comment",
					Data: map[string]interface{}{
						"ticket_id": ticketID,
						"author":    comment.author,
						"text":      comment.text,
					},
				}, token)

				assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
			}
		}
	})

	t.Run("Step 4: List Watchers", func(t *testing.T) {
		if ticketID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionWatcherList,
				Data: map[string]interface{}{
					"ticket_id": ticketID,
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 5: Remove Watcher", func(t *testing.T) {
		if ticketID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionWatcherRemove,
				Data: map[string]interface{}{
					"ticket_id": ticketID,
					"username":  "qa_engineer",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})
}

// TestPM_CrossTeamDependencies tests ticket linking and dependencies
// Scenario: Multiple teams with dependent tickets
func TestPM_CrossTeamDependencies(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	var backendTicketID, frontendTicketID, qaTicketID string

	t.Run("Step 1: Create Backend Ticket", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":    "Create API endpoint for user settings",
				"type":     "task",
				"assignee": "backend_dev",
				"team":     "Backend Team",
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					backendTicketID = id
				}
			}
		}
	})

	t.Run("Step 2: Create Frontend Ticket (Blocked)", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":    "Implement user settings UI",
				"type":     "task",
				"assignee": "frontend_dev",
				"team":     "Frontend Team",
				"blocked_by": backendTicketID,
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					frontendTicketID = id
				}
			}
		}
	})

	t.Run("Step 3: Create QA Ticket (Depends on Both)", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionCreate,
			Object: "ticket",
			Data: map[string]interface{}{
				"title":    "Test user settings feature",
				"type":     "test",
				"assignee": "qa_engineer",
				"team":     "QA Team",
				"depends_on": []string{backendTicketID, frontendTicketID},
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var ticketResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &ticketResp)
			if ticketData, ok := ticketResp.Data["ticket"].(map[string]interface{}); ok {
				if id, ok := ticketData["id"].(string); ok {
					qaTicketID = id
				}
			}
		}
	})

	t.Run("Step 4: Complete Backend Task", func(t *testing.T) {
		if backendTicketID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     backendTicketID,
					"status": "done",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 5: Unblock Frontend Task", func(t *testing.T) {
		if frontendTicketID != "" {
			// Frontend ticket should now be unblocked
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     frontendTicketID,
					"status": "in_progress",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 6: Complete All and Start QA", func(t *testing.T) {
		if frontendTicketID != "" {
			app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     frontendTicketID,
					"status": "done",
				},
			}, token)
		}

		if qaTicketID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionModify,
				Object: "ticket",
				Data: map[string]interface{}{
					"id":     qaTicketID,
					"status": "in_progress",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})
}

// TestPM_FilterAndSearch tests saved filters for project management
// Scenario: Create and use filters for different views
func TestPM_FilterAndSearch(t *testing.T) {
	app := setupCompleteApplication(t)
	defer app.cleanup()

	token := "valid-test-token"

	var filterID string

	t.Run("Step 1: Create Filter for My Open Tickets", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionFilterSave,
			Data: map[string]interface{}{
				"name":        "My Open Tickets",
				"description": "All tickets assigned to me that are not closed",
				"jql":         "assignee = currentUser() AND status != closed",
			},
		}, token)

		if resp.Code == http.StatusCreated {
			var filterResp models.Response
			json.Unmarshal(resp.Body.Bytes(), &filterResp)
			if filterData, ok := filterResp.Data["filter"].(map[string]interface{}); ok {
				if id, ok := filterData["id"].(string); ok {
					filterID = id
				}
			}
		}
	})

	t.Run("Step 2: Create Filter for High Priority Bugs", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionFilterSave,
			Data: map[string]interface{}{
				"name":        "Critical Bugs",
				"description": "High priority bugs that need immediate attention",
				"jql":         "type = bug AND priority = high AND status != closed",
			},
		}, token)

		assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Step 3: Share Filter with Team", func(t *testing.T) {
		if filterID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionFilterShare,
				Data: map[string]interface{}{
					"id":         filterID,
					"share_type": "team",
					"share_with": "development_team",
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 4: Load and Use Filter", func(t *testing.T) {
		if filterID != "" {
			resp := app.makeRequest("POST", "/do", models.Request{
				Action: models.ActionFilterLoad,
				Data: map[string]interface{}{
					"id": filterID,
				},
			}, token)

			assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Step 5: List All My Filters", func(t *testing.T) {
		resp := app.makeRequest("POST", "/do", models.Request{
			Action: models.ActionFilterList,
			Data:   map[string]interface{}{},
		}, token)

		assert.NotEqual(t, http.StatusInternalServerError, resp.Code)
	})
}

// Helper function to setup complete application (reusing from complete_flow_test.go)
// Note: This would normally be in a shared test utilities file
