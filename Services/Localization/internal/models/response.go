package models

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse creates a success response
func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(code int, message string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// LocalizationResponse represents a single localization response
type LocalizationResponse struct {
	Key      string            `json:"key"`
	Language string            `json:"language"`
	Value    string            `json:"value"`
	Variables []string         `json:"variables,omitempty"`
	Approved bool              `json:"approved"`
}

// BatchLocalizationResponse represents a batch localization response
type BatchLocalizationResponse struct {
	Language      string            `json:"language"`
	Localizations map[string]string `json:"localizations"`
}

// LanguageListResponse represents a list of languages response
type LanguageListResponse struct {
	Languages []Language `json:"languages"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string                 `json:"status"`
	Version string                 `json:"version"`
	Checks  map[string]interface{} `json:"checks"`
}
