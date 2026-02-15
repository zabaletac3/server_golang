package resources

import (
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	g := r.Group("/resources")

	g.POST("", handler.Create)
	g.GET("", handler.FindAll)
	g.GET("/:id", handler.FindByID)
	g.PATCH("/:id", handler.Update)
	g.DELETE("/:id", handler.Delete)
}
