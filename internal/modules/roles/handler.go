package roles

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary      Crear rol
// @Description  Crea un nuevo rol con permisos y recursos asignados
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        body  body      CreateRoleDTO  true  "Datos del rol"
// @Success      200   {object}  RoleResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      409   {object}  map[string]string "Nombre ya existe"
// @Router       /api/roles [post]
func (h *Handler) Create(c *gin.Context) (any, error) {
	var dto CreateRoleDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Create(c.Request.Context(), &dto)
}

// FindAll godoc
// @Summary      Listar roles
// @Description  Obtiene una lista paginada de roles con permisos y recursos poblados
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        skip   query     int  false  " "  default(0)
// @Param        limit  query     int  false  " "  default(10)
// @Success      200    {object}  PaginatedRolesResponse
// @Failure      401    {object}  map[string]string "No autorizado"
// @Router       /api/roles [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	return h.service.FindAll(c.Request.Context(), params)
}

// FindByID godoc
// @Summary      Obtener rol
// @Description  Obtiene un rol por su ID con permisos y recursos poblados
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  RoleResponse
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /api/roles/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Update godoc
// @Summary      Actualizar rol
// @Description  Actualiza un rol, sus permisos y/o recursos asignados
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      string         true  "Role ID"
// @Param        body  body      UpdateRoleDTO  true  "Datos a actualizar"
// @Success      200   {object}  RoleResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      404   {object}  map[string]string "No encontrado"
// @Router       /api/roles/{id} [patch]
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
// @Security     Bearer
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  map[string]bool
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /api/roles/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}
	return gin.H{"deleted": true}, nil
}
