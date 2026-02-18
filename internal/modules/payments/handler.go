package payments

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/validation"
)

type PaymentHandler struct {
	service *PaymentService
}

func NewPaymentHandler(service *PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// Create godoc
// @Summary      Crear pago
// @Description  Registra un nuevo pago en el sistema
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payment body CreatePaymentDTO true "Datos del pago"
// @Success      201 {object} PaymentResponse
// @Failure      400 {object} validation.ValidationError
// @Failure      500 {object} map[string]string
// @Router       /api/payments [post]
func (h *PaymentHandler) Create(c *gin.Context) (any, error) {
	var dto CreatePaymentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Create(c.Request.Context(), &dto)
}

// FindByID godoc
// @Summary      Obtener pago
// @Description  Obtiene un pago por ID
// @Tags         payments
// @Produce      json
// @Param        id path string true "Payment ID"
// @Success      200 {object} PaymentResponse
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/payments/{id} [get]
func (h *PaymentHandler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// FindByTenantID godoc
// @Summary      Historial de pagos de un tenant
// @Description  Obtiene el historial de pagos de un tenant
// @Tags         payments
// @Produce      json
// @Param        tenant_id path string true "Tenant ID"
// @Param        limit query int false "Límite de resultados" default(50)
// @Success      200 {array} PaymentResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/tenants/{tenant_id}/payments [get]
func (h *PaymentHandler) FindByTenantID(c *gin.Context) (any, error) {
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		tenantID = c.Param("id")
	}
	
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	return h.service.FindByTenantID(c.Request.Context(), tenantID, limit)
}

