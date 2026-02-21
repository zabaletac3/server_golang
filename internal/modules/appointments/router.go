package appointments

import (
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

// RegisterAdminRoutes registers admin-panel routes under /api/appointments (JWT + RBAC)
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB) {
	repo := NewAppointmentRepository(db)
	service := NewService(repo)
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
func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB) {
	repo := NewAppointmentRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	m := mobile.Group("/appointments")
	m.POST("/request", handler.RequestAppointment)
	m.GET("", handler.GetOwnerAppointments)
	m.GET("/:id", handler.GetOwnerAppointment)
}
