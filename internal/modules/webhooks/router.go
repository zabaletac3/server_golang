package webhooks

import (
	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB, paymentManager *payment.PaymentManager) {
	// Inicializar dependencias
	paymentRepo := payments.NewPaymentRepository(db)
	paymentService := payments.NewPaymentService(paymentRepo)
	tenantRepo := tenant.NewTenantRepository(db)
	
	handler := NewWebhookHandler(paymentManager, paymentService, tenantRepo)

	// Rutas públicas de webhooks (sin autenticación)
	webhooks := r.Group("/webhooks")
	webhooks.POST("/:provider", handler.ProcessWebhook)
}
