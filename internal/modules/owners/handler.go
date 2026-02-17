package owners

import (
	"net/http"

	"github.com/gin-gonic/gin"

	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe returns the authenticated owner's profile.
//
//	@Summary		Get my profile
//	@Tags			mobile/owners
//	@Produce		json
//	@Success		200	{object}	OwnerResponse
//	@Failure		401	{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/owners/me [get]
func (h *Handler) GetMe(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	return h.service.GetMe(c.Request.Context(), ownerID)
}

// UpdateMe updates the authenticated owner's profile.
//
//	@Summary		Update my profile
//	@Tags			mobile/owners
//	@Accept			json
//	@Produce		json
//	@Param			body	body		UpdateOwnerDTO	true	"Profile data"
//	@Success		200		{object}	OwnerResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/owners/me [patch]
func (h *Handler) UpdateMe(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}

	var dto UpdateOwnerDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	return h.service.UpdateMe(c.Request.Context(), ownerID, &dto)
}

// AddPushToken registers a push notification token for the owner.
//
//	@Summary		Register push token
//	@Tags			mobile/owners
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterPushTokenDTO	true	"Push token"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/owners/me/push-tokens [post]
func (h *Handler) AddPushToken(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}

	var dto RegisterPushTokenDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	if err := h.service.AddPushToken(c.Request.Context(), ownerID, &dto); err != nil {
		return nil, err
	}

	return gin.H{"message": "push token registered"}, nil
}

// RemovePushToken removes a push notification token.
//
//	@Summary		Remove push token
//	@Tags			mobile/owners
//	@Produce		json
//	@Param			token	path		string	true	"Push token value"
//	@Success		200		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/owners/me/push-tokens/{token} [delete]
func (h *Handler) RemovePushToken(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return nil, nil
	}

	if err := h.service.RemovePushToken(c.Request.Context(), ownerID, token); err != nil {
		return nil, err
	}

	return gin.H{"message": "push token removed"}, nil
}

// FindAll returns a paginated list of owners (admin panel use).
//
//	@Summary		List owners
//	@Tags			owners
//	@Produce		json
//	@Param			skip	query		int	false	"Skip"
//	@Param			limit	query		int	false	"Limit"
//	@Success		200		{object}	PaginatedOwnersResponse
//	@Failure		401		{object}	map[string]string
//	@Security		Bearer
//	@Router			/owners [get]
func (h *Handler) FindAll(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	return h.service.FindAll(c.Request.Context(), params)
}

// FindByID returns a single owner by ID (admin panel use).
//
//	@Summary		Get owner by ID
//	@Tags			owners
//	@Produce		json
//	@Param			id	path		string	true	"Owner ID"
//	@Success		200	{object}	OwnerResponse
//	@Failure		404	{object}	map[string]string
//	@Security		Bearer
//	@Router			/owners/{id} [get]
func (h *Handler) FindByID(c *gin.Context) (any, error) {
	return h.service.FindByID(c.Request.Context(), c.Param("id"))
}

// Delete soft-deletes an owner (admin panel use).
//
//	@Summary		Delete owner
//	@Tags			owners
//	@Produce		json
//	@Param			id	path		string	true	"Owner ID"
//	@Success		200	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		Bearer
//	@Router			/owners/{id} [delete]
func (h *Handler) Delete(c *gin.Context) (any, error) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		return nil, err
	}
	return gin.H{"message": "owner deleted"}, nil
}
