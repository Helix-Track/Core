package models

// Error code constants following the specification
const (
	// No error
	ErrorCodeNoError = -1

	// Request-related errors (100X)
	ErrorCodeInvalidRequest = 1000
	ErrorCodeInvalidAction  = 1001
	ErrorCodeMissingJWT     = 1002
	ErrorCodeInvalidJWT     = 1003
	ErrorCodeMissingObject  = 1004
	ErrorCodeInvalidObject  = 1005
	ErrorCodeMissingData    = 1006
	ErrorCodeInvalidData    = 1007
	ErrorCodeUnauthorized   = 1008
	ErrorCodeForbidden      = 1009

	// System-related errors (200X)
	ErrorCodeInternalError          = 2000
	ErrorCodeDatabaseError          = 2001
	ErrorCodeServiceUnavailable     = 2002
	ErrorCodeConfigurationError     = 2003
	ErrorCodeAuthServiceError       = 2004
	ErrorCodePermissionServiceError = 2005
	ErrorCodeExtensionServiceError  = 2006

	// Entity-related errors (300X)
	ErrorCodeEntityNotFound         = 3000
	ErrorCodeEntityAlreadyExists    = 3001
	ErrorCodeEntityValidationFailed = 3002
	ErrorCodeEntityDeleteFailed     = 3003
	ErrorCodeEntityUpdateFailed     = 3004
	ErrorCodeEntityCreateFailed     = 3005
	ErrorCodeVersionConflict        = 3006
)

// ErrorMessages provides default English error messages
var ErrorMessages = map[int]string{
	ErrorCodeNoError:                "Success",
	ErrorCodeInvalidRequest:         "Invalid request",
	ErrorCodeInvalidAction:          "Invalid action",
	ErrorCodeMissingJWT:             "Missing JWT token",
	ErrorCodeInvalidJWT:             "Invalid JWT token",
	ErrorCodeMissingObject:          "Missing object type",
	ErrorCodeInvalidObject:          "Invalid object type",
	ErrorCodeMissingData:            "Missing required data",
	ErrorCodeInvalidData:            "Invalid data format",
	ErrorCodeUnauthorized:           "Unauthorized",
	ErrorCodeForbidden:              "Forbidden - insufficient permissions",
	ErrorCodeInternalError:          "Internal server error",
	ErrorCodeDatabaseError:          "Database error",
	ErrorCodeServiceUnavailable:     "Service unavailable",
	ErrorCodeConfigurationError:     "Configuration error",
	ErrorCodeAuthServiceError:       "Authentication service error",
	ErrorCodePermissionServiceError: "Permission service error",
	ErrorCodeExtensionServiceError:  "Extension service error",
	ErrorCodeEntityNotFound:         "Entity not found",
	ErrorCodeEntityAlreadyExists:    "Entity already exists",
	ErrorCodeEntityValidationFailed: "Entity validation failed",
	ErrorCodeEntityDeleteFailed:     "Failed to delete entity",
	ErrorCodeEntityUpdateFailed:     "Failed to update entity",
	ErrorCodeEntityCreateFailed:     "Failed to create entity",
	ErrorCodeVersionConflict:        "Version conflict - entity was modified by another user",
}

// GetErrorMessage returns the error message for a given error code
func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}
