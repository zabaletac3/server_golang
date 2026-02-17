package webhooks

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/platform/logger"
	"github.com/eren_dev/go_server/internal/platform/payment"
)

type WebhookHandler struct {
	paymentManager  *payment.PaymentManager
	paymentService  *payments.PaymentService
	tenantRepo      tenant.TenantRepository
}

func NewWebhookHandler(
	paymentManager *payment.PaymentManager,
	paymentService *payments.PaymentService,
	tenantRepo tenant.TenantRepository,
) *WebhookHandler {
	return &WebhookHandler{
		paymentManager: paymentManager,
		paymentService: paymentService,
		tenantRepo:     tenantRepo,
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
	if err := h.handleWebhookEvent(c.Request.Context(), event); err != nil {
		logger.Default().Error(c.Request.Context(), "webhook_handler_error", "error", err)
		// No retornamos error al provider para evitar reintentos
	}
	
	return gin.H{
		"status": "received",
		"event":  event.EventType,
	}, nil
}

func (h *WebhookHandler) handleWebhookEvent(ctx any, event *payment.WebhookEvent) error {
	switch event.EventType {
	case "payment.succeeded", "transaction.updated":
		return h.handlePaymentSucceeded(event)
	case "payment.failed":
		return h.handlePaymentFailed(event)
	case "subscription.canceled":
		return h.handleSubscriptionCanceled(event)
	default:
		// ctx es any, pero logger espera context.Context, usamos nil
		logger.Default().Info(nil, "unhandled_webhook_event", "event_type", event.EventType)
	}
	return nil
}

func (h *WebhookHandler) handlePaymentSucceeded(event *payment.WebhookEvent) error {
	// Extraer tenant_id del metadata
	tenantIDStr, ok := event.Metadata["tenant_id"].(string)
	if !ok {
		logger.Default().Error(nil, "missing_tenant_id_in_webhook")
		return nil
	}
	
	tenantID, err := primitive.ObjectIDFromHex(tenantIDStr)
	if err != nil {
		return err
	}
	
	// Crear registro de pago
	now := time.Now()
	paymentDTO := &payments.CreatePaymentDTO{
		TenantID:              tenantIDStr,
		Amount:                event.Amount,
		Currency:              event.Currency,
		PaymentMethod:         string(event.Provider),
		Status:                payments.PaymentCompleted,
		ExternalTransactionID: event.TransactionID,
		Concept:               "Pago de suscripción",
		ProcessedAt:           &now,
		Metadata:              event.Metadata,
	}
	
	if _, err := h.paymentService.Create(nil, paymentDTO); err != nil {
		logger.Default().Error(nil, "failed_to_create_payment", "error", err)
		return err
	}
	
	// Actualizar estado del tenant
	tenantObj, err := h.tenantRepo.FindByID(nil, tenantID.Hex())
	if err != nil {
		return err
	}
	
	tenantObj.Subscription.BillingStatus = "active"
	tenantObj.Status = tenant.Active
	
	if err := h.tenantRepo.Update(nil, tenantObj); err != nil {
		logger.Default().Error(nil, "failed_to_update_tenant", "error", err)
		return err
	}
	
	logger.Default().Info(nil, "payment_processed", "tenant_id", tenantIDStr, "amount", event.Amount)
	return nil
}

func (h *WebhookHandler) handlePaymentFailed(event *payment.WebhookEvent) error {
	tenantIDStr, ok := event.Metadata["tenant_id"].(string)
	if !ok {
		return nil
	}
	
	// Registrar pago fallido
	paymentDTO := &payments.CreatePaymentDTO{
		TenantID:              tenantIDStr,
		Amount:                event.Amount,
		Currency:              event.Currency,
		PaymentMethod:         string(event.Provider),
		Status:                payments.PaymentFailed,
		ExternalTransactionID: event.TransactionID,
		Concept:               "Intento de pago fallido",
		FailureReason:         event.Status,
		Metadata:              event.Metadata,
	}
	
	if _, err := h.paymentService.Create(nil, paymentDTO); err != nil {
		logger.Default().Error(nil, "failed_to_create_payment", "error", err)
		return err
	}
	
	logger.Default().Warn(nil, "payment_failed", "tenant_id", tenantIDStr, "reason", event.Status)
	return nil
}

func (h *WebhookHandler) handleSubscriptionCanceled(event *payment.WebhookEvent) error {
	tenantIDStr, ok := event.Metadata["tenant_id"].(string)
	if !ok {
		return nil
	}
	
	tenantID, err := primitive.ObjectIDFromHex(tenantIDStr)
	if err != nil {
		return err
	}
	
	// Actualizar estado del tenant
	tenantObj, err := h.tenantRepo.FindByID(nil, tenantID.Hex())
	if err != nil {
		return err
	}
	
	tenantObj.Subscription.BillingStatus = "canceled"
	tenantObj.Status = tenant.Suspended
	
	if err := h.tenantRepo.Update(nil, tenantObj); err != nil {
		logger.Default().Error(nil, "failed_to_update_tenant", "error", err)
		return err
	}
	
	logger.Default().Info(nil, "subscription_canceled", "tenant_id", tenantIDStr)
	return nil
}
