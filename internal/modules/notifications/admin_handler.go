package notifications

import (
	"github.com/gin-gonic/gin"

	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
	sharedErrors "github.com/eren_dev/go_server/internal/shared/errors"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type AdminHandler struct {
	service *Service
}

func NewAdminHandler(service *Service) *AdminHandler {
	return &AdminHandler{service: service}
}

// GetAll returns the authenticated staff user's notifications (paginated, newest first).
//
//	@Summary		List my notifications (staff)
//	@Tags			admin/notifications
//	@Produce		json
//	@Param			skip	query		int	false	"Skip"
//	@Param			limit	query		int	false	"Limit"
//	@Success		200		{object}	PaginatedStaffNotificationsResponse
//	@Failure		401		{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/notifications [get]
func (h *AdminHandler) GetAll(c *gin.Context) (any, error) {
	userID := sharedAuth.GetUserID(c)
	if userID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	params := pagination.FromContext(c)
	return h.service.GetForUser(c.Request.Context(), userID, params)
}

// GetUnreadCount returns the number of unread notifications for the staff badge.
//
//	@Summary		Unread notification count (staff)
//	@Tags			admin/notifications
//	@Produce		json
//	@Success		200	{object}	UnreadCountResponse
//	@Failure		401	{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/notifications/unread-count [get]
func (h *AdminHandler) GetUnreadCount(c *gin.Context) (any, error) {
	userID := sharedAuth.GetUserID(c)
	if userID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	return h.service.GetUnreadCountStaff(c.Request.Context(), userID)
}

// MarkAsRead marks a single staff notification as read.
//
//	@Summary		Mark notification as read (staff)
//	@Tags			admin/notifications
//	@Produce		json
//	@Param			id	path		string	true	"Notification ID"
//	@Success		200	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/notifications/{id}/read [patch]
func (h *AdminHandler) MarkAsRead(c *gin.Context) (any, error) {
	userID := sharedAuth.GetUserID(c)
	if userID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	if err := h.service.MarkStaffAsRead(c.Request.Context(), userID, c.Param("id")); err != nil {
		return nil, err
	}
	return gin.H{"message": "notification marked as read"}, nil
}

// MarkAllAsRead marks all notifications as read for the authenticated staff user.
//
//	@Summary		Mark all notifications as read (staff)
//	@Tags			admin/notifications
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Security		Bearer
//	@Router			/api/notifications/read-all [patch]
func (h *AdminHandler) MarkAllAsRead(c *gin.Context) (any, error) {
	userID := sharedAuth.GetUserID(c)
	if userID == "" {
		return nil, sharedErrors.ErrUnauthorized
	}
	if err := h.service.MarkAllStaffAsRead(c.Request.Context(), userID); err != nil {
		return nil, err
	}
	return gin.H{"message": "all notifications marked as read"}, nil
}
