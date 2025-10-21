package models

// Response represents the unified API response format
type Response struct {
	ErrorCode              int                    `json:"errorCode"`                        // -1 means no error
	ErrorMessage           string                 `json:"errorMessage,omitempty"`           // Error message in English
	ErrorMessageLocalised  string                 `json:"errorMessageLocalised,omitempty"`  // Localized error message
	Data                   map[string]interface{} `json:"data,omitempty"`                   // Response data
}

// NewSuccessResponse creates a successful response with optional data
func NewSuccessResponse(data map[string]interface{}) *Response {
	return &Response{
		ErrorCode: ErrorCodeNoError,
		Data:      data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code int, message string, localisedMessage string) *Response {
	return &Response{
		ErrorCode:             code,
		ErrorMessage:          message,
		ErrorMessageLocalised: localisedMessage,
	}
}

// IsSuccess returns true if the response indicates success
func (r *Response) IsSuccess() bool {
	return r.ErrorCode == ErrorCodeNoError
}

// NewLocalizedErrorResponse creates an error response with localized message
// It uses the error code to look up both the default message and localization key
func NewLocalizedErrorResponse(code int, locale string, localizedMessage string) *Response {
	defaultMessage := GetErrorMessage(code)

	return &Response{
		ErrorCode:             code,
		ErrorMessage:          defaultMessage,
		ErrorMessageLocalised: localizedMessage,
	}
}

// NewErrorResponseFromCode creates an error response from just an error code
// This uses the default English message from the ErrorMessages map
func NewErrorResponseFromCode(code int) *Response {
	message := GetErrorMessage(code)

	return &Response{
		ErrorCode:    code,
		ErrorMessage: message,
	}
}
