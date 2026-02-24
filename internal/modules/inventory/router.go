package inventory

import (
	"context"
	"log"

	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

// RegisterAdminRoutes registers admin-panel routes under /api/products
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB) {
	repo := NewProductRepository(db)
	userRepo := users.NewRepository(db)
	notifSvc := notifications.NewService(
		notifications.NewRepository(db),
		notifications.NewStaffRepository(db),
		owners.NewRepository(db),
		nil,
	)

	// Ensure indexes
	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("failed to ensure indexes for inventory: %v", err)
	}

	service := NewService(repo, userRepo, notifSvc)
	handler := NewHandler(service)

	// Products routes
	products := private.Group("/products")
	products.POST("", handler.CreateProduct)
	products.GET("", handler.ListProducts)
	products.GET("/:id", handler.GetProduct)
	products.PUT("/:id", handler.UpdateProduct)
	products.DELETE("/:id", handler.DeleteProduct)
	products.POST("/:id/stock-in", handler.StockIn)
	products.POST("/:id/stock-out", handler.StockOut)
	products.GET("/low-stock", handler.GetLowStockProducts)
	products.GET("/expiring", handler.GetExpiringProducts)
	products.GET("/alerts", handler.GetProductAlerts)

	// Stock movements routes
	movements := private.Group("/stock-movements")
	movements.GET("", handler.GetStockMovements)

	// Categories routes
	categories := private.Group("/categories")
	categories.POST("", handler.CreateCategory)
	categories.GET("", handler.ListCategories)
	categories.GET("/:id", handler.GetCategory)
	categories.PUT("/:id", handler.UpdateCategory)
	categories.DELETE("/:id", handler.DeleteCategory)
}
