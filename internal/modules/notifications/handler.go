package notifications

import (
	"github.com/gin-gonic/gin"

	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetAll returns the owner's notifications (paginated, newest first).
//
//	@Summary		List my notifications
//	@Tags			mobile/notifications
//	@Produce		json
//	@Param			skip	query		int	false	"Skip"
//	@Param			limit	query		int	false	"Limit"
//	@Success		200		{object}	PaginatedNotificationsResponse
//	@Failure		401		{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/notifications [get]
func (h *Handler) GetAll(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	params := pagination.FromContext(c)
	return h.service.GetForOwner(c.Request.Context(), ownerID, params)
}

// GetUnreadCount returns the number of unread notifications (for badge).
//
//	@Summary		Unread notification count
//	@Tags			mobile/notifications
//	@Produce		json
//	@Success		200	{object}	UnreadCountResponse
//	@Failure		401	{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/notifications/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	return h.service.GetUnreadCount(c.Request.Context(), ownerID)
}

// MarkAsRead marks a single notification as read.
//
//	@Summary		Mark notification as read
//	@Tags			mobile/notifications
//	@Produce		json
//	@Param			id	path		string	true	"Notification ID"
//	@Success		200	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/notifications/{id}/read [patch]
func (h *Handler) MarkAsRead(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	if err := h.service.MarkAsRead(c.Request.Context(), ownerID, c.Param("id")); err != nil {
		return nil, err
	}
	return gin.H{"message": "notification marked as read"}, nil
}

// MarkAllAsRead marks all notifications as read for the owner.
//
//	@Summary		Mark all notifications as read
//	@Tags			mobile/notifications
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Security		Bearer
//	@Router			/mobile/notifications/read-all [patch]
func (h *Handler) MarkAllAsRead(c *gin.Context) (any, error) {
	ownerID := sharedAuth.GetUserID(c)
	if ownerID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	if err := h.service.MarkAllAsRead(c.Request.Context(), ownerID); err != nil {
		return nil, err
	}
	return gin.H{"message": "all notifications marked as read"}, nil
}
