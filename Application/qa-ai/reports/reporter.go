package reports

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"helixtrack.ru/core/qa-ai/agents"
	"helixtrack.ru/core/qa-ai/orchestrator"
)

// Reporter handles test report generation
type Reporter struct {
	OutputPath string
}

// NewReporter creates a new reporter
func NewReporter(outputPath string) *Reporter {
	return &Reporter{
		OutputPath: outputPath,
	}
}

// GenerateHTMLReport generates an HTML test report
func (r *Reporter) GenerateHTMLReport(orch *orchestrator.Orchestrator) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(r.OutputPath, fmt.Sprintf("qa-report-%s.html", timestamp))

	results := orch.GetResults()
	successRate := orch.GetSuccessRate()

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>HelixTrack QA Report - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 30px 0; }
        .stat { background: #f8f9fa; padding: 20px; border-radius: 4px; text-align: center; border-left: 4px solid #4CAF50; }
        .stat.failed { border-left-color: #f44336; }
        .stat h3 { margin: 0; font-size: 32px; color: #333; }
        .stat p { margin: 5px 0 0 0; color: #666; }
        table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
        th { background: #4CAF50; color: white; padding: 12px; text-align: left; }
        td { padding: 12px; border-bottom: 1px solid #ddd; }
        tr:hover { background: #f5f5f5; }
        .pass { color: #4CAF50; font-weight: bold; }
        .fail { color: #f44336; font-weight: bold; }
        .skip { color: #ff9800; font-weight: bold; }
        .error { color: #9c27b0; font-weight: bold; }
        .success-rate { font-size: 48px; color: %s; font-weight: bold; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>HelixTrack QA-AI Test Report</h1>
        <p><strong>Generated:</strong> %s</p>
        <p><strong>Duration:</strong> %v</p>

        <div class="success-rate">Success Rate: %.2f%%</div>

        <div class="summary">
            <div class="stat">
                <h3>%d</h3>
                <p>Total Tests</p>
            </div>
            <div class="stat">
                <h3 class="pass">%d</h3>
                <p>Passed</p>
            </div>
            <div class="stat failed">
                <h3 class="fail">%d</h3>
                <p>Failed</p>
            </div>
            <div class="stat">
                <h3 class="skip">%d</h3>
                <p>Skipped</p>
            </div>
        </div>

        <h2>Test Results</h2>
        <table>
            <thead>
                <tr>
                    <th>Test ID</th>
                    <th>Test Name</th>
                    <th>Status</th>
                    <th>Duration</th>
                    <th>Steps</th>
                    <th>Error</th>
                </tr>
            </thead>
            <tbody>
`,
		timestamp,
		getColorForRate(successRate),
		time.Now().Format("2006-01-02 15:04:05"),
		orch.EndTime.Sub(orch.StartTime),
		successRate,
		len(results),
		countStatus(results, "PASS"),
		countStatus(results, "FAIL"),
		countStatus(results, "SKIP"),
	)

	// Add test results
	for _, result := range results {
		statusClass := getStatusClass(result.Status)
		errorMsg := result.Error
		if errorMsg == "" {
			errorMsg = "-"
		}

		html += fmt.Sprintf(`
                <tr>
                    <td>%s</td>
                    <td>%s</td>
                    <td class="%s">%s</td>
                    <td>%v</td>
                    <td>%d</td>
                    <td>%s</td>
                </tr>
`, result.TestID, result.TestName, statusClass, result.Status, result.Duration, len(result.Steps), errorMsg)
	}

	html += `
            </tbody>
        </table>
    </div>
</body>
</html>
`

	// Ensure directory exists
	if err := os.MkdirAll(r.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	fmt.Printf("HTML report generated: %s\n", filename)
	return nil
}

// GenerateJSONReport generates a JSON test report
func (r *Reporter) GenerateJSONReport(orch *orchestrator.Orchestrator) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(r.OutputPath, fmt.Sprintf("qa-report-%s.json", timestamp))

	report := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"duration":     orch.EndTime.Sub(orch.StartTime).String(),
		"successRate":  orch.GetSuccessRate(),
		"totalTests":   len(orch.GetResults()),
		"passed":       countStatus(orch.GetResults(), "PASS"),
		"failed":       countStatus(orch.GetResults(), "FAIL"),
		"skipped":      countStatus(orch.GetResults(), "SKIP"),
		"errors":       countStatus(orch.GetResults(), "ERROR"),
		"testResults":  orch.GetResults(),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.MkdirAll(r.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	fmt.Printf("JSON report generated: %s\n", filename)
	return nil
}

// GenerateMarkdownReport generates a Markdown test report
func (r *Reporter) GenerateMarkdownReport(orch *orchestrator.Orchestrator) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(r.OutputPath, fmt.Sprintf("qa-report-%s.md", timestamp))

	results := orch.GetResults()
	successRate := orch.GetSuccessRate()

	markdown := fmt.Sprintf(`# HelixTrack QA-AI Test Report

**Generated:** %s
**Duration:** %v
**Success Rate:** %.2f%%

## Summary

| Metric | Count |
|--------|-------|
| Total Tests | %d |
| Passed | %d |
| Failed | %d |
| Skipped | %d |
| Errors | %d |

## Test Results

| Test ID | Test Name | Status | Duration | Error |
|---------|-----------|--------|----------|-------|
`,
		time.Now().Format("2006-01-02 15:04:05"),
		orch.EndTime.Sub(orch.StartTime),
		successRate,
		len(results),
		countStatus(results, "PASS"),
		countStatus(results, "FAIL"),
		countStatus(results, "SKIP"),
		countStatus(results, "ERROR"),
	)

	for _, result := range results {
		errorMsg := result.Error
		if errorMsg == "" {
			errorMsg = "-"
		}
		markdown += fmt.Sprintf("| %s | %s | %s | %v | %s |\n",
			result.TestID, result.TestName, result.Status, result.Duration, errorMsg)
	}

	markdown += "\n## Failed Tests\n\n"
	hasFailures := false
	for _, result := range results {
		if result.Status == "FAIL" || result.Status == "ERROR" {
			hasFailures = true
			markdown += fmt.Sprintf("### %s (%s)\n\n", result.TestName, result.TestID)
			markdown += fmt.Sprintf("**Error:** %s\n\n", result.Error)
			markdown += fmt.Sprintf("**Duration:** %v\n\n", result.Duration)
		}
	}

	if !hasFailures {
		markdown += "No failures detected. All tests passed!\n"
	}

	if err := os.MkdirAll(r.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(filename, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report: %w", err)
	}

	fmt.Printf("Markdown report generated: %s\n", filename)
	return nil
}

// Helper functions

func countStatus(results map[string]agents.TestResult, status string) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}

func getStatusClass(status string) string {
	switch status {
	case "PASS":
		return "pass"
	case "FAIL":
		return "fail"
	case "SKIP":
		return "skip"
	case "ERROR":
		return "error"
	default:
		return ""
	}
}

func getColorForRate(rate float64) string {
	if rate >= 95.0 {
		return "#4CAF50"
	} else if rate >= 80.0 {
		return "#ff9800"
	}
	return "#f44336"
}
