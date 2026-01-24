package roles

import "errors"

var (
	ErrRoleNotFound  = errors.New("role not found")
	ErrInvalidRoleID = errors.New("invalid role id")
	ErrRoleExists    = errors.New("role already exists")
)
