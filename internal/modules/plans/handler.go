package plans

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/validation"
)

type PlanHandler struct {
	service *PlanService
}

func NewPlanHandler(service *PlanService) *PlanHandler {
	return &PlanHandler{service: service}
}

// Create godoc
// @Summary      Crear plan
// @Description  Crea un nuevo plan de suscripción
// @Tags         plans
// @Accept       json
// @Produce      json
// @Param        plan body CreatePlanDTO true "Datos del plan"
// @Success      201 {object} PlanResponse
// @Failure      400 {object} validation.ValidationError
// @Failure      500 {object} map[string]string
// @Router       /api/plans [post]
func (h *PlanHandler) Create(c *gin.Context) (any, error) {
	var dto CreatePlanDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Create(c.Request.Context(), &dto)
}

// FindAll godoc
// @Summary      Listar planes
// @Description  Obtiene todos los planes (admin)
// @Tags         plans
// @Produce      json
// @Success      200 {array} PlanResponse
// @Failure      500 {object} map[string]string
// @Router       /api/plans [get]
func (h *PlanHandler) FindAll(c *gin.Context) (any, error) {
	return h.service.FindAll(c.Request.Context())
}

// FindVisible godoc
// @Summary      Listar planes visibles
// @Description  Obtiene solo los planes visibles (público)
// @Tags         plans
// @Produce      json
// @Success      200 {array} PlanResponse
// @Failure      500 {object} map[string]string
// @Router       /api/plans/visible [get]
func (h *PlanHandler) FindVisible(c *gin.Context) (any, error) {
	return h.service.FindVisible(c.Request.Context())
}

// FindByID godoc
// @Summary      Obtener plan
// @Description  Obtiene un plan por ID
// @Tags         plans
// @Produce      json
// @Param        id path string true "Plan ID"
// @Success      200 {object} PlanResponse
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/plans/{id} [get]
func (h *PlanHandler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Update godoc
// @Summary      Actualizar plan
// @Description  Actualiza un plan existente
// @Tags         plans
// @Accept       json
// @Produce      json
// @Param        id path string true "Plan ID"
// @Param        plan body UpdatePlanDTO true "Datos a actualizar"
// @Success      200 {object} PlanResponse
// @Failure      400 {object} validation.ValidationError
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/plans/{id} [patch]
func (h *PlanHandler) Update(c *gin.Context) (any, error) {
	id := c.Param("id")

	var dto UpdatePlanDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Update(c.Request.Context(), id, &dto)
}

// Delete godoc
// @Summary      Eliminar plan
// @Description  Elimina un plan
// @Tags         plans
// @Param        id path string true "Plan ID"
// @Success      204
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/plans/{id} [delete]
func (h *PlanHandler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")
	return nil, h.service.Delete(c.Request.Context(), id)
}
