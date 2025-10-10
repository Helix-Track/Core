package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helixtrack.ru/core/internal/models"
)

// setupCustomFieldTestHandler creates a test handler with custom field tables and dependencies
func setupCustomFieldTestHandler(t *testing.T) *Handler {
	handler := setupTestHandler(t)

	// Create custom_field table
	_, err := handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS custom_field (
			id TEXT PRIMARY KEY,
			field_name TEXT NOT NULL,
			field_type TEXT NOT NULL,
			description TEXT,
			project_id TEXT,
			is_required INTEGER NOT NULL DEFAULT 0,
			default_value TEXT,
			configuration TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create custom_field_option table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS custom_field_option (
			id TEXT PRIMARY KEY,
			custom_field_id TEXT NOT NULL,
			value TEXT NOT NULL,
			display_value TEXT NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			is_default INTEGER NOT NULL DEFAULT 0,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create ticket_custom_field_value table
	_, err = handler.db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ticket_custom_field_value (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			custom_field_id TEXT NOT NULL,
			value TEXT,
			created INTEGER NOT NULL,
			modified INTEGER NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Insert test ticket for value tests
	_, err = handler.db.Exec(context.Background(),
		"INSERT INTO ticket (id, title, description, status, created, modified, deleted) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"test-ticket-id", "Test Ticket", "Test ticket description", "open", 1000, 1000, 0)
	require.NoError(t, err)

	return handler
}

// ============================================================================
// Custom Field CRUD Tests
// ============================================================================

func TestCustomFieldHandler_Create_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldCreate,
		Data: map[string]interface{}{
			"fieldName":    "priority_score",
			"fieldType":    "number",
			"description":  "Priority score for tickets",
			"isRequired":   true,
			"defaultValue": "5",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeNoError, response.ErrorCode)

	customFieldData, ok := response.Data["customField"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "priority_score", customFieldData["FieldName"])
	assert.Equal(t, "number", customFieldData["FieldType"])
	assert.Equal(t, true, customFieldData["IsRequired"])
	assert.NotEmpty(t, customFieldData["ID"])

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM custom_field WHERE field_name = ?", "priority_score").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCustomFieldHandler_Create_AllFieldTypes(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	fieldTypes := []string{"text", "number", "date", "boolean", "select", "multiselect", "textarea", "url", "email"}

	for _, fieldType := range fieldTypes {
		reqBody := models.Request{
			Action: models.ActionCustomFieldCreate,
			Data: map[string]interface{}{
				"fieldName": "test_" + fieldType,
				"fieldType": fieldType,
			},
		}

		w := performRequest(handler, "POST", "/do", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code, "Field type %s should be valid", fieldType)
	}

	// Verify all were created
	var count int
	err := handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM custom_field WHERE deleted = 0").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(fieldTypes), count)
}

func TestCustomFieldHandler_Create_InvalidFieldType(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldCreate,
		Data: map[string]interface{}{
			"fieldName": "invalid_field",
			"fieldType": "invalid_type",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Invalid field type")
}

func TestCustomFieldHandler_Create_WithProjectId(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldCreate,
		Data: map[string]interface{}{
			"fieldName": "project_specific",
			"fieldType": "text",
			"projectId": "project-123",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify projectId was stored
	var projectID *string
	err := handler.db.QueryRow(context.Background(),
		"SELECT project_id FROM custom_field WHERE field_name = ?", "project_specific").Scan(&projectID)
	require.NoError(t, err)
	require.NotNil(t, projectID)
	assert.Equal(t, "project-123", *projectID)
}

func TestCustomFieldHandler_Create_WithConfiguration(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldCreate,
		Data: map[string]interface{}{
			"fieldName": "configured_field",
			"fieldType": "number",
			"configuration": map[string]interface{}{
				"min": 0,
				"max": 100,
			},
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify configuration was stored as JSON
	var config *string
	err := handler.db.QueryRow(context.Background(),
		"SELECT configuration FROM custom_field WHERE field_name = ?", "configured_field").Scan(&config)
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Contains(t, *config, "min")
	assert.Contains(t, *config, "max")
}

func TestCustomFieldHandler_Create_MissingFieldName(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldCreate,
		Data: map[string]interface{}{
			"fieldType": "text",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing fieldName")
}

func TestCustomFieldHandler_Create_MissingFieldType(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldCreate,
		Data: map[string]interface{}{
			"fieldName": "test_field",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "Missing fieldType")
}

func TestCustomFieldHandler_Read_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert test custom field
	fieldID := "test-field-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "test_field", "text", "Test description", nil, 1, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldRead,
		Data: map[string]interface{}{
			"id": fieldID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	customFieldData, ok := response.Data["customField"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, fieldID, customFieldData["ID"])
	assert.Equal(t, "test_field", customFieldData["FieldName"])
}

func TestCustomFieldHandler_Read_NotFound(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldRead,
		Data: map[string]interface{}{
			"id": "non-existent-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestCustomFieldHandler_List_Empty(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	customFields, ok := response.Data["customFields"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, customFields)
}

func TestCustomFieldHandler_List_Multiple(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert multiple custom fields
	fields := []string{"field1", "field2", "field3"}
	for i, fieldName := range fields {
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"field-"+string(rune('1'+i)), fieldName, "text", "Desc", nil, 0, nil, nil, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionCustomFieldList,
		Data:   map[string]interface{}{},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	customFields, ok := response.Data["customFields"].([]interface{})
	require.True(t, ok)
	assert.Len(t, customFields, 3)
}

func TestCustomFieldHandler_List_FilterByProject(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	projectID := "project-123"

	// Insert global field
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"global-field", "global", "text", "Global field", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert project-specific field
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"project-field", "project_specific", "text", "Project field", &projectID, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert field for different project
	otherProject := "other-project"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"other-field", "other", "text", "Other field", &otherProject, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldList,
		Data: map[string]interface{}{
			"projectId": projectID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	customFields, ok := response.Data["customFields"].([]interface{})
	require.True(t, ok)
	// Should include global field + project-specific field (not other project's field)
	assert.Len(t, customFields, 2)
}

func TestCustomFieldHandler_Modify_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert test custom field
	fieldID := "test-field-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "old_name", "text", "Old description", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldModify,
		Data: map[string]interface{}{
			"id":          fieldID,
			"fieldName":   "new_name",
			"description": "New description",
			"isRequired":  true,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["updated"].(bool))

	// Verify in database
	var fieldName, description string
	var isRequired bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT field_name, description, is_required FROM custom_field WHERE id = ?", fieldID).
		Scan(&fieldName, &description, &isRequired)
	require.NoError(t, err)
	assert.Equal(t, "new_name", fieldName)
	assert.Equal(t, "New description", description)
	assert.True(t, isRequired)
}

func TestCustomFieldHandler_Modify_NoFieldsToUpdate(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert test custom field
	fieldID := "test-field-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "test_field", "text", "Description", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldModify,
		Data: map[string]interface{}{
			"id": fieldID,
			// No fields to update
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeMissingData, response.ErrorCode)
}

func TestCustomFieldHandler_Remove_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert test custom field
	fieldID := "test-field-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "test_field", "text", "Description", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldRemove,
		Data: map[string]interface{}{
			"id": fieldID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["deleted"].(bool))

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM custom_field WHERE id = ?", fieldID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// ============================================================================
// Custom Field Option Tests
// ============================================================================

func TestCustomFieldHandler_OptionCreate_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert select-type custom field
	fieldID := "test-field-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "status", "select", "Status field", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldOptionCreate,
		Data: map[string]interface{}{
			"customFieldId": fieldID,
			"value":         "in_progress",
			"displayValue":  "In Progress",
			"position":      float64(1),
			"isDefault":     true,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	optionData, ok := response.Data["option"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "in_progress", optionData["Value"])
	assert.Equal(t, "In Progress", optionData["DisplayValue"])
	assert.Equal(t, true, optionData["IsDefault"])

	// Verify in database
	var count int
	err = handler.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM custom_field_option WHERE custom_field_id = ?", fieldID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCustomFieldHandler_OptionCreate_NonSelectType(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert text-type custom field (doesn't support options)
	fieldID := "test-field-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "description", "text", "Text field", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldOptionCreate,
		Data: map[string]interface{}{
			"customFieldId": fieldID,
			"value":         "option",
			"displayValue":  "Option",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeInvalidData, response.ErrorCode)
	assert.Contains(t, response.ErrorMessage, "does not support options")
}

func TestCustomFieldHandler_OptionModify_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert option
	optionID := "test-option-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field_option (id, custom_field_id, value, display_value, position, is_default, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		optionID, "field-id", "old_value", "Old Display", 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldOptionModify,
		Data: map[string]interface{}{
			"id":           optionID,
			"value":        "new_value",
			"displayValue": "New Display",
			"isDefault":    true,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify in database
	var value, displayValue string
	var isDefault bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT value, display_value, is_default FROM custom_field_option WHERE id = ?", optionID).
		Scan(&value, &displayValue, &isDefault)
	require.NoError(t, err)
	assert.Equal(t, "new_value", value)
	assert.Equal(t, "New Display", displayValue)
	assert.True(t, isDefault)
}

func TestCustomFieldHandler_OptionRemove_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert option
	optionID := "test-option-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field_option (id, custom_field_id, value, display_value, position, is_default, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		optionID, "field-id", "value", "Display", 0, 0, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldOptionRemove,
		Data: map[string]interface{}{
			"id": optionID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM custom_field_option WHERE id = ?", optionID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestCustomFieldHandler_OptionList_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	fieldID := "test-field-id"

	// Insert options with different positions
	options := []struct {
		id       string
		value    string
		position int
	}{
		{"opt-1", "low", 1},
		{"opt-2", "high", 3},
		{"opt-3", "medium", 2},
	}

	for _, opt := range options {
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO custom_field_option (id, custom_field_id, value, display_value, position, is_default, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			opt.id, fieldID, opt.value, opt.value, opt.position, 0, 1000, 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionCustomFieldOptionList,
		Data: map[string]interface{}{
			"customFieldId": fieldID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	optionsList, ok := response.Data["options"].([]interface{})
	require.True(t, ok)
	assert.Len(t, optionsList, 3)

	// Verify ordered by position
	values := make([]string, len(optionsList))
	for i, opt := range optionsList {
		optMap := opt.(map[string]interface{})
		values[i] = optMap["Value"].(string)
	}
	assert.Equal(t, "low", values[0])    // position 1
	assert.Equal(t, "medium", values[1]) // position 2
	assert.Equal(t, "high", values[2])   // position 3
}

// ============================================================================
// Custom Field Value Tests
// ============================================================================

func TestCustomFieldHandler_ValueSet_CreateNew(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert custom field
	fieldID := "test-field-id"
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "test_field", "text", "Description", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldValueSet,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
			"value":         "Test Value",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["created"].(bool))

	// Verify in database
	var value *string
	err = handler.db.QueryRow(context.Background(),
		"SELECT value FROM ticket_custom_field_value WHERE ticket_id = ? AND custom_field_id = ?",
		"test-ticket-id", fieldID).Scan(&value)
	require.NoError(t, err)
	require.NotNil(t, value)
	assert.Equal(t, "Test Value", *value)
}

func TestCustomFieldHandler_ValueSet_UpdateExisting(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	fieldID := "test-field-id"

	// Insert custom field
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO custom_field (id, field_name, field_type, description, project_id, is_required, default_value, configuration, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fieldID, "test_field", "text", "Description", nil, 0, nil, nil, 1000, 1000, 0)
	require.NoError(t, err)

	// Insert existing value
	oldValue := "Old Value"
	_, err = handler.db.Exec(context.Background(),
		`INSERT INTO ticket_custom_field_value (id, ticket_id, custom_field_id, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"value-id", "test-ticket-id", fieldID, &oldValue, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldValueSet,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
			"value":         "New Value",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["updated"].(bool))

	// Verify value was updated
	var value *string
	err = handler.db.QueryRow(context.Background(),
		"SELECT value FROM ticket_custom_field_value WHERE ticket_id = ? AND custom_field_id = ?",
		"test-ticket-id", fieldID).Scan(&value)
	require.NoError(t, err)
	require.NotNil(t, value)
	assert.Equal(t, "New Value", *value)
}

func TestCustomFieldHandler_ValueGet_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	fieldID := "test-field-id"
	testValue := "Test Value"

	// Insert custom field value
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO ticket_custom_field_value (id, ticket_id, custom_field_id, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"value-id", "test-ticket-id", fieldID, &testValue, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldValueGet,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	fieldValue, ok := response.Data["fieldValue"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test Value", *fieldValue["Value"].(*string))
}

func TestCustomFieldHandler_ValueGet_NotFound(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	reqBody := models.Request{
		Action: models.ActionCustomFieldValueGet,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": "non-existent-field",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrorCodeEntityNotFound, response.ErrorCode)
}

func TestCustomFieldHandler_ValueList_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// Insert multiple custom field values for the ticket
	values := []struct {
		fieldID string
		value   string
	}{
		{"field-1", "Value 1"},
		{"field-2", "Value 2"},
		{"field-3", "Value 3"},
	}

	for i, v := range values {
		_, err := handler.db.Exec(context.Background(),
			`INSERT INTO ticket_custom_field_value (id, ticket_id, custom_field_id, value, created, modified, deleted)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"value-"+string(rune('1'+i)), "test-ticket-id", v.fieldID, &v.value, 1000+int64(i), 1000, 0)
		require.NoError(t, err)
	}

	reqBody := models.Request{
		Action: models.ActionCustomFieldValueList,
		Data: map[string]interface{}{
			"ticketId": "test-ticket-id",
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	fieldValues, ok := response.Data["fieldValues"].([]interface{})
	require.True(t, ok)
	assert.Len(t, fieldValues, 3)
}

func TestCustomFieldHandler_ValueRemove_Success(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	fieldID := "test-field-id"
	testValue := "Test Value"

	// Insert custom field value
	_, err := handler.db.Exec(context.Background(),
		`INSERT INTO ticket_custom_field_value (id, ticket_id, custom_field_id, value, created, modified, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"value-id", "test-ticket-id", fieldID, &testValue, 1000, 1000, 0)
	require.NoError(t, err)

	reqBody := models.Request{
		Action: models.ActionCustomFieldValueRemove,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
		},
	}

	w := performRequest(handler, "POST", "/do", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Data["deleted"].(bool))

	// Verify soft delete
	var deleted bool
	err = handler.db.QueryRow(context.Background(),
		"SELECT deleted FROM ticket_custom_field_value WHERE ticket_id = ? AND custom_field_id = ?",
		"test-ticket-id", fieldID).Scan(&deleted)
	require.NoError(t, err)
	assert.True(t, deleted)
}

// ============================================================================
// Full Custom Field Lifecycle Test
// ============================================================================

func TestCustomFieldHandler_FullLifecycle(t *testing.T) {
	handler := setupCustomFieldTestHandler(t)

	// 1. Create custom field
	createFieldReq := models.Request{
		Action: models.ActionCustomFieldCreate,
		Data: map[string]interface{}{
			"fieldName":   "priority",
			"fieldType":   "select",
			"description": "Priority level",
			"isRequired":  true,
		},
	}

	w := performRequest(handler, "POST", "/do", createFieldReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp models.Response
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	fieldData := createResp.Data["customField"].(map[string]interface{})
	fieldID := fieldData["ID"].(string)

	// 2. Create options for the select field
	options := []struct {
		value   string
		display string
	}{
		{"low", "Low Priority"},
		{"medium", "Medium Priority"},
		{"high", "High Priority"},
	}

	for _, opt := range options {
		createOptionReq := models.Request{
			Action: models.ActionCustomFieldOptionCreate,
			Data: map[string]interface{}{
				"customFieldId": fieldID,
				"value":         opt.value,
				"displayValue":  opt.display,
			},
		}

		w = performRequest(handler, "POST", "/do", createOptionReq)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// 3. List options
	listOptionsReq := models.Request{
		Action: models.ActionCustomFieldOptionList,
		Data: map[string]interface{}{
			"customFieldId": fieldID,
		},
	}

	w = performRequest(handler, "POST", "/do", listOptionsReq)
	var listResp models.Response
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)

	optionsList, ok := listResp.Data["options"].([]interface{})
	require.True(t, ok)
	assert.Len(t, optionsList, 3)

	// 4. Set custom field value on ticket
	setValueReq := models.Request{
		Action: models.ActionCustomFieldValueSet,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
			"value":         "high",
		},
	}

	w = performRequest(handler, "POST", "/do", setValueReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 5. Get custom field value
	getValueReq := models.Request{
		Action: models.ActionCustomFieldValueGet,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
		},
	}

	w = performRequest(handler, "POST", "/do", getValueReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. Update custom field value
	updateValueReq := models.Request{
		Action: models.ActionCustomFieldValueSet,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
			"value":         "medium",
		},
	}

	w = performRequest(handler, "POST", "/do", updateValueReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 7. Remove custom field value
	removeValueReq := models.Request{
		Action: models.ActionCustomFieldValueRemove,
		Data: map[string]interface{}{
			"ticketId":      "test-ticket-id",
			"customFieldId": fieldID,
		},
	}

	w = performRequest(handler, "POST", "/do", removeValueReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// 8. Remove custom field
	removeFieldReq := models.Request{
		Action: models.ActionCustomFieldRemove,
		Data: map[string]interface{}{
			"id": fieldID,
		},
	}

	w = performRequest(handler, "POST", "/do", removeFieldReq)
	assert.Equal(t, http.StatusOK, w.Code)
}
