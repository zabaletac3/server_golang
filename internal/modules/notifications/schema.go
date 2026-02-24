package notifications

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Owner notification types ---

type NotificationType string

const (
	TypeAppointmentConfirmed NotificationType = "appointment_confirmed"
	TypeAppointmentCancelled NotificationType = "appointment_cancelled"
	TypeAppointmentReminder  NotificationType = "appointment_reminder"
	TypeVaccinationDue       NotificationType = "vaccination_due"
	TypeMedicalRecordCreated NotificationType = "medical_record_created"
	TypeMedicalRecordUpdated NotificationType = "medical_record_updated"
	TypePrescriptionReady    NotificationType = "prescription_ready"
	TypeGeneral              NotificationType = "general"
)

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	OwnerID   primitive.ObjectID `bson:"owner_id"`
	TenantID  primitive.ObjectID `bson:"tenant_id"`
	Type      NotificationType   `bson:"type"`
	Title     string             `bson:"title"`
	Body      string             `bson:"body"`
	// Data carries context for deep-linking (appointment_id, patient_id, etc.)
	Data       map[string]string `bson:"data,omitempty"`
	Read       bool              `bson:"read"`
	ReadAt     *time.Time        `bson:"read_at,omitempty"`
	PushSent   bool              `bson:"push_sent"`
	PushSentAt *time.Time        `bson:"push_sent_at,omitempty"`
	CreatedAt  time.Time         `bson:"created_at"`
}

// --- Staff notification types ---

type StaffNotificationType string

const (
	TypeStaffNewAppointment  StaffNotificationType = "new_appointment"
	TypeStaffPaymentReceived StaffNotificationType = "payment_received"
	TypeStaffNewPatient      StaffNotificationType = "new_patient"
	TypeStaffSystemAlert     StaffNotificationType = "system_alert"
	TypeStaffGeneral         StaffNotificationType = "general"
)

// StaffNotification is stored in the staff_notifications collection.
// Each record targets a specific staff user within a tenant.
type StaffNotification struct {
	ID        primitive.ObjectID    `bson:"_id,omitempty"`
	UserID    primitive.ObjectID    `bson:"user_id"`
	TenantID  primitive.ObjectID    `bson:"tenant_id"`
	Type      StaffNotificationType `bson:"type"`
	Title     string                `bson:"title"`
	Body      string                `bson:"body"`
	Data      map[string]string     `bson:"data,omitempty"`
	Read      bool                  `bson:"read"`
	ReadAt    *time.Time            `bson:"read_at,omitempty"`
	CreatedAt time.Time             `bson:"created_at"`
}
