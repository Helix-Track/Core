package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCodeConstants(t *testing.T) {
	// Test that error code ranges are correct
	assert.Equal(t, -1, ErrorCodeNoError)

	// Request errors (100X)
	assert.Equal(t, 1000, ErrorCodeInvalidRequest)
	assert.Equal(t, 1001, ErrorCodeInvalidAction)
	assert.Equal(t, 1002, ErrorCodeMissingJWT)
	assert.Equal(t, 1003, ErrorCodeInvalidJWT)
	assert.Equal(t, 1004, ErrorCodeMissingObject)
	assert.Equal(t, 1005, ErrorCodeInvalidObject)
	assert.Equal(t, 1006, ErrorCodeMissingData)
	assert.Equal(t, 1007, ErrorCodeInvalidData)
	assert.Equal(t, 1008, ErrorCodeUnauthorized)
	assert.Equal(t, 1009, ErrorCodeForbidden)

	// System errors (200X)
	assert.Equal(t, 2000, ErrorCodeInternalError)
	assert.Equal(t, 2001, ErrorCodeDatabaseError)
	assert.Equal(t, 2002, ErrorCodeServiceUnavailable)
	assert.Equal(t, 2003, ErrorCodeConfigurationError)
	assert.Equal(t, 2004, ErrorCodeAuthServiceError)
	assert.Equal(t, 2005, ErrorCodePermissionServiceError)
	assert.Equal(t, 2006, ErrorCodeExtensionServiceError)

	// Entity errors (300X)
	assert.Equal(t, 3000, ErrorCodeEntityNotFound)
	assert.Equal(t, 3001, ErrorCodeEntityAlreadyExists)
	assert.Equal(t, 3002, ErrorCodeEntityValidationFailed)
	assert.Equal(t, 3003, ErrorCodeEntityDeleteFailed)
	assert.Equal(t, 3004, ErrorCodeEntityUpdateFailed)
	assert.Equal(t, 3005, ErrorCodeEntityCreateFailed)
}

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "No error message",
			code:     ErrorCodeNoError,
			expected: "Success",
		},
		{
			name:     "Invalid request message",
			code:     ErrorCodeInvalidRequest,
			expected: "Invalid request",
		},
		{
			name:     "Invalid action message",
			code:     ErrorCodeInvalidAction,
			expected: "Invalid action",
		},
		{
			name:     "Missing JWT message",
			code:     ErrorCodeMissingJWT,
			expected: "Missing JWT token",
		},
		{
			name:     "Invalid JWT message",
			code:     ErrorCodeInvalidJWT,
			expected: "Invalid JWT token",
		},
		{
			name:     "Missing object message",
			code:     ErrorCodeMissingObject,
			expected: "Missing object type",
		},
		{
			name:     "Invalid object message",
			code:     ErrorCodeInvalidObject,
			expected: "Invalid object type",
		},
		{
			name:     "Missing data message",
			code:     ErrorCodeMissingData,
			expected: "Missing required data",
		},
		{
			name:     "Invalid data message",
			code:     ErrorCodeInvalidData,
			expected: "Invalid data format",
		},
		{
			name:     "Unauthorized message",
			code:     ErrorCodeUnauthorized,
			expected: "Unauthorized",
		},
		{
			name:     "Forbidden message",
			code:     ErrorCodeForbidden,
			expected: "Forbidden - insufficient permissions",
		},
		{
			name:     "Internal error message",
			code:     ErrorCodeInternalError,
			expected: "Internal server error",
		},
		{
			name:     "Database error message",
			code:     ErrorCodeDatabaseError,
			expected: "Database error",
		},
		{
			name:     "Service unavailable message",
			code:     ErrorCodeServiceUnavailable,
			expected: "Service unavailable",
		},
		{
			name:     "Configuration error message",
			code:     ErrorCodeConfigurationError,
			expected: "Configuration error",
		},
		{
			name:     "Auth service error message",
			code:     ErrorCodeAuthServiceError,
			expected: "Authentication service error",
		},
		{
			name:     "Permission service error message",
			code:     ErrorCodePermissionServiceError,
			expected: "Permission service error",
		},
		{
			name:     "Extension service error message",
			code:     ErrorCodeExtensionServiceError,
			expected: "Extension service error",
		},
		{
			name:     "Entity not found message",
			code:     ErrorCodeEntityNotFound,
			expected: "Entity not found",
		},
		{
			name:     "Entity already exists message",
			code:     ErrorCodeEntityAlreadyExists,
			expected: "Entity already exists",
		},
		{
			name:     "Entity validation failed message",
			code:     ErrorCodeEntityValidationFailed,
			expected: "Entity validation failed",
		},
		{
			name:     "Entity delete failed message",
			code:     ErrorCodeEntityDeleteFailed,
			expected: "Failed to delete entity",
		},
		{
			name:     "Entity update failed message",
			code:     ErrorCodeEntityUpdateFailed,
			expected: "Failed to update entity",
		},
		{
			name:     "Entity create failed message",
			code:     ErrorCodeEntityCreateFailed,
			expected: "Failed to create entity",
		},
		{
			name:     "Unknown error code",
			code:     9999,
			expected: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := GetErrorMessage(tt.code)
			assert.Equal(t, tt.expected, msg)
		})
	}
}

func TestErrorMessages_Completeness(t *testing.T) {
	// Ensure all error codes have messages
	errorCodes := []int{
		ErrorCodeNoError,
		ErrorCodeInvalidRequest,
		ErrorCodeInvalidAction,
		ErrorCodeMissingJWT,
		ErrorCodeInvalidJWT,
		ErrorCodeMissingObject,
		ErrorCodeInvalidObject,
		ErrorCodeMissingData,
		ErrorCodeInvalidData,
		ErrorCodeUnauthorized,
		ErrorCodeForbidden,
		ErrorCodeInternalError,
		ErrorCodeDatabaseError,
		ErrorCodeServiceUnavailable,
		ErrorCodeConfigurationError,
		ErrorCodeAuthServiceError,
		ErrorCodePermissionServiceError,
		ErrorCodeExtensionServiceError,
		ErrorCodeEntityNotFound,
		ErrorCodeEntityAlreadyExists,
		ErrorCodeEntityValidationFailed,
		ErrorCodeEntityDeleteFailed,
		ErrorCodeEntityUpdateFailed,
		ErrorCodeEntityCreateFailed,
	}

	for _, code := range errorCodes {
		t.Run("Error code has message", func(t *testing.T) {
			msg := GetErrorMessage(code)
			assert.NotEmpty(t, msg)
			assert.NotEqual(t, "Unknown error", msg)
		})
	}
}
