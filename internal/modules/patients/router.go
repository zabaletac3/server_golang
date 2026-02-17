package patients

import (
	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func newDeps(db *database.MongoDB) (*PatientService, *SpeciesService, owners.OwnerRepository) {
	patientRepo := NewPatientRepository(db)
	speciesRepo := NewSpeciesRepository(db)
	ownerRepo := owners.NewRepository(db)
	speciesSvc := NewSpeciesService(speciesRepo)
	patientSvc := NewService(patientRepo, speciesSvc)
	return patientSvc, speciesSvc, ownerRepo
}

// RegisterAdminRoutes registers admin-panel routes (JWT + RBAC).
func RegisterAdminRoutes(private *httpx.Router, db *database.MongoDB) {
	patientSvc, speciesSvc, ownerRepo := newDeps(db)
	h := NewHandler(patientSvc, speciesSvc, ownerRepo)

	p := private.Group("/patients")
	p.POST("", h.Create)
	p.GET("", h.FindAll)
	p.GET("/:id", h.FindByID)
	p.PATCH("/:id", h.Update)
	p.DELETE("/:id", h.Delete)

	s := private.Group("/species")
	s.GET("", h.ListSpecies)
	s.POST("", h.ResolveSpecies)
}

// RegisterMobileRoutes registers mobile routes (JWT + OwnerGuard).
func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB) {
	patientSvc, speciesSvc, ownerRepo := newDeps(db)
	h := NewHandler(patientSvc, speciesSvc, ownerRepo)

	mp := mobile.Group("/patients")
	mp.GET("", h.MobileFindAll)
	mp.GET("/:id", h.MobileFindByID)
}
