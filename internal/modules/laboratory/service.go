package laboratory

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

// Service provides business logic for laboratory
type Service struct {
	repo            LabOrderRepository
	patientRepo     PatientRepository
	userRepo        UserRepository
	notificationSvc NotificationSender
}

// NewService creates a new laboratory service
func NewService(repo LabOrderRepository, patientRepo PatientRepository, userRepo UserRepository, notificationSvc NotificationSender) *Service {
	return &Service{
		repo:            repo,
		patientRepo:     patientRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
	}
}

// CreateLabOrder creates a new lab order
func (s *Service) CreateLabOrder(ctx context.Context, dto *CreateLabOrderDTO, tenantID primitive.ObjectID) (*LabOrder, error) {
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

	// Validate test type
	testType := LabTestType(dto.TestType)
	if !IsValidLabTestType(dto.TestType) {
		return nil, ErrValidation("test_type", "invalid test type")
	}

	// Create lab order
	now := time.Now()
	order := &LabOrder{
		ID:             primitive.NewObjectID(),
		TenantID:       tenantID,
		PatientID:      patientID,
		OwnerID:        patient.OwnerID,
		VeterinarianID: vetID,
		OrderDate:      now,
		LabID:          dto.LabID,
		TestType:       testType,
		Status:         LabOrderStatusPending,
		Notes:          dto.Notes,
		Cost:           dto.Cost,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Send notification to owner
	s.notificationSvc.Send(ctx, &notifications.SendDTO{
		OwnerID:  patient.OwnerID.Hex(),
		TenantID: tenantID.Hex(),
		Type:     notifications.TypeGeneral,
		Title:    "Orden de Laboratorio Creada",
		Body:     fmt.Sprintf("Se ha creado una orden de laboratorio (%s) para %s", dto.TestType, patient.Name),
		Data: map[string]string{
			"order_id":   order.ID.Hex(),
			"patient_id": order.PatientID.Hex(),
			"test_type":  dto.TestType,
		},
		SendPush: true,
	})

	return order, nil
}

// GetLabOrder gets a lab order by ID
func (s *Service) GetLabOrder(ctx context.Context, id string, tenantID primitive.ObjectID) (*LabOrder, error) {
	orderID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid order ID format")
	}

	order, err := s.repo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// ListLabOrders lists lab orders with filters
func (s *Service) ListLabOrders(ctx context.Context, filters LabOrderListFilters, tenantID primitive.ObjectID, params pagination.Params) ([]LabOrder, int64, error) {
	return s.repo.FindByFilters(ctx, tenantID, filters, params)
}

// GetPatientLabOrders gets all lab orders for a patient
func (s *Service) GetPatientLabOrders(ctx context.Context, patientID string, tenantID primitive.ObjectID, params pagination.Params) ([]LabOrder, int64, error) {
	pID, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return nil, 0, ErrValidation("patient_id", "invalid patient ID format")
	}

	return s.repo.FindByPatient(ctx, pID, tenantID, params)
}

// UpdateLabOrder updates a lab order
func (s *Service) UpdateLabOrder(ctx context.Context, id string, dto *UpdateLabOrderDTO, tenantID primitive.ObjectID) (*LabOrder, error) {
	orderID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid order ID format")
	}

	_, err = s.repo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}

	if dto.LabID != "" {
		updates["lab_id"] = dto.LabID
	}

	if dto.TestType != "" {
		if !IsValidLabTestType(dto.TestType) {
			return nil, ErrValidation("test_type", "invalid test type")
		}
		updates["test_type"] = dto.TestType
	}

	if dto.Notes != "" {
		updates["notes"] = dto.Notes
	}

	if dto.Cost > 0 {
		updates["cost"] = dto.Cost
	}

	if err := s.repo.Update(ctx, orderID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedOrder, err := s.repo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedOrder, nil
}

// UpdateLabOrderStatus updates the status of a lab order
func (s *Service) UpdateLabOrderStatus(ctx context.Context, id string, dto *UpdateLabOrderStatusDTO, tenantID primitive.ObjectID) (*LabOrder, error) {
	orderID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid order ID format")
	}

	order, err := s.repo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	newStatus := LabOrderStatus(dto.Status)
	if !IsValidStatusTransition(order.Status, newStatus) {
		return nil, ErrInvalidStatusTransition
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, orderID, newStatus, tenantID); err != nil {
		return nil, err
	}

	// Send notification if status is processed
	if newStatus == LabOrderStatusProcessed {
		updatedOrder, err := s.repo.FindByID(ctx, orderID, tenantID)
		if err != nil {
			return nil, err
		}

		patient, err := s.patientRepo.FindByID(ctx, tenantID, updatedOrder.PatientID.Hex())
		if err == nil {
			s.notificationSvc.Send(ctx, &notifications.SendDTO{
				OwnerID:  patient.OwnerID.Hex(),
				TenantID: tenantID.Hex(),
				Type:     notifications.TypeGeneral,
				Title:    "Resultados de Laboratorio Listos",
				Body:     fmt.Sprintf("Los resultados de %s de %s están listos", updatedOrder.TestType, patient.Name),
				Data: map[string]string{
					"order_id":   updatedOrder.ID.Hex(),
					"patient_id": updatedOrder.PatientID.Hex(),
				},
				SendPush: true,
			})
		}
	}

	updatedOrder, err := s.repo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedOrder, nil
}

// UploadLabResult uploads a lab result file
func (s *Service) UploadLabResult(ctx context.Context, id string, dto *UploadLabResultDTO, tenantID primitive.ObjectID) (*LabOrder, error) {
	orderID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid order ID format")
	}

	order, err := s.repo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}

	// Check if result already uploaded
	if order.ResultFileID != "" {
		return nil, ErrResultAlreadyUploaded
	}

	updates := bson.M{
		"result_file_id": dto.ResultFileID,
	}

	if dto.Notes != "" {
		updates["notes"] = order.Notes + "\n" + dto.Notes
	}

	if err := s.repo.Update(ctx, orderID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedOrder, err := s.repo.FindByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedOrder, nil
}

// DeleteLabOrder soft deletes a lab order
func (s *Service) DeleteLabOrder(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	orderID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid order ID format")
	}

	return s.repo.Delete(ctx, orderID, tenantID)
}

// GetOverdueLabOrders gets overdue lab orders
func (s *Service) GetOverdueLabOrders(ctx context.Context, tenantID primitive.ObjectID, turnaroundDays int) ([]LabOrder, error) {
	return s.repo.FindOverdueOrders(ctx, tenantID, turnaroundDays)
}

// SendOverdueReminders sends reminders for overdue lab orders
func (s *Service) SendOverdueReminders(ctx context.Context, tenantID primitive.ObjectID, turnaroundDays int) error {
	orders, err := s.GetOverdueLabOrders(ctx, tenantID, turnaroundDays)
	if err != nil {
		return err
	}

	for _, o := range orders {
		daysOverdue := o.DaysSinceOrder() - turnaroundDays

		// Send to staff
		s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
			UserID:   primitive.NilObjectID.Hex(), // Broadcast to all staff
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeStaffSystemAlert,
			Title:    "Orden de Laboratorio Vencida",
			Body:     fmt.Sprintf("La orden de %s de %s está vencida hace %d días", o.TestType, o.PatientID.Hex(), daysOverdue),
			Data: map[string]string{
				"order_id":       o.ID.Hex(),
				"patient_id":     o.PatientID.Hex(),
				"days_overdue":   fmt.Sprintf("%d", daysOverdue),
			},
		})
	}

	return nil
}

// CreateLabTest creates a new lab test in the catalog
func (s *Service) CreateLabTest(ctx context.Context, dto *CreateLabTestDTO, tenantID primitive.ObjectID) (*LabTest, error) {
	test := &LabTest{
		ID:             primitive.NewObjectID(),
		TenantID:       tenantID,
		Name:           dto.Name,
		Description:    dto.Description,
		Category:       dto.Category,
		Price:          dto.Price,
		TurnaroundTime: dto.TurnaroundTime,
		Active:         dto.Active,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateLabTest(ctx, test); err != nil {
		return nil, err
	}

	return test, nil
}

// GetLabTest gets a lab test by ID
func (s *Service) GetLabTest(ctx context.Context, id string, tenantID primitive.ObjectID) (*LabTest, error) {
	testID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid test ID format")
	}

	test, err := s.repo.FindLabTestByID(ctx, testID, tenantID)
	if err != nil {
		return nil, err
	}

	return test, nil
}

// ListLabTests lists all lab tests for a tenant
func (s *Service) ListLabTests(ctx context.Context, filters LabTestListFilters, tenantID primitive.ObjectID) ([]LabTest, error) {
	return s.repo.FindLabTests(ctx, tenantID, filters)
}

// UpdateLabTest updates a lab test
func (s *Service) UpdateLabTest(ctx context.Context, id string, dto *UpdateLabTestDTO, tenantID primitive.ObjectID) (*LabTest, error) {
	testID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid test ID format")
	}

	_, err = s.repo.FindLabTestByID(ctx, testID, tenantID)
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

	if dto.Category != "" {
		updates["category"] = dto.Category
	}

	if dto.Price > 0 {
		updates["price"] = dto.Price
	}

	if dto.TurnaroundTime > 0 {
		updates["turnaround_time"] = dto.TurnaroundTime
	}

	updates["active"] = dto.Active

	if err := s.repo.UpdateLabTest(ctx, testID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedTest, err := s.repo.FindLabTestByID(ctx, testID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedTest, nil
}

// DeleteLabTest soft deletes a lab test
func (s *Service) DeleteLabTest(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	testID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid test ID format")
	}

	return s.repo.DeleteLabTest(ctx, testID, tenantID)
}
