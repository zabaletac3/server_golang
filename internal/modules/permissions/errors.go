package permissions

import "errors"

var (
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrPermissionExists    = errors.New("permission already exists")
	ErrInvalidPermissionID = errors.New("invalid permission id")
	ErrInvalidAction       = errors.New("invalid action: must be get, post, put, patch or delete")
)
