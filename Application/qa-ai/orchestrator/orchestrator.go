package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"helixtrack.ru/core/qa-ai/agents"
	"helixtrack.ru/core/qa-ai/config"
	"helixtrack.ru/core/qa-ai/testcases"
)

// Orchestrator manages and coordinates QA test execution
type Orchestrator struct {
	Config      config.QAConfig
	TestCases   []testcases.TestCase
	Agents      []*agents.QAAgent
	Results     map[string]agents.TestResult
	mu          sync.Mutex
	StartTime   time.Time
	EndTime     time.Time
}

// NewOrchestrator creates a new test orchestrator
func NewOrchestrator(cfg config.QAConfig) *Orchestrator {
	return &Orchestrator{
		Config:    cfg,
		TestCases: testcases.GetAllTestCases(),
		Agents:    make([]*agents.QAAgent, 0),
		Results:   make(map[string]agents.TestResult),
	}
}

// Initialize prepares the orchestrator for testing
func (o *Orchestrator) Initialize() error {
	log.Println("Initializing QA Orchestrator...")

	// Create test agents for each profile
	profiles := config.GetTestProfiles()
	for _, profile := range profiles {
		agent := agents.NewQAAgent(
			fmt.Sprintf("Agent-%s", profile.Username),
			profile.Role,
			o.Config.ServerURL,
		)
		o.Agents = append(o.Agents, agent)
		log.Printf("Created agent: %s (Profile: %s)", agent.Name, profile.Role)
	}

	// Login all agents and share JWT tokens
	o.loginAllAgents()

	// TODO: Reset database if configured
	// TODO: Start server if configured
	// TODO: Wait for server to be ready

	log.Printf("Initialized with %d test cases and %d agents", len(o.TestCases), len(o.Agents))

	return nil
}

// loginAllAgents logs in all agents and shares JWT tokens
func (o *Orchestrator) loginAllAgents() {
	profiles := config.GetTestProfiles()
	for i, agent := range o.Agents {
		if i < len(profiles) {
			profile := profiles[i]
			// Login and get JWT
			testCase := testcases.TestCase{
				ID: "LOGIN-" + profile.Username,
				Steps: []testcases.TestStep{
					{
						ID:       "LOGIN",
						Action:   "http_request",
						Method:   "POST",
						Endpoint: "/do",
						Payload: map[string]interface{}{
							"action": "authenticate",
							"data": map[string]interface{}{
								"username": profile.Username,
								"password": profile.Password,
							},
						},
						SaveResponse: "login_response",
					},
				},
			}

			result := agent.ExecuteTestCase(context.Background(), testCase)
			if result.Status == "PASS" && agent.JWT != "" {
				// Share JWT token with all other agents using a common variable
				// e.g., viewer_jwt_token, admin_jwt_token, etc.
				tokenVarName := profile.Username + "_jwt_token"
				for _, otherAgent := range o.Agents {
					otherAgent.Variables[tokenVarName] = agent.JWT
				}
				log.Printf("Agent %s logged in successfully, JWT shared as {{%s}}", agent.Name, tokenVarName)
			}
		}
	}
}

// RunAllTests executes all test cases
func (o *Orchestrator) RunAllTests(ctx context.Context) error {
	o.StartTime = time.Now()
	defer func() {
		o.EndTime = time.Now()
	}()

	log.Printf("Starting QA test execution: %d test cases", len(o.TestCases))

	// Group tests by suite
	suites := make(map[string][]testcases.TestCase)
	for _, testCase := range o.TestCases {
		suites[testCase.Suite] = append(suites[testCase.Suite], testCase)
	}

	// Execute each suite
	for suiteName, suiteTests := range suites {
		log.Printf("\n========== Executing Suite: %s ==========", suiteName)
		log.Printf("Test cases in suite: %d", len(suiteTests))

		for _, testCase := range suiteTests {
			if err := o.executeTestCase(ctx, testCase); err != nil {
				if o.Config.StopOnFirstFail {
					return fmt.Errorf("test failed and StopOnFirstFail is enabled: %w", err)
				}
				log.Printf("Test case %s failed: %v", testCase.ID, err)
			}
		}
	}

	log.Printf("\n========== Test Execution Complete ==========")
	o.PrintSummary()

	return nil
}

// RunTestSuite executes a specific test suite
func (o *Orchestrator) RunTestSuite(ctx context.Context, suiteName string) error {
	log.Printf("Running test suite: %s", suiteName)

	for _, testCase := range o.TestCases {
		if testCase.Suite == suiteName {
			if err := o.executeTestCase(ctx, testCase); err != nil {
				if o.Config.StopOnFirstFail {
					return err
				}
			}
		}
	}

	return nil
}

// executeTestCase executes a single test case
func (o *Orchestrator) executeTestCase(ctx context.Context, testCase testcases.TestCase) error {
	// Select appropriate agent based on test requirements
	agent := o.selectAgent(testCase)
	if agent == nil {
		return fmt.Errorf("no suitable agent found for test case %s", testCase.ID)
	}

	log.Printf("\n--- Test: %s ---", testCase.Name)
	log.Printf("Description: %s", testCase.Description)
	log.Printf("Agent: %s", agent.Name)

	// Execute test with timeout
	testCtx, cancel := context.WithTimeout(ctx, testCase.Timeout)
	defer cancel()

	result := agent.ExecuteTestCase(testCtx, testCase)

	// Store result
	o.mu.Lock()
	o.Results[testCase.ID] = result
	o.mu.Unlock()

	// Log result
	status := "✓"
	if result.Status != "PASS" {
		status = "✗"
	}
	log.Printf("%s %s (%s) - Duration: %v", status, testCase.Name, result.Status, result.Duration)

	if result.Error != "" {
		log.Printf("   Error: %s", result.Error)
	}

	// Retry if configured and test failed
	if result.Status == "FAIL" && o.Config.RetryFailedTests {
		for retry := 1; retry <= o.Config.MaxRetries; retry++ {
			log.Printf("   Retrying... (Attempt %d/%d)", retry, o.Config.MaxRetries)
			time.Sleep(time.Second * time.Duration(retry))

			result = agent.ExecuteTestCase(ctx, testCase)
			o.mu.Lock()
			o.Results[testCase.ID] = result
			o.mu.Unlock()

			if result.Status == "PASS" {
				log.Printf("   ✓ Retry successful")
				break
			}
		}
	}

	return nil
}

// selectAgent selects an appropriate agent for a test case
func (o *Orchestrator) selectAgent(testCase testcases.TestCase) *agents.QAAgent {
	// For now, return the first agent (admin)
	// TODO: Implement smarter agent selection based on test requirements
	if len(o.Agents) > 0 {
		return o.Agents[0]
	}
	return nil
}

// PrintSummary prints a summary of test execution
func (o *Orchestrator) PrintSummary() {
	totalTests := len(o.Results)
	passed := 0
	failed := 0
	skipped := 0
	errors := 0

	for _, result := range o.Results {
		switch result.Status {
		case "PASS":
			passed++
		case "FAIL":
			failed++
		case "SKIP":
			skipped++
		case "ERROR":
			errors++
		}
	}

	duration := o.EndTime.Sub(o.StartTime)
	successRate := float64(passed) / float64(totalTests) * 100

	fmt.Printf("\n")
	fmt.Printf("============================================\n")
	fmt.Printf("         QA TEST EXECUTION SUMMARY         \n")
	fmt.Printf("============================================\n")
	fmt.Printf("Total Tests:     %d\n", totalTests)
	fmt.Printf("Passed:          %d (%.1f%%)\n", passed, successRate)
	fmt.Printf("Failed:          %d\n", failed)
	fmt.Printf("Skipped:         %d\n", skipped)
	fmt.Printf("Errors:          %d\n", errors)
	fmt.Printf("Duration:        %v\n", duration)
	fmt.Printf("Success Rate:    %.2f%%\n", successRate)
	fmt.Printf("============================================\n")

	if failed > 0 || errors > 0 {
		fmt.Printf("\nFailed Tests:\n")
		for _, result := range o.Results {
			if result.Status == "FAIL" || result.Status == "ERROR" {
				fmt.Printf("  - %s (%s): %s\n", result.TestName, result.TestID, result.Error)
			}
		}
	}
}

// GetResults returns all test results
func (o *Orchestrator) GetResults() map[string]agents.TestResult {
	o.mu.Lock()
	defer o.mu.Unlock()

	results := make(map[string]agents.TestResult)
	for k, v := range o.Results {
		results[k] = v
	}

	return results
}

// GetSuccessRate returns the overall success rate
func (o *Orchestrator) GetSuccessRate() float64 {
	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.Results) == 0 {
		return 0.0
	}

	passed := 0
	for _, result := range o.Results {
		if result.Status == "PASS" {
			passed++
		}
	}

	return float64(passed) / float64(len(o.Results)) * 100
}
