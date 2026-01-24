package permissions

import (
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB) {
	repo := NewPermissionRepository(db)
	service := NewPermissionService(repo)
	handler := NewHandler(service)

	permissions := r.Group("/permissions")

	permissions.GET("/options", handler.GetOptions)
	permissions.POST("", handler.Create)
	permissions.GET("", handler.FindAll)
	permissions.GET("/:id", handler.FindByID)
	permissions.DELETE("/:id", handler.Delete)
}
