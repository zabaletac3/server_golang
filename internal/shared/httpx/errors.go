package httpx

import (
	"errors"
	"net/http"
	"strings"

	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

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
