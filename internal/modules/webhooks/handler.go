package webhooks

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/platform/logger"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/platform/webhook"
)

type WebhookHandler struct {
	paymentManager *payment.PaymentManager
	paymentService *payments.PaymentService
	tenantRepo     tenant.TenantRepository
	planRepo       plans.PlanRepository
	validator      *webhook.SignatureValidator
}

func NewWebhookHandler(
	paymentManager *payment.PaymentManager,
	paymentService *payments.PaymentService,
	tenantRepo tenant.TenantRepository,
	planRepo plans.PlanRepository,
	validator *webhook.SignatureValidator,
) *WebhookHandler {
	return &WebhookHandler{
		paymentManager: paymentManager,
		paymentService: paymentService,
		tenantRepo:     tenantRepo,
		planRepo:       planRepo,
		validator:      validator,
	}
}

// ProcessWebhook godoc
// @Summary      Procesar webhook de pago
// @Description  Endpoint para recibir webhooks de proveedores de pago
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        provider path string true "Provider name" Enums(wompi, stripe)
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /api/webhooks/{provider} [post]
func (h *WebhookHandler) ProcessWebhook(c *gin.Context) (any, error) {
	providerName := c.Param("provider")

	// Leer el body raw para validación de firma
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Default().Error(c.Request.Context(), "webhook_read_error", "error", err)
		return nil, err
	}

	// Obtener firma del header (puede variar según provider)
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Wompi-Signature")
	}

	// Validar que la firma no esté vacía
	if signature == "" {
		logger.Default().Warn(c.Request.Context(), "webhook_missing_signature", "provider", providerName)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing signature header"})
		return nil, nil
	}

	// Validar firma según el provider
	if err := h.validateSignature(providerName, signature, string(bodyBytes)); err != nil {
		logger.Default().Error(c.Request.Context(), "webhook_signature_invalid", "error", err, "provider", providerName)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return nil, nil
	}

	// Convertir provider name a tipo
	var providerType payment.ProviderType
	switch providerName {
	case "wompi":
		providerType = payment.ProviderWompi
	case "stripe":
		providerType = payment.ProviderStripe
	default:
		logger.Default().Error(c.Request.Context(), "unknown_provider", "provider", providerName)
		return gin.H{"error": "unknown provider"}, nil
	}

	// Procesar webhook usando el PaymentManager
	event, err := h.paymentManager.ProcessWebhook(c.Request.Context(), providerType, bodyBytes, signature)
	if err != nil {
		logger.Default().Error(c.Request.Context(), "webhook_processing_error", "error", err, "provider", providerName)
		return nil, err
	}

	logger.Default().Info(c.Request.Context(), "webhook_received",
		"provider", providerName,
		"event_type", event.EventType,
		"transaction_id", event.TransactionID,
		"status", event.Status,
	)

	// Procesar el evento según el tipo
	ctx := context.Background()
	if err := h.handleWebhookEvent(ctx, event); err != nil {
		logger.Default().Error(ctx, "webhook_handler_error", "error", err)
		// No retornamos error al provider para evitar reintentos
	}

	return gin.H{
		"status": "received",
		"event":  event.EventType,
	}, nil
}

// validateSignature validates the webhook signature based on provider
func (h *WebhookHandler) validateSignature(provider, signature, payload string) error {
	if h.validator == nil {
		return nil // Skip validation if validator not configured
	}

	switch strings.ToLower(provider) {
	case "stripe":
		return h.validator.ValidateStripe(signature, payload)
	default:
		// Wompi and other providers use standard HMAC-SHA256
		return h.validator.Validate(provider, signature, payload)
	}
}

func (h *WebhookHandler) handleWebhookEvent(ctx context.Context, event *payment.WebhookEvent) error {
	switch event.EventType {
	case "payment.succeeded", "transaction.updated":
		if event.Status == "APPROVED" {
			return h.handlePaymentSucceeded(ctx, event)
		}
		if event.Status == "DECLINED" || event.Status == "ERROR" || event.Status == "VOIDED" {
			return h.handlePaymentFailed(ctx, event)
		}
	case "payment.failed":
		return h.handlePaymentFailed(ctx, event)
	case "subscription.canceled":
		return h.handleSubscriptionCanceled(ctx, event)
	default:
		logger.Default().Info(ctx, "unhandled_webhook_event", "event_type", event.EventType)
	}
	return nil
}

func (h *WebhookHandler) handlePaymentSucceeded(ctx context.Context, event *payment.WebhookEvent) error {
	// Buscar el payment pendiente por external_transaction_id (payment link ID)
	tenantObj, err := h.findTenantBySubscriptionID(ctx, event.SubscriptionID)
	if err != nil {
		logger.Default().Error(ctx, "tenant_not_found_for_webhook", "subscription_id", event.SubscriptionID)
		return err
	}

	// Actualizar el pago pendiente
	now := time.Now()

	// Actualizar estado del tenant
	tenantObj.Subscription.BillingStatus = "active"
	tenantObj.Status = tenant.Active
	tenantObj.UpdatedAt = now

	// Actualizar límites según el plan
	if !tenantObj.Subscription.PlanID.IsZero() {
		plan, err := h.planRepo.FindByID(ctx, tenantObj.Subscription.PlanID.Hex())
		if err == nil {
			tenantObj.Usage.UsersLimit = plan.MaxUsers
			tenantObj.Usage.StorageLimitMB = plan.StorageLimitGB * 1024
		}
	}

	if err := h.tenantRepo.Update(ctx, tenantObj); err != nil {
		logger.Default().Error(ctx, "failed_to_update_tenant", "error", err)
		return err
	}

	logger.Default().Info(ctx, "payment_processed", "tenant_id", tenantObj.ID.Hex(), "amount", event.Amount)
	return nil
}

func (h *WebhookHandler) handlePaymentFailed(ctx context.Context, event *payment.WebhookEvent) error {
	tenantObj, err := h.findTenantBySubscriptionID(ctx, event.SubscriptionID)
	if err != nil {
		logger.Default().Error(ctx, "tenant_not_found_for_webhook", "subscription_id", event.SubscriptionID)
		return nil
	}

	// Registrar pago fallido
	paymentDTO := &payments.CreatePaymentDTO{
		TenantID:              tenantObj.ID.Hex(),
		Amount:                event.Amount,
		Currency:              event.Currency,
		PaymentMethod:         string(event.Provider),
		Status:                payments.PaymentFailed,
		ExternalTransactionID: event.TransactionID,
		Concept:               "Intento de pago fallido",
		FailureReason:         event.Status,
		Metadata:              event.Metadata,
	}

	if _, err := h.paymentService.Create(ctx, paymentDTO); err != nil {
		logger.Default().Error(ctx, "failed_to_create_payment", "error", err)
		return err
	}

	logger.Default().Warn(ctx, "payment_failed", "tenant_id", tenantObj.ID.Hex(), "reason", event.Status)
	return nil
}

func (h *WebhookHandler) handleSubscriptionCanceled(ctx context.Context, event *payment.WebhookEvent) error {
	tenantObj, err := h.findTenantBySubscriptionID(ctx, event.SubscriptionID)
	if err != nil {
		return err
	}

	tenantObj.Subscription.BillingStatus = "canceled"
	tenantObj.Status = tenant.Suspended
	tenantObj.UpdatedAt = time.Now()

	if err := h.tenantRepo.Update(ctx, tenantObj); err != nil {
		logger.Default().Error(ctx, "failed_to_update_tenant", "error", err)
		return err
	}

	logger.Default().Info(ctx, "subscription_canceled", "tenant_id", tenantObj.ID.Hex())
	return nil
}

// findTenantBySubscriptionID busca el tenant que tiene el external_subscription_id del payment link
func (h *WebhookHandler) findTenantBySubscriptionID(ctx context.Context, subscriptionID string) (*tenant.Tenant, error) {
	return h.tenantRepo.FindByExternalSubscriptionID(ctx, subscriptionID)
}
