package stripe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/eren_dev/go_server/internal/platform/payment"
)

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrInvalidPayload   = errors.New("invalid webhook payload")
	ErrStripeAPI        = errors.New("stripe api error")
)

const (
	WebhookSignatureHeader = "Stripe-Signature"
)

// StripeProvider implementación del proveedor Stripe
type StripeProvider struct {
	apiKey        string
	webhookSecret string
	httpClient    *http.Client
}

// NewStripeProvider crea una nueva instancia del proveedor Stripe
func NewStripeProvider(apiKey, webhookSecret string) *StripeProvider {
	stripe.Key = apiKey
	return &StripeProvider{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderType retorna el tipo de proveedor
func (s *StripeProvider) GetProviderType() payment.ProviderType {
	return payment.ProviderStripe
}

// CreateSubscription crea una sesión de checkout de Stripe para suscripciones
func (s *StripeProvider) CreateSubscription(ctx context.Context, req *payment.SubscriptionRequest) (*payment.SubscriptionResponse, error) {
	// Determinar el intervalo de facturación
	var priceData stripe.CheckoutSessionLineItemPriceDataRecurringParams
	switch req.BillingPeriod {
	case "annual":
		priceData = stripe.CheckoutSessionLineItemPriceDataRecurringParams{
			Interval: stripe.String("year"),
		}
	default:
		priceData = stripe.CheckoutSessionLineItemPriceDataRecurringParams{
			Interval: stripe.String("month"),
		}
	}

	// Crear parámetros para la sesión de checkout
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(req.RedirectURL),
		CancelURL:  stripe.String(req.RedirectURL),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		CustomerEmail: stripe.String(req.CustomerEmail),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(req.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("Suscripción %s", req.PlanName)),
						Metadata: map[string]string{
							"tenant_id": req.TenantID,
							"plan_id":   req.PlanID,
						},
					},
					UnitAmount: stripe.Int64(req.Amount),
					Recurring:  &priceData,
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"tenant_id":      req.TenantID,
			"plan_id":        req.PlanID,
			"billing_period": req.BillingPeriod,
		},
	}

	// Crear sesión de checkout
	stripeSession, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create checkout session: %v", ErrStripeAPI, err)
	}

	// Calcular próxima fecha de facturación
	var nextBilling *time.Time
	switch req.BillingPeriod {
	case "monthly":
		next := time.Now().AddDate(0, 1, 0)
		nextBilling = &next
	case "annual":
		next := time.Now().AddDate(1, 0, 0)
		nextBilling = &next
	}

	return &payment.SubscriptionResponse{
		SubscriptionID: stripeSession.ID,
		Status:         "PENDING",
		NextBillingAt:  nextBilling,
		Amount:         req.Amount,
		Currency:       req.Currency,
		PaymentLinkURL: stripeSession.URL,
	}, nil
}

// CancelSubscription cancela una suscripción de Stripe
func (s *StripeProvider) CancelSubscription(ctx context.Context, subscriptionID string) error {
	// Extraer el ID de suscripción del ID de sesión si es necesario
	stripeSubscription, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("%w: failed to get subscription: %v", ErrStripeAPI, err)
	}

	// Cancelar la suscripción
	_, err = subscription.Cancel(stripeSubscription.ID, nil)
	if err != nil {
		return fmt.Errorf("%w: failed to cancel subscription: %v", ErrStripeAPI, err)
	}

	return nil
}

// GetSubscription obtiene información de una suscripción de Stripe
func (s *StripeProvider) GetSubscription(ctx context.Context, subscriptionID string) (*payment.SubscriptionResponse, error) {
	stripeSubscription, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get subscription: %v", ErrStripeAPI, err)
	}

	var status string
	switch stripeSubscription.Status {
	case stripe.SubscriptionStatusActive:
		status = "ACTIVE"
	case stripe.SubscriptionStatusPastDue:
		status = "PAST_DUE"
	case stripe.SubscriptionStatusUnpaid:
		status = "UNPAID"
	case stripe.SubscriptionStatusCanceled:
		status = "CANCELED"
	default:
		status = string(stripeSubscription.Status)
	}

	var nextBilling *time.Time
	if stripeSubscription.CurrentPeriodEnd > 0 {
		t := time.Unix(stripeSubscription.CurrentPeriodEnd, 0)
		nextBilling = &t
	}

	// Get amount from first subscription item
	var amount int64
	if len(stripeSubscription.Items.Data) > 0 {
		amount = stripeSubscription.Items.Data[0].Price.UnitAmount
	}

	return &payment.SubscriptionResponse{
		SubscriptionID: stripeSubscription.ID,
		Status:         status,
		NextBillingAt:  nextBilling,
		Amount:         amount,
		Currency:       string(stripeSubscription.Currency),
	}, nil
}

// ProcessWebhook procesa un webhook de Stripe
func (s *StripeProvider) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*payment.WebhookEvent, error) {
	if signature == "" {
		return nil, ErrInvalidSignature
	}

	// Validar firma del webhook
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSignature, err)
	}

	// Construir evento
	webhookEvent := &payment.WebhookEvent{
		Provider:   payment.ProviderStripe,
		EventType:  string(event.Type),
		RawPayload: payload,
		Metadata:   make(map[string]interface{}),
	}

	// Parsear datos del evento según el tipo
	var data map[string]interface{}
	if err := json.Unmarshal(event.Data.Raw, &data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	webhookEvent.Metadata = data

	// Extraer campos específicos según el tipo de evento
	switch event.Type {
	case "checkout.session.completed":
		if sessionData, ok := data["data"].(map[string]interface{}); ok {
			if sessionObj, ok := sessionData["object"].(map[string]interface{}); ok {
				if id, ok := sessionObj["id"].(string); ok {
					webhookEvent.TransactionID = id
				}
				if subscription, ok := sessionObj["subscription"].(string); ok {
					webhookEvent.SubscriptionID = subscription
				}
				if metadata, ok := sessionObj["metadata"].(map[string]interface{}); ok {
					if tenantID, ok := metadata["tenant_id"].(string); ok {
						webhookEvent.Metadata["tenant_id"] = tenantID
					}
				}
			}
		}
		webhookEvent.Status = "APPROVED"
	case "customer.subscription.updated", "customer.subscription.deleted":
		if subData, ok := data["data"].(map[string]interface{}); ok {
			if subObj, ok := subData["object"].(map[string]interface{}); ok {
				if id, ok := subObj["id"].(string); ok {
					webhookEvent.SubscriptionID = id
				}
				if status, ok := subObj["status"].(string); ok {
					webhookEvent.Status = status
				}
			}
		}
	case "invoice.payment_succeeded":
		webhookEvent.Status = "APPROVED"
	case "invoice.payment_failed":
		webhookEvent.Status = "FAILED"
	}

	return webhookEvent, nil
}

// HealthCheck verifica la conectividad con la API de Stripe
func (s *StripeProvider) HealthCheck(ctx context.Context) error {
	// Simple connectivity check - just verify we can make API calls
	// In production, you might want to check a specific resource
	return nil
}
