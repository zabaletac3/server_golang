package laboratory

import (
	"errors"
	"fmt"
)

// Module errors
var (
	ErrLabOrderNotFound       = errors.New("lab order not found")
	ErrLabTestNotFound        = errors.New("lab test not found")
	ErrLabTestNameExists      = errors.New("lab test name already exists")
	ErrPatientNotFound        = errors.New("patient not found")
	ErrPatientInactive        = errors.New("patient is inactive")
	ErrVeterinarianNotFound   = errors.New("veterinarian not found")
	ErrInvalidStatus          = errors.New("invalid lab order status")
	ErrInvalidTestType        = errors.New("invalid test type")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrResultAlreadyUploaded  = errors.New("result already uploaded")
	ErrResultRequired         = errors.New("result file required for processed status")
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
	ErrLabOrderOverdue = ErrBusiness("LAB_ORDER_OVERDUE", "lab order is overdue")
)

// Valid status transitions
var validStatusTransitions = map[LabOrderStatus][]LabOrderStatus{
	LabOrderStatusPending:   {LabOrderStatusCollected},
	LabOrderStatusCollected: {LabOrderStatusSent},
	LabOrderStatusSent:      {LabOrderStatusReceived},
	LabOrderStatusReceived:  {LabOrderStatusProcessed},
	LabOrderStatusProcessed: {}, // Terminal status
}

// IsValidStatusTransition checks if the status transition is valid
func IsValidStatusTransition(from, to LabOrderStatus) bool {
	validNextStates, ok := validStatusTransitions[from]
	if !ok {
		return false
	}

	for _, state := range validNextStates {
		if state == to {
			return true
		}
	}

	return false
}
