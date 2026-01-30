package payments

import "errors"

var (
	ErrPaymentNotFound   = errors.New("payment not found")
	ErrInvalidPaymentID  = errors.New("invalid payment ID")
	ErrInvalidTenantID   = errors.New("invalid tenant ID")
)
