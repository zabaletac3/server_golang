package vaccinations

import (
	"context"
	"fmt"
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

// Service provides business logic for vaccinations
type Service struct {
	repo            VaccinationRepository
	patientRepo     PatientRepository
	userRepo        UserRepository
	notificationSvc NotificationSender
}

// NewService creates a new vaccinations service
func NewService(repo VaccinationRepository, patientRepo PatientRepository, userRepo UserRepository, notificationSvc NotificationSender) *Service {
	return &Service{
		repo:            repo,
		patientRepo:     patientRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
	}
}

// CreateVaccination creates a new vaccination record
func (s *Service) CreateVaccination(ctx context.Context, dto *CreateVaccinationDTO, tenantID primitive.ObjectID) (*Vaccination, error) {
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

	// Parse and validate application date
	applicationDate, err := dto.ParseApplicationDate()
	if err != nil {
		return nil, ErrValidation("application_date", "invalid date format, use RFC3339")
	}

	// Application date cannot be in the future
	if applicationDate.After(time.Now()) {
		return nil, ErrInvalidApplicationDate
	}

	// Parse and validate next due date
	nextDueDate, err := dto.ParseNextDueDate()
	if err != nil {
		return nil, ErrValidation("next_due_date", "invalid date format, use RFC3339")
	}

	if nextDueDate != nil && nextDueDate.Before(applicationDate) {
		return nil, ErrInvalidNextDueDate
	}

	// Determine status based on dates
	status := VaccinationStatusApplied
	if nextDueDate != nil {
		if nextDueDate.Before(time.Now()) {
			status = VaccinationStatusOverdue
		} else if nextDueDate.Before(time.Now().AddDate(0, 0, 30)) {
			status = VaccinationStatusDue
		}
	}

	// Generate certificate number
	certificateNumber := s.generateCertificateNumber(tenantID, applicationDate)

	// Create vaccination
	now := time.Now()
	vaccination := &Vaccination{
		ID:                primitive.NewObjectID(),
		TenantID:          tenantID,
		PatientID:         patientID,
		OwnerID:           patient.OwnerID,
		VeterinarianID:    vetID,
		VaccineName:       dto.VaccineName,
		Manufacturer:      dto.Manufacturer,
		LotNumber:         dto.LotNumber,
		ApplicationDate:   applicationDate,
		NextDueDate:       nextDueDate,
		Status:            status,
		CertificateNumber: certificateNumber,
		Notes:             dto.Notes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.Create(ctx, vaccination); err != nil {
		return nil, err
	}

	// Send notification to owner
	s.notificationSvc.Send(ctx, &notifications.SendDTO{
		OwnerID:  patient.OwnerID.Hex(),
		TenantID: tenantID.Hex(),
		Type:     notifications.TypeVaccinationDue,
		Title:    "Vacunación Registrada",
		Body:     fmt.Sprintf("Se ha registrado la vacunación %s para %s", dto.VaccineName, patient.Name),
		Data: map[string]string{
			"vaccination_id": vaccination.ID.Hex(),
			"patient_id":     vaccination.PatientID.Hex(),
			"vaccine_name":   dto.VaccineName,
		},
		SendPush: true,
	})

	return vaccination, nil
}

// GetVaccination gets a vaccination by ID
func (s *Service) GetVaccination(ctx context.Context, id string, tenantID primitive.ObjectID) (*Vaccination, error) {
	vaccinationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid vaccination ID format")
	}

	vaccination, err := s.repo.FindByID(ctx, vaccinationID, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccination, nil
}

// ListVaccinations lists vaccinations with filters
func (s *Service) ListVaccinations(ctx context.Context, filters VaccinationListFilters, tenantID primitive.ObjectID, params pagination.Params) ([]Vaccination, int64, error) {
	return s.repo.FindByFilters(ctx, tenantID, filters, params)
}

// GetPatientVaccinations gets all vaccinations for a patient
func (s *Service) GetPatientVaccinations(ctx context.Context, patientID string, tenantID primitive.ObjectID, params pagination.Params) ([]Vaccination, int64, error) {
	pID, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return nil, 0, ErrValidation("patient_id", "invalid patient ID format")
	}

	return s.repo.FindByPatient(ctx, pID, tenantID, params)
}

// UpdateVaccination updates a vaccination
func (s *Service) UpdateVaccination(ctx context.Context, id string, dto *UpdateVaccinationDTO, tenantID primitive.ObjectID) (*Vaccination, error) {
	vaccinationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid vaccination ID format")
	}

	vaccination, err := s.repo.FindByID(ctx, vaccinationID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}

	if dto.VaccineName != "" {
		updates["vaccine_name"] = dto.VaccineName
	}

	if dto.Manufacturer != "" {
		updates["manufacturer"] = dto.Manufacturer
	}

	if dto.LotNumber != "" {
		updates["lot_number"] = dto.LotNumber
	}

	if dto.NextDueDate != "" {
		nextDueDate, err := time.Parse(time.RFC3339, dto.NextDueDate)
		if err != nil {
			return nil, ErrValidation("next_due_date", "invalid date format, use RFC3339")
		}
		if nextDueDate.Before(vaccination.ApplicationDate) {
			return nil, ErrInvalidNextDueDate
		}
		updates["next_due_date"] = nextDueDate

		// Update status based on new due date
		if nextDueDate.Before(time.Now()) {
			updates["status"] = VaccinationStatusOverdue
		} else if nextDueDate.Before(time.Now().AddDate(0, 0, 30)) {
			updates["status"] = VaccinationStatusDue
		}
	}

	if dto.Notes != "" {
		updates["notes"] = dto.Notes
	}

	if err := s.repo.Update(ctx, vaccinationID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedVaccination, err := s.repo.FindByID(ctx, vaccinationID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedVaccination, nil
}

// UpdateVaccinationStatus updates the status of a vaccination
func (s *Service) UpdateVaccinationStatus(ctx context.Context, id string, dto *UpdateVaccinationStatusDTO, tenantID primitive.ObjectID) (*Vaccination, error) {
	vaccinationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid vaccination ID format")
	}

	status := VaccinationStatus(dto.Status)
	if !IsValidVaccinationStatus(string(status)) {
		return nil, ErrInvalidStatus
	}

	if err := s.repo.UpdateStatus(ctx, vaccinationID, status, tenantID); err != nil {
		return nil, err
	}

	updatedVaccination, err := s.repo.FindByID(ctx, vaccinationID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedVaccination, nil
}

// DeleteVaccination soft deletes a vaccination
func (s *Service) DeleteVaccination(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	vaccinationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid vaccination ID format")
	}

	return s.repo.Delete(ctx, vaccinationID, tenantID)
}

// GetDueVaccinations gets vaccinations due within the specified days
func (s *Service) GetDueVaccinations(ctx context.Context, tenantID primitive.ObjectID, days int) ([]Vaccination, error) {
	return s.repo.FindDueVaccinations(ctx, tenantID, days)
}

// GetOverdueVaccinations gets overdue vaccinations
func (s *Service) GetOverdueVaccinations(ctx context.Context, tenantID primitive.ObjectID) ([]Vaccination, error) {
	return s.repo.FindOverdueVaccinations(ctx, tenantID)
}

// SendDueReminders sends reminders for vaccinations due within the specified days
func (s *Service) SendDueReminders(ctx context.Context, tenantID primitive.ObjectID, days int) error {
	vaccinations, err := s.GetDueVaccinations(ctx, tenantID, days)
	if err != nil {
		return err
	}

	for _, v := range vaccinations {
		daysUntil := v.DaysUntilDue()

		var title, body string
		if daysUntil == 0 {
			title = "Vacuna Vence Hoy"
			body = fmt.Sprintf("La vacuna %s de tu mascota vence hoy", v.VaccineName)
		} else if daysUntil == 1 {
			title = "Vacuna Vence Mañana"
			body = fmt.Sprintf("La vacuna %s de tu mascota vence mañana", v.VaccineName)
		} else {
			title = fmt.Sprintf("Vacuna Vence en %d días", daysUntil)
			body = fmt.Sprintf("La vacuna %s de tu mascota vence en %d días", v.VaccineName, daysUntil)
		}

		s.notificationSvc.Send(ctx, &notifications.SendDTO{
			OwnerID:  v.OwnerID.Hex(),
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeVaccinationDue,
			Title:    title,
			Body:     body,
			Data: map[string]string{
				"vaccination_id": v.ID.Hex(),
				"patient_id":     v.PatientID.Hex(),
				"vaccine_name":   v.VaccineName,
				"days_until_due": fmt.Sprintf("%d", daysUntil),
			},
			SendPush: true,
		})
	}

	return nil
}

// SendOverdueReminders sends reminders for overdue vaccinations
func (s *Service) SendOverdueReminders(ctx context.Context, tenantID primitive.ObjectID) error {
	vaccinations, err := s.GetOverdueVaccinations(ctx, tenantID)
	if err != nil {
		return err
	}

	for _, v := range vaccinations {
		daysOverdue := -v.DaysUntilDue()

		s.notificationSvc.Send(ctx, &notifications.SendDTO{
			OwnerID:  v.OwnerID.Hex(),
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeVaccinationDue,
			Title:    "Vacuna Vencida",
			Body:     fmt.Sprintf("La vacuna %s de tu mascota venció hace %d días. ¡Programa una cita!", v.VaccineName, daysOverdue),
			Data: map[string]string{
				"vaccination_id": v.ID.Hex(),
				"patient_id":     v.PatientID.Hex(),
				"vaccine_name":   v.VaccineName,
				"days_overdue":   fmt.Sprintf("%d", daysOverdue),
			},
			SendPush: true,
		})
	}

	return nil
}

// CreateVaccine creates a new vaccine in the catalog
func (s *Service) CreateVaccine(ctx context.Context, dto *CreateVaccineDTO, tenantID primitive.ObjectID) (*Vaccine, error) {
	// Validate dose type
	doseNumber := VaccineDoseType(dto.DoseNumber)
	if !IsValidVaccineDoseType(dto.DoseNumber) {
		return nil, ErrValidation("dose_number", "invalid dose type, must be first, second, or booster")
	}

	vaccine := &Vaccine{
		ID:             primitive.NewObjectID(),
		TenantID:       tenantID,
		Name:           dto.Name,
		Description:    dto.Description,
		Manufacturer:   dto.Manufacturer,
		DoseNumber:     doseNumber,
		ValidityMonths: dto.ValidityMonths,
		TargetSpecies:  dto.TargetSpecies,
		Active:         dto.Active,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateVaccine(ctx, vaccine); err != nil {
		return nil, err
	}

	return vaccine, nil
}

// GetVaccine gets a vaccine by ID
func (s *Service) GetVaccine(ctx context.Context, id string, tenantID primitive.ObjectID) (*Vaccine, error) {
	vaccineID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid vaccine ID format")
	}

	vaccine, err := s.repo.FindVaccineByID(ctx, vaccineID, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccine, nil
}

// ListVaccines lists all vaccines for a tenant
func (s *Service) ListVaccines(ctx context.Context, filters VaccineListFilters, tenantID primitive.ObjectID) ([]Vaccine, error) {
	return s.repo.FindVaccines(ctx, tenantID, filters)
}

// UpdateVaccine updates a vaccine
func (s *Service) UpdateVaccine(ctx context.Context, id string, dto *UpdateVaccineDTO, tenantID primitive.ObjectID) (*Vaccine, error) {
	vaccineID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid vaccine ID format")
	}

	_, err = s.repo.FindVaccineByID(ctx, vaccineID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}

	if dto.Name != "" {
		updates["name"] = dto.Name
	}

	if dto.Description != "" {
		updates["description"] = dto.Description
	}

	if dto.Manufacturer != "" {
		updates["manufacturer"] = dto.Manufacturer
	}

	if dto.DoseNumber != "" {
		if !IsValidVaccineDoseType(dto.DoseNumber) {
			return nil, ErrValidation("dose_number", "invalid dose type")
		}
		updates["dose_number"] = dto.DoseNumber
	}

	if dto.ValidityMonths > 0 {
		updates["validity_months"] = dto.ValidityMonths
	}

	if len(dto.TargetSpecies) > 0 {
		updates["target_species"] = dto.TargetSpecies
	}

	updates["active"] = dto.Active

	if err := s.repo.UpdateVaccine(ctx, vaccineID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedVaccine, err := s.repo.FindVaccineByID(ctx, vaccineID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedVaccine, nil
}

// DeleteVaccine soft deletes a vaccine
func (s *Service) DeleteVaccine(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	vaccineID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid vaccine ID format")
	}

	return s.repo.DeleteVaccine(ctx, vaccineID, tenantID)
}

// generateCertificateNumber generates a unique certificate number
func (s *Service) generateCertificateNumber(tenantID primitive.ObjectID, applicationDate time.Time) string {
	// Format: VAC-TENANT-YYYYMMDD-RANDOM
	// Example: VAC-507f-20260224-A1B2
	return fmt.Sprintf("VAC-%s-%s",
		tenantID.Hex()[:4],
		applicationDate.Format("20060102"),
	)
}
