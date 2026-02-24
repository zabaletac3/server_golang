package appointments

import (
	"context"
	"fmt"
	"time"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/modules/patients"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationSender interface {
	Send(ctx context.Context, dto *notifications.SendDTO) error
	SendToStaff(ctx context.Context, dto *notifications.SendStaffDTO) error
}

// Service provides business logic for appointments
type Service struct {
	repo            AppointmentRepository
	patientRepo     patients.PatientRepository
	ownerRepo       owners.OwnerRepository
	userRepo        users.UserRepository
	notificationSvc NotificationSender
	cfg             *config.Config
}

// NewService creates a new appointment service
func NewService(repo AppointmentRepository, patientRepo patients.PatientRepository, ownerRepo owners.OwnerRepository, userRepo users.UserRepository, notificationSvc NotificationSender, cfg *config.Config) *Service {
	return &Service{
		repo:            repo,
		patientRepo:     patientRepo,
		ownerRepo:       ownerRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
		cfg:             cfg,
	}
}

// populateAppointment populates references for an appointment
func (s *Service) populateAppointment(ctx context.Context, appointment *Appointment, tenantID primitive.ObjectID) (*AppointmentResponse, error) {
	resp := appointment.ToResponse()

	// Populate patient
	if patient, err := s.patientRepo.FindByID(ctx, tenantID, appointment.PatientID.Hex()); err == nil {
		resp.Patient = &PatientSummary{ID: patient.ID.Hex(), Name: patient.Name}
	}

	// Populate owner
	if owner, err := s.ownerRepo.FindByID(ctx, appointment.OwnerID.Hex()); err == nil {
		resp.Owner = &OwnerSummary{ID: owner.ID.Hex(), Name: owner.Name, Email: owner.Email, Phone: owner.Phone}
	}

	// Populate veterinarian (only if assigned)
	if !appointment.VeterinarianID.IsZero() {
		if vet, err := s.userRepo.FindByID(ctx, appointment.VeterinarianID.Hex()); err == nil {
			resp.Veterinarian = &VeterinarianSummary{ID: vet.ID.Hex(), Name: vet.Name, Email: vet.Email}
		}
	}

	return resp, nil
}

// validateAppointmentTime checks if the appointment time is valid
func (s *Service) validateAppointmentTime(scheduledAt time.Time) error {
	if scheduledAt.Before(time.Now()) {
		return ErrPastAppointmentTime
	}

	hour := scheduledAt.Hour()
	if hour < s.cfg.AppointmentBusinessStartHour || hour >= s.cfg.AppointmentBusinessEndHour {
		return ErrInvalidAppointmentTime
	}

	if scheduledAt.Weekday() == time.Sunday {
		return ErrInvalidAppointmentTime
	}

	return nil
}

// CreateAppointment creates a new appointment
func (s *Service) CreateAppointment(ctx context.Context, dto CreateAppointmentDTO, tenantID primitive.ObjectID, createdBy primitive.ObjectID) (*AppointmentResponse, error) {
	patientID, err := primitive.ObjectIDFromHex(dto.PatientID)
	if err != nil {
		return nil, ErrValidationFailed("patient_id", "invalid patient ID format")
	}

	veterinarianID, err := primitive.ObjectIDFromHex(dto.VeterinarianID)
	if err != nil {
		return nil, ErrValidationFailed("veterinarian_id", "invalid veterinarian ID format")
	}

	if err := s.validateAppointmentTime(dto.ScheduledAt); err != nil {
		return nil, err
	}

	patient, err := s.patientRepo.FindByID(ctx, tenantID, patientID.Hex())
	if err != nil {
		return nil, ErrPatientNotFound
	}

	if !veterinarianID.IsZero() {
		_, err := s.userRepo.FindByID(ctx, veterinarianID.Hex())
		if err != nil {
			return nil, ErrVeterinarianNotFound
		}
	}

	hasConflict, err := s.repo.CheckConflicts(ctx, veterinarianID, dto.ScheduledAt, dto.Duration, nil, tenantID)
	if err != nil {
		return nil, err
	}

	if hasConflict {
		return nil, ErrAppointmentConflict
	}

	priority := dto.Priority
	if priority == "" {
		priority = AppointmentPriorityNormal
	}

	now := time.Now()
	appointment := &Appointment{
		TenantID:       tenantID,
		PatientID:      patientID,
		OwnerID:        patient.OwnerID,
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

	if err := s.repo.Create(ctx, appointment); err != nil {
		return nil, err
	}

	s.notificationSvc.Send(ctx, &notifications.SendDTO{
		OwnerID:  appointment.OwnerID.Hex(),
		TenantID: tenantID.Hex(),
		Type:     notifications.TypeAppointmentReminder,
		Title:    "Nueva cita agendada",
		Body:     fmt.Sprintf("Se ha agendado una cita para %s el %s", patient.Name, appointment.ScheduledAt.Format("02/01/2006 15:04")),
		Data:     map[string]string{"appointment_id": appointment.ID.Hex(), "patient_id": appointment.PatientID.Hex()},
		SendPush: true,
	})

	if !appointment.VeterinarianID.IsZero() {
		s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
			UserID:   appointment.VeterinarianID.Hex(),
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeStaffNewAppointment,
			Title:    "Nueva cita asignada",
			Body:     fmt.Sprintf("Cita con %s el %s", patient.Name, appointment.ScheduledAt.Format("02/01/2006 15:04")),
			Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
		})
	}

	return appointment.ToResponse(), nil
}

// GetAppointment gets an appointment by ID
func (s *Service) GetAppointment(ctx context.Context, id string, tenantID primitive.ObjectID, populate bool) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	appointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	if populate {
		return s.populateAppointment(ctx, appointment, tenantID)
	}
	return appointment.ToResponse(), nil
}

// ListAppointments lists appointments with filters and pagination
func (s *Service) ListAppointments(ctx context.Context, filters map[string]interface{}, tenantID primitive.ObjectID, params pagination.Params, populate bool) (*PaginatedAppointmentsResponse, error) {
	appointmentFilters := s.parseFilters(filters)

	appointments, total, err := s.repo.List(ctx, appointmentFilters, tenantID, params)
	if err != nil {
		return nil, err
	}

	return CreatePaginatedResponse(appointments, params, total), nil
}

// UpdateAppointment updates an appointment
func (s *Service) UpdateAppointment(ctx context.Context, id string, dto UpdateAppointmentDTO, tenantID primitive.ObjectID, updatedBy primitive.ObjectID) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	appointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}

	if dto.ScheduledAt != nil {
		if err := s.validateAppointmentTime(*dto.ScheduledAt); err != nil {
			return nil, err
		}

		duration := appointment.Duration
		if dto.Duration != nil {
			duration = *dto.Duration
		}

		hasConflict, err := s.repo.CheckConflicts(ctx, appointment.VeterinarianID, *dto.ScheduledAt, duration, &appointmentID, tenantID)
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

	if err := s.repo.Update(ctx, appointmentID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedAppointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedAppointment.ToResponse(), nil
}

// UpdateStatus updates an appointment status
func (s *Service) UpdateStatus(ctx context.Context, id string, dto UpdateStatusDTO, tenantID primitive.ObjectID, changedBy primitive.ObjectID) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	appointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	if !appointment.CanTransitionTo(dto.Status) {
		return nil, ErrInvalidStatus(appointment.Status, dto.Status)
	}

	now := time.Now()
	updates := bson.M{
		"status":     dto.Status,
		"updated_at": now,
	}

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

	if err := s.repo.Update(ctx, appointmentID, updates, tenantID); err != nil {
		return nil, err
	}

	transition := &AppointmentStatusTransition{
		TenantID:      tenantID,
		AppointmentID: appointmentID,
		FromStatus:    appointment.Status,
		ToStatus:      dto.Status,
		ChangedBy:     changedBy,
		Reason:        dto.Reason,
		CreatedAt:     now,
	}

	s.repo.CreateStatusTransition(ctx, transition)

	if dto.Status == AppointmentStatusConfirmed {
		s.notificationSvc.Send(ctx, &notifications.SendDTO{
			OwnerID:  appointment.OwnerID.Hex(),
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeAppointmentConfirmed,
			Title:    "Cita confirmada",
			Body:     fmt.Sprintf("Tu cita del %s ha sido confirmada", appointment.ScheduledAt.Format("02/01/2006 15:04")),
			Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
			SendPush: true,
		})
	} else if dto.Status == AppointmentStatusCancelled {
		s.notificationSvc.Send(ctx, &notifications.SendDTO{
			OwnerID:  appointment.OwnerID.Hex(),
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeAppointmentCancelled,
			Title:    "Cita cancelada",
			Body:     fmt.Sprintf("La cita del %s ha sido cancelada. Razón: %s", appointment.ScheduledAt.Format("02/01/2006 15:04"), dto.Reason),
			Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
			SendPush: true,
		})
	}

	updatedAppointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedAppointment.ToResponse(), nil
}

// DeleteAppointment deletes an appointment
func (s *Service) DeleteAppointment(ctx context.Context, id string, tenantID primitive.ObjectID, deletedBy primitive.ObjectID) error {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidationFailed("id", "invalid appointment ID format")
	}

	_, err = s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, appointmentID, tenantID)
}

// RequestAppointment creates an appointment request from mobile
func (s *Service) RequestAppointment(ctx context.Context, dto MobileAppointmentRequestDTO, tenantID primitive.ObjectID, ownerID primitive.ObjectID) (*AppointmentResponse, error) {
	patientID, err := primitive.ObjectIDFromHex(dto.PatientID)
	if err != nil {
		return nil, ErrValidationFailed("patient_id", "invalid patient ID format")
	}

	if err := s.validateAppointmentTime(dto.ScheduledAt); err != nil {
		return nil, err
	}

	patient, err := s.patientRepo.FindByID(ctx, tenantID, patientID.Hex())
	if err != nil {
		return nil, ErrPatientNotFound
	}
	if patient.OwnerID != ownerID {
		return nil, ErrOwnerMismatch
	}

	owner, err := s.ownerRepo.FindByID(ctx, ownerID.Hex())
	if err != nil {
		return nil, ErrOwnerNotFound
	}

	priority := dto.Priority
	if priority == "" {
		priority = AppointmentPriorityNormal
	}

	now := time.Now()
	appointment := &Appointment{
		TenantID:       tenantID,
		PatientID:      patientID,
		OwnerID:        ownerID,
		VeterinarianID: primitive.NilObjectID,
		ScheduledAt:    dto.ScheduledAt,
		Duration:       30,
		Type:           dto.Type,
		Status:         AppointmentStatusScheduled,
		Priority:       priority,
		Reason:         dto.Reason,
		OwnerNotes:     dto.OwnerNotes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, appointment); err != nil {
		return nil, err
	}

	s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
		UserID:   primitive.NilObjectID.Hex(),
		TenantID: tenantID.Hex(),
		Type:     notifications.TypeStaffNewAppointment,
		Title:    "Nueva solicitud de cita",
		Body:     fmt.Sprintf("Solicitud de cita de %s para %s", owner.Name, patient.Name),
		Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
	})

	return appointment.ToResponse(), nil
}

// GetOwnerAppointments gets appointments for a specific owner
func (s *Service) GetOwnerAppointments(ctx context.Context, ownerID primitive.ObjectID, tenantID primitive.ObjectID, params pagination.Params) (*PaginatedAppointmentsResponse, error) {
	appointments, total, err := s.repo.FindByOwner(ctx, ownerID, tenantID, params)
	if err != nil {
		return nil, err
	}

	return CreatePaginatedResponse(appointments, params, total), nil
}

// GetOwnerAppointment gets a specific appointment for an owner
func (s *Service) GetOwnerAppointment(ctx context.Context, id string, tenantID primitive.ObjectID, ownerID primitive.ObjectID, populate bool) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	appointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	if appointment.OwnerID != ownerID {
		return nil, ErrOwnerMismatch
	}

	if populate {
		return s.populateAppointment(ctx, appointment, tenantID)
	}

	return appointment.ToResponse(), nil
}

// CancelAppointment cancels an appointment (used by mobile)
func (s *Service) CancelAppointment(ctx context.Context, id string, reason string, tenantID primitive.ObjectID, ownerID primitive.ObjectID) (*AppointmentResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	appointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	if appointment.OwnerID != ownerID {
		return nil, ErrOwnerMismatch
	}

	if appointment.Status == AppointmentStatusCancelled {
		return appointment.ToResponse(), nil
	}

	if !appointment.CanTransitionTo(AppointmentStatusCancelled) {
		return nil, ErrInvalidStatus(appointment.Status, AppointmentStatusCancelled)
	}

	now := time.Now()
	updates := bson.M{
		"status":        AppointmentStatusCancelled,
		"cancelled_at":  now,
		"cancel_reason": reason,
		"updated_at":    now,
	}

	if err := s.repo.Update(ctx, appointmentID, updates, tenantID); err != nil {
		return nil, err
	}

	s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
		UserID:   primitive.NilObjectID.Hex(),
		TenantID: tenantID.Hex(),
		Type:     notifications.TypeStaffSystemAlert,
		Title:    "Cita cancelada por cliente",
		Body:     fmt.Sprintf("El cliente ha cancelado la cita del %s. Razón: %s", appointment.ScheduledAt.Format("02/01/2006 15:04"), reason),
		Data:     map[string]string{"appointment_id": appointment.ID.Hex()},
	})

	updatedAppointment, err := s.repo.FindByID(ctx, appointmentID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedAppointment.ToResponse(), nil
}

// GetCalendarView gets a calendar view of appointments
func (s *Service) GetCalendarView(ctx context.Context, from, to time.Time, veterinarianID *string, tenantID primitive.ObjectID) ([]AppointmentResponse, error) {
	var appointments []Appointment
	var err error

	if veterinarianID != nil {
		vetID, parseErr := primitive.ObjectIDFromHex(*veterinarianID)
		if parseErr != nil {
			return nil, ErrValidationFailed("veterinarian_id", "invalid veterinarian ID format")
		}
		appointments, err = s.repo.FindByVeterinarian(ctx, vetID, from, to, tenantID)
	} else {
		appointments, err = s.repo.FindByDateRange(ctx, from, to, tenantID)
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
func (s *Service) CheckAvailability(ctx context.Context, vetID string, scheduledAt time.Time, duration int, excludeID *string, tenantID primitive.ObjectID) (*AvailabilityResponse, error) {
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

	hasConflict, err := s.repo.CheckConflicts(ctx, veterinarianID, scheduledAt, duration, excludeOID, tenantID)
	if err != nil {
		return nil, err
	}

	response := &AvailabilityResponse{
		Available: !hasConflict,
	}

	if hasConflict {
		endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)
		conflicts, err := s.repo.FindByVeterinarian(ctx, veterinarianID, scheduledAt.Add(-1*time.Hour), endTime.Add(1*time.Hour), tenantID)
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
func (s *Service) GetStatusHistory(ctx context.Context, id string, tenantID primitive.ObjectID) ([]AppointmentStatusTransitionResponse, error) {
	appointmentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidationFailed("id", "invalid appointment ID format")
	}

	_, err = s.repo.FindByID(ctx, appointmentID, tenantID)
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
