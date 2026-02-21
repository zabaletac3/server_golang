package appointments

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service provides business logic for appointments
type Service struct {
	repo AppointmentRepository
}

// NewService creates a new appointment service
func NewService(repo AppointmentRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateAppointment creates a new appointment
func (s *Service) CreateAppointment(ctx context.Context, dto CreateAppointmentDTO, createdBy string) (*AppointmentResponse, error) {
	// Convert string IDs to ObjectIDs
	patientID, err := primitive.ObjectIDFromHex(dto.PatientID)
	if err != nil {
		return nil, ErrValidationFailed("patient_id", "invalid patient ID format")
	}

	veterinarianID, err := primitive.ObjectIDFromHex(dto.VeterinarianID)
	if err != nil {
		return nil, ErrValidationFailed("veterinarian_id", "invalid veterinarian ID format")
	}

	// Note: createdBy tracking can be added later if needed

	// Validate appointment time is not in the past
	if dto.ScheduledAt.Before(time.Now()) {
		return nil, ErrPastAppointmentTime
	}

	// Check for conflicts
	hasConflict, err := s.repo.CheckConflicts(ctx, veterinarianID, dto.ScheduledAt, dto.Duration, nil, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	if hasConflict {
		return nil, ErrAppointmentConflict
	}

	// Set default priority if not provided
	priority := dto.Priority
	if priority == "" {
		priority = AppointmentPriorityNormal
	}

	// Create appointment entity
	now := time.Now()
	appointment := &Appointment{
		PatientID:      patientID,
		OwnerID:        primitive.NewObjectID(), // This should be set properly in a real implementation
		VeterinarianID: veterinarianID,
		ScheduledAt:    dto.ScheduledAt,
		Duration:       dto.Duration,
		Type:           dto.Type,
		Status:         AppointmentStatusScheduled,
		Priority:       priority,
		Reason:         dto.Reason,
		Notes:          dto.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Save appointment
	if err := s.repo.Create(ctx, appointment); err != nil {
		return nil, err
	}

	return appointment.ToResponse(), nil
}

// GetAppointment gets an appointment by ID
func (s *Service) GetAppointment(ctx context.Context, id string, populate bool) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	appointment, err := s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	return appointment.ToResponse(), nil
}

// ListAppointments lists appointments with filters and pagination
func (s *Service) ListAppointments(ctx context.Context, filters map[string]interface{}, page, limit int, populate bool) (*PaginatedAppointmentsResponse, error) {
	// Convert filters map to internal filter struct
	appointmentFilters := s.parseFilters(filters)

	appointments, total, err := s.repo.List(ctx, appointmentFilters, []primitive.ObjectID{}, page, limit)
	if err != nil {
		return nil, err
	}

	return CreatePaginatedResponse(appointments, page, limit, total), nil
}

// UpdateAppointment updates an appointment
func (s *Service) UpdateAppointment(ctx context.Context, id string, dto UpdateAppointmentDTO, updatedBy string) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	// Get current appointment
	appointment, err := s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	// Build updates
	updates := bson.M{}

	if dto.ScheduledAt != nil {
		if dto.ScheduledAt.Before(time.Now()) {
			return nil, ErrPastAppointmentTime
		}

		duration := appointment.Duration
		if dto.Duration != nil {
			duration = *dto.Duration
		}

		// Check for conflicts
		hasConflict, err := s.repo.CheckConflicts(ctx, appointment.VeterinarianID, *dto.ScheduledAt, duration, &appointmentID, []primitive.ObjectID{})
		if err != nil {
			return nil, err
		}

		if hasConflict {
			return nil, ErrAppointmentConflict
		}

		updates["scheduled_at"] = *dto.ScheduledAt
	}

	if dto.Duration != nil {
		updates["duration"] = *dto.Duration
	}

	if dto.Type != nil {
		updates["type"] = *dto.Type
	}

	if dto.Priority != nil {
		updates["priority"] = *dto.Priority
	}

	if dto.Reason != nil {
		updates["reason"] = *dto.Reason
	}

	if dto.Notes != nil {
		updates["notes"] = *dto.Notes
	}

	updates["updated_at"] = time.Now()

	// Update appointment
	if err := s.repo.Update(ctx, appointmentID, updates, []primitive.ObjectID{}); err != nil {
		return nil, err
	}

	// Get updated appointment
	updatedAppointment, err := s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	return updatedAppointment.ToResponse(), nil
}

// UpdateStatus updates an appointment status
func (s *Service) UpdateStatus(ctx context.Context, id string, dto UpdateStatusDTO, changedBy string) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	changedByID, err := primitive.ObjectIDFromHex(changedBy)
	if err != nil {
		return nil, ErrValidationFailed("changed_by", "invalid user ID format")
	}

	// Get current appointment
	appointment, err := s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !appointment.CanTransitionTo(dto.Status) {
		return nil, ErrInvalidStatus(appointment.Status, dto.Status)
	}

	// Build updates based on new status
	updates := bson.M{
		"status":     dto.Status,
		"updated_at": time.Now(),
	}

	now := time.Now()
	switch dto.Status {
	case AppointmentStatusConfirmed:
		updates["confirmed_at"] = now
	case AppointmentStatusInProgress:
		updates["started_at"] = now
	case AppointmentStatusCompleted:
		updates["completed_at"] = now
	case AppointmentStatusCancelled, AppointmentStatusNoShow:
		updates["cancelled_at"] = now
		if dto.Reason != "" {
			updates["cancel_reason"] = dto.Reason
		}
	}

	// Update appointment
	if err := s.repo.Update(ctx, appointmentID, updates, []primitive.ObjectID{}); err != nil {
		return nil, err
	}

	// Record status transition
	transition := &AppointmentStatusTransition{
		AppointmentID: appointmentID,
		FromStatus:    appointment.Status,
		ToStatus:      dto.Status,
		ChangedBy:     changedByID,
		Reason:        dto.Reason,
		CreatedAt:     now,
	}

	s.repo.CreateStatusTransition(ctx, transition) // Ignore error for now

	// Get updated appointment
	updatedAppointment, err := s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	return updatedAppointment.ToResponse(), nil
}

// DeleteAppointment deletes an appointment
func (s *Service) DeleteAppointment(ctx context.Context, id string, deletedBy string) error {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidationFailed("id", "invalid appointment ID format")
	}

	// Check if appointment exists
	_, err = s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, appointmentID, []primitive.ObjectID{})
}

// RequestAppointment creates an appointment request from mobile
func (s *Service) RequestAppointment(ctx context.Context, dto MobileAppointmentRequestDTO, ownerID string) (*AppointmentResponse, error) {
	ownerOID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, ErrValidationFailed("owner_id", "invalid owner ID format")
	}

	patientID, err := primitive.ObjectIDFromHex(dto.PatientID)
	if err != nil {
		return nil, ErrValidationFailed("patient_id", "invalid patient ID format")
	}

	// Validate appointment time is not in the past
	if dto.ScheduledAt.Before(time.Now()) {
		return nil, ErrPastAppointmentTime
	}

	// Set default priority if not provided
	priority := dto.Priority
	if priority == "" {
		priority = AppointmentPriorityNormal
	}

	// Create appointment entity (requested status)
	now := time.Now()
	appointment := &Appointment{
		PatientID:      patientID,
		OwnerID:        ownerOID,
		VeterinarianID: primitive.NewObjectID(), // Should be assigned by staff
		ScheduledAt:    dto.ScheduledAt,
		Duration:       30, // Default duration for requests
		Type:           dto.Type,
		Status:         AppointmentStatusScheduled, // Requested status
		Priority:       priority,
		Reason:         dto.Reason,
		OwnerNotes:     dto.OwnerNotes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Save appointment
	if err := s.repo.Create(ctx, appointment); err != nil {
		return nil, err
	}

	return appointment.ToResponse(), nil
}

// GetOwnerAppointments gets appointments for a specific owner
func (s *Service) GetOwnerAppointments(ctx context.Context, ownerID string, page, limit int) (*PaginatedAppointmentsResponse, error) {
	ownerOID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, ErrValidationFailed("owner_id", "invalid owner ID format")
	}

	appointments, total, err := s.repo.FindByOwner(ctx, ownerOID, []primitive.ObjectID{}, page, limit)
	if err != nil {
		return nil, err
	}

	return CreatePaginatedResponse(appointments, page, limit, total), nil
}

// GetOwnerAppointment gets a specific appointment for an owner
func (s *Service) GetOwnerAppointment(ctx context.Context, id string, ownerID string) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	ownerOID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, ErrValidationFailed("owner_id", "invalid owner ID format")
	}

	appointment, err := s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if appointment.OwnerID != ownerOID {
		return nil, ErrOwnerMismatch
	}

	return appointment.ToResponse(), nil
}

// GetCalendarView gets a calendar view of appointments
func (s *Service) GetCalendarView(ctx context.Context, from, to time.Time, veterinarianID *string) ([]AppointmentResponse, error) {
	var appointments []Appointment
	var err error

	if veterinarianID != nil {
		vetID, parseErr := primitive.ObjectIDFromHex(*veterinarianID)
		if parseErr != nil {
			return nil, ErrValidationFailed("veterinarian_id", "invalid veterinarian ID format")
		}
		appointments, err = s.repo.FindByVeterinarian(ctx, vetID, from, to, []primitive.ObjectID{})
	} else {
		appointments, err = s.repo.FindByDateRange(ctx, from, to, []primitive.ObjectID{})
	}

	if err != nil {
		return nil, err
	}

	response := make([]AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		response[i] = *appointment.ToResponse()
	}

	return response, nil
}

// CheckAvailability checks veterinarian availability
func (s *Service) CheckAvailability(ctx context.Context, vetID string, scheduledAt time.Time, duration int, excludeID *string) (*AvailabilityResponse, error) {
	veterinarianID, err := primitive.ObjectIDFromHex(vetID)
	if err != nil {
		return nil, ErrValidationFailed("veterinarian_id", "invalid veterinarian ID format")
	}

	var excludeOID *primitive.ObjectID
	if excludeID != nil {
		oid, err := primitive.ObjectIDFromHex(*excludeID)
		if err != nil {
			return nil, ErrValidationFailed("exclude_id", "invalid exclude ID format")
		}
		excludeOID = &oid
	}

	hasConflict, err := s.repo.CheckConflicts(ctx, veterinarianID, scheduledAt, duration, excludeOID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	response := &AvailabilityResponse{
		Available: !hasConflict,
	}

	if hasConflict {
		// Find conflicting appointments
		endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)
		conflicts, err := s.repo.FindByVeterinarian(ctx, veterinarianID, scheduledAt.Add(-1*time.Hour), endTime.Add(1*time.Hour), []primitive.ObjectID{})
		if err == nil {
			conflictTimes := make([]string, len(conflicts))
			for i, conflict := range conflicts {
				start := conflict.ScheduledAt.Format("15:04")
				end := conflict.ScheduledAt.Add(time.Duration(conflict.Duration) * time.Minute).Format("15:04")
				conflictTimes[i] = start + "-" + end
			}
			response.ConflictTimes = conflictTimes
		}
	}

	return response, nil
}

// GetStatusHistory gets the status history for an appointment
func (s *Service) GetStatusHistory(ctx context.Context, id string) ([]AppointmentStatusTransitionResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	// Verify appointment exists
	_, err = s.repo.FindByID(ctx, appointmentID, []primitive.ObjectID{})
	if err != nil {
		return nil, err
	}

	transitions, err := s.repo.GetStatusHistory(ctx, appointmentID)
	if err != nil {
		return nil, err
	}

	response := make([]AppointmentStatusTransitionResponse, len(transitions))
	for i, transition := range transitions {
		response[i] = *transition.ToResponse()
	}

	return response, nil
}

// parseFilters converts a map of filters to internal filter struct
func (s *Service) parseFilters(filters map[string]interface{}) appointmentFilters {
	result := appointmentFilters{}

	if status, ok := filters["status"].([]string); ok {
		result.Status = status
	}

	if types, ok := filters["type"].([]string); ok {
		result.Type = types
	}

	if vetID, ok := filters["veterinarian_id"].(string); ok {
		if oid, err := primitive.ObjectIDFromHex(vetID); err == nil {
			result.VeterinarianID = &oid
		}
	}

	if patientID, ok := filters["patient_id"].(string); ok {
		if oid, err := primitive.ObjectIDFromHex(patientID); err == nil {
			result.PatientID = &oid
		}
	}

	if ownerID, ok := filters["owner_id"].(string); ok {
		if oid, err := primitive.ObjectIDFromHex(ownerID); err == nil {
			result.OwnerID = &oid
		}
	}

	if dateFrom, ok := filters["date_from"].(time.Time); ok {
		result.DateFrom = &dateFrom
	}

	if dateTo, ok := filters["date_to"].(time.Time); ok {
		result.DateTo = &dateTo
	}

	if priority, ok := filters["priority"].(string); ok {
		result.Priority = &priority
	}

	return result
}
