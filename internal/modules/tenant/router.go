package tenant

import (
	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB, paymentManager *payment.PaymentManager) {
	repo := NewTenantRepository(db)
	userRepo := users.NewRepository(db)
	service := NewTenantService(repo, userRepo, paymentManager)
	handler := NewHandler(service)

	// Payment service para historial
	paymentRepo := payments.NewPaymentRepository(db)
	paymentService := payments.NewPaymentService(paymentRepo)
	paymentHandler := payments.NewPaymentHandler(paymentService)

	tenants := r.Group("/tenants")

	tenants.POST("", handler.Create)
	tenants.GET("", handler.FindAll)
	tenants.GET("/:id", handler.FindByID)
	tenants.PATCH("/:id", handler.Update)
	tenants.DELETE("/:id", handler.Delete)
	
	// Historial de pagos del tenant
	tenants.GET("/:id/payments", paymentHandler.FindByTenantID)
}