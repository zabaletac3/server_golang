package resources

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
// @Summary      Crear recurso
// @Description  Crea un nuevo recurso del sistema (debe coincidir con el segmento de ruta)
// @Tags         resources
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        body  body      CreateResourceDTO  true  "Datos del recurso"
// @Success      200   {object}  ResourceResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      409   {object}  map[string]string "Nombre ya existe"
// @Router       /api/resources [post]
func (h *Handler) Create(c *gin.Context) (any, error) {
	var dto CreateResourceDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Create(c.Request.Context(), &dto)
}

// FindAll godoc
// @Summary      Listar recursos
// @Description  Obtiene una lista paginada de recursos
// @Tags         resources
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        skip   query     int  false  " "  default(0)
// @Param        limit  query     int  false  " "  default(10)
// @Success      200    {object}  PaginatedResourcesResponse
// @Failure      401    {object}  map[string]string "No autorizado"
// @Router       /api/resources [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	return h.service.FindAll(c.Request.Context(), params)
}

// FindByID godoc
// @Summary      Obtener recurso
// @Description  Obtiene un recurso por su ID
// @Tags         resources
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Resource ID"
// @Success      200  {object}  ResourceResponse
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /api/resources/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Update godoc
// @Summary      Actualizar recurso
// @Description  Actualiza un recurso existente
// @Tags         resources
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      string             true  "Resource ID"
// @Param        body  body      UpdateResourceDTO  true  "Datos a actualizar"
// @Success      200   {object}  ResourceResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      404   {object}  map[string]string "No encontrado"
// @Router       /api/resources/{id} [patch]
func (h *Handler) Update(c *gin.Context) (any, error) {
	id := c.Param("id")
	var dto UpdateResourceDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Update(c.Request.Context(), id, &dto)
}

// Delete godoc
// @Summary      Eliminar recurso
// @Description  Elimina un recurso por su ID (soft delete)
// @Tags         resources
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "Resource ID"
// @Success      200  {object}  map[string]bool
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /api/resources/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}
	return gin.H{"deleted": true}, nil
}
