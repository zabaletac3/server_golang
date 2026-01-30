package wompi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eren_dev/go_server/internal/platform/payment"
)

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrInvalidPayload   = errors.New("invalid webhook payload")
)

// WompiProvider implementación del proveedor Wompi
type WompiProvider struct {
	publicKey     string
	privateKey    string
	webhookSecret string
	baseURL       string
}

// NewWompiProvider crea una nueva instancia del proveedor Wompi
func NewWompiProvider(publicKey, privateKey, webhookSecret string) *WompiProvider {
	return &WompiProvider{
		publicKey:     publicKey,
		privateKey:    privateKey,
		webhookSecret: webhookSecret,
		baseURL:       "https://production.wompi.co/v1",
	}
}

// GetProviderType retorna el tipo de proveedor
func (w *WompiProvider) GetProviderType() payment.ProviderType {
	return payment.ProviderWompi
}

// CreateSubscription crea una nueva suscripción en Wompi
func (w *WompiProvider) CreateSubscription(ctx context.Context, req *payment.SubscriptionRequest) (*payment.SubscriptionResponse, error) {
	// TODO: Implementar llamada real a API de Wompi
	// Por ahora retornamos una respuesta simulada
	
	return &payment.SubscriptionResponse{
		SubscriptionID: fmt.Sprintf("wompi_sub_%s", req.TenantID),
		Status:         "active",
		Amount:         req.Amount,
		Currency:       req.Currency,
	}, nil
}

// CancelSubscription cancela una suscripción en Wompi
func (w *WompiProvider) CancelSubscription(ctx context.Context, subscriptionID string) error {
	// TODO: Implementar llamada real a API de Wompi
	return nil
}

// GetSubscription obtiene información de una suscripción
func (w *WompiProvider) GetSubscription(ctx context.Context, subscriptionID string) (*payment.SubscriptionResponse, error) {
	// TODO: Implementar llamada real a API de Wompi
	
	return &payment.SubscriptionResponse{
		SubscriptionID: subscriptionID,
		Status:         "active",
	}, nil
}

// ProcessWebhook procesa un webhook de Wompi
func (w *WompiProvider) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	// Validar firma del webhook
	if !w.validateSignature(payload, signature) {
		return nil, ErrInvalidSignature
	}

	// Parsear payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, ErrInvalidPayload
	}

	// Extraer datos del evento
	event := &payment.WebhookEvent{
		Provider:   payment.ProviderWompi,
		RawPayload: payload,
		Metadata:   webhookData,
	}

	// Extraer campos específicos de Wompi
	if eventType, ok := webhookData["event"].(string); ok {
		event.EventType = eventType
	}

	if data, ok := webhookData["data"].(map[string]interface{}); ok {
		if transactionID, ok := data["id"].(string); ok {
			event.TransactionID = transactionID
		}
		if status, ok := data["status"].(string); ok {
			event.Status = status
		}
		if amountInCents, ok := data["amount_in_cents"].(float64); ok {
			event.Amount = int64(amountInCents)
		}
		if currency, ok := data["currency"].(string); ok {
			event.Currency = currency
		}
	}

	return event, nil
}

// validateSignature valida la firma del webhook
func (w *WompiProvider) validateSignature(payload []byte, signature string) bool {
	if w.webhookSecret == "" {
		// Si no hay secret configurado, aceptar (solo para desarrollo)
		return true
	}

	// Calcular HMAC SHA256
	mac := hmac.New(sha256.New, []byte(w.webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
