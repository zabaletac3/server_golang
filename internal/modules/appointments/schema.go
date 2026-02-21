package appointments

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Appointment represents an appointment in the veterinary system
type Appointment struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`
	TenantIds []primitive.ObjectID `bson:"tenant_ids,omitempty"`

	// Core appointment data
	PatientID      primitive.ObjectID `bson:"patient_id"`
	OwnerID        primitive.ObjectID `bson:"owner_id"`
	VeterinarianID primitive.ObjectID `bson:"veterinarian_id"`

	// Scheduling
	ScheduledAt time.Time `bson:"scheduled_at"`
	Duration    int       `bson:"duration"` // minutes

	// Appointment details
	Type     string `bson:"type"`     // consultation, surgery, vaccination, etc.
	Status   string `bson:"status"`   // scheduled, confirmed, in_progress, completed, cancelled, no_show
	Priority string `bson:"priority"` // low, normal, high, emergency

	// Notes and observations
	Reason     string `bson:"reason"`                // Reason for visit
	Notes      string `bson:"notes,omitempty"`       // Staff notes
	OwnerNotes string `bson:"owner_notes,omitempty"` // Owner notes

	// Status tracking
	ConfirmedAt  *time.Time `bson:"confirmed_at,omitempty"`
	StartedAt    *time.Time `bson:"started_at,omitempty"`
	CompletedAt  *time.Time `bson:"completed_at,omitempty"`
	CancelledAt  *time.Time `bson:"cancelled_at,omitempty"`
	CancelReason string     `bson:"cancel_reason,omitempty"`

	// Standard fields
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
}

// AppointmentStatusTransition tracks status changes for audit purposes
type AppointmentStatusTransition struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty"`
	TenantIds     []primitive.ObjectID `bson:"tenant_ids,omitempty"`
	AppointmentID primitive.ObjectID   `bson:"appointment_id"`
	FromStatus    string               `bson:"from_status"`
	ToStatus      string               `bson:"to_status"`
	ChangedBy     primitive.ObjectID   `bson:"changed_by"`
	Reason        string               `bson:"reason,omitempty"`
	CreatedAt     time.Time            `bson:"created_at"`
}

// AppointmentType constants
const (
	AppointmentTypeConsultation = "consultation"
	AppointmentTypeSurgery      = "surgery"
	AppointmentTypeVaccination  = "vaccination"
	AppointmentTypeEmergency    = "emergency"
	AppointmentTypeCheckup      = "checkup"
	AppointmentTypeGrooming     = "grooming"
)

// AppointmentStatus constants
const (
	AppointmentStatusScheduled  = "scheduled"
	AppointmentStatusConfirmed  = "confirmed"
	AppointmentStatusInProgress = "in_progress"
	AppointmentStatusCompleted  = "completed"
	AppointmentStatusCancelled  = "cancelled"
	AppointmentStatusNoShow     = "no_show"
)

// AppointmentPriority constants
const (
	AppointmentPriorityLow       = "low"
	AppointmentPriorityNormal    = "normal"
	AppointmentPriorityHigh      = "high"
	AppointmentPriorityEmergency = "emergency"
)

// Valid status transitions map
var ValidStatusTransitions = map[string][]string{
	AppointmentStatusScheduled:  {AppointmentStatusConfirmed, AppointmentStatusCancelled, AppointmentStatusNoShow},
	AppointmentStatusConfirmed:  {AppointmentStatusInProgress, AppointmentStatusCancelled, AppointmentStatusNoShow},
	AppointmentStatusInProgress: {AppointmentStatusCompleted, AppointmentStatusCancelled},
	AppointmentStatusCompleted:  {}, // Terminal status
	AppointmentStatusCancelled:  {}, // Terminal status
	AppointmentStatusNoShow:     {}, // Terminal status
}

// GetValidNextStatuses returns the valid statuses that can be transitioned to from current status
func (a *Appointment) GetValidNextStatuses() []string {
	return ValidStatusTransitions[a.Status]
}

// CanTransitionTo checks if the appointment can transition to the given status
func (a *Appointment) CanTransitionTo(status string) bool {
	validStatuses := ValidStatusTransitions[a.Status]
	for _, validStatus := range validStatuses {
		if validStatus == status {
			return true
		}
	}
	return false
}

// IsActive returns true if the appointment is in an active state (not cancelled, completed, or no-show)
func (a *Appointment) IsActive() bool {
	return a.Status != AppointmentStatusCancelled &&
		a.Status != AppointmentStatusCompleted &&
		a.Status != AppointmentStatusNoShow
}

// IsUpcoming returns true if the appointment is scheduled in the future
func (a *Appointment) IsUpcoming() bool {
	return a.ScheduledAt.After(time.Now()) && a.IsActive()
}

// IsPast returns true if the appointment was scheduled in the past
func (a *Appointment) IsPast() bool {
	return a.ScheduledAt.Before(time.Now())
}
