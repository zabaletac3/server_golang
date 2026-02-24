package laboratory

import (
	"context"
	"log"

	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/modules/patients"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

// RegisterAdminRoutes registers admin-panel routes under /api/lab-orders
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB) {
	repo := NewLabOrderRepository(db)
	patientRepo := patients.NewPatientRepository(db)
	userRepo := users.NewRepository(db)
	notifSvc := notifications.NewService(
		notifications.NewRepository(db),
		notifications.NewStaffRepository(db),
		owners.NewRepository(db),
		nil,
	)

	// Ensure indexes
	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("failed to ensure indexes for laboratory: %v", err)
	}

	service := NewService(repo, patientRepo, userRepo, notifSvc)
	handler := NewHandler(service)

	// Lab Orders routes
	orders := private.Group("/lab-orders")
	orders.POST("", handler.CreateLabOrder)
	orders.GET("", handler.ListLabOrders)
	orders.GET("/:id", handler.GetLabOrder)
	orders.PUT("/:id", handler.UpdateLabOrder)
	orders.PATCH("/:id/status", handler.UpdateLabOrderStatus)
	orders.POST("/:id/result", handler.UploadLabResult)
	orders.DELETE("/:id", handler.DeleteLabOrder)
	orders.GET("/patient/:patient_id", handler.GetPatientLabOrders)
	orders.GET("/overdue", handler.GetOverdueLabOrders)

	// Lab Tests catalog routes
	tests := private.Group("/lab-tests")
	tests.POST("", handler.CreateLabTest)
	tests.GET("", handler.ListLabTests)
	tests.GET("/:id", handler.GetLabTest)
	tests.PUT("/:id", handler.UpdateLabTest)
	tests.DELETE("/:id", handler.DeleteLabTest)
}

// RegisterMobileRoutes registers mobile (owner-facing) routes
func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB) {
	repo := NewLabOrderRepository(db)
	patientRepo := patients.NewPatientRepository(db)
	userRepo := users.NewRepository(db)
	notifSvc := notifications.NewService(
		notifications.NewRepository(db),
		notifications.NewStaffRepository(db),
		owners.NewRepository(db),
		nil,
	)

	service := NewService(repo, patientRepo, userRepo, notifSvc)
	handler := NewHandler(service)

	// Mobile routes - read only for owners
	m := mobile.Group("/lab-orders")
	m.GET("/patient/:patient_id", handler.GetPatientLabOrders)
}
