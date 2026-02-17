package mobile_auth

import (
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
	"github.com/eren_dev/go_server/internal/shared/validation"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register creates a new owner account and returns JWT tokens.
//
//	@Summary		Register (owner)
//	@Tags			mobile/auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterDTO		true	"Registration data"
//	@Success		201		{object}	TokenResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		409		{object}	map[string]string
//	@Router			/mobile/auth/register [post]
func (h *Handler) Register(c *gin.Context) (any, error) {
	var dto RegisterDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Register(c.Request.Context(), &dto)
}

// Login authenticates an owner and returns JWT tokens.
//
//	@Summary		Login (owner)
//	@Tags			mobile/auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginDTO	true	"Credentials"
//	@Success		200		{object}	TokenResponse
//	@Failure		401		{object}	map[string]string
//	@Router			/mobile/auth/login [post]
func (h *Handler) Login(c *gin.Context) (any, error) {
	var dto LoginDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Login(c.Request.Context(), &dto)
}

// Refresh issues a new token pair from a valid refresh token.
//
//	@Summary		Refresh token (owner)
//	@Tags			mobile/auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RefreshDTO	true	"Refresh token"
//	@Success		200		{object}	TokenResponse
//	@Failure		401		{object}	map[string]string
//	@Router			/mobile/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) (any, error) {
	var dto RefreshDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}
	return h.service.Refresh(c.Request.Context(), dto.RefreshToken)
}

// Me returns the authenticated owner's basic info.
//
//	@Summary		Get current owner
//	@Tags			mobile/auth
//	@Produce		json
//	@Success		200	{object}	OwnerInfo
//	@Failure		401	{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/auth/me [get]
func (h *Handler) Me(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	return h.service.GetOwnerInfo(c.Request.Context(), ownerID)
}
