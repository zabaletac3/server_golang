package payments

import (
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB) {
	repo := NewPaymentRepository(db)
	service := NewPaymentService(repo)
	handler := NewPaymentHandler(service)

	// Rutas de pagos (protegidas)
	payments := r.Group("/payments")
	payments.POST("", handler.Create)
	payments.GET("/:id", handler.FindByID)
	
	// Nota: El historial por tenant está en /tenants/:tenant_id/payments
	// Nota: Los webhooks están en el módulo webhooks en /webhooks/:provider
}
