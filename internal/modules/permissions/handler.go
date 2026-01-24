package permissions

import (
	"github.com/gin-gonic/gin"

	"github.com/eren_dev/go_server/internal/shared/validation"
)

type Handler struct {
	service *PermissionService
}

func NewHandler(service *PermissionService) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary      Crear permiso
// @Description  Crea un nuevo permiso
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        body  body      CreatePermissionDTO  true  "Datos del permiso"
// @Success      200   {object}  PermissionResponse
// @Failure      400   {object}  validation.ValidationError
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
// @Description  Obtiene todos los permisos
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Success      200  {array}  PermissionResponse
// @Router       /permissions [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	return h.service.FindAll(c.Request.Context())
}

// FindByID godoc
// @Summary      Obtener permiso
// @Description  Obtiene un permiso por su ID
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  PermissionResponse
// @Failure      404  {object}  map[string]string
// @Router       /permissions/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Delete godoc
// @Summary      Eliminar permiso
// @Description  Elimina un permiso por su ID
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Router       /permissions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}

	return gin.H{"deleted": true}, nil
}

// GetOptions godoc
// @Summary      Opciones disponibles
// @Description  Obtiene recursos y acciones disponibles para crear permisos
// @Tags         permissions
// @Produce      json
// @Success      200  {object}  AvailableOptionsResponse
// @Router       /permissions/options [get]
func (h *Handler) GetOptions(c *gin.Context) (any, error) {
	return h.service.GetAvailableOptions(c.Request.Context()), nil
}
