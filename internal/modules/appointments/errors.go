package appointments

import "errors"

var (
	// General appointment errors
	ErrAppointmentNotFound      = errors.New("appointment not found")
	ErrAppointmentAlreadyExists = errors.New("appointment already exists at this time")

	// Validation errors
	ErrInvalidAppointmentTime = errors.New("invalid appointment time")
	ErrPastAppointmentTime    = errors.New("cannot schedule appointment in the past")
	ErrInvalidDuration        = errors.New("invalid appointment duration")
	ErrInvalidTimeRange       = errors.New("invalid time range for appointment")
	ErrOutsideBusinessHours   = errors.New("appointment must be scheduled during business hours")

	// Conflict errors
	ErrAppointmentConflict      = errors.New("appointment time conflicts with existing appointment")
	ErrVeterinarianNotAvailable = errors.New("veterinarian is not available at the requested time")
	ErrPatientNotAvailable      = errors.New("patient already has an appointment at this time")

	// Status transition errors
	ErrInvalidStatusTransition     = errors.New("invalid status transition")
	ErrAppointmentAlreadyStarted   = errors.New("appointment already started")
	ErrAppointmentAlreadyCompleted = errors.New("appointment already completed")
	ErrAppointmentAlreadyCancelled = errors.New("appointment already cancelled")
	ErrAppointmentNotConfirmed     = errors.New("appointment must be confirmed before starting")
	ErrCannotCancelPastAppointment = errors.New("cannot cancel past appointments")

	// Business logic errors
	ErrPatientNotFound        = errors.New("patient not found for appointment")
	ErrOwnerNotFound          = errors.New("owner not found for appointment")
	ErrVeterinarianNotFound   = errors.New("veterinarian not found")
	ErrInvalidAppointmentType = errors.New("invalid appointment type")
	ErrInvalidPriority        = errors.New("invalid appointment priority")

	// Permission errors
	ErrUnauthorizedAccess      = errors.New("unauthorized access to appointment")
	ErrInsufficientPermissions = errors.New("insufficient permissions to perform this action")
	ErrOwnerMismatch           = errors.New("appointment does not belong to this owner")

	// Mobile-specific errors
	ErrAppointmentRequestLimit = errors.New("appointment request limit reached")
	ErrTooManyPendingRequests  = errors.New("too many pending appointment requests")
	ErrRequestTooSoon          = errors.New("cannot request appointment with less than 24 hours notice")

	// System errors
	ErrDatabaseConnection  = errors.New("database connection error")
	ErrNotificationFailed  = errors.New("failed to send notification")
	ErrInternalServerError = errors.New("internal server error")
)

// AppointmentError wraps appointment-specific errors with additional context
type AppointmentError struct {
	Code    string
	Message string
	Details map[string]interface{}
	Err     error
}

func (e AppointmentError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e AppointmentError) Unwrap() error {
	return e.Err
}

// NewAppointmentError creates a new AppointmentError
func NewAppointmentError(code, message string, details map[string]interface{}, err error) *AppointmentError {
	return &AppointmentError{
		Code:    code,
		Message: message,
		Details: details,
		Err:     err,
	}
}

// Common error constructors
func ErrConflictingAppointment(conflictTime string) *AppointmentError {
	return NewAppointmentError(
		"APPOINTMENT_CONFLICT",
		"Appointment conflicts with existing booking",
		map[string]interface{}{
			"conflict_time": conflictTime,
		},
		ErrAppointmentConflict,
	)
}

func ErrInvalidStatus(currentStatus, targetStatus string) *AppointmentError {
	return NewAppointmentError(
		"INVALID_STATUS_TRANSITION",
		"Cannot transition from current status to target status",
		map[string]interface{}{
			"current_status": currentStatus,
			"target_status":  targetStatus,
		},
		ErrInvalidStatusTransition,
	)
}

func ErrResourceNotFound(resourceType, resourceID string) *AppointmentError {
	return NewAppointmentError(
		"RESOURCE_NOT_FOUND",
		resourceType+" not found",
		map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   resourceID,
		},
		nil,
	)
}

func ErrValidationFailed(field, reason string) *AppointmentError {
	return NewAppointmentError(
		"VALIDATION_ERROR",
		"Validation failed for field: "+field,
		map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
		nil,
	)
}
