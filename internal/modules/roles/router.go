package roles

import (
	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB) {
	repo := NewRoleRepository(db)
	permRepo := permissions.NewPermissionRepository(db)
	service := NewRoleService(repo, permRepo)
	handler := NewHandler(service)

	roles := r.Group("/roles")

	roles.POST("", handler.Create)
	roles.GET("", handler.FindByTenantID)
	roles.GET("/:id", handler.FindByID)
	roles.PATCH("/:id", handler.Update)
	roles.DELETE("/:id", handler.Delete)
}
