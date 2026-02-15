package permissions

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
// @Summary      Crear permiso
// @Description  Crea un nuevo permiso (combinación recurso + acción)
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        body  body      CreatePermissionDTO  true  "Datos del permiso"
// @Success      200   {object}  PermissionResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      409   {object}  map[string]string "Permiso ya existe"
// @Router       /permissions [post]
func (h *Handler) Create(c *gin.Context) (any, error) {
	var dto CreatePermissionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Create(c.Request.Context(), &dto)
}

// FindAll godoc
// @Summary      Listar permisos
// @Description  Obtiene una lista paginada de permisos
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        skip   query     int  false  " "  default(0)
// @Param        limit  query     int  false  " "  default(10)
// @Success      200    {object}  PaginatedPermissionsResponse
// @Failure      401    {object}  map[string]string "No autorizado"
// @Router       /permissions [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	return h.service.FindAll(c.Request.Context(), params)
}

// FindByID godoc
// @Summary      Obtener permiso
// @Description  Obtiene un permiso por su ID con el recurso poblado
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  PermissionResponse
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /permissions/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Update godoc
// @Summary      Actualizar permiso
// @Description  Actualiza la acción de un permiso existente
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      string               true  "Permission ID"
// @Param        body  body      UpdatePermissionDTO  true  "Datos a actualizar"
// @Success      200   {object}  PermissionResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      404   {object}  map[string]string "No encontrado"
// @Router       /permissions/{id} [patch]
func (h *Handler) Update(c *gin.Context) (any, error) {
	id := c.Param("id")
	var dto UpdatePermissionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Update(c.Request.Context(), id, &dto)
}

// Delete godoc
// @Summary      Eliminar permiso
// @Description  Elimina un permiso por su ID (soft delete)
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  map[string]bool
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /permissions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}
	return gin.H{"deleted": true}, nil
}
