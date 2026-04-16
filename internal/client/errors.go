package client

import (
	"fmt"
)

// BiliError represents standard normalized error types
type BiliError struct {
	Code    string
	Message string
	Details interface{}
}

func (e *BiliError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeNotAuthenticated = "not_authenticated"
	ErrCodePermissionDenied = "permission_denied"
	ErrCodeInvalidInput     = "invalid_input"
	ErrCodeNetworkError     = "network_error"
	ErrCodeUpstreamError    = "upstream_error"
	ErrCodeNotFound         = "not_found"
	ErrCodeRateLimited      = "rate_limited"
	ErrCodeInternalError    = "internal_error"
)

// NewError creates a new BiliError
func NewError(code, message string) *BiliError {
	return &BiliError{
		Code:    code,
		Message: message,
	}
}

// MapAPIError maps Bilibili API codes to normalized BiliError codes
func MapAPIError(action string, apiCode int, message string) error {
	var code string
	switch apiCode {
	case -101, -111:
		code = ErrCodeNotAuthenticated
	case -404, 62002, 62004:
		code = ErrCodeNotFound
	case -412, 412:
		code = ErrCodeRateLimited
	default:
		code = ErrCodeUpstreamError
	}

	return &BiliError{
		Code:    code,
		Message: fmt.Sprintf("%s: [%d] %s", action, apiCode, message),
	}
}

func IsAuthError(err error) bool {
	if bErr, ok := err.(*BiliError); ok {
		return bErr.Code == ErrCodeNotAuthenticated
	}
	return false
}
