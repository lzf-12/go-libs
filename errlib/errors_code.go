package errlib

import "net/http"

type ErrorCode string

// category
const (
	// general
	ErrInternal           ErrorCode = "ERR_INTERNAL"
	ErrServiceUnavailable ErrorCode = "ERR_SERVICE_UNAVAILABLE"
	ErrTimeout            ErrorCode = "ERR_TIMEOUT"
	ErrBadRequest         ErrorCode = "ERR_INVALID_INPUT"

	// validation
	ErrInvalidInput  ErrorCode = "ERR_INVALID_INPUT"
	ErrMissingField  ErrorCode = "ERR_MISSING_FIELD"
	ErrInvalidFormat ErrorCode = "ERR_INVALID_FORMAT"
	ErrOutOfRange    ErrorCode = "ERR_OUT_OF_RANGE"

	// authentication & authorization
	ErrUnauthorized   ErrorCode = "ERR_UNAUTHORIZED"
	ErrForbidden      ErrorCode = "ERR_FORBIDDEN"
	ErrSessionExpired ErrorCode = "ERR_SESSION_EXPIRED"

	// db
	ErrDBConnection ErrorCode = "ERR_DB_CONNECTION"
	ErrDBTimeout    ErrorCode = "ERR_DB_TIMEOUT"
	ErrNotFound     ErrorCode = "ERR_NOT_FOUND"
	ErrDuplicate    ErrorCode = "ERR_DUPLICATE"

	// external services
	ErrAPIUnavailable ErrorCode = "ERR_API_UNAVAILABLE"
	ErrAPIError       ErrorCode = "ERR_API_ERROR"
	ErrRateLimited    ErrorCode = "ERR_RATE_LIMITED"

	// network
	ErrNetwork       ErrorCode = "ERR_NETWORK"
	ErrDNSResolution ErrorCode = "ERR_DNS_RESOLUTION"
)

// http status map
func ErrorCodeToHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrInvalidInput, ErrMissingField, ErrInvalidFormat:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrDuplicate:
		return http.StatusConflict
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
