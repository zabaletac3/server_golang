package tenant

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/validation"
)

type Handler struct {
	service *TenantService
}

func NewHandler(service *TenantService) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary      Crear tenant
// @Description  Crea un nuevo tenant en el sistema
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Param        body  body      CreateTenantDTO  true  "Datos del tenant"
// @Success      200   {object}  TenantResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      409   {object}  map[string]string
// @Router       /tenants [post]
func (h *Handler) Create(c *gin.Context) (any, error) {
	var dto CreateTenantDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Create(c.Request.Context(), &dto)
}

// FindAll godoc
// @Summary      Listar tenants
// @Description  Obtiene una lista de tenants
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Success      200  {array}   TenantResponse
// @Router       /tenants [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	return h.service.FindAll(c.Request.Context())
}

// FindByID godoc
// @Summary      Obtener tenant
// @Description  Obtiene un tenant por su ID
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Tenant ID"
// @Success      200  {object}  TenantResponse
// @Failure      404  {object}  map[string]string
// @Router       /tenants/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Update godoc
// @Summary      Actualizar tenant
// @Description  Actualiza un tenant existente
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Param        id    path      string          true  "Tenant ID"
// @Param        body  body      UpdateTenantDTO true  "Datos a actualizar"
// @Success      200   {object}  TenantResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      404   {object}  map[string]string
// @Router       /tenants/{id} [patch]
func (h *Handler) Update(c *gin.Context) (any, error) {
	id := c.Param("id")

	var dto UpdateTenantDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Update(c.Request.Context(), id, &dto)
}

// Delete godoc
// @Summary      Eliminar tenant
// @Description  Elimina un tenant por su ID (soft delete)
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Tenant ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Router       /tenants/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}

	return gin.H{"deleted": true}, nil
}