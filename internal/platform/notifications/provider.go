package notifications

import "context"

// PushPayload is the data sent to the device.
type PushPayload struct {
	Title string
	Body  string
	// Data is extra key-value context for the mobile app (appointment_id, patient_id, etc.)
	Data map[string]string
}

// PushProvider sends push notifications to device tokens.
// The implementation is nil-safe: callers should check IsEnabled() before sending.
type PushProvider interface {
	// Send delivers a push notification to one or more FCM/APNs tokens.
	// Tokens that are no longer valid should be silently ignored by the implementation.
	Send(ctx context.Context, tokens []string, payload PushPayload) error
	// IsEnabled returns false when the provider was not configured (e.g. no credentials).
	IsEnabled() bool
}
