package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "with wrapped error",
			appErr: &AppError{
				Code:    1001,
				Message: "validation failed",
				Err:     errors.New("field is required"),
			},
			expected: "validation failed: field is required",
		},
		{
			name: "without wrapped error",
			appErr: &AppError{
				Code:    1001,
				Message: "validation failed",
			},
			expected: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appErr.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	appErr := &AppError{
		Code:    1001,
		Message: "test error",
		Err:     wrappedErr,
	}

	unwrapped := appErr.Unwrap()
	assert.Equal(t, wrappedErr, unwrapped)
}

func TestNewAppError(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	appErr := NewAppError(1001, "test message", wrappedErr)

	assert.Equal(t, 1001, appErr.Code)
	assert.Equal(t, "test message", appErr.Message)
	assert.Equal(t, wrappedErr, appErr.Err)
}

func TestErrorFactoryFunctions(t *testing.T) {
	tests := []struct {
		name         string
		createError  func() error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "ErrValidationFailed",
			createError:  func() error { return ErrValidationFailed("test validation") },
			expectedCode: ErrCodeValidationFailed,
			expectedMsg:  "test validation",
		},
		{
			name:         "ErrResourceNotFound",
			createError:  func() error { return ErrResourceNotFound("user") },
			expectedCode: ErrCodeNotFound,
			expectedMsg:  "user not found",
		},
		{
			name:         "ErrResourceAlreadyExists",
			createError:  func() error { return ErrResourceAlreadyExists("user") },
			expectedCode: ErrCodeAlreadyExists,
			expectedMsg:  "user already exists",
		},
		{
			name:         "ErrDatabase",
			createError:  func() error { return ErrDatabase(errors.New("db error")) },
			expectedCode: ErrCodeDatabaseError,
			expectedMsg:  "database operation failed",
		},
		{
			name:         "ErrCache",
			createError:  func() error { return ErrCache(errors.New("cache error")) },
			expectedCode: ErrCodeCacheError,
			expectedMsg:  "cache operation failed",
		},
		{
			name:         "ErrAuth",
			createError:  func() error { return ErrAuth("invalid credentials") },
			expectedCode: ErrCodeUnauthorized,
			expectedMsg:  "invalid credentials",
		},
		{
			name:         "ErrAccessDenied",
			createError:  func() error { return ErrAccessDenied("insufficient permissions") },
			expectedCode: ErrCodeForbidden,
			expectedMsg:  "insufficient permissions",
		},
		{
			name:         "ErrInternal",
			createError:  func() error { return ErrInternal(errors.New("internal error")) },
			expectedCode: ErrCodeInternalError,
			expectedMsg:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()
			assert.True(t, IsAppError(err))

			appErr := GetAppError(err)
			assert.NotNil(t, appErr)
			assert.Equal(t, tt.expectedCode, appErr.Code)
			assert.Contains(t, appErr.Message, tt.expectedMsg)
		})
	}
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is AppError",
			err:      NewAppError(1001, "test", nil),
			expected: true,
		},
		{
			name:     "is not AppError",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAppError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected *AppError
	}{
		{
			name:     "is AppError",
			err:      NewAppError(1001, "test", nil),
			expected: &AppError{Code: 1001, Message: "test"},
		},
		{
			name:     "is not AppError",
			err:      errors.New("regular error"),
			expected: nil,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAppError(tt.err)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Code, result.Code)
				assert.Equal(t, tt.expected.Message, result.Message)
			}
		})
	}
}
