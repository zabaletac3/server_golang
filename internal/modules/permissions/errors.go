package permissions

import "errors"

var (
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrInvalidPermissionID = errors.New("invalid permission id")
	ErrInvalidResource     = errors.New("invalid resource")
	ErrInvalidAction       = errors.New("invalid action")
	ErrPermissionExists    = errors.New("permission already exists")
)
