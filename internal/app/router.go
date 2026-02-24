package app

import (
	"os"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/audit"
	"github.com/eren_dev/go_server/internal/modules/appointments"
	"github.com/eren_dev/go_server/internal/modules/auth"
	"github.com/eren_dev/go_server/internal/modules/medical_records"
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
	"github.com/eren_dev/go_server/internal/platform/cache"
	platformNotifications "github.com/eren_dev/go_server/internal/platform/notifications"
	"github.com/eren_dev/go_server/internal/platform/payment"
	"github.com/eren_dev/go_server/internal/platform/ratelimit"
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
)

// initializeCache creates a Redis cache client if REDIS_ADDR is configured
func initializeCache() cache.Cache {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil // Cache disabled
	}

	password := os.Getenv("REDIS_PASSWORD")
	db := 0

	c, err := cache.NewRedisCache(cache.Config{
		Addr:     addr,
		Password: password,
		DB:       db,
		Prefix:   "vetsify",
	})
	if err != nil {
		// Log error but continue without cache
		return nil
	}

	return c
}

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
		// Initialize optional Redis cache (disabled if REDIS_ADDR not configured)
		redisCache := initializeCache()

		// Initialize hierarchical rate limiter
		rateLimiterCfg := ratelimit.DefaultConfig()
		rateLimiter := ratelimit.NewLimiter(rateLimiterCfg)

		// Initialize audit service
		auditRepo := audit.NewRepository(db)
		auditService := audit.NewService(auditRepo)

		// RBAC middleware aplicado a rutas de staff
		rbacMiddleware := sharedMiddleware.RBACMiddleware(sharedMiddleware.RBACConfig{
			UserRepo:       users.NewRepository(db),
			RoleRepo:       roles.NewRepository(db),
			PermissionRepo: permissions.NewRepository(db),
			ResourceRepo:   resources.NewRepository(db),
			Cache:          redisCache,
		})
		private.Use(rbacMiddleware)
		privateTenant.Use(rbacMiddleware)

		// Apply tenant rate limiting to tenant-scoped routes
		privateTenant.Use(sharedMiddleware.TenantRateLimitMiddleware(rateLimiter))
		mobileTenant.Use(sharedMiddleware.TenantRateLimitMiddleware(rateLimiter))

		// Staff auth: rutas públicas + /auth/me sin RBAC
		auth.RegisterRoutes(public, authPrivate, db, cfg)

		// Users module (JWT + RBAC)
		users.RegisterRoutes(private, db)

		// Owners admin routes (JWT + RBAC)
		owners.RegisterAdminRoutes(private, db)

		// Tenant module (JWT + RBAC)
		tenant.RegisterRoutes(private, db, paymentManager, cfg, auditService)

		// Plans module (JWT + RBAC)
		plans.RegisterRoutes(private, db)

		// Payments module (JWT + RBAC)
		payments.RegisterRoutes(private, db)

		// Webhooks module (público)
		webhooks.RegisterRoutes(public, db, paymentManager, cfg)

		// RBAC modules (JWT + RBAC)
		resources.RegisterRoutes(private, db)
		permissions.RegisterRoutes(private, db)
		roles.RegisterRoutes(private, db)

		// Patients + Species (JWT + Tenant + RBAC)
		patients.RegisterAdminRoutes(privateTenant, db)

		// Appointments (JWT + Tenant + RBAC)
		appointments.RegisterAdminRoutes(privateTenant, db, pushProvider, cfg)

		// Medical Records (JWT + Tenant + RBAC)
		medical_records.RegisterAdminRoutes(privateTenant, db)

		// Mobile auth routes (public + owner-private)
		mobileAuth.RegisterRoutes(mobilePublic, mobilePrivate, db, cfg)

		// Mobile owner profile routes (owner-private)
		owners.RegisterMobileRoutes(mobilePrivate, db)

		// Mobile patients (owner-private + tenant)
		patients.RegisterMobileRoutes(mobileTenant, db)

		// Mobile appointments (owner-private + tenant)
		appointments.RegisterMobileRoutes(mobileTenant, db, pushProvider, cfg)

		// Mobile medical records (owner-private + tenant, read-only)
		medical_records.RegisterMobileRoutes(mobileTenant, db)

		// Mobile notifications (owner-private)
		notifications.RegisterMobileRoutes(mobilePrivate, db, pushProvider)

		// Admin notifications (JWT only, no RBAC — any staff member can read their own)
		notifications.RegisterAdminRoutes(authPrivate, db, pushProvider)
	}
}
