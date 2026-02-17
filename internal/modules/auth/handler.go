package auth

import (
	"github.com/gin-gonic/gin"

	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register godoc
// @Summary      Registrar usuario
// @Description  Crea una nueva cuenta de usuario
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterDTO  true  "Datos de registro"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      409   {object}  map[string]string "Email ya existe"
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) (any, error) {
	var dto RegisterDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Register(c.Request.Context(), &dto)
}

// Login godoc
// @Summary      Iniciar sesión
// @Description  Autentica un usuario y retorna tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginDTO  true  "Credenciales"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "Credenciales inválidas"
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) (any, error) {
	var dto LoginDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Login(c.Request.Context(), &dto)
}

// Refresh godoc
// @Summary      Renovar token
// @Description  Genera un nuevo access token usando el refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RefreshDTO  true  "Refresh token"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  validation.ValidationError
// @Failure      401   {object}  map[string]string "Token inválido"
// @Router       /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) (any, error) {
	var dto RefreshDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.Refresh(c.Request.Context(), dto.RefreshToken)
}

// Me godoc
// @Summary      Usuario actual
// @Description  Obtiene información del usuario autenticado
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  UserInfo
// @Failure      401  {object}  map[string]string "No autorizado"
// @Router       /auth/me [get]
func (h *Handler) Me(c *gin.Context) (any, error) {
	userID := sharedAuth.GetUserID(c)
	if userID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	return h.service.GetUserInfo(c.Request.Context(), userID)
}
