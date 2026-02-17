package notifications

import (
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// SendDTO is used internally by other modules to trigger a notification.
type SendDTO struct {
	OwnerID  string
	TenantID string
	Type     NotificationType
	Title    string
	Body     string
	// Data is forwarded as FCM data payload and stored for deep-linking.
	Data     map[string]string
	// SendPush controls whether a push notification is also sent via FCM.
	SendPush bool
}

// --- Response DTOs ---

type NotificationResponse struct {
	ID         string            `json:"id"`
	Type       NotificationType  `json:"type"`
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	Data       map[string]string `json:"data,omitempty"`
	Read       bool              `json:"read"`
	ReadAt     *time.Time        `json:"read_at,omitempty"`
	PushSent   bool              `json:"push_sent"`
	CreatedAt  time.Time         `json:"created_at"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

type PaginatedNotificationsResponse struct {
	Data       []NotificationResponse    `json:"data"`
	Pagination pagination.PaginationInfo `json:"pagination"`
}

func toResponse(n *Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID.Hex(),
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Data:      n.Data,
		Read:      n.Read,
		ReadAt:    n.ReadAt,
		PushSent:  n.PushSent,
		CreatedAt: n.CreatedAt,
	}
}

// --- Staff DTOs ---

// SendStaffDTO is used internally to send a notification to a specific staff user.
type SendStaffDTO struct {
	UserID   string
	TenantID string
	Type     StaffNotificationType
	Title    string
	Body     string
	Data     map[string]string
}

type StaffNotificationResponse struct {
	ID        string                `json:"id"`
	Type      StaffNotificationType `json:"type"`
	Title     string                `json:"title"`
	Body      string                `json:"body"`
	Data      map[string]string     `json:"data,omitempty"`
	Read      bool                  `json:"read"`
	ReadAt    *time.Time            `json:"read_at,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
}

type PaginatedStaffNotificationsResponse struct {
	Data       []StaffNotificationResponse `json:"data"`
	Pagination pagination.PaginationInfo   `json:"pagination"`
}

func toStaffResponse(n *StaffNotification) StaffNotificationResponse {
	return StaffNotificationResponse{
		ID:        n.ID.Hex(),
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Data:      n.Data,
		Read:      n.Read,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}
