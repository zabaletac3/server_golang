package app

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/auth"
	mobileAuth "github.com/eren_dev/go_server/internal/modules/mobile_auth"
	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/modules/patients"
	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/modules/webhooks"
	platformNotifications "github.com/eren_dev/go_server/internal/platform/notifications"
	"github.com/eren_dev/go_server/internal/platform/payment"
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
)

func registerRoutes(engine *gin.Engine, db *database.MongoDB, cfg *config.Config, paymentManager *payment.PaymentManager, pushProvider platformNotifications.PushProvider) {
	r := httpx.NewRouter(engine)

	// Public routes (sin autenticación)
	public := r.Group("/api")

	// Rutas con JWT pero sin RBAC (accesibles para cualquier usuario autenticado)
	authPrivate := r.Group("/api")
	authPrivate.Use(sharedAuth.JWTMiddleware(cfg))

	// Rutas protegidas con JWT + RBAC
	private := r.Group("/api")
	private.Use(sharedAuth.JWTMiddleware(cfg))

	// Mobile routes: /mobile/auth/... y /mobile/owners/...
	mobilePublic := r.Group("/mobile")
	mobilePrivate := r.Group("/mobile")
	mobilePrivate.Use(sharedAuth.JWTMiddleware(cfg))
	mobilePrivate.Use(sharedMiddleware.OwnerGuardMiddleware())

	// Tenant-scoped staff routes (JWT + X-Tenant-ID + RBAC)
	privateTenant := r.Group("/api")
	privateTenant.Use(sharedAuth.JWTMiddleware(cfg))
	privateTenant.Use(sharedMiddleware.TenantMiddleware())

	// Tenant-scoped mobile routes (JWT + X-Tenant-ID + OwnerGuard)
	mobileTenant := r.Group("/mobile")
	mobileTenant.Use(sharedAuth.JWTMiddleware(cfg))
	mobileTenant.Use(sharedMiddleware.TenantMiddleware())
	mobileTenant.Use(sharedMiddleware.OwnerGuardMiddleware())

	if db != nil {
		// RBAC middleware aplicado a rutas de staff
		rbacMiddleware := sharedMiddleware.RBACMiddleware(sharedMiddleware.RBACConfig{
			UserRepo:       users.NewRepository(db),
			RoleRepo:       roles.NewRepository(db),
			PermissionRepo: permissions.NewRepository(db),
			ResourceRepo:   resources.NewRepository(db),
		})
		private.Use(rbacMiddleware)
		privateTenant.Use(rbacMiddleware)

		// Staff auth: rutas públicas + /auth/me sin RBAC
		auth.RegisterRoutes(public, authPrivate, db, cfg)

		// Users module (JWT + RBAC)
		users.RegisterRoutes(private, db)

		// Owners admin routes (JWT + RBAC)
		owners.RegisterAdminRoutes(private, db)

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

		// Patients + Species (JWT + Tenant + RBAC)
		patients.RegisterAdminRoutes(privateTenant, db)

		// Mobile auth routes (public + owner-private)
		mobileAuth.RegisterRoutes(mobilePublic, mobilePrivate, db, cfg)

		// Mobile owner profile routes (owner-private)
		owners.RegisterMobileRoutes(mobilePrivate, db)

		// Mobile patients (owner-private + tenant)
		patients.RegisterMobileRoutes(mobileTenant, db)

		// Mobile notifications (owner-private)
		notifications.RegisterMobileRoutes(mobilePrivate, db, pushProvider)

		// Admin notifications (JWT only, no RBAC — any staff member can read their own)
		notifications.RegisterAdminRoutes(authPrivate, db, pushProvider)
	}
}
