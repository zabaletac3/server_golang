package vaccinations

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

// RegisterAdminRoutes registers admin-panel routes under /api/vaccinations
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB) {
	repo := NewVaccinationRepository(db)
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
		log.Printf("failed to ensure indexes for vaccinations: %v", err)
	}

	service := NewService(repo, patientRepo, userRepo, notifSvc)
	handler := NewHandler(service)

	// Vaccinations routes
	vaccinations := private.Group("/vaccinations")
	vaccinations.POST("", handler.CreateVaccination)
	vaccinations.GET("", handler.ListVaccinations)
	vaccinations.GET("/:id", handler.GetVaccination)
	vaccinations.PUT("/:id", handler.UpdateVaccination)
	vaccinations.PATCH("/:id/status", handler.UpdateVaccinationStatus)
	vaccinations.DELETE("/:id", handler.DeleteVaccination)
	vaccinations.GET("/patient/:patient_id", handler.GetPatientVaccinations)
	vaccinations.GET("/due", handler.GetDueVaccinations)
	vaccinations.GET("/overdue", handler.GetOverdueVaccinations)

	// Vaccine catalog routes
	vaccines := private.Group("/vaccines")
	vaccines.POST("", handler.CreateVaccine)
	vaccines.GET("", handler.ListVaccines)
	vaccines.GET("/:id", handler.GetVaccine)
	vaccines.PUT("/:id", handler.UpdateVaccine)
	vaccines.DELETE("/:id", handler.DeleteVaccine)
}

// RegisterMobileRoutes registers mobile (owner-facing) routes
func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB) {
	repo := NewVaccinationRepository(db)
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
	m := mobile.Group("/vaccinations")
	m.GET("/patient/:patient_id", handler.GetPatientVaccinations)
}
