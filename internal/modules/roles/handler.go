package roles

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/validation"
)

type Handler struct {
	service *RoleService
}

func NewHandler(service *RoleService) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary      Crear rol
// @Description  Crea un nuevo rol para un tenant
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        body  body      CreateRoleDTO  true  "Datos del rol"
// @Success      200   {object}  RoleSimpleResponse
// @Failure      400   {object}  validation.ValidationError
// @Router       /roles [post]
func (h *Handler) Create(c *gin.Context) (any, error) {
	var dto CreateRoleDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Create(c.Request.Context(), &dto)
}

// FindByTenantID godoc
// @Summary      Listar roles por tenant
// @Description  Obtiene todos los roles de un tenant
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        tenant_id  query     string  true  "Tenant ID"
// @Success      200        {array}   RoleSimpleResponse
// @Router       /roles [get]
func (h *Handler) FindByTenantID(c *gin.Context) (any, error) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		return nil, ErrInvalidRoleID
	}
	return h.service.FindByTenantID(c.Request.Context(), tenantID)
}

// FindByID godoc
// @Summary      Obtener rol
// @Description  Obtiene un rol por su ID con permisos expandidos
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  RoleResponse
// @Failure      404  {object}  map[string]string
// @Router       /roles/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Update godoc
// @Summary      Actualizar rol
// @Description  Actualiza un rol existente
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id    path      string        true  "Role ID"
// @Param        body  body      UpdateRoleDTO true  "Datos a actualizar"
// @Success      200   {object}  RoleSimpleResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      404   {object}  map[string]string
// @Router       /roles/{id} [patch]
func (h *Handler) Update(c *gin.Context) (any, error) {
	id := c.Param("id")

	var dto UpdateRoleDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Update(c.Request.Context(), id, &dto)
}

// Delete godoc
// @Summary      Eliminar rol
// @Description  Elimina un rol por su ID (soft delete)
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Router       /roles/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}

	return gin.H{"deleted": true}, nil
}
