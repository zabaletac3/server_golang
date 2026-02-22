package appointments

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/shared/auth"
	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

// Handler handles HTTP requests for appointments
type Handler struct {
	service *Service
}

// NewHandler creates a new appointment handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Admin/Staff endpoints

// CreateAppointment creates a new appointment
// @Summary Create appointment
// @Description Create a new appointment in the system
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param appointment body CreateAppointmentDTO true "Appointment data"
// @Success 201 {object} AppointmentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments [post]
func (h *Handler) CreateAppointment(c *gin.Context) (any, error) {
	var dto CreateAppointmentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	userIDStr := auth.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrValidationFailed("user_id", "invalid user ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	appointment, err := h.service.CreateAppointment(c.Request.Context(), dto, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// GetAppointment gets an appointment by ID
// @Summary Get appointment
// @Description Get appointment details by ID
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param populate query bool false "Populate related data"
// @Success 200 {object} AppointmentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments/{id} [get]
func (h *Handler) GetAppointment(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidationFailed("id", "appointment ID is required")
	}

	populate := c.Query("populate") == "true"
	tenantID := sharedMiddleware.GetTenantID(c)

	appointment, err := h.service.GetAppointment(c.Request.Context(), id, tenantID, populate)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// ListAppointments lists appointments with filters and pagination
// @Summary List appointments
// @Description Get a paginated list of appointments with optional filters
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query []string false "Filter by status"
// @Param type query []string false "Filter by appointment type"
// @Param veterinarian_id query string false "Filter by veterinarian ID"
// @Param patient_id query string false "Filter by patient ID"
// @Param owner_id query string false "Filter by owner ID"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Param priority query string false "Filter by priority"
// @Param populate query bool false "Populate related data"
// @Success 200 {object} PaginatedAppointmentsResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments [get]
func (h *Handler) ListAppointments(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	populate := c.Query("populate") == "true"
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := make(map[string]interface{})

	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		filters["status"] = statuses
	}

	if types := c.QueryArray("type"); len(types) > 0 {
		filters["type"] = types
	}

	if vetID := c.Query("veterinarian_id"); vetID != "" {
		filters["veterinarian_id"] = vetID
	}

	if patientID := c.Query("patient_id"); patientID != "" {
		filters["patient_id"] = patientID
	}

	if ownerID := c.Query("owner_id"); ownerID != "" {
		filters["owner_id"] = ownerID
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		df, err := time.Parse(time.RFC3339, dateFrom)
		if err != nil {
			return nil, ErrValidationFailed("date_from", "invalid date format, use RFC3339")
		}
		filters["date_from"] = df
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		dt, err := time.Parse(time.RFC3339, dateTo)
		if err != nil {
			return nil, ErrValidationFailed("date_to", "invalid date format, use RFC3339")
		}
		filters["date_to"] = dt
	}

	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}

	appointments, err := h.service.ListAppointments(c.Request.Context(), filters, tenantID, params, populate)
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

// UpdateAppointment updates an appointment
// @Summary Update appointment
// @Description Update appointment details
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param appointment body UpdateAppointmentDTO true "Updated appointment data"
// @Success 200 {object} AppointmentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments/{id} [put]
func (h *Handler) UpdateAppointment(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidationFailed("id", "appointment ID is required")
	}

	var dto UpdateAppointmentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	userIDStr := auth.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrValidationFailed("user_id", "invalid user ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	appointment, err := h.service.UpdateAppointment(c.Request.Context(), id, dto, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// DeleteAppointment deletes an appointment
// @Summary Delete appointment
// @Description Soft delete an appointment
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments/{id} [delete]
func (h *Handler) DeleteAppointment(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidationFailed("id", "appointment ID is required")
	}

	userIDStr := auth.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrValidationFailed("user_id", "invalid user ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err = h.service.DeleteAppointment(c.Request.Context(), id, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Appointment deleted successfully"}, nil
}

// UpdateStatus updates an appointment status
// @Summary Update appointment status
// @Description Update the status of an appointment
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param status body UpdateStatusDTO true "New status data"
// @Success 200 {object} AppointmentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments/{id}/status [patch]
func (h *Handler) UpdateStatus(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidationFailed("id", "appointment ID is required")
	}

	var dto UpdateStatusDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	userIDStr := auth.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrValidationFailed("user_id", "invalid user ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	appointment, err := h.service.UpdateStatus(c.Request.Context(), id, dto, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// GetCalendarView gets a calendar view of appointments
// @Summary Get calendar view
// @Description Get appointments organized by date for calendar display
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param date_from query string true "Start date (RFC3339)"
// @Param date_to query string true "End date (RFC3339)"
// @Param veterinarian_id query string false "Filter by veterinarian ID"
// @Success 200 {object} []CalendarViewResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments/calendar [get]
func (h *Handler) GetCalendarView(c *gin.Context) (any, error) {
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	if dateFromStr == "" || dateToStr == "" {
		return nil, ErrValidationFailed("date_range", "both date_from and date_to are required")
	}

	dateFrom, err := time.Parse(time.RFC3339, dateFromStr)
	if err != nil {
		return nil, ErrValidationFailed("date_from", "invalid date format, use RFC3339")
	}

	dateTo, err := time.Parse(time.RFC3339, dateToStr)
	if err != nil {
		return nil, ErrValidationFailed("date_to", "invalid date format, use RFC3339")
	}

	var veterinarianID *string
	if vetID := c.Query("veterinarian_id"); vetID != "" {
		veterinarianID = &vetID
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	calendar, err := h.service.GetCalendarView(c.Request.Context(), dateFrom, dateTo, veterinarianID, tenantID)
	if err != nil {
		return nil, err
	}

	return calendar, nil
}

// CheckAvailability checks veterinarian availability
// @Summary Check veterinarian availability
// @Description Check if a veterinarian is available at a specific time
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param veterinarian_id query string true "Veterinarian ID"
// @Param scheduled_at query string true "Scheduled time (RFC3339)"
// @Param duration query int true "Duration in minutes"
// @Param exclude_id query string false "Exclude appointment ID (for updates)"
// @Success 200 {object} AvailabilityResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments/availability [get]
func (h *Handler) CheckAvailability(c *gin.Context) (any, error) {
	vetID := c.Query("veterinarian_id")
	scheduledAtStr := c.Query("scheduled_at")
	durationStr := c.Query("duration")

	if vetID == "" || scheduledAtStr == "" || durationStr == "" {
		return nil, ErrValidationFailed("required_params", "veterinarian_id, scheduled_at, and duration are required")
	}

	scheduledAt, err := time.Parse(time.RFC3339, scheduledAtStr)
	if err != nil {
		return nil, ErrValidationFailed("scheduled_at", "invalid date format, use RFC3339")
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		return nil, ErrValidationFailed("duration", "duration must be a valid integer")
	}

	var excludeID *string
	if exclude := c.Query("exclude_id"); exclude != "" {
		excludeID = &exclude
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	available, err := h.service.CheckAvailability(c.Request.Context(), vetID, scheduledAt, duration, excludeID, tenantID)
	if err != nil {
		return nil, err
	}

	return available, nil
}

// GetStatusHistory gets the status history for an appointment
// @Summary Get status history
// @Description Get the status change history for an appointment
// @Tags admin-appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Success 200 {object} []AppointmentStatusTransitionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/appointments/{id}/history [get]
func (h *Handler) GetStatusHistory(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidationFailed("id", "appointment ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	history, err := h.service.GetStatusHistory(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// Mobile endpoints

// RequestAppointment creates an appointment request from mobile
// @Summary Request appointment
// @Description Request a new appointment from mobile app
// @Tags mobile-appointments
// @Accept json
// @Produce json
// @Param appointment body MobileAppointmentRequestDTO true "Appointment request data"
// @Success 201 {object} AppointmentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security MobileBearerAuth
// @Router /mobile/appointments/request [post]
func (h *Handler) RequestAppointment(c *gin.Context) (any, error) {
	var dto MobileAppointmentRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	ownerIDStr := auth.GetUserID(c)
	ownerID, err := primitive.ObjectIDFromHex(ownerIDStr)
	if err != nil {
		return nil, ErrValidationFailed("owner_id", "invalid owner ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	appointment, err := h.service.RequestAppointment(c.Request.Context(), dto, tenantID, ownerID)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// GetOwnerAppointments gets appointments for a specific owner
// @Summary Get owner appointments
// @Description Get appointments for the authenticated owner
// @Tags mobile-appointments
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} PaginatedAppointmentsResponse
// @Failure 400 {object} map[string]interface{}
// @Security MobileBearerAuth
// @Router /mobile/appointments [get]
func (h *Handler) GetOwnerAppointments(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)

	ownerIDStr := auth.GetUserID(c)
	ownerID, err := primitive.ObjectIDFromHex(ownerIDStr)
	if err != nil {
		return nil, ErrValidationFailed("owner_id", "invalid owner ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	appointments, err := h.service.GetOwnerAppointments(c.Request.Context(), ownerID, tenantID, params)
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

// GetOwnerAppointment gets a specific appointment for an owner
// @Summary Get owner appointment
// @Description Get appointment details for the authenticated owner
// @Tags mobile-appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param populate query bool false "Populate related data"
// @Success 200 {object} AppointmentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security MobileBearerAuth
// @Router /mobile/appointments/{id} [get]
func (h *Handler) GetOwnerAppointment(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidationFailed("id", "appointment ID is required")
	}

	ownerIDStr := auth.GetUserID(c)
	ownerID, err := primitive.ObjectIDFromHex(ownerIDStr)
	if err != nil {
		return nil, ErrValidationFailed("owner_id", "invalid owner ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)
	populate := c.Query("populate") == "true"

	appointment, err := h.service.GetOwnerAppointment(c.Request.Context(), id, tenantID, ownerID, populate)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// CancelOwnerAppointment cancels an appointment by the owner
// @Summary Cancel appointment
// @Description Cancel an appointment from mobile app
// @Tags mobile-appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Param cancel body AppointmentCancelDTO true "Cancellation reason"
// @Success 200 {object} AppointmentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security MobileBearerAuth
// @Router /mobile/appointments/{id}/cancel [patch]
func (h *Handler) CancelOwnerAppointment(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidationFailed("id", "appointment ID is required")
	}

	var dto AppointmentCancelDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	ownerIDStr := auth.GetUserID(c)
	ownerID, err := primitive.ObjectIDFromHex(ownerIDStr)
	if err != nil {
		return nil, ErrValidationFailed("owner_id", "invalid owner ID format")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	appointment, err := h.service.CancelAppointment(c.Request.Context(), id, dto.Reason, tenantID, ownerID)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}
