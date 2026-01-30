package plans

import (
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func RegisterRoutes(r *httpx.Router, db *database.MongoDB) {
	repo := NewPlanRepository(db)
	service := NewPlanService(repo)
	handler := NewPlanHandler(service)

	plans := r.Group("/plans")

	plans.POST("", handler.Create)
	plans.GET("", handler.FindAll)
	plans.GET("/visible", handler.FindVisible)
	plans.GET("/:id", handler.FindByID)
	plans.PATCH("/:id", handler.Update)
	plans.DELETE("/:id", handler.Delete)
}
