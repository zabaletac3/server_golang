package appointments

import (
	"context"
	"log"

	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/modules/patients"
	"github.com/eren_dev/go_server/internal/modules/users"
	platformNotifications "github.com/eren_dev/go_server/internal/platform/notifications"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

// RegisterAdminRoutes registers admin-panel routes under /api/appointments (JWT + RBAC)
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB, pushProvider platformNotifications.PushProvider) {
	repo := NewAppointmentRepository(db)
	patientRepo := patients.NewPatientRepository(db)
	ownerRepo := owners.NewRepository(db)
	userRepo := users.NewRepository(db)
	notifSvc := notifications.NewService(notifications.NewRepository(db), notifications.NewStaffRepository(db), ownerRepo, pushProvider)

	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("failed to ensure indexes for appointments: %v", err)
	}

	service := NewService(repo, patientRepo, ownerRepo, userRepo, notifSvc)
	handler := NewHandler(service)

	p := private.Group("/appointments")
	p.POST("", handler.CreateAppointment)
	p.GET("", handler.ListAppointments)
	p.GET("/calendar", handler.GetCalendarView)
	p.GET("/availability", handler.CheckAvailability)
	p.GET("/:id", handler.GetAppointment)
	p.PUT("/:id", handler.UpdateAppointment)
	p.DELETE("/:id", handler.DeleteAppointment)
	p.PATCH("/:id/status", handler.UpdateStatus)
	p.GET("/:id/history", handler.GetStatusHistory)
}

// RegisterMobileRoutes registers mobile (owner-facing) routes under /mobile/appointments
func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB, pushProvider platformNotifications.PushProvider) {
	repo := NewAppointmentRepository(db)
	patientRepo := patients.NewPatientRepository(db)
	ownerRepo := owners.NewRepository(db)
	userRepo := users.NewRepository(db)
	notifSvc := notifications.NewService(notifications.NewRepository(db), notifications.NewStaffRepository(db), ownerRepo, pushProvider)

	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("failed to ensure indexes for appointments: %v", err)
	}

	service := NewService(repo, patientRepo, ownerRepo, userRepo, notifSvc)
	handler := NewHandler(service)

	m := mobile.Group("/appointments")
	m.POST("/request", handler.RequestAppointment)
	m.GET("", handler.GetOwnerAppointments)
	m.GET("/:id", handler.GetOwnerAppointment)
	m.PATCH("/:id/cancel", handler.CancelOwnerAppointment)
}
