package owners

import (
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

// RegisterMobileRoutes registers mobile (owner-facing) routes under /mobile/owners
func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	me := mobile.Group("/owners/me")
	me.GET("", handler.GetMe)
	me.PATCH("", handler.UpdateMe)
	me.POST("/push-tokens", handler.AddPushToken)
	me.DELETE("/push-tokens/:token", handler.RemovePushToken)
}

// RegisterAdminRoutes registers admin-panel routes under /api/owners (JWT + RBAC)
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	owners := private.Group("/owners")
	owners.GET("", handler.FindAll)
	owners.GET("/:id", handler.FindByID)
	owners.DELETE("/:id", handler.Delete)
}
