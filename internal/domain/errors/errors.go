package errors

import "errors"

var (
	ErrInvalidInput = errors.New("invalid_input")
	ErrBadRequest   = errors.New("bad_request")
	ErrNotFound     = errors.New("not_found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInternal     = errors.New("internal_error")
)
