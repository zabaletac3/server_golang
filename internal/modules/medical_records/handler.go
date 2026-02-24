package medical_records

import (
	"strconv"

	"github.com/gin-gonic/gin"

	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

// Handler handles HTTP requests for medical records
type Handler struct {
	service *Service
}

// NewHandler creates a new medical records handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ==================== MEDICAL RECORDS ====================

// CreateMedicalRecord creates a new medical record
// @Summary Create medical record
// @Description Create a new medical record for a patient
// @Tags medical-records
// @Accept json
// @Produce json
// @Param record body CreateMedicalRecordDTO true "Medical record data"
// @Success 201 {object} MedicalRecordResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-records [post]
func (h *Handler) CreateMedicalRecord(c *gin.Context) (any, error) {
	var dto CreateMedicalRecordDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	record, err := h.service.CreateMedicalRecord(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return record.ToResponse(), nil
}

// GetMedicalRecord gets a medical record by ID
// @Summary Get medical record
// @Description Get medical record details by ID
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "Record ID"
// @Success 200 {object} MedicalRecordResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-records/{id} [get]
func (h *Handler) GetMedicalRecord(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "record ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	record, err := h.service.GetMedicalRecord(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return record.ToResponse(), nil
}

// ListMedicalRecords lists medical records with filters
// @Summary List medical records
// @Description Get a paginated list of medical records with optional filters
// @Tags medical-records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param patient_id query string false "Filter by patient ID"
// @Param veterinarian_id query string false "Filter by veterinarian ID"
// @Param type query string false "Filter by type"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Param has_attachments query bool false "Filter records with attachments"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-records [get]
func (h *Handler) ListMedicalRecords(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := MedicalRecordListFilters{
		PatientID:      c.Query("patient_id"),
		VeterinarianID: c.Query("veterinarian_id"),
		Type:           c.Query("type"),
		DateFrom:       c.Query("date_from"),
		DateTo:         c.Query("date_to"),
		HasAttachments: c.Query("has_attachments") == "true",
	}

	records, total, err := h.service.ListMedicalRecords(c.Request.Context(), filters, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]MedicalRecordResponse, len(records))
	for i, r := range records {
		data[i] = *r.ToResponse()
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

// GetPatientRecords gets all records for a patient
// @Summary Get patient records
// @Description Get all medical records for a specific patient
// @Tags medical-records
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-records/patient/{patient_id} [get]
func (h *Handler) GetPatientRecords(c *gin.Context) (any, error) {
	patientID := c.Param("patient_id")
	if patientID == "" {
		return nil, ErrValidation("patient_id", "patient ID is required")
	}

	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	records, total, err := h.service.GetPatientRecords(c.Request.Context(), patientID, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]MedicalRecordResponse, len(records))
	for i, r := range records {
		data[i] = *r.ToResponse()
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

// UpdateMedicalRecord updates a medical record
// @Summary Update medical record
// @Description Update medical record (only within 24 hours of creation)
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "Record ID"
// @Param record body UpdateMedicalRecordDTO true "Updated record data"
// @Success 200 {object} MedicalRecordResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-records/{id} [put]
func (h *Handler) UpdateMedicalRecord(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "record ID is required")
	}

	var dto UpdateMedicalRecordDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	record, err := h.service.UpdateMedicalRecord(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return record.ToResponse(), nil
}

// DeleteMedicalRecord deletes a medical record
// @Summary Delete medical record
// @Description Soft delete a medical record
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "Record ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-records/{id} [delete]
func (h *Handler) DeleteMedicalRecord(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "record ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteMedicalRecord(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Medical record deleted successfully"}, nil
}

// GetPatientTimeline gets the medical timeline for a patient
// @Summary Get patient timeline
// @Description Get chronological medical timeline for a patient
// @Tags medical-records
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Param record_type query string false "Filter by record type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} MedicalTimeline
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-records/patient/{patient_id}/timeline [get]
func (h *Handler) GetPatientTimeline(c *gin.Context) (any, error) {
	patientID := c.Param("patient_id")
	if patientID == "" {
		return nil, ErrValidation("patient_id", "patient ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	filters := TimelineFilters{
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		RecordType: c.Query("record_type"),
	}

	// Parse limit and skip
	limitStr := c.DefaultQuery("limit", "50")
	skipStr := c.DefaultQuery("skip", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	skip, err := strconv.Atoi(skipStr)
	if err != nil {
		skip = 0
	}
	filters.Limit = limit
	filters.Skip = skip

	timeline, err := h.service.GetPatientTimeline(c.Request.Context(), patientID, tenantID, filters)
	if err != nil {
		return nil, err
	}

	return timeline, nil
}

// ==================== ALLERGIES ====================

// CreateAllergy creates a new allergy
// @Summary Create allergy
// @Description Register a new allergy for a patient
// @Tags medical-records
// @Accept json
// @Produce json
// @Param allergy body CreateAllergyDTO true "Allergy data"
// @Success 201 {object} AllergyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/allergies [post]
func (h *Handler) CreateAllergy(c *gin.Context) (any, error) {
	var dto CreateAllergyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	allergy, err := h.service.CreateAllergy(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return allergy.ToResponse(), nil
}

// GetPatientAllergies gets all allergies for a patient
// @Summary Get patient allergies
// @Description Get all allergies for a specific patient
// @Tags medical-records
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Success 200 {object} []AllergyResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/allergies/patient/{patient_id} [get]
func (h *Handler) GetPatientAllergies(c *gin.Context) (any, error) {
	patientID := c.Param("patient_id")
	if patientID == "" {
		return nil, ErrValidation("patient_id", "patient ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	allergies, err := h.service.GetPatientAllergies(c.Request.Context(), patientID, tenantID)
	if err != nil {
		return nil, err
	}

	data := make([]AllergyResponse, len(allergies))
	for i, a := range allergies {
		data[i] = *a.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// UpdateAllergy updates an allergy
// @Summary Update allergy
// @Description Update allergy details
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "Allergy ID"
// @Param allergy body UpdateAllergyDTO true "Updated allergy data"
// @Success 200 {object} AllergyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/allergies/{id} [put]
func (h *Handler) UpdateAllergy(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "allergy ID is required")
	}

	var dto UpdateAllergyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	allergy, err := h.service.UpdateAllergy(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return allergy.ToResponse(), nil
}

// DeleteAllergy deletes an allergy
// @Summary Delete allergy
// @Description Soft delete an allergy
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "Allergy ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/allergies/{id} [delete]
func (h *Handler) DeleteAllergy(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "allergy ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteAllergy(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Allergy deleted successfully"}, nil
}

// ==================== MEDICAL HISTORY ====================

// CreateMedicalHistory creates or updates medical history
// @Summary Create/Update medical history
// @Description Create or update medical history summary for a patient
// @Tags medical-records
// @Accept json
// @Produce json
// @Param history body CreateMedicalHistoryDTO true "Medical history data"
// @Success 201 {object} MedicalHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-history [post]
func (h *Handler) CreateMedicalHistory(c *gin.Context) (any, error) {
	var dto CreateMedicalHistoryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	history, err := h.service.CreateMedicalHistory(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return history.ToResponse(), nil
}

// GetMedicalHistory gets medical history for a patient
// @Summary Get medical history
// @Description Get medical history summary for a specific patient
// @Tags medical-records
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Success 200 {object} MedicalHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-history/patient/{patient_id} [get]
func (h *Handler) GetMedicalHistory(c *gin.Context) (any, error) {
	patientID := c.Param("patient_id")
	if patientID == "" {
		return nil, ErrValidation("patient_id", "patient ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	history, err := h.service.GetMedicalHistory(c.Request.Context(), patientID, tenantID)
	if err != nil {
		return nil, err
	}

	return history.ToResponse(), nil
}

// UpdateMedicalHistory updates medical history
// @Summary Update medical history
// @Description Update medical history summary
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "History ID"
// @Param history body UpdateMedicalHistoryDTO true "Updated history data"
// @Success 200 {object} MedicalHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-history/{id} [put]
func (h *Handler) UpdateMedicalHistory(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "history ID is required")
	}

	var dto UpdateMedicalHistoryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	history, err := h.service.UpdateMedicalHistory(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return history.ToResponse(), nil
}

// DeleteMedicalHistory deletes medical history
// @Summary Delete medical history
// @Description Soft delete medical history
// @Tags medical-records
// @Accept json
// @Produce json
// @Param id path string true "History ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/medical-history/{id} [delete]
func (h *Handler) DeleteMedicalHistory(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "history ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteMedicalHistory(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Medical history deleted successfully"}, nil
}
