package medical_records

import (
	"errors"
	"fmt"
)

// Module errors
var (
	ErrPatientNotFound       = errors.New("patient not found")
	ErrPatientInactive       = errors.New("patient is inactive")
	ErrVeterinarianNotFound  = errors.New("veterinarian not found")
	ErrOwnerNotFound         = errors.New("owner not found")
	ErrRecordNotFound        = errors.New("medical record not found")
	ErrRecordNotEditable     = errors.New("medical record cannot be edited after 24 hours")
	ErrAllergyNotFound       = errors.New("allergy not found")
	ErrMedicalHistoryNotFound = errors.New("medical history not found")
	ErrInvalidTemperature    = errors.New("invalid temperature: must be between 30 and 45°C")
	ErrInvalidWeight         = errors.New("invalid weight: must be greater than 0")
	ErrInvalidNextVisitDate  = errors.New("next visit date cannot be in the past")
	ErrInvalidMedicalRecordType = errors.New("invalid medical record type")
	ErrInvalidAllergySeverity   = errors.New("invalid allergy severity")
	ErrAttachmentNotFound    = errors.New("attachment not found")
	ErrDuplicateHistory      = errors.New("medical history already exists for this patient")
)

// Error types for validation
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

// Business logic errors
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
	ErrPatientHasActiveHospitalization = ErrBusiness("PATIENT_HOSPITALIZED", "patient is currently hospitalized")
	ErrSevereAllergyAlert             = ErrBusiness("SEVERE_ALLERGY", "patient has severe allergies - review before proceeding")
)
