package patients

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/owners"
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

type Handler struct {
	service        *PatientService
	speciesService *SpeciesService
	ownerRepo      owners.OwnerRepository
}

func NewHandler(service *PatientService, speciesService *SpeciesService, ownerRepo owners.OwnerRepository) *Handler {
	return &Handler{
		service:        service,
		speciesService: speciesService,
		ownerRepo:      ownerRepo,
	}
}

// --- Admin: Patient CRUD ---

// Create creates a new patient.
//
//	@Summary		Create patient
//	@Tags			patients
//	@Accept			json
//	@Produce		json
//	@Param			X-Tenant-ID	header		string				true	"Tenant ID"
//	@Param			body		body		CreatePatientDTO	true	"Patient data"
//	@Success		200			{object}	PatientResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		409			{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/patients [post]
func (h *Handler) Create(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	var dto CreatePatientDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Create(c.Request.Context(), tenantID, &dto)
}

// FindAll lists all patients for a tenant.
//
//	@Summary		List patients
//	@Tags			patients
//	@Produce		json
//	@Param			X-Tenant-ID	header		string	true	"Tenant ID"
//	@Param			skip		query		int		false	"Skip"
//	@Param			limit		query		int		false	"Limit"
//	@Success		200			{object}	PaginatedPatientsResponse
//	@Failure		400			{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/patients [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)
	params := pagination.FromContext(c)
	return h.service.FindAll(c.Request.Context(), tenantID, params)
}

// FindByID returns a patient by ID.
//
//	@Summary		Get patient by ID
//	@Tags			patients
//	@Produce		json
//	@Param			X-Tenant-ID	header		string	true	"Tenant ID"
//	@Param			id			path		string	true	"Patient ID"
//	@Success		200			{object}	PatientResponse
//	@Failure		404			{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/patients/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)
	return h.service.FindByID(c.Request.Context(), tenantID, c.Param("id"))
}

// Update updates a patient.
//
//	@Summary		Update patient
//	@Tags			patients
//	@Accept			json
//	@Produce		json
//	@Param			X-Tenant-ID	header		string				true	"Tenant ID"
//	@Param			id			path		string				true	"Patient ID"
//	@Param			body		body		UpdatePatientDTO	true	"Update data"
//	@Success		200			{object}	PatientResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/patients/{id} [patch]
func (h *Handler) Update(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	var dto UpdatePatientDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Update(c.Request.Context(), tenantID, c.Param("id"), &dto)
}

// Delete soft-deletes a patient.
//
//	@Summary		Delete patient
//	@Tags			patients
//	@Produce		json
//	@Param			X-Tenant-ID	header		string	true	"Tenant ID"
//	@Param			id			path		string	true	"Patient ID"
//	@Success		200			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/patients/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	if err := h.service.Delete(c.Request.Context(), tenantID, c.Param("id")); err != nil {
		return nil, err
	}
	return gin.H{"message": "patient deleted"}, nil
}

// --- Admin: Species ---

// ListSpecies lists all species for a tenant.
//
//	@Summary		List species
//	@Tags			species
//	@Produce		json
//	@Param			X-Tenant-ID	header		string	true	"Tenant ID"
//	@Success		200			{array}		SpeciesResponse
//	@Failure		400			{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/species [get]
func (h *Handler) ListSpecies(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)
	return h.speciesService.ListByTenant(c.Request.Context(), tenantID)
}

// ResolveSpecies resolves or creates a species with trigram dedup.
//
//	@Summary		Resolve or create species
//	@Description	Finds an exact/similar match or creates a new species. Returns 409 with suggestions if ambiguous.
//	@Tags			species
//	@Accept			json
//	@Produce		json
//	@Param			X-Tenant-ID	header		string				true	"Tenant ID"
//	@Param			body		body		ResolveSpeciesDTO	true	"Species name"
//	@Success		200			{object}	SpeciesResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		409			{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/species [post]
func (h *Handler) ResolveSpecies(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	var dto ResolveSpeciesDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	result, conflict, err := h.speciesService.Resolve(c.Request.Context(), tenantID, dto.Name)
	if err != nil {
		if errors.Is(err, ErrSpeciesConflict) {
			c.JSON(http.StatusConflict, conflict)
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// --- Mobile: Owner's patients ---

// MobileFindAll returns the owner's patients.
//
//	@Summary		List my patients
//	@Tags			mobile/patients
//	@Produce		json
//	@Param			X-Tenant-ID	header		string	true	"Tenant ID"
//	@Param			skip		query		int		false	"Skip"
//	@Param			limit		query		int		false	"Limit"
//	@Success		200			{object}	PaginatedPatientsResponse
//	@Failure		401			{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/patients [get]
func (h *Handler) MobileFindAll(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}

	tenantID := sharedMiddleware.GetTenantID(c)
	ownerOID, _ := primitive.ObjectIDFromHex(ownerID)
	params := pagination.FromContext(c)
	return h.service.FindByOwner(c.Request.Context(), tenantID, ownerOID, params)
}

// MobileFindByID returns a single patient detail for the owner.
//
//	@Summary		Get my patient by ID
//	@Tags			mobile/patients
//	@Produce		json
//	@Param			X-Tenant-ID	header		string	true	"Tenant ID"
//	@Param			id			path		string	true	"Patient ID"
//	@Success		200			{object}	PatientResponse
//	@Failure		401			{object}	map[string]string
//	@Failure		403			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/patients/{id} [get]
func (h *Handler) MobileFindByID(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	resp, err := h.service.FindByID(c.Request.Context(), tenantID, c.Param("id"))
	if err != nil {
		return nil, err
	}

	// Ensure the patient belongs to this owner
	if resp.OwnerID != ownerID {
		return nil, sharedErrors.ErrForbidden
	}

	return resp, nil
}
