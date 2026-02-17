package owners

import "errors"

var (
	ErrOwnerNotFound  = errors.New("owner not found")
	ErrEmailExists    = errors.New("email already exists")
	ErrInvalidOwnerID = errors.New("invalid owner id")
)
