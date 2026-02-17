package mobile_auth

import (
	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/owners"
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

// RegisterRoutes mounts mobile auth routes.
//   - mobilePublic: /mobile/auth/register, /mobile/auth/login, /mobile/auth/refresh
//   - mobilePrivate: /mobile/auth/me (JWT + owner guard already applied)
func RegisterRoutes(mobilePublic, mobilePrivate *httpx.Router, db *database.MongoDB, cfg *config.Config) {
	ownerRepo := owners.NewRepository(db)
	jwtService := sharedAuth.NewJWTService(cfg)
	service := NewService(ownerRepo, jwtService)
	handler := NewHandler(service)

	pub := mobilePublic.Group("/auth")
	pub.POST("/register", handler.Register)
	pub.POST("/login", handler.Login)
	pub.POST("/refresh", handler.Refresh)

	priv := mobilePrivate.Group("/auth")
	priv.GET("/me", handler.Me)
}
