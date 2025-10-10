package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"helixtrack.ru/core/qa-ai/testcases"
)

// QAAgent represents an AI-driven test agent
type QAAgent struct {
	Name         string
	Profile      string
	HTTPClient   *http.Client
	BaseURL      string
	JWT          string
	Variables    map[string]interface{}
	TestResults  []TestResult
	CurrentTest  *testcases.TestCase
}

// TestResult stores the result of a test execution
type TestResult struct {
	TestID       string
	TestName     string
	Status       string // "PASS", "FAIL", "SKIP", "ERROR"
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Steps        []StepResult
	Error        string
	DatabaseOK   bool
	Screenshot   string
}

// StepResult stores the result of a test step
type StepResult struct {
	StepID      string
	Description string
	Status      string
	Request     *http.Request
	Response    *http.Response
	ResponseBody string
	Error       string
	Duration    time.Duration
}

// NewQAAgent creates a new QA agent
func NewQAAgent(name, profile, baseURL string) *QAAgent {
	return &QAAgent{
		Name:    name,
		Profile: profile,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL:     baseURL,
		Variables:   make(map[string]interface{}),
		TestResults: make([]TestResult, 0),
	}
}

// ExecuteTestCase executes a single test case
func (agent *QAAgent) ExecuteTestCase(ctx context.Context, testCase testcases.TestCase) TestResult {
	result := TestResult{
		TestID:    testCase.ID,
		TestName:  testCase.Name,
		StartTime: time.Now(),
		Steps:     make([]StepResult, 0),
	}

	agent.CurrentTest = &testCase

	// Check prerequisites
	if !agent.checkPrerequisites(testCase.Prerequisites) {
		result.Status = "SKIP"
		result.Error = "Prerequisites not met"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// Execute test steps
	for _, step := range testCase.Steps {
		stepResult := agent.executeStep(ctx, step)
		result.Steps = append(result.Steps, stepResult)

		if stepResult.Status == "FAIL" || stepResult.Status == "ERROR" {
			result.Status = stepResult.Status
			result.Error = stepResult.Error
			break
		}
	}

	// If all steps passed, check database
	if result.Status == "" {
		if agent.verifyDatabase(testCase.DatabaseChecks) {
			result.Status = "PASS"
			result.DatabaseOK = true
		} else {
			result.Status = "FAIL"
			result.Error = "Database verification failed"
			result.DatabaseOK = false
		}
	}

	// Execute cleanup steps
	for _, step := range testCase.CleanupSteps {
		agent.executeStep(ctx, step)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	agent.TestResults = append(agent.TestResults, result)

	return result
}

// executeStep executes a single test step
func (agent *QAAgent) executeStep(ctx context.Context, step testcases.TestStep) StepResult {
	result := StepResult{
		StepID:      step.ID,
		Description: step.Description,
		Status:      "PASS",
	}

	startTime := time.Now()

	switch step.Action {
	case "http_request":
		result = agent.executeHTTPRequest(ctx, step)
	case "wait":
		time.Sleep(2 * time.Second)
	case "set_variable":
		// Set a variable for later use
		for key, value := range step.Payload.(map[string]interface{}) {
			agent.Variables[key] = value
		}
	default:
		result.Status = "ERROR"
		result.Error = fmt.Sprintf("Unknown action: %s", step.Action)
	}

	result.Duration = time.Since(startTime)

	return result
}

// executeHTTPRequest executes an HTTP request step
func (agent *QAAgent) executeHTTPRequest(ctx context.Context, step testcases.TestStep) StepResult {
	result := StepResult{
		StepID:      step.ID,
		Description: step.Description,
	}

	// Replace variables in payload
	payload := agent.replaceVariables(step.Payload)

	// Marshal payload to JSON
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			result.Status = "ERROR"
			result.Error = fmt.Sprintf("Failed to marshal payload: %v", err)
			return result
		}
	}

	// Create request
	url := agent.BaseURL + step.Endpoint
	req, err := http.NewRequestWithContext(ctx, step.Method, url, bytes.NewBuffer(body))
	if err != nil {
		result.Status = "ERROR"
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		return result
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range step.Headers {
		value = agent.replaceVariablesInString(value)
		req.Header.Set(key, value)
	}

	// Add JWT if available
	if agent.JWT != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+agent.JWT)
	}

	result.Request = req

	// Execute request
	resp, err := agent.HTTPClient.Do(req)
	if err != nil {
		result.Status = "ERROR"
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Response = resp

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = "ERROR"
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		return result
	}
	result.ResponseBody = string(bodyBytes)

	// Save response if requested
	if step.SaveResponse != "" {
		var responseData interface{}
		json.Unmarshal(bodyBytes, &responseData)
		agent.Variables[step.SaveResponse] = responseData

		// Extract JWT token from login response
		if step.SaveResponse == "login_response" {
			if respMap, ok := responseData.(map[string]interface{}); ok {
				if data, ok := respMap["data"].(map[string]interface{}); ok {
					if token, ok := data["token"].(string); ok {
						agent.JWT = token
					}
				}
			}
		}

		// Extract project ID from project creation response
		if step.SaveResponse == "project_response" {
			if respMap, ok := responseData.(map[string]interface{}); ok {
				if data, ok := respMap["data"].(map[string]interface{}); ok {
					if project, ok := data["project"].(map[string]interface{}); ok {
						if projectID, ok := project["id"].(string); ok {
							agent.Variables["project_id"] = projectID
						}
					}
				}
			}
		}

		// Extract ticket ID from ticket creation response
		if step.SaveResponse == "ticket_response" {
			if respMap, ok := responseData.(map[string]interface{}); ok {
				if data, ok := respMap["data"].(map[string]interface{}); ok {
					if ticket, ok := data["ticket"].(map[string]interface{}); ok {
						if ticketID, ok := ticket["id"].(string); ok {
							agent.Variables["ticket_id"] = ticketID
						}
					}
				}
			}
		}
	}

	// Verify expected results
	if !agent.verifyExpectedResult(step.Expected, resp, string(bodyBytes)) {
		result.Status = "FAIL"
		result.Error = fmt.Sprintf("Expected result not met. Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
		return result
	}

	result.Status = "PASS"
	return result
}

// verifyExpectedResult verifies if the response matches expected result
func (agent *QAAgent) verifyExpectedResult(expected testcases.ExpectedResult, resp *http.Response, body string) bool {
	// Check status code
	if expected.StatusCode != 0 && resp.StatusCode != expected.StatusCode {
		return false
	}

	// Check body contains
	for _, contains := range expected.BodyContains {
		if !strings.Contains(body, contains) {
			return false
		}
	}

	// Check body not contains
	for _, notContains := range expected.BodyNotContains {
		if strings.Contains(body, notContains) {
			return false
		}
	}

	// Check headers
	for key, value := range expected.HeadersContain {
		if resp.Header.Get(key) != value {
			return false
		}
	}

	// TODO: Implement JSONPath verification
	// TODO: Implement response time check

	return true
}

// replaceVariables replaces variable placeholders in payload
func (agent *QAAgent) replaceVariables(payload interface{}) interface{} {
	if payload == nil {
		return nil
	}

	switch v := payload.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = agent.replaceVariables(value)
		}
		return result
	case string:
		return agent.replaceVariablesInString(v)
	default:
		return payload
	}
}

// replaceVariablesInString replaces variable placeholders in a string
func (agent *QAAgent) replaceVariablesInString(s string) string {
	// Replace {{timestamp}}
	s = strings.ReplaceAll(s, "{{timestamp}}", fmt.Sprintf("%d", time.Now().Unix()))

	// Replace {{jwt_token}}
	if agent.JWT != "" {
		s = strings.ReplaceAll(s, "{{jwt_token}}", agent.JWT)
	}

	// Replace other variables
	for key, value := range agent.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		s = strings.ReplaceAll(s, placeholder, replacement)
	}

	return s
}

// checkPrerequisites checks if test prerequisites are met
func (agent *QAAgent) checkPrerequisites(prerequisites []string) bool {
	// Check if all prerequisite tests have passed
	for _, prereqID := range prerequisites {
		found := false
		for _, result := range agent.TestResults {
			if result.TestID == prereqID && result.Status == "PASS" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// verifyDatabase verifies database checks
func (agent *QAAgent) verifyDatabase(checks []testcases.DatabaseCheck) bool {
	// TODO: Implement actual database verification
	// For now, assume all checks pass
	return true
}

// GetTestSummary returns a summary of all test results
func (agent *QAAgent) GetTestSummary() TestSummary {
	summary := TestSummary{
		TotalTests: len(agent.TestResults),
		AgentName:  agent.Name,
	}

	for _, result := range agent.TestResults {
		switch result.Status {
		case "PASS":
			summary.Passed++
		case "FAIL":
			summary.Failed++
		case "SKIP":
			summary.Skipped++
		case "ERROR":
			summary.Errors++
		}
		summary.TotalDuration += result.Duration
	}

	summary.SuccessRate = float64(summary.Passed) / float64(summary.TotalTests) * 100

	return summary
}

// TestSummary provides a summary of test results
type TestSummary struct {
	AgentName     string
	TotalTests    int
	Passed        int
	Failed        int
	Skipped       int
	Errors        int
	SuccessRate   float64
	TotalDuration time.Duration
}
