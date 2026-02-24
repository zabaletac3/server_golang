package audit

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventType represents the type of audit event
type EventType string

const (
	// Authentication events
	EventLogin           EventType = "auth.login"
	EventLogout          EventType = "auth.logout"
	EventPasswordChanged EventType = "auth.password_changed"

	// User management events
	EventUserCreated     EventType = "user.created"
	EventUserUpdated     EventType = "user.updated"
	EventUserDeleted     EventType = "user.deleted"
	EventRoleAssigned    EventType = "user.role_assigned"
	EventRoleRevoked     EventType = "user.role_revoked"

	// Tenant events
	EventTenantCreated      EventType = "tenant.created"
	EventTenantUpdated      EventType = "tenant.updated"
	EventTenantDeleted      EventType = "tenant.deleted"
	EventTenantSubscription EventType = "tenant.subscription_changed"
	EventTenantStatusChange EventType = "tenant.status_changed"

	// Payment events
	EventPaymentCreated   EventType = "payment.created"
	EventPaymentSucceeded EventType = "payment.succeeded"
	EventPaymentFailed    EventType = "payment.failed"
	EventPaymentRefunded  EventType = "payment.refunded"

	// Appointment events
	EventAppointmentCreated   EventType = "appointment.created"
	EventAppointmentUpdated   EventType = "appointment.updated"
	EventAppointmentCancelled EventType = "appointment.cancelled"
	EventAppointmentCompleted EventType = "appointment.completed"

	// Patient events
	EventPatientCreated EventType = "patient.created"
	EventPatientUpdated EventType = "patient.updated"
	EventPatientDeleted EventType = "patient.deleted"

	// RBAC events
	EventRoleCreated    EventType = "role.created"
	EventRoleUpdated    EventType = "role.updated"
	EventRoleDeleted    EventType = "role.deleted"
	EventPermissionGrant EventType = "permission.granted"
	EventPermissionRevoke EventType = "permission.revoked"

	// System events
	EventSystemAlert EventType = "system.alert"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID          primitive.ObjectID       `bson:"_id"`
	TenantID    primitive.ObjectID       `bson:"tenant_id"`
	UserID      primitive.ObjectID       `bson:"user_id"`
	EventType   EventType                `bson:"event_type"`
	Resource    string                   `bson:"resource"`
	ResourceID  primitive.ObjectID       `bson:"resource_id,omitempty"`
	Action      string                   `bson:"action"`
	Description string                   `bson:"description"`
	Metadata    map[string]interface{}   `bson:"metadata,omitempty"`
	IPAddress   string                   `bson:"ip_address,omitempty"`
	UserAgent   string                   `bson:"user_agent,omitempty"`
	CreatedAt   time.Time                `bson:"created_at"`
}

// AuditFilter represents filters for querying audit events
type AuditFilter struct {
	TenantID    primitive.ObjectID
	UserID      *primitive.ObjectID
	EventType   *EventType
	Resource    *string
	DateFrom    *time.Time
	DateTo      *time.Time
	Limit       int
	Skip        int
}

// AuditEventResponse represents an audit event in API responses
type AuditEventResponse struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	UserID      string                 `json:"user_id"`
	EventType   string                 `json:"event_type"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ToResponse converts an AuditEvent to AuditEventResponse
func (e *AuditEvent) ToResponse() *AuditEventResponse {
	return &AuditEventResponse{
		ID:          e.ID.Hex(),
		TenantID:    e.TenantID.Hex(),
		UserID:      e.UserID.Hex(),
		EventType:   string(e.EventType),
		Resource:    e.Resource,
		ResourceID:  e.ResourceID.Hex(),
		Action:      e.Action,
		Description: e.Description,
		Metadata:    e.Metadata,
		IPAddress:   e.IPAddress,
		CreatedAt:   e.CreatedAt,
	}
}
