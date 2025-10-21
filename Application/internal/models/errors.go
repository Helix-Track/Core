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

// ErrorCodeToLocalizationKey maps error codes to localization keys
var ErrorCodeToLocalizationKey = map[int]string{
	ErrorCodeNoError:                "error.success",
	ErrorCodeInvalidRequest:         "error.invalid_request",
	ErrorCodeInvalidAction:          "error.invalid_action",
	ErrorCodeMissingJWT:             "error.missing_jwt",
	ErrorCodeInvalidJWT:             "error.invalid_jwt",
	ErrorCodeMissingObject:          "error.missing_object",
	ErrorCodeInvalidObject:          "error.invalid_object",
	ErrorCodeMissingData:            "error.missing_data",
	ErrorCodeInvalidData:            "error.invalid_data",
	ErrorCodeUnauthorized:           "error.unauthorized",
	ErrorCodeForbidden:              "error.forbidden",
	ErrorCodeInternalError:          "error.internal_error",
	ErrorCodeDatabaseError:          "error.database_error",
	ErrorCodeServiceUnavailable:     "error.service_unavailable",
	ErrorCodeConfigurationError:     "error.configuration_error",
	ErrorCodeAuthServiceError:       "error.auth_service_error",
	ErrorCodePermissionServiceError: "error.permission_service_error",
	ErrorCodeExtensionServiceError:  "error.extension_service_error",
	ErrorCodeEntityNotFound:         "error.not_found",
	ErrorCodeEntityAlreadyExists:    "error.already_exists",
	ErrorCodeEntityValidationFailed: "error.validation_failed",
	ErrorCodeEntityDeleteFailed:     "error.delete_failed",
	ErrorCodeEntityUpdateFailed:     "error.update_failed",
	ErrorCodeEntityCreateFailed:     "error.create_failed",
	ErrorCodeVersionConflict:        "error.version_conflict",
}

// GetLocalizationKey returns the localization key for a given error code
func GetLocalizationKey(code int) string {
	if key, ok := ErrorCodeToLocalizationKey[code]; ok {
		return key
	}
	return "error.unknown"
}

// ErrorCodeFromKey returns the error code for a given localization key (fallback helper)
func ErrorCodeFromKey(key string) int {
	for code, k := range ErrorCodeToLocalizationKey {
		if k == key {
			return code
		}
	}
	return ErrorCodeInternalError
}
