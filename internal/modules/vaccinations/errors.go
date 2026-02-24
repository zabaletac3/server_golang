package vaccinations

import (
	"errors"
	"fmt"
)

// Module errors
var (
	ErrVaccinationNotFound    = errors.New("vaccination not found")
	ErrVaccineNotFound        = errors.New("vaccine not found")
	ErrVaccineNameExists      = errors.New("vaccine name already exists")
	ErrPatientNotFound        = errors.New("patient not found")
	ErrPatientInactive        = errors.New("patient is inactive")
	ErrVeterinarianNotFound   = errors.New("veterinarian not found")
	ErrInvalidApplicationDate = errors.New("application date cannot be in the future")
	ErrInvalidNextDueDate     = errors.New("next due date must be after application date")
	ErrInvalidStatus          = errors.New("invalid vaccination status")
	ErrInvalidDoseType        = errors.New("invalid dose type")
	ErrSpeciesMismatch        = errors.New("vaccine is not for this species")
	ErrCertificateNotFound    = errors.New("certificate not found")
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
	ErrVaccinationOverdue = ErrBusiness("VACCINATION_OVERDUE", "vaccination is overdue")
)
