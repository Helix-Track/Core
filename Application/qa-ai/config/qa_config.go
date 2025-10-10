package config

import (
	"time"
)

// QAConfig holds the configuration for QA testing
type QAConfig struct {
	// Server configuration
	ServerURL      string
	ServerStartCmd string
	ServerStopCmd  string
	StartupTimeout time.Duration

	// Database configuration
	DatabasePath   string
	DatabaseType   string
	ResetBeforeRun bool

	// Test configuration
	ConcurrentTests   int
	TestTimeout       time.Duration
	RetryFailedTests  bool
	MaxRetries        int
	StopOnFirstFail   bool
	GenerateReport    bool
	ReportPath        string
	ScreenshotOnFail  bool
	VerboseLogging    bool

	// AI Agent configuration
	AIModel           string
	AITemperature     float64
	AIMaxTokens       int
	AIThinkingEnabled bool

	// Test data
	TestDataPath      string
	FixturesPath      string
	CleanupAfterTests bool
}

// DefaultQAConfig returns default QA configuration
func DefaultQAConfig() QAConfig {
	return QAConfig{
		ServerURL:         "http://localhost:8080",
		ServerStartCmd:    "", // Server must be started manually
		ServerStopCmd:     "",
		StartupTimeout:    30 * time.Second,

		DatabasePath:      "../Database/Definition.sqlite", // Relative to qa-ai directory
		DatabaseType:      "sqlite",
		ResetBeforeRun:    false, // Use existing database

		ConcurrentTests:   1, // Sequential by default for deterministic results
		TestTimeout:       5 * time.Minute,
		RetryFailedTests:  true,
		MaxRetries:        3,
		StopOnFirstFail:   false,
		GenerateReport:    true,
		ReportPath:        "./reports",
		ScreenshotOnFail:  false,
		VerboseLogging:    true,

		AIModel:           "claude-sonnet-4",
		AITemperature:     0.7,
		AIMaxTokens:       4096,
		AIThinkingEnabled: true,

		TestDataPath:      "./fixtures",
		FixturesPath:      "./fixtures",
		CleanupAfterTests: false, // Keep data for analysis
	}
}

// TestProfile represents a user profile for testing
type TestProfile struct {
	Username    string
	Password    string
	Email       string
	Role        string
	Permissions []string
	Description string
}

// GetTestProfiles returns all test user profiles
func GetTestProfiles() []TestProfile {
	return []TestProfile{
		{
			Username:    "admin_user",
			Password:    "Admin@123456",
			Email:       "admin@helixtrack.test",
			Role:        "admin",
			Permissions: []string{"ALL"},
			Description: "Administrator with full permissions",
		},
		{
			Username:    "project_manager",
			Password:    "PM@123456",
			Email:       "pm@helixtrack.test",
			Role:        "manager",
			Permissions: []string{"CREATE_PROJECT", "UPDATE_PROJECT", "CREATE_TICKET", "UPDATE_TICKET", "DELETE_TICKET"},
			Description: "Project manager with project and ticket management",
		},
		{
			Username:    "developer",
			Password:    "Dev@123456",
			Email:       "dev@helixtrack.test",
			Role:        "developer",
			Permissions: []string{"CREATE_TICKET", "UPDATE_TICKET", "CREATE_COMMENT"},
			Description: "Developer with ticket creation and update",
		},
		{
			Username:    "reporter",
			Password:    "Reporter@123456",
			Email:       "reporter@helixtrack.test",
			Role:        "reporter",
			Permissions: []string{"CREATE_TICKET", "CREATE_COMMENT"},
			Description: "Reporter who can create tickets and comments",
		},
		{
			Username:    "viewer",
			Password:    "Viewer@123456",
			Email:       "viewer@helixtrack.test",
			Role:        "viewer",
			Permissions: []string{"READ"},
			Description: "Read-only user",
		},
		{
			Username:    "qa_tester",
			Password:    "QA@123456",
			Email:       "qa@helixtrack.test",
			Role:        "tester",
			Permissions: []string{"CREATE_TICKET", "UPDATE_TICKET", "CREATE_COMMENT", "UPDATE_COMMENT"},
			Description: "QA tester with testing-specific permissions",
		},
	}
}

// TestSuite represents a collection of related tests
type TestSuite struct {
	Name        string
	Description string
	Tests       []string
	Priority    int
	Tags        []string
}

// GetTestSuites returns all test suites
func GetTestSuites() []TestSuite {
	return []TestSuite{
		{
			Name:        "authentication",
			Description: "User authentication and authorization tests",
			Tests:       []string{"register", "login", "logout", "jwt_validation", "password_reset"},
			Priority:    1,
			Tags:        []string{"critical", "security"},
		},
		{
			Name:        "projects",
			Description: "Project management tests",
			Tests:       []string{"create_project", "update_project", "delete_project", "list_projects", "project_permissions"},
			Priority:    2,
			Tags:        []string{"core", "crud"},
		},
		{
			Name:        "tickets",
			Description: "Ticket/Issue management tests",
			Tests:       []string{"create_ticket", "update_ticket", "delete_ticket", "assign_ticket", "ticket_lifecycle", "ticket_search"},
			Priority:    2,
			Tags:        []string{"core", "crud"},
		},
		{
			Name:        "comments",
			Description: "Comment system tests",
			Tests:       []string{"create_comment", "update_comment", "delete_comment", "nested_comments"},
			Priority:    3,
			Tags:        []string{"core"},
		},
		{
			Name:        "attachments",
			Description: "File attachment tests",
			Tests:       []string{"upload_file", "download_file", "delete_file", "large_file", "multiple_files"},
			Priority:    3,
			Tags:        []string{"core", "files"},
		},
		{
			Name:        "permissions",
			Description: "Permission system tests",
			Tests:       []string{"role_permissions", "user_permissions", "project_permissions", "forbidden_access"},
			Priority:    1,
			Tags:        []string{"critical", "security"},
		},
		{
			Name:        "search",
			Description: "Search and filter tests",
			Tests:       []string{"search_tickets", "filter_projects", "advanced_search", "search_performance"},
			Priority:    4,
			Tags:        []string{"feature"},
		},
		{
			Name:        "security",
			Description: "Security feature tests",
			Tests:       []string{"csrf_protection", "xss_prevention", "sql_injection", "rate_limiting", "brute_force"},
			Priority:    1,
			Tags:        []string{"critical", "security"},
		},
		{
			Name:        "edge_cases",
			Description: "Edge case and error handling tests",
			Tests:       []string{"invalid_input", "missing_data", "concurrent_updates", "race_conditions", "large_datasets"},
			Priority:    5,
			Tags:        []string{"edge", "stress"},
		},
		{
			Name:        "database",
			Description: "Database integrity tests",
			Tests:       []string{"data_consistency", "foreign_keys", "transactions", "rollback", "concurrent_writes"},
			Priority:    2,
			Tags:        []string{"critical", "database"},
		},
	}
}
