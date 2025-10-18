package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ================================================================
// DocumentTemplate Tests
// ================================================================

func TestDocumentTemplate_Validate(t *testing.T) {
	tests := []struct {
		name      string
		template  *DocumentTemplate
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid template",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "Meeting Notes",
				TypeID:          "type-page",
				ContentTemplate: "# Meeting Notes\n\n## Attendees\n...",
				CreatorID:       "user-123",
				IsPublic:        true,
				UseCount:        0,
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			template: &DocumentTemplate{
				ID:              "",
				Name:            "Meeting Notes",
				TypeID:          "type-page",
				ContentTemplate: "# Meeting Notes",
				CreatorID:       "user-123",
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "template ID cannot be empty",
		},
		{
			name: "Empty Name",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "",
				TypeID:          "type-page",
				ContentTemplate: "# Meeting Notes",
				CreatorID:       "user-123",
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "template name cannot be empty",
		},
		{
			name: "Empty TypeID",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "Meeting Notes",
				TypeID:          "",
				ContentTemplate: "# Meeting Notes",
				CreatorID:       "user-123",
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "template type ID cannot be empty",
		},
		{
			name: "Empty ContentTemplate",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "Meeting Notes",
				TypeID:          "type-page",
				ContentTemplate: "",
				CreatorID:       "user-123",
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "template content cannot be empty",
		},
		{
			name: "Empty CreatorID",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "Meeting Notes",
				TypeID:          "type-page",
				ContentTemplate: "# Meeting Notes",
				CreatorID:       "",
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "template creator ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "Meeting Notes",
				TypeID:          "type-page",
				ContentTemplate: "# Meeting Notes",
				CreatorID:       "user-123",
				Created:         0,
				Modified:        time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "template created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "Meeting Notes",
				TypeID:          "type-page",
				ContentTemplate: "# Meeting Notes",
				CreatorID:       "user-123",
				Created:         time.Now().Unix(),
				Modified:        0,
			},
			wantError: true,
			errorMsg:  "template modified timestamp cannot be zero",
		},
		{
			name: "Template with all optional fields",
			template: &DocumentTemplate{
				ID:              "template-123",
				Name:            "Project Plan",
				Description:     stringPtr("Template for project planning"),
				SpaceID:         stringPtr("space-123"),
				TypeID:          "type-page",
				ContentTemplate: "# {{project_name}}\n\n## Objectives\n...",
				VariablesJSON:   stringPtr(`{"project_name": "string"}`),
				CreatorID:       "user-123",
				IsPublic:        false,
				UseCount:        25,
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentTemplate_SetTimestamps(t *testing.T) {
	tests := []struct {
		name     string
		template *DocumentTemplate
		checkFn  func(*testing.T, *DocumentTemplate, int64)
	}{
		{
			name: "Set both timestamps when zero",
			template: &DocumentTemplate{
				Created:  0,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dt *DocumentTemplate, before int64) {
				assert.GreaterOrEqual(t, dt.Created, before)
				assert.GreaterOrEqual(t, dt.Modified, before)
				assert.Equal(t, dt.Created, dt.Modified)
			},
		},
		{
			name: "Only update modified when created exists",
			template: &DocumentTemplate{
				Created:  1234567890,
				Modified: 0,
			},
			checkFn: func(t *testing.T, dt *DocumentTemplate, before int64) {
				assert.Equal(t, int64(1234567890), dt.Created)
				assert.GreaterOrEqual(t, dt.Modified, before)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.template.SetTimestamps()
			tt.checkFn(t, tt.template, before)
		})
	}
}

func TestDocumentTemplate_IncrementUseCount(t *testing.T) {
	tests := []struct {
		name         string
		initialCount int
		increments   int
		expectedCount int
	}{
		{"From zero", 0, 1, 1},
		{"From one", 1, 1, 2},
		{"Multiple increments", 0, 5, 5},
		{"From high count", 100, 10, 110},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := &DocumentTemplate{UseCount: tt.initialCount}

			for i := 0; i < tt.increments; i++ {
				template.IncrementUseCount()
			}

			assert.Equal(t, tt.expectedCount, template.UseCount)
		})
	}
}

func TestDocumentTemplate_Structure(t *testing.T) {
	description := "Template for sprint planning"
	spaceID := "space-tech"
	variablesJSON := `{"sprint_number": "int", "team_name": "string"}`

	template := DocumentTemplate{
		ID:              "template-123",
		Name:            "Sprint Planning",
		Description:     &description,
		SpaceID:         &spaceID,
		TypeID:          "type-page",
		ContentTemplate: "# Sprint {{sprint_number}} - {{team_name}}",
		VariablesJSON:   &variablesJSON,
		CreatorID:       "user-123",
		IsPublic:        true,
		UseCount:        42,
		Created:         time.Now().Unix(),
		Modified:        time.Now().Unix(),
		Deleted:         false,
	}

	assert.Equal(t, "template-123", template.ID)
	assert.Equal(t, "Sprint Planning", template.Name)
	assert.NotNil(t, template.Description)
	assert.Equal(t, "Template for sprint planning", *template.Description)
	assert.NotNil(t, template.SpaceID)
	assert.Equal(t, "space-tech", *template.SpaceID)
	assert.Equal(t, "type-page", template.TypeID)
	assert.Contains(t, template.ContentTemplate, "{{sprint_number}}")
	assert.NotNil(t, template.VariablesJSON)
	assert.Contains(t, *template.VariablesJSON, "sprint_number")
	assert.Equal(t, "user-123", template.CreatorID)
	assert.True(t, template.IsPublic)
	assert.Equal(t, 42, template.UseCount)
	assert.Greater(t, template.Created, int64(0))
	assert.Greater(t, template.Modified, int64(0))
	assert.False(t, template.Deleted)
}

// ================================================================
// DocumentBlueprint Tests
// ================================================================

func TestDocumentBlueprint_Validate(t *testing.T) {
	tests := []struct {
		name      string
		blueprint *DocumentBlueprint
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid blueprint",
			blueprint: &DocumentBlueprint{
				ID:          "blueprint-123",
				Name:        "Project Setup Wizard",
				TemplateID:  "template-123",
				CreatorID:   "user-123",
				IsPublic:    true,
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			blueprint: &DocumentBlueprint{
				ID:          "",
				Name:        "Project Setup Wizard",
				TemplateID:  "template-123",
				CreatorID:   "user-123",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "blueprint ID cannot be empty",
		},
		{
			name: "Empty Name",
			blueprint: &DocumentBlueprint{
				ID:          "blueprint-123",
				Name:        "",
				TemplateID:  "template-123",
				CreatorID:   "user-123",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "blueprint name cannot be empty",
		},
		{
			name: "Empty TemplateID",
			blueprint: &DocumentBlueprint{
				ID:          "blueprint-123",
				Name:        "Project Setup Wizard",
				TemplateID:  "",
				CreatorID:   "user-123",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "blueprint template ID cannot be empty",
		},
		{
			name: "Empty CreatorID",
			blueprint: &DocumentBlueprint{
				ID:          "blueprint-123",
				Name:        "Project Setup Wizard",
				TemplateID:  "template-123",
				CreatorID:   "",
				Created:     time.Now().Unix(),
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "blueprint creator ID cannot be empty",
		},
		{
			name: "Zero Created timestamp",
			blueprint: &DocumentBlueprint{
				ID:          "blueprint-123",
				Name:        "Project Setup Wizard",
				TemplateID:  "template-123",
				CreatorID:   "user-123",
				Created:     0,
				Modified:    time.Now().Unix(),
			},
			wantError: true,
			errorMsg:  "blueprint created timestamp cannot be zero",
		},
		{
			name: "Zero Modified timestamp",
			blueprint: &DocumentBlueprint{
				ID:          "blueprint-123",
				Name:        "Project Setup Wizard",
				TemplateID:  "template-123",
				CreatorID:   "user-123",
				Created:     time.Now().Unix(),
				Modified:    0,
			},
			wantError: true,
			errorMsg:  "blueprint modified timestamp cannot be zero",
		},
		{
			name: "Blueprint with all optional fields",
			blueprint: &DocumentBlueprint{
				ID:              "blueprint-123",
				Name:            "Project Setup Wizard",
				Description:     stringPtr("Multi-step wizard for project setup"),
				SpaceID:         stringPtr("space-123"),
				TemplateID:      "template-123",
				WizardStepsJSON: stringPtr(`[{"step": 1, "title": "Basic Info"}]`),
				CreatorID:       "user-123",
				IsPublic:        false,
				Created:         time.Now().Unix(),
				Modified:        time.Now().Unix(),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.blueprint.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocumentBlueprint_SetTimestamps(t *testing.T) {
	tests := []struct {
		name      string
		blueprint *DocumentBlueprint
		checkFn   func(*testing.T, *DocumentBlueprint, int64)
	}{
		{
			name: "Set both timestamps when zero",
			blueprint: &DocumentBlueprint{
				Created:  0,
				Modified: 0,
			},
			checkFn: func(t *testing.T, db *DocumentBlueprint, before int64) {
				assert.GreaterOrEqual(t, db.Created, before)
				assert.GreaterOrEqual(t, db.Modified, before)
				assert.Equal(t, db.Created, db.Modified)
			},
		},
		{
			name: "Only update modified when created exists",
			blueprint: &DocumentBlueprint{
				Created:  1234567890,
				Modified: 0,
			},
			checkFn: func(t *testing.T, db *DocumentBlueprint, before int64) {
				assert.Equal(t, int64(1234567890), db.Created)
				assert.GreaterOrEqual(t, db.Modified, before)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			tt.blueprint.SetTimestamps()
			tt.checkFn(t, tt.blueprint, before)
		})
	}
}

func TestDocumentBlueprint_Structure(t *testing.T) {
	description := "Complete project setup with team creation"
	spaceID := "space-projects"
	wizardStepsJSON := `[
		{"step": 1, "title": "Project Name"},
		{"step": 2, "title": "Team Selection"},
		{"step": 3, "title": "Timeline"}
	]`

	blueprint := DocumentBlueprint{
		ID:              "blueprint-123",
		Name:            "Project Setup Wizard",
		Description:     &description,
		SpaceID:         &spaceID,
		TemplateID:      "template-project",
		WizardStepsJSON: &wizardStepsJSON,
		CreatorID:       "user-admin",
		IsPublic:        true,
		Created:         time.Now().Unix(),
		Modified:        time.Now().Unix(),
		Deleted:         false,
	}

	assert.Equal(t, "blueprint-123", blueprint.ID)
	assert.Equal(t, "Project Setup Wizard", blueprint.Name)
	assert.NotNil(t, blueprint.Description)
	assert.Equal(t, "Complete project setup with team creation", *blueprint.Description)
	assert.NotNil(t, blueprint.SpaceID)
	assert.Equal(t, "space-projects", *blueprint.SpaceID)
	assert.Equal(t, "template-project", blueprint.TemplateID)
	assert.NotNil(t, blueprint.WizardStepsJSON)
	assert.Contains(t, *blueprint.WizardStepsJSON, "Project Name")
	assert.Equal(t, "user-admin", blueprint.CreatorID)
	assert.True(t, blueprint.IsPublic)
	assert.Greater(t, blueprint.Created, int64(0))
	assert.Greater(t, blueprint.Modified, int64(0))
	assert.False(t, blueprint.Deleted)
}

// ================================================================
// Benchmark Tests
// ================================================================

func BenchmarkDocumentTemplate_Validate(b *testing.B) {
	template := &DocumentTemplate{
		ID:              "template-123",
		Name:            "Test Template",
		TypeID:          "type-page",
		ContentTemplate: "# Test",
		CreatorID:       "user-123",
		Created:         time.Now().Unix(),
		Modified:        time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = template.Validate()
	}
}

func BenchmarkDocumentTemplate_IncrementUseCount(b *testing.B) {
	template := &DocumentTemplate{UseCount: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		template.IncrementUseCount()
	}
}

func BenchmarkDocumentBlueprint_Validate(b *testing.B) {
	blueprint := &DocumentBlueprint{
		ID:          "blueprint-123",
		Name:        "Test Blueprint",
		TemplateID:  "template-123",
		CreatorID:   "user-123",
		Created:     time.Now().Unix(),
		Modified:    time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = blueprint.Validate()
	}
}
