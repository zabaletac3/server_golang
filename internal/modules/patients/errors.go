package patients

import "errors"

var (
	ErrPatientNotFound  = errors.New("patient not found")
	ErrInvalidPatientID = errors.New("invalid patient id")
	ErrInvalidOwnerID   = errors.New("invalid owner id")
	ErrInvalidSpeciesID = errors.New("invalid species id")
	ErrInvalidTenantID  = errors.New("invalid tenant id")
	ErrMicrochipExists  = errors.New("microchip already exists")
	ErrSpeciesNotFound  = errors.New("species not found")
	ErrSpeciesConflict  = errors.New("similar species already exists")
)
