package notifications

import "errors"

var (
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrInvalidNotificationID = errors.New("invalid notification id")
)
