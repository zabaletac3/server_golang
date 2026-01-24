package tenant

import (
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB) {
	repo := NewTenantRepository(db)
	userRepo := users.NewRepository(db)
	service := NewTenantService(repo, userRepo)
	handler := NewHandler(service)

	tenants := r.Group("/tenants")

	tenants.POST("", handler.Create)
	tenants.GET("", handler.FindAll)
	tenants.GET("/:id", handler.FindByID)
	tenants.PATCH("/:id", handler.Update)
	tenants.DELETE("/:id", handler.Delete)
}