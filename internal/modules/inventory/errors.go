package inventory

import (
	"errors"
	"fmt"
)

// Module errors
var (
	ErrProductNotFound       = errors.New("product not found")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryNameExists    = errors.New("category name already exists")
	ErrSKUAlreadyExists      = errors.New("SKU already exists")
	ErrBarcodeAlreadyExists  = errors.New("barcode already exists")
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrInvalidStockMovement  = errors.New("invalid stock movement")
	ErrCannotDeleteWithStock = errors.New("cannot delete product with stock > 0")
	ErrInvalidCategory       = errors.New("invalid category")
	ErrInvalidUnit           = errors.New("invalid unit")
	ErrInvalidPrice          = errors.New("invalid price: must be >= 0")
	ErrInvalidQuantity       = errors.New("invalid quantity: must be > 0")
	ErrSalePriceTooLow       = errors.New("sale price must be >= purchase price")
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// ErrValidation creates a new validation error
func ErrValidation(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// BusinessError represents a business logic error
type BusinessError struct {
	Code    string
	Message string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("business error: %s - %s", e.Code, e.Message)
}

// ErrBusiness creates a new business error
func ErrBusiness(code, message string) error {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// Specific business errors
var (
	ErrProductExpired = ErrBusiness("PRODUCT_EXPIRED", "product has expired")
)
