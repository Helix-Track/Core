package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomField_IsValidFieldType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		expected  bool
	}{
		{"Text type", string(CustomFieldTypeText), true},
		{"TextArea type", string(CustomFieldTypeTextArea), true},
		{"Number type", string(CustomFieldTypeNumber), true},
		{"Date type", string(CustomFieldTypeDate), true},
		{"DateTime type", string(CustomFieldTypeDateTime), true},
		{"Select type", string(CustomFieldTypeSelect), true},
		{"MultiSelect type", string(CustomFieldTypeMultiSelect), true},
		{"User type", string(CustomFieldTypeUser), true},
		{"URL type", string(CustomFieldTypeURL), true},
		{"Checkbox type", string(CustomFieldTypeCheckbox), true},
		{"Radio type", string(CustomFieldTypeRadio), true},
		{"Invalid type", "invalid-type", false},
		{"Empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := &CustomField{
				FieldType: tt.fieldType,
			}
			result := cf.IsValidFieldType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomField_IsGlobal(t *testing.T) {
	projectID := "project-123"

	tests := []struct {
		name      string
		projectID *string
		expected  bool
	}{
		{"Global field (nil project)", nil, true},
		{"Project-specific field", &projectID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := &CustomField{
				ProjectID: tt.projectID,
			}
			result := cf.IsGlobal()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomField_IsSelectType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		expected  bool
	}{
		{"Select type", string(CustomFieldTypeSelect), true},
		{"MultiSelect type", string(CustomFieldTypeMultiSelect), true},
		{"Text type (not select)", string(CustomFieldTypeText), false},
		{"Number type (not select)", string(CustomFieldTypeNumber), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := &CustomField{
				FieldType: tt.fieldType,
			}
			result := cf.IsSelectType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomField_RequiresOptions(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		expected  bool
	}{
		{"Select type", string(CustomFieldTypeSelect), true},
		{"MultiSelect type", string(CustomFieldTypeMultiSelect), true},
		{"Radio type", string(CustomFieldTypeRadio), true},
		{"Text type (no options)", string(CustomFieldTypeText), false},
		{"Number type (no options)", string(CustomFieldTypeNumber), false},
		{"Checkbox type (no options)", string(CustomFieldTypeCheckbox), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := &CustomField{
				FieldType: tt.fieldType,
			}
			result := cf.RequiresOptions()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomFieldOption_Structure(t *testing.T) {
	option := CustomFieldOption{
		ID:            "option-1",
		CustomFieldID: "field-1",
		Value:         "value1",
		DisplayValue:  "Value 1",
		Position:      1,
		IsDefault:     true,
		Created:       1234567890,
		Modified:      1234567890,
		Deleted:       false,
	}

	assert.Equal(t, "option-1", option.ID)
	assert.Equal(t, "field-1", option.CustomFieldID)
	assert.Equal(t, "value1", option.Value)
	assert.Equal(t, "Value 1", option.DisplayValue)
	assert.Equal(t, 1, option.Position)
	assert.True(t, option.IsDefault)
}

func TestTicketCustomFieldValue_Structure(t *testing.T) {
	value := "test-value"
	fieldValue := TicketCustomFieldValue{
		ID:            "value-1",
		TicketID:      "ticket-1",
		CustomFieldID: "field-1",
		Value:         &value,
		Created:       1234567890,
		Modified:      1234567890,
		Deleted:       false,
	}

	assert.Equal(t, "value-1", fieldValue.ID)
	assert.Equal(t, "ticket-1", fieldValue.TicketID)
	assert.Equal(t, "field-1", fieldValue.CustomFieldID)
	assert.NotNil(t, fieldValue.Value)
	assert.Equal(t, "test-value", *fieldValue.Value)
}

func TestCustomFieldTypes_Constants(t *testing.T) {
	// Verify all field type constants are defined
	assert.Equal(t, CustomFieldType("text"), CustomFieldTypeText)
	assert.Equal(t, CustomFieldType("textarea"), CustomFieldTypeTextArea)
	assert.Equal(t, CustomFieldType("number"), CustomFieldTypeNumber)
	assert.Equal(t, CustomFieldType("date"), CustomFieldTypeDate)
	assert.Equal(t, CustomFieldType("datetime"), CustomFieldTypeDateTime)
	assert.Equal(t, CustomFieldType("select"), CustomFieldTypeSelect)
	assert.Equal(t, CustomFieldType("multi_select"), CustomFieldTypeMultiSelect)
	assert.Equal(t, CustomFieldType("user"), CustomFieldTypeUser)
	assert.Equal(t, CustomFieldType("url"), CustomFieldTypeURL)
	assert.Equal(t, CustomFieldType("checkbox"), CustomFieldTypeCheckbox)
	assert.Equal(t, CustomFieldType("radio"), CustomFieldTypeRadio)
}

func TestCustomField_Complete(t *testing.T) {
	projectID := "project-1"
	defaultValue := "default"
	config := `{"max": 100}`

	cf := CustomField{
		ID:            "field-1",
		FieldName:     "Priority Level",
		FieldType:     string(CustomFieldTypeNumber),
		Description:   "Priority level from 1-5",
		ProjectID:     &projectID,
		IsRequired:    true,
		DefaultValue:  &defaultValue,
		Configuration: &config,
		Created:       1234567890,
		Modified:      1234567890,
		Deleted:       false,
	}

	// Test all fields
	assert.Equal(t, "field-1", cf.ID)
	assert.Equal(t, "Priority Level", cf.FieldName)
	assert.Equal(t, string(CustomFieldTypeNumber), cf.FieldType)
	assert.Equal(t, "Priority level from 1-5", cf.Description)
	assert.NotNil(t, cf.ProjectID)
	assert.Equal(t, "project-1", *cf.ProjectID)
	assert.True(t, cf.IsRequired)
	assert.NotNil(t, cf.DefaultValue)
	assert.Equal(t, "default", *cf.DefaultValue)
	assert.NotNil(t, cf.Configuration)
	assert.Equal(t, `{"max": 100}`, *cf.Configuration)

	// Test methods
	assert.True(t, cf.IsValidFieldType())
	assert.False(t, cf.IsGlobal())
	assert.False(t, cf.IsSelectType())
	assert.False(t, cf.RequiresOptions())
}

func BenchmarkCustomField_IsValidFieldType(b *testing.B) {
	cf := &CustomField{FieldType: string(CustomFieldTypeSelect)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cf.IsValidFieldType()
	}
}

func BenchmarkCustomField_IsSelectType(b *testing.B) {
	cf := &CustomField{FieldType: string(CustomFieldTypeSelect)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cf.IsSelectType()
	}
}
