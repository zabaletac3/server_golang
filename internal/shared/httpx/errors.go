package httpx

import (
	"errors"
	"net/http"

	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func FromError(err error) (int, ErrorResponse) {
	switch {
	case errors.Is(err, sharedErrors.ErrInvalidInput):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_INPUT",
			Message: err.Error(),
		}

	case errors.Is(err, sharedErrors.ErrBadRequest):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		}

	case errors.Is(err, sharedErrors.ErrUnauthorized):
		return http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "unauthorized",
		}

	case errors.Is(err, sharedErrors.ErrNotFound):
		return http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: err.Error(),
		}

	case errors.Is(err, sharedErrors.ErrConflict):
		return http.StatusConflict, ErrorResponse{
			Code:    "CONFLICT",
			Message: err.Error(),
		}

	default:
		return http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		}
	}
}
