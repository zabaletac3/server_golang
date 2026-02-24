package laboratory

import (
	"github.com/gin-gonic/gin"

	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

// Handler handles HTTP requests for laboratory
type Handler struct {
	service *Service
}

// NewHandler creates a new laboratory handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ==================== LAB ORDERS ====================

// CreateLabOrder creates a new lab order
// @Summary Create lab order
// @Description Create a new laboratory order for a patient
// @Tags laboratory
// @Accept json
// @Produce json
// @Param order body CreateLabOrderDTO true "Lab order data"
// @Success 201 {object} LabOrderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders [post]
func (h *Handler) CreateLabOrder(c *gin.Context) (any, error) {
	var dto CreateLabOrderDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	order, err := h.service.CreateLabOrder(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return order.ToResponse(), nil
}

// GetLabOrder gets a lab order by ID
// @Summary Get lab order
// @Description Get lab order details by ID
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} LabOrderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders/{id} [get]
func (h *Handler) GetLabOrder(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "order ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	order, err := h.service.GetLabOrder(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return order.ToResponse(), nil
}

// ListLabOrders lists lab orders with filters
// @Summary List lab orders
// @Description Get a paginated list of lab orders with optional filters
// @Tags laboratory
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param patient_id query string false "Filter by patient ID"
// @Param veterinarian_id query string false "Filter by veterinarian ID"
// @Param status query string false "Filter by status"
// @Param test_type query string false "Filter by test type"
// @Param lab_id query string false "Filter by lab ID"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Param overdue query bool false "Filter overdue orders"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders [get]
func (h *Handler) ListLabOrders(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := LabOrderListFilters{
		PatientID:      c.Query("patient_id"),
		VeterinarianID: c.Query("veterinarian_id"),
		Status:         c.Query("status"),
		TestType:       c.Query("test_type"),
		LabID:          c.Query("lab_id"),
		DateFrom:       c.Query("date_from"),
		DateTo:         c.Query("date_to"),
		Overdue:        c.Query("overdue") == "true",
	}

	orders, total, err := h.service.ListLabOrders(c.Request.Context(), filters, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]LabOrderResponse, len(orders))
	for i, o := range orders {
		data[i] = *o.ToResponse()
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

// UpdateLabOrder updates a lab order
// @Summary Update lab order
// @Description Update lab order details
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param order body UpdateLabOrderDTO true "Updated order data"
// @Success 200 {object} LabOrderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders/{id} [put]
func (h *Handler) UpdateLabOrder(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "order ID is required")
	}

	var dto UpdateLabOrderDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	order, err := h.service.UpdateLabOrder(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return order.ToResponse(), nil
}

// UpdateLabOrderStatus updates lab order status
// @Summary Update lab order status
// @Description Update the status of a lab order
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body UpdateLabOrderStatusDTO true "New status"
// @Success 200 {object} LabOrderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders/{id}/status [patch]
func (h *Handler) UpdateLabOrderStatus(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "order ID is required")
	}

	var dto UpdateLabOrderStatusDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	order, err := h.service.UpdateLabOrderStatus(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return order.ToResponse(), nil
}

// UploadLabResult uploads a lab result file
// @Summary Upload lab result
// @Description Upload a result file for a lab order
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param result body UploadLabResultDTO true "Result file ID"
// @Success 200 {object} LabOrderResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders/{id}/result [post]
func (h *Handler) UploadLabResult(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "order ID is required")
	}

	var dto UploadLabResultDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	order, err := h.service.UploadLabResult(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return order.ToResponse(), nil
}

// DeleteLabOrder deletes a lab order
// @Summary Delete lab order
// @Description Soft delete a lab order
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders/{id} [delete]
func (h *Handler) DeleteLabOrder(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "order ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteLabOrder(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Lab order deleted successfully"}, nil
}

// GetPatientLabOrders gets all lab orders for a patient
// @Summary Get patient lab orders
// @Description Get all lab orders for a specific patient
// @Tags laboratory
// @Accept json
// @Produce json
// @Param patient_id path string true "Patient ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-orders/patient/{patient_id} [get]
func (h *Handler) GetPatientLabOrders(c *gin.Context) (any, error) {
	patientID := c.Param("patient_id")
	if patientID == "" {
		return nil, ErrValidation("patient_id", "patient ID is required")
	}

	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	orders, total, err := h.service.GetPatientLabOrders(c.Request.Context(), patientID, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]LabOrderResponse, len(orders))
	for i, o := range orders {
		data[i] = *o.ToResponse()
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

// GetOverdueLabOrders gets overdue lab orders
// @Summary Get overdue lab orders
// @Description Get lab orders that are past their turnaround time
// @Tags laboratory
// @Accept json
// @Produce json
// @Success 200 {object} []LabOrderResponse
// @Security BearerAuth
// @Router /api/lab-orders/overdue [get]
func (h *Handler) GetOverdueLabOrders(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	// Default turnaround time: 5 days
	orders, err := h.service.GetOverdueLabOrders(c.Request.Context(), tenantID, 5)
	if err != nil {
		return nil, err
	}

	data := make([]LabOrderResponse, len(orders))
	for i, o := range orders {
		data[i] = *o.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// ==================== LAB TESTS CATALOG ====================

// CreateLabTest creates a new lab test in the catalog
// @Summary Create lab test
// @Description Create a new lab test catalog entry
// @Tags laboratory
// @Accept json
// @Produce json
// @Param test body CreateLabTestDTO true "Lab test data"
// @Success 201 {object} LabTestResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-tests [post]
func (h *Handler) CreateLabTest(c *gin.Context) (any, error) {
	var dto CreateLabTestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	test, err := h.service.CreateLabTest(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return test.ToResponse(), nil
}

// GetLabTest gets a lab test by ID
// @Summary Get lab test
// @Description Get lab test details by ID
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} LabTestResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-tests/{id} [get]
func (h *Handler) GetLabTest(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "test ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	test, err := h.service.GetLabTest(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return test.ToResponse(), nil
}

// ListLabTests lists all lab tests
// @Summary List lab tests
// @Description Get all lab tests for the tenant
// @Tags laboratory
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param active query bool false "Filter by active status"
// @Param search query string false "Search by name"
// @Success 200 {object} []LabTestResponse
// @Security BearerAuth
// @Router /api/lab-tests [get]
func (h *Handler) ListLabTests(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := LabTestListFilters{
		Category: c.Query("category"),
		Search:   c.Query("search"),
	}

	if active := c.Query("active"); active != "" {
		activeBool := active == "true"
		filters.Active = &activeBool
	}

	tests, err := h.service.ListLabTests(c.Request.Context(), filters, tenantID)
	if err != nil {
		return nil, err
	}

	data := make([]LabTestResponse, len(tests))
	for i, t := range tests {
		data[i] = *t.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// UpdateLabTest updates a lab test
// @Summary Update lab test
// @Description Update lab test details
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Param test body UpdateLabTestDTO true "Updated test data"
// @Success 200 {object} LabTestResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-tests/{id} [put]
func (h *Handler) UpdateLabTest(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "test ID is required")
	}

	var dto UpdateLabTestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	test, err := h.service.UpdateLabTest(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return test.ToResponse(), nil
}

// DeleteLabTest deletes a lab test
// @Summary Delete lab test
// @Description Soft delete a lab test from the catalog
// @Tags laboratory
// @Accept json
// @Produce json
// @Param id path string true "Test ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/lab-tests/{id} [delete]
func (h *Handler) DeleteLabTest(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "test ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteLabTest(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Lab test deleted successfully"}, nil
}
