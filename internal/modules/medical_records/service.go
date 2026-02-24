package medical_records

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/patients"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// NotificationSender defines the interface for sending notifications
type NotificationSender interface {
	Send(ctx context.Context, dto *notifications.SendDTO) error
	SendToStaff(ctx context.Context, dto *notifications.SendStaffDTO) error
}

// PatientRepository defines the interface for patient data access
type PatientRepository interface {
	FindByID(ctx context.Context, tenantID primitive.ObjectID, id string) (*patients.Patient, error)
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*users.User, error)
}

// Service provides business logic for medical records
type Service struct {
	repo            MedicalRecordRepository
	patientRepo     PatientRepository
	userRepo        UserRepository
	notificationSvc NotificationSender
}

// NewService creates a new medical records service
func NewService(repo MedicalRecordRepository, patientRepo PatientRepository, userRepo UserRepository, notificationSvc NotificationSender) *Service {
	return &Service{
		repo:            repo,
		patientRepo:     patientRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
	}
}

// CreateMedicalRecord creates a new medical record
func (s *Service) CreateMedicalRecord(ctx context.Context, dto *CreateMedicalRecordDTO, tenantID primitive.ObjectID) (*MedicalRecord, error) {
	// Validate patient
	patientID, err := primitive.ObjectIDFromHex(dto.PatientID)
	if err != nil {
		return nil, ErrValidation("patient_id", "invalid patient ID format")
	}

	patient, err := s.patientRepo.FindByID(ctx, tenantID, patientID.Hex())
	if err != nil {
		return nil, ErrPatientNotFound
	}

	// Validate veterinarian
	vetID, err := primitive.ObjectIDFromHex(dto.VeterinarianID)
	if err != nil {
		return nil, ErrValidation("veterinarian_id", "invalid veterinarian ID format")
	}

	_, err = s.userRepo.FindByID(ctx, vetID.Hex())
	if err != nil {
		return nil, ErrVeterinarianNotFound
	}

	// Validate appointment if provided
	var appointmentID *primitive.ObjectID
	if dto.AppointmentID != "" {
		aptID, err := primitive.ObjectIDFromHex(dto.AppointmentID)
		if err != nil {
			return nil, ErrValidation("appointment_id", "invalid appointment ID format")
		}
		appointmentID = &aptID
	}

	// Validate next visit date
	var nextVisitDate *time.Time
	if dto.NextVisitDate != "" {
		nvd, err := time.Parse(time.RFC3339, dto.NextVisitDate)
		if err != nil {
			return nil, ErrValidation("next_visit_date", "invalid date format, use RFC3339")
		}
		if nvd.Before(time.Now()) {
			return nil, ErrInvalidNextVisitDate
		}
		nextVisitDate = &nvd
	}

	// Validate temperature if provided
	if dto.Temperature != 0 && (dto.Temperature < 30 || dto.Temperature > 45) {
		return nil, ErrInvalidTemperature
	}

	// Validate weight if provided
	if dto.Weight < 0 {
		return nil, ErrInvalidWeight
	}

	// Create medical record
	now := time.Now()
	record := &MedicalRecord{
		ID:             primitive.NewObjectID(),
		TenantID:       tenantID,
		PatientID:      patientID,
		OwnerID:        patient.OwnerID,
		VeterinarianID: vetID,
		AppointmentID:  appointmentID,
		Type:           MedicalRecordType(dto.Type),
		ChiefComplaint: dto.ChiefComplaint,
		Diagnosis:      dto.Diagnosis,
		Symptoms:       dto.Symptoms,
		Weight:         dto.Weight,
		Temperature:    dto.Temperature,
		Treatment:      dto.Treatment,
		Medications:    dto.ToMedications(),
		EvolutionNotes: dto.EvolutionNotes,
		AttachmentIDs:  dto.AttachmentIDs,
		NextVisitDate:  nextVisitDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, err
	}

	// Send notification to owner about new medical record
	s.notificationSvc.Send(ctx, &notifications.SendDTO{
		OwnerID:  patient.OwnerID.Hex(),
		TenantID: tenantID.Hex(),
		Type:     notifications.TypeMedicalRecordCreated,
		Title:    "Nuevo registro médico",
		Body:     "Se ha creado un nuevo registro médico para " + patient.Name,
		Data: map[string]string{
			"record_id":   record.ID.Hex(),
			"patient_id":  record.PatientID.Hex(),
			"record_type": string(record.Type),
		},
		SendPush: true,
	})

	// Send notification if next visit is scheduled
	if nextVisitDate != nil {
		s.notificationSvc.Send(ctx, &notifications.SendDTO{
			OwnerID:  patient.OwnerID.Hex(),
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeAppointmentReminder,
			Title:    "Próxima visita programada",
			Body:     "Próxima visita programada para el " + nextVisitDate.Format("02/01/2006"),
			Data: map[string]string{
				"record_id":      record.ID.Hex(),
				"next_visit":     nextVisitDate.Format(time.RFC3339),
				"patient_id":     record.PatientID.Hex(),
			},
			SendPush: true,
		})
	}

	return record, nil
}

// GetMedicalRecord gets a medical record by ID
func (s *Service) GetMedicalRecord(ctx context.Context, id string, tenantID primitive.ObjectID) (*MedicalRecord, error) {
	recordID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid record ID format")
	}

	record, err := s.repo.FindByID(ctx, recordID, tenantID)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ListMedicalRecords lists medical records with filters
func (s *Service) ListMedicalRecords(ctx context.Context, filters MedicalRecordListFilters, tenantID primitive.ObjectID, params pagination.Params) ([]MedicalRecord, int64, error) {
	return s.repo.FindByFilters(ctx, tenantID, filters, params)
}

// GetPatientRecords gets all medical records for a patient
func (s *Service) GetPatientRecords(ctx context.Context, patientID string, tenantID primitive.ObjectID, params pagination.Params) ([]MedicalRecord, int64, error) {
	pID, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return nil, 0, ErrValidation("patient_id", "invalid patient ID format")
	}

	return s.repo.FindByPatient(ctx, pID, tenantID, params)
}

// UpdateMedicalRecord updates a medical record (only within 24 hours)
func (s *Service) UpdateMedicalRecord(ctx context.Context, id string, dto *UpdateMedicalRecordDTO, tenantID primitive.ObjectID) (*MedicalRecord, error) {
	recordID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid record ID format")
	}

	record, err := s.repo.FindByID(ctx, recordID, tenantID)
	if err != nil {
		return nil, err
	}

	// Check if record is editable (within 24 hours)
	if !record.IsEditable() {
		return nil, ErrRecordNotEditable
	}

	updates := bson.M{}

	if dto.Diagnosis != "" {
		updates["diagnosis"] = dto.Diagnosis
	}

	if dto.Symptoms != "" {
		updates["symptoms"] = dto.Symptoms
	}

	if dto.Weight != 0 {
		if dto.Weight < 0 {
			return nil, ErrInvalidWeight
		}
		updates["weight"] = dto.Weight
	}

	if dto.Temperature != 0 {
		if dto.Temperature < 30 || dto.Temperature > 45 {
			return nil, ErrInvalidTemperature
		}
		updates["temperature"] = dto.Temperature
	}

	if dto.Treatment != "" {
		updates["treatment"] = dto.Treatment
	}

	if len(dto.Medications) > 0 {
		meds := make([]Medication, len(dto.Medications))
		for i, m := range dto.Medications {
			meds[i] = Medication{
				Name:     m.Name,
				Dose:     m.Dose,
				Frequency: m.Frequency,
				Duration: m.Duration,
			}
		}
		updates["medications"] = meds
	}

	if dto.EvolutionNotes != "" {
		updates["evolution_notes"] = dto.EvolutionNotes
	}

	if len(dto.AttachmentIDs) > 0 {
		updates["attachment_ids"] = dto.AttachmentIDs
	}

	if dto.NextVisitDate != "" {
		nvd, err := time.Parse(time.RFC3339, dto.NextVisitDate)
		if err != nil {
			return nil, ErrValidation("next_visit_date", "invalid date format, use RFC3339")
		}
		if nvd.Before(time.Now()) {
			return nil, ErrInvalidNextVisitDate
		}
		updates["next_visit_date"] = nvd
	}

	if err := s.repo.Update(ctx, recordID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedRecord, err := s.repo.FindByID(ctx, recordID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedRecord, nil
}

// DeleteMedicalRecord soft deletes a medical record
func (s *Service) DeleteMedicalRecord(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	recordID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid record ID format")
	}

	_, err = s.repo.FindByID(ctx, recordID, tenantID)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, recordID, tenantID)
}

// GetPatientTimeline gets the medical timeline for a patient
func (s *Service) GetPatientTimeline(ctx context.Context, patientID string, tenantID primitive.ObjectID, filters TimelineFilters) (*MedicalTimeline, error) {
	pID, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return nil, ErrValidation("patient_id", "invalid patient ID format")
	}

	// Get patient to retrieve name
	patient, err := s.patientRepo.FindByID(ctx, tenantID, patientID)
	if err != nil {
		return nil, ErrPatientNotFound
	}

	entries, total, err := s.repo.FindTimeline(ctx, pID, tenantID, filters)
	if err != nil {
		return nil, err
	}

	return &MedicalTimeline{
		PatientID:   patientID,
		PatientName: patient.Name,
		Entries:     entries,
		TotalCount:  total,
	}, nil
}

// CreateAllergy creates a new allergy for a patient
func (s *Service) CreateAllergy(ctx context.Context, dto *CreateAllergyDTO, tenantID primitive.ObjectID) (*Allergy, error) {
	// Validate patient
	patientID, err := primitive.ObjectIDFromHex(dto.PatientID)
	if err != nil {
		return nil, ErrValidation("patient_id", "invalid patient ID format")
	}

	patient, err := s.patientRepo.FindByID(ctx, tenantID, patientID.Hex())
	if err != nil {
		return nil, ErrPatientNotFound
	}

	// Validate severity
	if !IsValidAllergySeverity(dto.Severity) {
		return nil, ErrValidation("severity", "invalid severity, must be mild, moderate, or severe")
	}

	allergy := &Allergy{
		ID:          primitive.NewObjectID(),
		TenantID:    tenantID,
		PatientID:   patientID,
		Allergen:    dto.Allergen,
		Severity:    AllergySeverity(dto.Severity),
		Description: dto.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateAllergy(ctx, allergy); err != nil {
		return nil, err
	}

	// Send alert if severe allergy
	if allergy.Severity == AllergySeveritySevere {
		s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
			UserID:   primitive.NilObjectID.Hex(), // Broadcast to all staff
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeStaffSystemAlert,
			Title:    "Alerta: Alergia Severa Registrada",
			Body:     "El paciente " + patient.Name + " tiene una nueva alergia severa: " + allergy.Allergen,
			Data: map[string]string{
				"allergy_id":  allergy.ID.Hex(),
				"patient_id":  patient.ID.Hex(),
				"allergen":    allergy.Allergen,
				"severity":    string(allergy.Severity),
			},
		})
	}

	return allergy, nil
}

// GetPatientAllergies gets all allergies for a patient
func (s *Service) GetPatientAllergies(ctx context.Context, patientID string, tenantID primitive.ObjectID) ([]Allergy, error) {
	pID, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return nil, ErrValidation("patient_id", "invalid patient ID format")
	}

	return s.repo.FindAllergiesByPatient(ctx, pID, tenantID)
}

// UpdateAllergy updates an allergy
func (s *Service) UpdateAllergy(ctx context.Context, id string, dto *UpdateAllergyDTO, tenantID primitive.ObjectID) (*Allergy, error) {
	allergyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid allergy ID format")
	}

	_, err = s.repo.FindAllergyByID(ctx, allergyID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}

	if dto.Allergen != "" {
		updates["allergen"] = dto.Allergen
	}

	if dto.Severity != "" {
		if !IsValidAllergySeverity(dto.Severity) {
			return nil, ErrValidation("severity", "invalid severity, must be mild, moderate, or severe")
		}
		updates["severity"] = dto.Severity
	}

	if dto.Description != "" {
		updates["description"] = dto.Description
	}

	if err := s.repo.UpdateAllergy(ctx, allergyID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedAllergy, err := s.repo.FindAllergyByID(ctx, allergyID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedAllergy, nil
}

// DeleteAllergy soft deletes an allergy
func (s *Service) DeleteAllergy(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	allergyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid allergy ID format")
	}

	return s.repo.DeleteAllergy(ctx, allergyID, tenantID)
}

// CreateMedicalHistory creates or updates medical history for a patient
func (s *Service) CreateMedicalHistory(ctx context.Context, dto *CreateMedicalHistoryDTO, tenantID primitive.ObjectID) (*MedicalHistory, error) {
	// Validate patient
	patientID, err := primitive.ObjectIDFromHex(dto.PatientID)
	if err != nil {
		return nil, ErrValidation("patient_id", "invalid patient ID format")
	}

	_, err = s.patientRepo.FindByID(ctx, tenantID, patientID.Hex())
	if err != nil {
		return nil, ErrPatientNotFound
	}

	// Check if history already exists
	existing, _ := s.repo.FindHistoryByPatient(ctx, patientID, tenantID)

	now := time.Now()
	if existing != nil {
		// Update existing
		updates := bson.M{
			"chronic_conditions": dto.ChronicConditions,
			"previous_surgeries": dto.PreviousSurgeries,
			"risk_factors":       dto.RiskFactors,
			"blood_type":         dto.BloodType,
		}

		if err := s.repo.UpdateHistory(ctx, existing.ID, updates, tenantID); err != nil {
			return nil, err
		}

		return s.repo.FindHistoryByPatient(ctx, patientID, tenantID)
	}

	// Create new
	history := &MedicalHistory{
		ID:                primitive.NewObjectID(),
		TenantID:          tenantID,
		PatientID:         patientID,
		ChronicConditions: dto.ChronicConditions,
		PreviousSurgeries: dto.PreviousSurgeries,
		RiskFactors:       dto.RiskFactors,
		BloodType:         dto.BloodType,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.CreateHistory(ctx, history); err != nil {
		if errors.Is(err, ErrDuplicateHistory) {
			// Race condition - try to get existing
			return s.repo.FindHistoryByPatient(ctx, patientID, tenantID)
		}
		return nil, err
	}

	return history, nil
}

// GetMedicalHistory gets medical history for a patient
func (s *Service) GetMedicalHistory(ctx context.Context, patientID string, tenantID primitive.ObjectID) (*MedicalHistory, error) {
	pID, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return nil, ErrValidation("patient_id", "invalid patient ID format")
	}

	return s.repo.FindHistoryByPatient(ctx, pID, tenantID)
}

// UpdateMedicalHistory updates medical history
func (s *Service) UpdateMedicalHistory(ctx context.Context, id string, dto *UpdateMedicalHistoryDTO, tenantID primitive.ObjectID) (*MedicalHistory, error) {
	historyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid history ID format")
	}

	_, err = s.repo.FindHistoryByPatient(ctx, primitive.NilObjectID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{
		"chronic_conditions": dto.ChronicConditions,
		"previous_surgeries": dto.PreviousSurgeries,
		"risk_factors":       dto.RiskFactors,
		"blood_type":         dto.BloodType,
	}

	if err := s.repo.UpdateHistory(ctx, historyID, updates, tenantID); err != nil {
		return nil, err
	}

	// Find the updated history by patient ID (we need to get patient ID from the original record)
	// For simplicity, we'll query by the history ID directly
	filter := bson.M{"_id": historyID, "tenant_id": tenantID, "deleted_at": nil}
	var updatedHistory MedicalHistory
	err = s.repo.(*medicalRecordRepository).historyCollection.FindOne(ctx, filter).Decode(&updatedHistory)
	if err != nil {
		return nil, err
	}

	return &updatedHistory, nil
}

// DeleteMedicalHistory soft deletes medical history
func (s *Service) DeleteMedicalHistory(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	historyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid history ID format")
	}

	return s.repo.DeleteHistory(ctx, historyID, tenantID)
}

// GetSevereAllergies gets all severe allergies for a patient (helper for other modules)
func (s *Service) GetSevereAllergies(ctx context.Context, patientID string, tenantID primitive.ObjectID) ([]Allergy, error) {
	allergies, err := s.GetPatientAllergies(ctx, patientID, tenantID)
	if err != nil {
		return nil, err
	}

	severe := make([]Allergy, 0)
	for _, a := range allergies {
		if a.Severity == AllergySeveritySevere {
			severe = append(severe, a)
		}
	}

	return severe, nil
}
