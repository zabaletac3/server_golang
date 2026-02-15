package resources

import "errors"

var (
	ErrResourceNotFound   = errors.New("resource not found")
	ErrResourceNameExists = errors.New("resource name already exists")
	ErrInvalidResourceID  = errors.New("invalid resource id")
)
