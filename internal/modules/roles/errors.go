package roles

import "errors"

var (
	ErrRoleNotFound   = errors.New("role not found")
	ErrRoleNameExists = errors.New("role name already exists")
	ErrInvalidRoleID  = errors.New("invalid role id")
)
