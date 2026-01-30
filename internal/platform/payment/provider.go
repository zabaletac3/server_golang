package payment

import (
	"context"
	"time"
)

// ProviderType tipo de proveedor de pago
type ProviderType string

const (
	ProviderWompi  ProviderType = "wompi"
	ProviderStripe ProviderType = "stripe"
)

// SubscriptionRequest datos para crear suscripción
type SubscriptionRequest struct {
	TenantID       string
	PlanID         string
	CustomerEmail  string
	CustomerName   string
	PaymentMethod  string
	BillingPeriod  string // monthly, annual
	Amount         int64  // en centavos
	Currency       string
	TrialDays      int
}

// SubscriptionResponse respuesta de suscripción creada
type SubscriptionResponse struct {
	SubscriptionID string
	Status         string
	TrialEndsAt    *time.Time
	NextBillingAt  *time.Time
	Amount         int64
	Currency       string
}

// WebhookEvent evento de webhook
type WebhookEvent struct {
	Provider      ProviderType
	EventType     string
	SubscriptionID string
	TransactionID string
	Status        string
	Amount        int64
	Currency      string
	Metadata      map[string]interface{}
	RawPayload    []byte
}

// PaymentProvider interfaz que todos los proveedores deben implementar
type PaymentProvider interface {
	// CreateSubscription crea una nueva suscripción
	CreateSubscription(ctx context.Context, req *SubscriptionRequest) (*SubscriptionResponse, error)
	
	// CancelSubscription cancela una suscripción existente
	CancelSubscription(ctx context.Context, subscriptionID string) error
	
	// GetSubscription obtiene información de una suscripción
	GetSubscription(ctx context.Context, subscriptionID string) (*SubscriptionResponse, error)
	
	// ProcessWebhook procesa un webhook del proveedor
	ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)
	
	// GetProviderType retorna el tipo de proveedor
	GetProviderType() ProviderType
}
