package appointments

import (
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Input DTOs

// CreateAppointmentDTO defines the structure for creating appointments
type CreateAppointmentDTO struct {
	PatientID      string    `json:"patient_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	VeterinarianID string    `json:"veterinarian_id" binding:"required" example:"507f1f77bcf86cd799439012"`
	ScheduledAt    time.Time `json:"scheduled_at" binding:"required" example:"2024-01-15T10:30:00Z"`
	Duration       int       `json:"duration" binding:"required,min=15,max=480" example:"30"`
	Type           string    `json:"type" binding:"required,oneof=consultation surgery vaccination emergency checkup grooming" example:"consultation"`
	Priority       string    `json:"priority" binding:"omitempty,oneof=low normal high emergency" example:"normal"`
	Reason         string    `json:"reason" binding:"required,max=500" example:"Annual checkup"`
	Notes          string    `json:"notes" binding:"omitempty,max=1000" example:"First visit for this patient"`
}

// UpdateAppointmentDTO defines the structure for updating appointments
type UpdateAppointmentDTO struct {
	ScheduledAt *time.Time `json:"scheduled_at" binding:"omitempty" example:"2024-01-15T11:00:00Z"`
	Duration    *int       `json:"duration" binding:"omitempty,min=15,max=480" example:"45"`
	Type        *string    `json:"type" binding:"omitempty,oneof=consultation surgery vaccination emergency checkup grooming" example:"surgery"`
	Priority    *string    `json:"priority" binding:"omitempty,oneof=low normal high emergency" example:"high"`
	Reason      *string    `json:"reason" binding:"omitempty,max=500" example:"Updated reason"`
	Notes       *string    `json:"notes" binding:"omitempty,max=1000" example:"Updated notes"`
}

// UpdateStatusDTO defines the structure for updating appointment status
type UpdateStatusDTO struct {
	Status string `json:"status" binding:"required,oneof=scheduled confirmed in_progress completed cancelled no_show" example:"confirmed"`
	Reason string `json:"reason" binding:"omitempty,max=200" example:"Patient confirmed by phone"`
}

// AppointmentCancelDTO defines the structure for cancelling an appointment
type AppointmentCancelDTO struct {
	Reason string `json:"reason" binding:"required,max=200" example:"Ya no necesito la cita"`
}

// MobileAppointmentRequestDTO defines the structure for mobile appointment requests
type MobileAppointmentRequestDTO struct {
	PatientID   string    `json:"patient_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required" example:"2024-01-15T10:30:00Z"`
	Type        string    `json:"type" binding:"required,oneof=consultation surgery vaccination emergency checkup grooming" example:"consultation"`
	Priority    string    `json:"priority" binding:"omitempty,oneof=low normal high emergency" example:"normal"`
	Reason      string    `json:"reason" binding:"required,max=500" example:"My pet is not feeling well"`
	OwnerNotes  string    `json:"owner_notes" binding:"omitempty,max=1000" example:"Additional information"`
}

// Response DTOs

// PatientSummary provides a summary of patient details
type PatientSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Species string `json:"species,omitempty"`
	Breed   string `json:"breed,omitempty"`
}

// OwnerSummary provides a summary of owner details
type OwnerSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// VeterinarianSummary provides a summary of veterinarian details
type VeterinarianSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// AppointmentResponse defines the structure for appointment responses
type AppointmentResponse struct {
	ID             string     `json:"id" example:"507f1f77bcf86cd799439011"`
	PatientID      string     `json:"patient_id" example:"507f1f77bcf86cd799439012"`
	OwnerID        string     `json:"owner_id" example:"507f1f77bcf86cd799439013"`
	VeterinarianID string     `json:"veterinarian_id" example:"507f1f77bcf86cd799439014"`
	ScheduledAt    time.Time  `json:"scheduled_at" example:"2024-01-15T10:30:00Z"`
	Duration       int        `json:"duration" example:"30"`
	Type           string     `json:"type" example:"consultation"`
	Status         string     `json:"status" example:"scheduled"`
	Priority       string     `json:"priority" example:"normal"`
	Reason         string     `json:"reason" example:"Annual checkup"`
	Notes          string     `json:"notes,omitempty" example:"First visit"`
	OwnerNotes     string     `json:"owner_notes,omitempty" example:"Patient anxious"`
	ConfirmedAt    *time.Time `json:"confirmed_at,omitempty" example:"2024-01-14T15:00:00Z"`
	StartedAt      *time.Time `json:"started_at,omitempty" example:"2024-01-15T10:35:00Z"`
	CompletedAt    *time.Time `json:"completed_at,omitempty" example:"2024-01-15T11:05:00Z"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
	CancelReason   string     `json:"cancel_reason,omitempty"`
	CreatedAt      time.Time  `json:"created_at" example:"2024-01-10T14:20:00Z"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2024-01-14T15:00:00Z"`

	// Populated data (will be filled when populate=true)
	Patient      *PatientSummary      `json:"patient,omitempty"`
	Owner        *OwnerSummary        `json:"owner,omitempty"`
	Veterinarian *VeterinarianSummary `json:"veterinarian,omitempty"`
}

// AppointmentStatusTransitionResponse defines the structure for status transition responses
type AppointmentStatusTransitionResponse struct {
	ID            string    `json:"id" example:"507f1f77bcf86cd799439011"`
	AppointmentID string    `json:"appointment_id" example:"507f1f77bcf86cd799439012"`
	FromStatus    string    `json:"from_status" example:"scheduled"`
	ToStatus      string    `json:"to_status" example:"confirmed"`
	ChangedBy     string    `json:"changed_by" example:"507f1f77bcf86cd799439013"`
	Reason        string    `json:"reason,omitempty" example:"Confirmed by staff"`
	CreatedAt     time.Time `json:"created_at" example:"2024-01-14T15:00:00Z"`
}

// PaginatedAppointmentsResponse defines the structure for paginated appointment responses
type PaginatedAppointmentsResponse struct {
	Data       []AppointmentResponse     `json:"data"`
	Pagination pagination.PaginationInfo `json:"pagination"`
}

// CalendarViewResponse defines the structure for calendar view responses
type CalendarViewResponse struct {
	Date         string                `json:"date" example:"2024-01-15"`
	Appointments []AppointmentResponse `json:"appointments"`
}

// AvailabilityResponse defines the structure for availability check responses
type AvailabilityResponse struct {
	Available     bool     `json:"available" example:"true"`
	ConflictTimes []string `json:"conflict_times,omitempty" example:"[\"10:30-11:00\", \"14:00-14:30\"]"`
	Suggestions   []string `json:"suggestions,omitempty" example:"[\"11:00\", \"15:00\", \"16:30\"]"`
}

// Internal DTOs for filtering and querying

// appointmentFilters defines internal filtering options
type appointmentFilters struct {
	Status         []string
	Type           []string
	VeterinarianID *primitive.ObjectID
	PatientID      *primitive.ObjectID
	OwnerID        *primitive.ObjectID
	DateFrom       *time.Time
	DateTo         *time.Time
	Priority       *string
}

// Conversion functions

// ToResponse converts Appointment entity to AppointmentResponse
func (a *Appointment) ToResponse() *AppointmentResponse {
	response := &AppointmentResponse{
		ID:             a.ID.Hex(),
		PatientID:      a.PatientID.Hex(),
		OwnerID:        a.OwnerID.Hex(),
		VeterinarianID: a.VeterinarianID.Hex(),
		ScheduledAt:    a.ScheduledAt,
		Duration:       a.Duration,
		Type:           a.Type,
		Status:         a.Status,
		Priority:       a.Priority,
		Reason:         a.Reason,
		Notes:          a.Notes,
		OwnerNotes:     a.OwnerNotes,
		ConfirmedAt:    a.ConfirmedAt,
		StartedAt:      a.StartedAt,
		CompletedAt:    a.CompletedAt,
		CancelledAt:    a.CancelledAt,
		CancelReason:   a.CancelReason,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}

	return response
}

// ToResponse converts AppointmentStatusTransition to AppointmentStatusTransitionResponse
func (t *AppointmentStatusTransition) ToResponse() *AppointmentStatusTransitionResponse {
	return &AppointmentStatusTransitionResponse{
		ID:            t.ID.Hex(),
		AppointmentID: t.AppointmentID.Hex(),
		FromStatus:    t.FromStatus,
		ToStatus:      t.ToStatus,
		ChangedBy:     t.ChangedBy.Hex(),
		Reason:        t.Reason,
		CreatedAt:     t.CreatedAt,
	}
}

// CreatePaginatedResponse creates a paginated response
func CreatePaginatedResponse(appointments []Appointment, params pagination.Params, total int64) *PaginatedAppointmentsResponse {
	data := make([]AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		data[i] = *appointment.ToResponse()
	}

	return &PaginatedAppointmentsResponse{
		Data:       data,
		Pagination: pagination.NewPaginationInfo(params, total),
	}
}
