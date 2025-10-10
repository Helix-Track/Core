package models

// CustomField represents a user-defined custom field for tickets
type CustomField struct {
	ID            string  `json:"id" db:"id"`
	FieldName     string  `json:"fieldName" db:"field_name" binding:"required"`
	FieldType     string  `json:"fieldType" db:"field_type" binding:"required"`
	Description   string  `json:"description,omitempty" db:"description"`
	ProjectID     *string `json:"projectId,omitempty" db:"project_id"` // NULL for global fields
	IsRequired    bool    `json:"isRequired" db:"is_required"`
	DefaultValue  *string `json:"defaultValue,omitempty" db:"default_value"`
	Configuration *string `json:"configuration,omitempty" db:"configuration"` // JSON configuration
	Created       int64   `json:"created" db:"created"`
	Modified      int64   `json:"modified" db:"modified"`
	Deleted       bool    `json:"deleted" db:"deleted"`
}

// CustomFieldType represents the type of custom field
type CustomFieldType string

const (
	CustomFieldTypeText        CustomFieldType = "text"
	CustomFieldTypeTextArea    CustomFieldType = "textarea"
	CustomFieldTypeNumber      CustomFieldType = "number"
	CustomFieldTypeDate        CustomFieldType = "date"
	CustomFieldTypeDateTime    CustomFieldType = "datetime"
	CustomFieldTypeSelect      CustomFieldType = "select"
	CustomFieldTypeMultiSelect CustomFieldType = "multi_select"
	CustomFieldTypeUser        CustomFieldType = "user"
	CustomFieldTypeURL         CustomFieldType = "url"
	CustomFieldTypeCheckbox    CustomFieldType = "checkbox"
	CustomFieldTypeRadio       CustomFieldType = "radio"
)

// CustomFieldOption represents an option for select/multi-select custom fields
type CustomFieldOption struct {
	ID            string `json:"id" db:"id"`
	CustomFieldID string `json:"customFieldId" db:"custom_field_id" binding:"required"`
	Value         string `json:"value" db:"value" binding:"required"`
	DisplayValue  string `json:"displayValue" db:"display_value" binding:"required"`
	Position      int    `json:"position" db:"position"`
	IsDefault     bool   `json:"isDefault" db:"is_default"`
	Created       int64  `json:"created" db:"created"`
	Modified      int64  `json:"modified" db:"modified"`
	Deleted       bool   `json:"deleted" db:"deleted"`
}

// TicketCustomFieldValue represents the value of a custom field for a specific ticket
type TicketCustomFieldValue struct {
	ID            string  `json:"id" db:"id"`
	TicketID      string  `json:"ticketId" db:"ticket_id" binding:"required"`
	CustomFieldID string  `json:"customFieldId" db:"custom_field_id" binding:"required"`
	Value         *string `json:"value,omitempty" db:"value"` // NULL for empty/unset fields
	Created       int64   `json:"created" db:"created"`
	Modified      int64   `json:"modified" db:"modified"`
	Deleted       bool    `json:"deleted" db:"deleted"`
}

// IsValidFieldType checks if the field type is valid
func (cf *CustomField) IsValidFieldType() bool {
	validTypes := []CustomFieldType{
		CustomFieldTypeText,
		CustomFieldTypeTextArea,
		CustomFieldTypeNumber,
		CustomFieldTypeDate,
		CustomFieldTypeDateTime,
		CustomFieldTypeSelect,
		CustomFieldTypeMultiSelect,
		CustomFieldTypeUser,
		CustomFieldTypeURL,
		CustomFieldTypeCheckbox,
		CustomFieldTypeRadio,
	}

	for _, validType := range validTypes {
		if CustomFieldType(cf.FieldType) == validType {
			return true
		}
	}

	return false
}

// IsGlobal checks if the custom field is global (applies to all projects)
func (cf *CustomField) IsGlobal() bool {
	return cf.ProjectID == nil
}

// IsSelectType checks if the field type is select or multi-select
func (cf *CustomField) IsSelectType() bool {
	return CustomFieldType(cf.FieldType) == CustomFieldTypeSelect ||
		CustomFieldType(cf.FieldType) == CustomFieldTypeMultiSelect
}

// RequiresOptions checks if the field type requires options (select types)
func (cf *CustomField) RequiresOptions() bool {
	return cf.IsSelectType() || CustomFieldType(cf.FieldType) == CustomFieldTypeRadio
}
