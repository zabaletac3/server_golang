package users

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
// @Summary      Crear usuario
// @Description  Crea un nuevo usuario en el sistema
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        body  body      CreateUserDTO  true  "Datos del usuario"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      409   {object}  map[string]string "Email ya existe"
// @Router       /api/users [post]
func (h *Handler) Create(c *gin.Context) (any, error) {
	var dto CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Create(c.Request.Context(), &dto)
}

// FindAll godoc
// @Summary      Listar usuarios
// @Description  Obtiene una lista paginada de usuarios
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        skip   query     int  false  " "  default(0)
// @Param        limit  query     int  false  " "  default(10)
// @Success      200    {object}  PaginatedUsersResponse
// @Failure      401    {object}  map[string]string "No autorizado"
// @Router       /api/users [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	return h.service.FindAll(c.Request.Context(), params)
}

// FindByID godoc
// @Summary      Obtener usuario
// @Description  Obtiene un usuario por su ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /api/users/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	id := c.Param("id")
	return h.service.FindByID(c.Request.Context(), id)
}

// Update godoc
// @Summary      Actualizar usuario
// @Description  Actualiza un usuario existente
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      string         true  "User ID"
// @Param        body  body      UpdateUserDTO  true  "Datos a actualizar"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "No autorizado"
// @Failure      404   {object}  map[string]string "No encontrado"
// @Router       /api/users/{id} [patch]
func (h *Handler) Update(c *gin.Context) (any, error) {
	id := c.Param("id")

	var dto UpdateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Update(c.Request.Context(), id, &dto)
}

// Delete godoc
// @Summary      Eliminar usuario
// @Description  Elimina un usuario por su ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]bool
// @Failure      401  {object}  map[string]string "No autorizado"
// @Failure      404  {object}  map[string]string "No encontrado"
// @Router       /api/users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return nil, err
	}

	return gin.H{"deleted": true}, nil
}
