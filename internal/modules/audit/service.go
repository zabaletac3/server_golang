package audit

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service provides audit logging functionality
type Service struct {
	repo Repository
}

// NewService creates a new audit service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// LogOptions contains optional fields for audit logging
type LogOptions struct {
	IPAddress  string
	UserAgent  string
	Metadata   map[string]interface{}
	ResourceID primitive.ObjectID
}

// Log creates a new audit log entry
func (s *Service) Log(ctx context.Context, tenantID, userID primitive.ObjectID, eventType EventType, resource, action, description string, opts *LogOptions) error {
	event := &AuditEvent{
		ID:          primitive.NewObjectID(),
		TenantID:    tenantID,
		UserID:      userID,
		EventType:   eventType,
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}

	if opts != nil {
		event.IPAddress = opts.IPAddress
		event.UserAgent = opts.UserAgent
		event.Metadata = opts.Metadata
		event.ResourceID = opts.ResourceID
	}

	return s.repo.Create(ctx, event)
}

// LogUserAction logs a user-specific action
func (s *Service) LogUserAction(ctx context.Context, tenantID, userID primitive.ObjectID, eventType EventType, action, description string) error {
	return s.Log(ctx, tenantID, userID, eventType, "user", action, description, nil)
}

// LogTenantAction logs a tenant-specific action
func (s *Service) LogTenantAction(ctx context.Context, tenantID, userID primitive.ObjectID, eventType EventType, action, description string, metadata map[string]interface{}) error {
	return s.Log(ctx, tenantID, userID, eventType, "tenant", action, description, &LogOptions{
		Metadata: metadata,
	})
}

// LogPaymentAction logs a payment-related action
func (s *Service) LogPaymentAction(ctx context.Context, tenantID, userID, paymentID primitive.ObjectID, eventType EventType, action, description string) error {
	return s.Log(ctx, tenantID, userID, eventType, "payment", action, description, &LogOptions{
		ResourceID: paymentID,
	})
}

// LogAppointmentAction logs an appointment-related action
func (s *Service) LogAppointmentAction(ctx context.Context, tenantID, userID, appointmentID primitive.ObjectID, eventType EventType, action, description string) error {
	return s.Log(ctx, tenantID, userID, eventType, "appointment", action, description, &LogOptions{
		ResourceID: appointmentID,
	})
}

// LogPatientAction logs a patient-related action
func (s *Service) LogPatientAction(ctx context.Context, tenantID, userID, patientID primitive.ObjectID, eventType EventType, action, description string) error {
	return s.Log(ctx, tenantID, userID, eventType, "patient", action, description, &LogOptions{
		ResourceID: patientID,
	})
}

// LogRBACAction logs a role/permission-related action
func (s *Service) LogRBACAction(ctx context.Context, tenantID, userID primitive.ObjectID, eventType EventType, resource, action, description string, roleID primitive.ObjectID) error {
	return s.Log(ctx, tenantID, userID, eventType, resource, action, description, &LogOptions{
		ResourceID: roleID,
	})
}

// GetEvents retrieves audit events with filters
func (s *Service) GetEvents(ctx context.Context, filter AuditFilter) ([]AuditEvent, int64, error) {
	return s.repo.FindByFilter(ctx, filter)
}

// GetEventsByResource retrieves audit events for a specific resource
func (s *Service) GetEventsByResource(ctx context.Context, tenantID, resourceID primitive.ObjectID, resource string) ([]AuditEvent, error) {
	return s.repo.FindByResource(ctx, tenantID, resourceID, resource)
}

// GetEventsByUser retrieves audit events for a specific user
func (s *Service) GetEventsByUser(ctx context.Context, tenantID, userID primitive.ObjectID) ([]AuditEvent, error) {
	return s.repo.FindByUser(ctx, tenantID, userID)
}

// GetEventsByType retrieves audit events of a specific type
func (s *Service) GetEventsByType(ctx context.Context, tenantID primitive.ObjectID, eventType EventType) ([]AuditEvent, error) {
	return s.repo.FindByEventType(ctx, tenantID, eventType)
}
