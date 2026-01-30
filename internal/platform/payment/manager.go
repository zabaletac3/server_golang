package payment

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrProviderNotFound      = errors.New("payment provider not found")
	ErrProviderAlreadyExists = errors.New("payment provider already exists")
	ErrNoDefaultProvider     = errors.New("no default payment provider configured")
)

// PaymentManager gestiona múltiples proveedores de pago
type PaymentManager struct {
	providers       map[ProviderType]PaymentProvider
	defaultProvider ProviderType
}

// NewPaymentManager crea un nuevo manager de pagos
func NewPaymentManager(defaultProvider ProviderType) *PaymentManager {
	return &PaymentManager{
		providers:       make(map[ProviderType]PaymentProvider),
		defaultProvider: defaultProvider,
	}
}

// RegisterProvider registra un proveedor de pago
func (m *PaymentManager) RegisterProvider(provider PaymentProvider) error {
	providerType := provider.GetProviderType()
	
	if _, exists := m.providers[providerType]; exists {
		return ErrProviderAlreadyExists
	}
	
	m.providers[providerType] = provider
	return nil
}

// GetProvider obtiene un proveedor específico
func (m *PaymentManager) GetProvider(providerType ProviderType) (PaymentProvider, error) {
	provider, exists := m.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerType)
	}
	return provider, nil
}

// GetDefaultProvider obtiene el proveedor por defecto
func (m *PaymentManager) GetDefaultProvider() (PaymentProvider, error) {
	if m.defaultProvider == "" {
		return nil, ErrNoDefaultProvider
	}
	return m.GetProvider(m.defaultProvider)
}

// CreateSubscription crea una suscripción usando el proveedor especificado o el default
func (m *PaymentManager) CreateSubscription(ctx context.Context, req *SubscriptionRequest, providerType *ProviderType) (*SubscriptionResponse, error) {
	var provider PaymentProvider
	var err error
	
	if providerType != nil && *providerType != "" {
		provider, err = m.GetProvider(*providerType)
	} else {
		provider, err = m.GetDefaultProvider()
	}
	
	if err != nil {
		return nil, err
	}
	
	return provider.CreateSubscription(ctx, req)
}

// CancelSubscription cancela una suscripción
func (m *PaymentManager) CancelSubscription(ctx context.Context, subscriptionID string, providerType ProviderType) error {
	provider, err := m.GetProvider(providerType)
	if err != nil {
		return err
	}
	
	return provider.CancelSubscription(ctx, subscriptionID)
}

// GetSubscription obtiene información de una suscripción
func (m *PaymentManager) GetSubscription(ctx context.Context, subscriptionID string, providerType ProviderType) (*SubscriptionResponse, error) {
	provider, err := m.GetProvider(providerType)
	if err != nil {
		return nil, err
	}
	
	return provider.GetSubscription(ctx, subscriptionID)
}

// ProcessWebhook procesa un webhook del proveedor especificado
func (m *PaymentManager) ProcessWebhook(ctx context.Context, providerType ProviderType, payload []byte, signature string) (*WebhookEvent, error) {
	provider, err := m.GetProvider(providerType)
	if err != nil {
		return nil, err
	}
	
	return provider.ProcessWebhook(ctx, payload, signature)
}

// ListProviders lista todos los proveedores registrados
func (m *PaymentManager) ListProviders() []ProviderType {
	providers := make([]ProviderType, 0, len(m.providers))
	for providerType := range m.providers {
		providers = append(providers, providerType)
	}
	return providers
}
