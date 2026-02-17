package app

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/auth"
	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/modules/webhooks"
	"github.com/eren_dev/go_server/internal/platform/payment"
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
)

func registerRoutes(engine *gin.Engine, db *database.MongoDB, cfg *config.Config, paymentManager *payment.PaymentManager) {
	r := httpx.NewRouter(engine)

	// Public routes (sin autenticación)
	public := r.Group("/api")

	// Rutas con JWT pero sin RBAC (accesibles para cualquier usuario autenticado)
	authPrivate := r.Group("/api")
	authPrivate.Use(sharedAuth.JWTMiddleware(cfg))

	// Rutas protegidas con JWT + RBAC
	private := r.Group("/api")
	private.Use(sharedAuth.JWTMiddleware(cfg))

	if db != nil {
		// RBAC middleware aplicado solo a rutas de recursos
		private.Use(sharedMiddleware.RBACMiddleware(sharedMiddleware.RBACConfig{
			UserRepo:       users.NewRepository(db),
			RoleRepo:       roles.NewRepository(db),
			PermissionRepo: permissions.NewRepository(db),
			ResourceRepo:   resources.NewRepository(db),
		}))

		// Auth module: rutas públicas + /auth/me sin RBAC
		auth.RegisterRoutes(public, authPrivate, db, cfg)

		// Users module (JWT + RBAC)
		users.RegisterRoutes(private, db)

		// Tenant module (JWT + RBAC)
		tenant.RegisterRoutes(private, db, paymentManager)

		// Plans module (JWT + RBAC)
		plans.RegisterRoutes(private, db)

		// Payments module (JWT + RBAC)
		payments.RegisterRoutes(private, db)

		// Webhooks module (público)
		webhooks.RegisterRoutes(public, db, paymentManager)

		// RBAC modules (JWT + RBAC)
		resources.RegisterRoutes(private, db)
		permissions.RegisterRoutes(private, db)
		roles.RegisterRoutes(private, db)
	}
}
