package models

import (
	"errors"
	"fmt"
)

// Error codes
const (
	ErrCodeValidationFailed  = 1001
	ErrCodeNotFound          = 1002
	ErrCodeAlreadyExists     = 1003
	ErrCodeDatabaseError     = 2001
	ErrCodeCacheError        = 2002
	ErrCodeUnauthorized      = 3001
	ErrCodeForbidden         = 3002
	ErrCodeInvalidToken      = 3003
	ErrCodeExpiredToken      = 3004
	ErrCodeInternalError     = 5001
)

// Predefined errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("expired token")
	ErrDatabaseError     = errors.New("database error")
	ErrCacheError        = errors.New("cache error")
	ErrInternalError     = errors.New("internal server error")
)

// AppError represents an application error with code and message
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error factory functions
func ErrValidationFailed(message string) error {
	return NewAppError(ErrCodeValidationFailed, message, nil)
}

func ErrResourceNotFound(resource string) error {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), ErrNotFound)
}

func ErrResourceAlreadyExists(resource string) error {
	return NewAppError(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), ErrAlreadyExists)
}

func ErrDatabase(err error) error {
	return NewAppError(ErrCodeDatabaseError, "database operation failed", err)
}

func ErrCache(err error) error {
	return NewAppError(ErrCodeCacheError, "cache operation failed", err)
}

func ErrAuth(message string) error {
	return NewAppError(ErrCodeUnauthorized, message, ErrUnauthorized)
}

func ErrAccessDenied(message string) error {
	return NewAppError(ErrCodeForbidden, message, ErrForbidden)
}

func ErrInternal(err error) error {
	return NewAppError(ErrCodeInternalError, "internal server error", err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError returns the AppError if it exists
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
