package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSuccessResponse(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "Success response with data",
			data: map[string]interface{}{
				"id":   123,
				"name": "Test",
			},
		},
		{
			name: "Success response with nil data",
			data: nil,
		},
		{
			name: "Success response with empty data",
			data: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewSuccessResponse(tt.data)
			assert.NotNil(t, resp)
			assert.Equal(t, ErrorCodeNoError, resp.ErrorCode)
			assert.Empty(t, resp.ErrorMessage)
			assert.Empty(t, resp.ErrorMessageLocalised)
			assert.Equal(t, tt.data, resp.Data)
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name              string
		code              int
		message           string
		localisedMessage  string
	}{
		{
			name:              "Error response with all fields",
			code:              ErrorCodeInvalidRequest,
			message:           "Invalid request",
			localisedMessage:  "Некорректный запрос",
		},
		{
			name:              "Error response with empty localised message",
			code:              ErrorCodeInternalError,
			message:           "Internal error",
			localisedMessage:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponse(tt.code, tt.message, tt.localisedMessage)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.code, resp.ErrorCode)
			assert.Equal(t, tt.message, resp.ErrorMessage)
			assert.Equal(t, tt.localisedMessage, resp.ErrorMessageLocalised)
			assert.Nil(t, resp.Data)
		})
	}
}

func TestResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name      string
		errorCode int
		expected  bool
	}{
		{
			name:      "No error is success",
			errorCode: ErrorCodeNoError,
			expected:  true,
		},
		{
			name:      "Error code 1000 is not success",
			errorCode: ErrorCodeInvalidRequest,
			expected:  false,
		},
		{
			name:      "Error code 2000 is not success",
			errorCode: ErrorCodeInternalError,
			expected:  false,
		},
		{
			name:      "Error code 3000 is not success",
			errorCode: ErrorCodeEntityNotFound,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{ErrorCode: tt.errorCode}
			assert.Equal(t, tt.expected, resp.IsSuccess())
		})
	}
}

func TestResponse_Structure(t *testing.T) {
	resp := &Response{
		ErrorCode:             ErrorCodeNoError,
		ErrorMessage:          "",
		ErrorMessageLocalised: "",
		Data: map[string]interface{}{
			"version": "1.0.0",
			"build":   "12345",
		},
	}

	assert.Equal(t, ErrorCodeNoError, resp.ErrorCode)
	assert.Empty(t, resp.ErrorMessage)
	assert.Empty(t, resp.ErrorMessageLocalised)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, "1.0.0", resp.Data["version"])
	assert.Equal(t, "12345", resp.Data["build"])
	assert.True(t, resp.IsSuccess())
}
