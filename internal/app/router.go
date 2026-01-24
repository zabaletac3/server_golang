package app

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/auth"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func registerRoutes(engine *gin.Engine, db *database.MongoDB, cfg *config.Config) {
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

		// Tenant module (protected)
		tenant.RegisterRoutes(private, db)
	}
}
