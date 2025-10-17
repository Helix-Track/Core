package models

import "errors"

// Error codes
const (
	ErrorCodeSuccess            = -1
	ErrorCodeInvalidRequest     = 1001
	ErrorCodeMissingParameter   = 1002
	ErrorCodeInvalidParameter   = 1003
	ErrorCodeUnauthorized       = 1004
	ErrorCodeForbidden          = 1005
	ErrorCodeDatabaseError      = 2001
	ErrorCodeInternalError      = 2002
	ErrorCodeServiceUnavailable = 2003
	ErrorCodeNotFound           = 3001
	ErrorCodeAlreadyExists      = 3002
	ErrorCodeConflict           = 3003
	ErrorCodeRateLimitExceeded  = 4001
	ErrorCodeMessageTooLarge    = 4002
	ErrorCodeAttachmentTooLarge = 4003
)

// Predefined errors
var (
	ErrInvalidRequest          = errors.New("invalid request")
	ErrMissingParameter        = errors.New("missing required parameter")
	ErrInvalidParameter        = errors.New("invalid parameter")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrDatabaseError           = errors.New("database error")
	ErrInternalError           = errors.New("internal server error")
	ErrServiceUnavailable      = errors.New("service unavailable")
	ErrNotFound                = errors.New("not found")
	ErrAlreadyExists           = errors.New("already exists")
	ErrConflict                = errors.New("conflict")
	ErrRateLimitExceeded       = errors.New("rate limit exceeded")
	ErrMessageTooLarge         = errors.New("message too large")
	ErrAttachmentTooLarge      = errors.New("attachment too large")
	ErrInvalidChatRoomType     = errors.New("invalid chat room type")
	ErrInvalidMessageType      = errors.New("invalid message type")
	ErrInvalidContentFormat    = errors.New("invalid content format")
	ErrInvalidParticipantRole  = errors.New("invalid participant role")
	ErrInvalidPresenceStatus   = errors.New("invalid presence status")
	ErrReplyNeedsParent        = errors.New("reply message needs parent_id")
	ErrQuoteNeedsQuotedMessage = errors.New("quote message needs quoted_message_id")
	ErrNotParticipant          = errors.New("user is not a participant")
	ErrChatRoomArchived        = errors.New("chat room is archived")
	ErrMessageDeleted          = errors.New("message is deleted")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

// ErrorResponse creates an API error response
func ErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

// SuccessResponse creates an API success response
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		ErrorCode: ErrorCodeSuccess,
		Data:      data,
	}
}

// GetErrorCode returns the error code for a given error
func GetErrorCode(err error) int {
	switch err {
	case ErrInvalidRequest:
		return ErrorCodeInvalidRequest
	case ErrMissingParameter:
		return ErrorCodeMissingParameter
	case ErrInvalidParameter, ErrInvalidChatRoomType, ErrInvalidMessageType,
	     ErrInvalidContentFormat, ErrInvalidParticipantRole, ErrInvalidPresenceStatus,
	     ErrReplyNeedsParent, ErrQuoteNeedsQuotedMessage:
		return ErrorCodeInvalidParameter
	case ErrUnauthorized:
		return ErrorCodeUnauthorized
	case ErrForbidden, ErrNotParticipant, ErrInsufficientPermissions:
		return ErrorCodeForbidden
	case ErrDatabaseError:
		return ErrorCodeDatabaseError
	case ErrInternalError:
		return ErrorCodeInternalError
	case ErrServiceUnavailable:
		return ErrorCodeServiceUnavailable
	case ErrNotFound:
		return ErrorCodeNotFound
	case ErrAlreadyExists:
		return ErrorCodeAlreadyExists
	case ErrConflict, ErrChatRoomArchived, ErrMessageDeleted:
		return ErrorCodeConflict
	case ErrRateLimitExceeded:
		return ErrorCodeRateLimitExceeded
	case ErrMessageTooLarge:
		return ErrorCodeMessageTooLarge
	case ErrAttachmentTooLarge:
		return ErrorCodeAttachmentTooLarge
	default:
		return ErrorCodeInternalError
	}
}
