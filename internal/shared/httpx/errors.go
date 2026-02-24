package httpx

import (
	"errors"
	"net/http"
	"strings"
	"time"

	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	RequestID string            `json:"request_id"`
	Timestamp string            `json:"timestamp"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Path      string            `json:"path"`
	Details   map[string]string `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response with request ID
func NewErrorResponse(requestID, path, code, message string) ErrorResponse {
	return ErrorResponse{
		RequestID: requestID,
		Timestamp: time.Now().Format(time.RFC3339),
		Code:      code,
		Message:   message,
		Path:      path,
	}
}

// FromError converts an error to HTTP status code and ErrorResponse
func FromError(err error) (int, ErrorResponse) {
	errMsg := err.Error()

	if strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "binding") {
		return http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: errMsg,
		}
	}

	if strings.Contains(errMsg, "not found") {
		return http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: errMsg,
		}
	}

	if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
		return http.StatusConflict, ErrorResponse{
			Code:    "CONFLICT",
			Message: errMsg,
		}
	}

	if strings.Contains(errMsg, "invalid credentials") {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: errMsg,
		}
	}

	if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "access denied") {
		return http.StatusForbidden, ErrorResponse{
			Code:    "FORBIDDEN",
			Message: errMsg,
		}
	}

	if strings.Contains(errMsg, "token") && (strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "expired")) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: errMsg,
		}
	}

	if strings.Contains(errMsg, "invalid") {
		return http.StatusBadRequest, ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: errMsg,
		}
	}

	switch {
	case errors.Is(err, sharedErrors.ErrInvalidInput):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_INPUT",
			Message: errMsg,
		}

	case errors.Is(err, sharedErrors.ErrBadRequest):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: errMsg,
		}

	case errors.Is(err, sharedErrors.ErrUnauthorized):
		return http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "unauthorized",
		}

	case errors.Is(err, sharedErrors.ErrNotFound):
		return http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: errMsg,
		}

	case errors.Is(err, sharedErrors.ErrConflict):
		return http.StatusConflict, ErrorResponse{
			Code:    "CONFLICT",
			Message: errMsg,
		}

	case errors.Is(err, sharedErrors.ErrForbidden):
		return http.StatusForbidden, ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "access denied",
		}

	default:
		return http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		}
	}
}

// Error codes for standard responses
const (
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeInvalidInput    = "INVALID_INPUT"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)
