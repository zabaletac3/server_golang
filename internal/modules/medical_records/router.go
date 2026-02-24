package medical_records

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

// RegisterAdminRoutes registers admin-panel routes under /api/medical-records
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB) {
	repo := NewMedicalRecordRepository(db)
	patientRepo := patients.NewPatientRepository(db)
	userRepo := users.NewRepository(db)
	notifSvc := notifications.NewService(
		notifications.NewRepository(db),
		notifications.NewStaffRepository(db),
		owners.NewRepository(db),
		nil, // push provider not needed for medical records
	)

	// Ensure indexes
	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("failed to ensure indexes for medical_records: %v", err)
	}

	service := NewService(repo, patientRepo, userRepo, notifSvc)
	handler := NewHandler(service)

	// Medical Records routes
	mr := private.Group("/medical-records")
	mr.POST("", handler.CreateMedicalRecord)
	mr.GET("", handler.ListMedicalRecords)
	mr.GET("/:id", handler.GetMedicalRecord)
	mr.PUT("/:id", handler.UpdateMedicalRecord)
	mr.DELETE("/:id", handler.DeleteMedicalRecord)
	mr.GET("/patient/:patient_id", handler.GetPatientRecords)
	mr.GET("/patient/:patient_id/timeline", handler.GetPatientTimeline)

	// Allergies routes
	allergies := private.Group("/allergies")
	allergies.POST("", handler.CreateAllergy)
	allergies.GET("/patient/:patient_id", handler.GetPatientAllergies)
	allergies.PUT("/:id", handler.UpdateAllergy)
	allergies.DELETE("/:id", handler.DeleteAllergy)

	// Medical History routes
	history := private.Group("/medical-history")
	history.POST("", handler.CreateMedicalHistory)
	history.GET("/patient/:patient_id", handler.GetMedicalHistory)
	history.PUT("/:id", handler.UpdateMedicalHistory)
	history.DELETE("/:id", handler.DeleteMedicalHistory)
}

// RegisterMobileRoutes registers mobile (owner-facing) routes
// Owners can only VIEW records, not create/update/delete
func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB) {
	repo := NewMedicalRecordRepository(db)
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
	m := mobile.Group("/medical-records")
	m.GET("/patient/:patient_id", handler.GetPatientRecords)
	m.GET("/patient/:patient_id/timeline", handler.GetPatientTimeline)

	// Mobile allergies - read only
	m.GET("/allergies/patient/:patient_id", handler.GetPatientAllergies)

	// Mobile medical history - read only
	m.GET("/medical-history/patient/:patient_id", handler.GetMedicalHistory)
}
