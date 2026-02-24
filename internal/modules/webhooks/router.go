package webhooks

import (
	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/platform/webhook"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB, paymentManager *payment.PaymentManager, cfg *config.Config) {
	// Inicializar dependencias
	paymentRepo := payments.NewPaymentRepository(db)
	paymentService := payments.NewPaymentService(paymentRepo)
	tenantRepo := tenant.NewTenantRepository(db)
	planRepo := plans.NewPlanRepository(db)

	// Crear validador de firmas y registrar secretos
	validator := webhook.NewSignatureValidator()
	if cfg.WompiWebhookSecret != "" {
		validator.RegisterSecret("wompi", cfg.WompiWebhookSecret)
	}
	if cfg.StripeWebhookSecret != "" {
		validator.RegisterSecret("stripe", cfg.StripeWebhookSecret)
	}

	handler := NewWebhookHandler(paymentManager, paymentService, tenantRepo, planRepo, validator)

	// Rutas públicas de webhooks (sin autenticación)
	webhooks := r.Group("/webhooks")
	webhooks.POST("/:provider", handler.ProcessWebhook)
}
