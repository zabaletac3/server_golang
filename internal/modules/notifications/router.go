package notifications

import (
	"github.com/eren_dev/go_server/internal/modules/owners"
	platformNotifications "github.com/eren_dev/go_server/internal/platform/notifications"
	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/httpx"
)

func newService(db *database.MongoDB, pushProvider platformNotifications.PushProvider) *Service {
	return NewService(
		NewRepository(db),
		NewStaffRepository(db),
		owners.NewRepository(db),
		pushProvider,
	)
}

func RegisterMobileRoutes(mobile *httpx.Router, db *database.MongoDB, pushProvider platformNotifications.PushProvider) {
	service := newService(db, pushProvider)
	handler := NewHandler(service)

	notifs := mobile.Group("/notifications")
	notifs.GET("", handler.GetAll)
	notifs.GET("/unread-count", handler.GetUnreadCount)
	notifs.PATCH("/read-all", handler.MarkAllAsRead)
	notifs.PATCH("/:id/read", handler.MarkAsRead)
}

func RegisterAdminRoutes(authPrivate *httpx.Router, db *database.MongoDB, pushProvider platformNotifications.PushProvider) {
	service := newService(db, pushProvider)
	handler := NewAdminHandler(service)

	notifs := authPrivate.Group("/notifications")
	notifs.GET("", handler.GetAll)
	notifs.GET("/unread-count", handler.GetUnreadCount)
	notifs.PATCH("/read-all", handler.MarkAllAsRead)
	notifs.PATCH("/:id/read", handler.MarkAsRead)
}
