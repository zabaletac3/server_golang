package vaccinations

import (
	"github.com/gin-gonic/gin"

	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

// Handler handles HTTP requests for vaccinations
type Handler struct {
	service *Service
}

// NewHandler creates a new vaccinations handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ==================== VACCINATIONS ====================

// CreateVaccination creates a new vaccination record
// @Summary Create vaccination
// @Description Register a new vaccination for a patient
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param vaccination body CreateVaccinationDTO true "Vaccination data"
// @Success 201 {object} VaccinationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccinations [post]
func (h *Handler) CreateVaccination(c *gin.Context) (any, error) {
	var dto CreateVaccinationDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	vaccination, err := h.service.CreateVaccination(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccination.ToResponse(), nil
}

// GetVaccination gets a vaccination by ID
// @Summary Get vaccination
// @Description Get vaccination details by ID
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "Vaccination ID"
// @Success 200 {object} VaccinationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccinations/{id} [get]
func (h *Handler) GetVaccination(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "vaccination ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	vaccination, err := h.service.GetVaccination(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccination.ToResponse(), nil
}

// ListVaccinations lists vaccinations with filters
// @Summary List vaccinations
// @Description Get a paginated list of vaccinations with optional filters
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param patient_id query string false "Filter by patient ID"
// @Param veterinarian_id query string false "Filter by veterinarian ID"
// @Param status query string false "Filter by status (applied, due, overdue)"
// @Param vaccine_name query string false "Filter by vaccine name"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Param due_soon query bool false "Filter vaccinations due within 30 days"
// @Param overdue query bool false "Filter overdue vaccinations"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccinations [get]
func (h *Handler) ListVaccinations(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := VaccinationListFilters{
		PatientID:      c.Query("patient_id"),
		VeterinarianID: c.Query("veterinarian_id"),
		Status:         c.Query("status"),
		VaccineName:    c.Query("vaccine_name"),
		DateFrom:       c.Query("date_from"),
		DateTo:         c.Query("date_to"),
		DueSoon:        c.Query("due_soon") == "true",
		Overdue:        c.Query("overdue") == "true",
	}

	vaccinations, total, err := h.service.ListVaccinations(c.Request.Context(), filters, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]VaccinationResponse, len(vaccinations))
	for i, v := range vaccinations {
		data[i] = *v.ToResponse()
	}

	return gin.H{
		"data": data,
		"pagination": gin.H{
			"skip":        params.Skip,
			"limit":       params.Limit,
			"total":       total,
			"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
		},
	}, nil
}

// UpdateVaccination updates a vaccination
// @Summary Update vaccination
// @Description Update vaccination details
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "Vaccination ID"
// @Param vaccination body UpdateVaccinationDTO true "Updated vaccination data"
// @Success 200 {object} VaccinationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccinations/{id} [put]
func (h *Handler) UpdateVaccination(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "vaccination ID is required")
	}

	var dto UpdateVaccinationDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	vaccination, err := h.service.UpdateVaccination(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccination.ToResponse(), nil
}

// UpdateVaccinationStatus updates vaccination status
// @Summary Update vaccination status
// @Description Update the status of a vaccination
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "Vaccination ID"
// @Param status body UpdateVaccinationStatusDTO true "New status"
// @Success 200 {object} VaccinationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccinations/{id}/status [patch]
func (h *Handler) UpdateVaccinationStatus(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "vaccination ID is required")
	}

	var dto UpdateVaccinationStatusDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	vaccination, err := h.service.UpdateVaccinationStatus(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccination.ToResponse(), nil
}

// DeleteVaccination deletes a vaccination
// @Summary Delete vaccination
// @Description Soft delete a vaccination
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "Vaccination ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccinations/{id} [delete]
func (h *Handler) DeleteVaccination(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "vaccination ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteVaccination(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Vaccination deleted successfully"}, nil
}

// GetPatientVaccinations gets all vaccinations for a patient
// @Summary Get patient vaccinations
// @Description Get all vaccinations for a specific patient
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccinations/patient/{patient_id} [get]
func (h *Handler) GetPatientVaccinations(c *gin.Context) (any, error) {
	patientID := c.Param("patient_id")
	if patientID == "" {
		return nil, ErrValidation("patient_id", "patient ID is required")
	}

	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	vaccinations, total, err := h.service.GetPatientVaccinations(c.Request.Context(), patientID, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]VaccinationResponse, len(vaccinations))
	for i, v := range vaccinations {
		data[i] = *v.ToResponse()
	}

	return gin.H{
		"data": data,
		"pagination": gin.H{
			"skip":        params.Skip,
			"limit":       params.Limit,
			"total":       total,
			"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
		},
	}, nil
}

// GetDueVaccinations gets vaccinations due soon
// @Summary Get due vaccinations
// @Description Get vaccinations due within the next 30 days
// @Tags vaccinations
// @Accept json
// @Produce json
// @Success 200 {object} []VaccinationResponse
// @Security BearerAuth
// @Router /api/vaccinations/due [get]
func (h *Handler) GetDueVaccinations(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	vaccinations, err := h.service.GetDueVaccinations(c.Request.Context(), tenantID, 30)
	if err != nil {
		return nil, err
	}

	data := make([]VaccinationResponse, len(vaccinations))
	for i, v := range vaccinations {
		data[i] = *v.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// GetOverdueVaccinations gets overdue vaccinations
// @Summary Get overdue vaccinations
// @Description Get vaccinations that are past their due date
// @Tags vaccinations
// @Accept json
// @Produce json
// @Success 200 {object} []VaccinationResponse
// @Security BearerAuth
// @Router /api/vaccinations/overdue [get]
func (h *Handler) GetOverdueVaccinations(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	vaccinations, err := h.service.GetOverdueVaccinations(c.Request.Context(), tenantID)
	if err != nil {
		return nil, err
	}

	data := make([]VaccinationResponse, len(vaccinations))
	for i, v := range vaccinations {
		data[i] = *v.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// ==================== VACCINE CATALOG ====================

// CreateVaccine creates a new vaccine in the catalog
// @Summary Create vaccine
// @Description Create a new vaccine catalog entry
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param vaccine body CreateVaccineDTO true "Vaccine data"
// @Success 201 {object} VaccineResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccines [post]
func (h *Handler) CreateVaccine(c *gin.Context) (any, error) {
	var dto CreateVaccineDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	vaccine, err := h.service.CreateVaccine(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccine.ToResponse(), nil
}

// GetVaccine gets a vaccine by ID
// @Summary Get vaccine
// @Description Get vaccine details by ID
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "Vaccine ID"
// @Success 200 {object} VaccineResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccines/{id} [get]
func (h *Handler) GetVaccine(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "vaccine ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	vaccine, err := h.service.GetVaccine(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccine.ToResponse(), nil
}

// ListVaccines lists all vaccines
// @Summary List vaccines
// @Description Get all vaccines for the tenant
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param dose_number query string false "Filter by dose type (first, second, booster)"
// @Param target_species query string false "Filter by target species"
// @Param active query bool false "Filter by active status"
// @Param search query string false "Search by name or manufacturer"
// @Success 200 {object} []VaccineResponse
// @Security BearerAuth
// @Router /api/vaccines [get]
func (h *Handler) ListVaccines(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := VaccineListFilters{
		DoseNumber:    c.Query("dose_number"),
		TargetSpecies: c.Query("target_species"),
		Search:        c.Query("search"),
	}

	if active := c.Query("active"); active != "" {
		activeBool := active == "true"
		filters.Active = &activeBool
	}

	vaccines, err := h.service.ListVaccines(c.Request.Context(), filters, tenantID)
	if err != nil {
		return nil, err
	}

	data := make([]VaccineResponse, len(vaccines))
	for i, v := range vaccines {
		data[i] = *v.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// UpdateVaccine updates a vaccine
// @Summary Update vaccine
// @Description Update vaccine details
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "Vaccine ID"
// @Param vaccine body UpdateVaccineDTO true "Updated vaccine data"
// @Success 200 {object} VaccineResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccines/{id} [put]
func (h *Handler) UpdateVaccine(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "vaccine ID is required")
	}

	var dto UpdateVaccineDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	vaccine, err := h.service.UpdateVaccine(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return vaccine.ToResponse(), nil
}

// DeleteVaccine deletes a vaccine
// @Summary Delete vaccine
// @Description Soft delete a vaccine from the catalog
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "Vaccine ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/vaccines/{id} [delete]
func (h *Handler) DeleteVaccine(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "vaccine ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteVaccine(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Vaccine deleted successfully"}, nil
}
