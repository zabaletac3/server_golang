package fcm

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/platform/circuitbreaker"
	"github.com/eren_dev/go_server/internal/platform/notifications"
)

type fcmProvider struct {
	client *messaging.Client
}

// NewProvider initializes the FCM provider from the credentials file path in config.
// Returns nil (disabled provider) if no credentials path is set.
func NewProvider(ctx context.Context, cfg *config.Config) (notifications.PushProvider, error) {
	if cfg.FirebaseCredentialsPath == "" {
		slog.Info("FCM disabled: FIREBASE_CREDENTIALS_PATH not set")
		return &fcmProvider{client: nil}, nil
	}

	// Extract project_id from the service account JSON so the Messaging client
	// can be initialized without requiring an extra environment variable.
	projectID, err := readProjectID(cfg.FirebaseCredentialsPath)
	if err != nil {
		return nil, err
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, option.WithCredentialsFile(cfg.FirebaseCredentialsPath))
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	slog.Info("FCM push notifications enabled")
	return &fcmProvider{client: client}, nil
}

func readProjectID(credentialsPath string) (string, error) {
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return "", err
	}
	var creds struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", err
	}
	return creds.ProjectID, nil
}

func (p *fcmProvider) IsEnabled() bool {
	return p.client != nil
}

func (p *fcmProvider) Send(ctx context.Context, tokens []string, payload notifications.PushPayload) error {
	if !p.IsEnabled() || len(tokens) == 0 {
		return nil
	}

	_, err := circuitbreaker.ExecuteWithFCMBreaker(func() (interface{}, error) {
		return nil, p.sendImpl(ctx, tokens, payload)
	})
	return err
}

func (p *fcmProvider) sendImpl(ctx context.Context, tokens []string, payload notifications.PushPayload) error {
	// Use SendEachForMulticast for batching (up to 500 tokens per call)
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: payload.Title,
			Body:  payload.Body,
		},
		Data: payload.Data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	resp, err := p.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return err
	}

	if resp.FailureCount > 0 {
		for i, r := range resp.Responses {
			if !r.Success {
				slog.Warn("FCM send failed for token",
					"token_index", i,
					"error", r.Error,
				)
			}
		}
	}

	return nil
}
