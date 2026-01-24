package tenant

import "errors"

var (
	ErrTenantNotFound  = errors.New("tenant not found")
	ErrInvalidTenantID = errors.New("invalid tenant id")
	ErrOwnerNotFound   = errors.New("owner not found")
	ErrInvalidOwnerID  = errors.New("invalid owner id")
)
