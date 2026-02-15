package roles

import (
	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB) {
	repo := NewRepository(db)
	permissionRepo := permissions.NewRepository(db)
	resourceRepo := resources.NewRepository(db)
	service := NewService(repo, permissionRepo, resourceRepo)
	handler := NewHandler(service)

	g := r.Group("/roles")

	g.POST("", handler.Create)
	g.GET("", handler.FindAll)
	g.GET("/:id", handler.FindByID)
	g.PATCH("/:id", handler.Update)
	g.DELETE("/:id", handler.Delete)
}
