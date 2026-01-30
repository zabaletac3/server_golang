package app

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/auth"
	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/modules/webhooks"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func registerRoutes(engine *gin.Engine, db *database.MongoDB, cfg *config.Config, paymentManager *payment.PaymentManager) {
	r := httpx.NewRouter(engine)

	// Public routes (no auth required)
	public := r.Group("/api")

	// Protected routes (auth required)
	private := r.Group("/api")
	// private.Use(sharedAuth.JWTMiddleware(cfg))

	if db != nil {
		// Auth module (public + private)
		auth.RegisterRoutes(public, private, db, cfg)

		// Users module (protected)
		users.RegisterRoutes(private, db)

		// Tenant module (protected) - incluye /tenants/:id/payments
		tenant.RegisterRoutes(private, db, paymentManager)

		// Plans module (protected)
		plans.RegisterRoutes(private, db)

		// Payments module (protected) - solo CRUD de pagos
		payments.RegisterRoutes(private, db)

		// Webhooks module (public) - procesa webhooks de providers
		webhooks.RegisterRoutes(public, db, paymentManager)
	}
}
