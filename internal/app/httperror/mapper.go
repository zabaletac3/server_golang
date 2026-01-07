package httperror

import (
	"errors"
	"net/http"

	domainErrors "github.com/eren_dev/go_server/internal/domain/errors"
)

func FromError(err error) (int, ErrorResponse) {
	switch {
	case errors.Is(err, domainErrors.ErrInvalidInput):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_INPUT",
			Message: err.Error(),
		}

	case errors.Is(err, domainErrors.ErrInvalidInput):
		return http.StatusBadRequest, ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		}

	case errors.Is(err, domainErrors.ErrUnauthorized):
		return http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "unauthorized",
		}

	case errors.Is(err, domainErrors.ErrNotFound):
		return http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: err.Error(),
		}

	case errors.Is(err, domainErrors.ErrConflict):
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
