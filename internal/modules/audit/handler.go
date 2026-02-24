package audit

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// Handler handles HTTP requests for audit logs
type Handler struct {
	service *Service
}

// NewHandler creates a new audit handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetAuditLogs retrieves audit logs with filters
// @Summary Get audit logs
// @Description Get paginated audit logs with optional filters
// @Tags audit
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param user_id query string false "Filter by user ID"
// @Param event_type query string false "Filter by event type"
// @Param resource query string false "Filter by resource"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/audit-logs [get]
func (h *Handler) GetAuditLogs(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	filter := AuditFilter{
		TenantID: tenantID,
		Skip:     int(params.Skip),
		Limit:    int(params.Limit),
	}

	if userID := c.Query("user_id"); userID != "" {
		if oid, err := primitive.ObjectIDFromHex(userID); err == nil {
			filter.UserID = &oid
		}
	}

	if eventType := c.Query("event_type"); eventType != "" {
		et := EventType(eventType)
		filter.EventType = &et
	}

	if resource := c.Query("resource"); resource != "" {
		filter.Resource = &resource
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if df, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filter.DateFrom = &df
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if dt, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filter.DateTo = &dt
		}
	}

	events, total, err := h.service.GetEvents(c.Request.Context(), filter)
	if err != nil {
		return nil, err
	}

	data := make([]AuditEventResponse, len(events))
	for i, e := range events {
		data[i] = *e.ToResponse()
	}

	return gin.H{
		"data": data,
		"pagination": gin.H{
			"skip":       params.Skip,
			"limit":      params.Limit,
			"total":      total,
			"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
		},
	}, nil
}

// GetResourceAuditLogs retrieves audit logs for a specific resource
// @Summary Get resource audit logs
// @Description Get audit logs for a specific resource
// @Tags audit
// @Accept json
// @Produce json
// @Param resource path string true "Resource name"
// @Param resource_id path string true "Resource ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/audit-logs/{resource}/{resource_id} [get]
func (h *Handler) GetResourceAuditLogs(c *gin.Context) (any, error) {
	resource := c.Param("resource")
	resourceID := c.Param("resource_id")
	tenantID := sharedMiddleware.GetTenantID(c)

	rid, err := primitive.ObjectIDFromHex(resourceID)
	if err != nil {
		return nil, err
	}

	events, err := h.service.GetEventsByResource(c.Request.Context(), tenantID, rid, resource)
	if err != nil {
		return nil, err
	}

	data := make([]AuditEventResponse, len(events))
	for i, e := range events {
		data[i] = *e.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// RegisterRoutes registers audit log routes
func RegisterRoutes(r *gin.RouterGroup, db interface{}) {
	// Routes would be registered here
	// For now, this is a placeholder for future implementation
}
