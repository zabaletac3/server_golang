package notifications

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/owners"
	"github.com/eren_dev/go_server/internal/platform/notifications"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type Service struct {
	repo         Repository
	staffRepo    StaffRepository
	ownerRepo    owners.OwnerRepository
	pushProvider notifications.PushProvider
}

func NewService(repo Repository, staffRepo StaffRepository, ownerRepo owners.OwnerRepository, pushProvider notifications.PushProvider) *Service {
	return &Service{
		repo:         repo,
		staffRepo:    staffRepo,
		ownerRepo:    ownerRepo,
		pushProvider: pushProvider,
	}
}

// Send persists the notification and optionally delivers a push via FCM.
// This is the single entry point for ALL other modules to trigger notifications.
func (s *Service) Send(ctx context.Context, dto *SendDTO) error {
	ownerID, err := primitive.ObjectIDFromHex(dto.OwnerID)
	if err != nil {
		return err
	}
	tenantID, err := primitive.ObjectIDFromHex(dto.TenantID)
	if err != nil {
		return err
	}

	notif := &Notification{
		ID:        primitive.NewObjectID(),
		OwnerID:   ownerID,
		TenantID:  tenantID,
		Type:      dto.Type,
		Title:     dto.Title,
		Body:      dto.Body,
		Data:      dto.Data,
		Read:      false,
		PushSent:  false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, notif); err != nil {
		return err
	}

	if dto.SendPush && s.pushProvider != nil && s.pushProvider.IsEnabled() {
		s.sendPushAsync(notif)
	}

	return nil
}

// sendPushAsync collects active push tokens for the owner and fires FCM in the background.
func (s *Service) sendPushAsync(notif *Notification) {
	go func() {
		ctx := context.Background()

		owner, err := s.ownerRepo.FindByID(ctx, notif.OwnerID.Hex())
		if err != nil {
			slog.Warn("push: owner not found", "owner_id", notif.OwnerID.Hex())
			return
		}

		tokens := make([]string, 0, len(owner.PushTokens))
		for _, pt := range owner.PushTokens {
			if pt.Active {
				tokens = append(tokens, pt.Token)
			}
		}

		if len(tokens) == 0 {
			return
		}

		payload := notifications.PushPayload{
			Title: notif.Title,
			Body:  notif.Body,
			Data:  notif.Data,
		}

		if err := s.pushProvider.Send(ctx, tokens, payload); err != nil {
			slog.Error("push: FCM send failed", "notification_id", notif.ID.Hex(), "error", err)
			return
		}

		if err := s.repo.MarkPushSent(ctx, notif.ID); err != nil {
			slog.Warn("push: failed to mark push_sent", "notification_id", notif.ID.Hex())
		}
	}()
}

func (s *Service) GetForOwner(ctx context.Context, ownerID string, params pagination.Params) (*PaginatedNotificationsResponse, error) {
	oid, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.FindByOwner(ctx, oid, params)
	if err != nil {
		return nil, err
	}

	data := make([]NotificationResponse, len(items))
	for i, n := range items {
		data[i] = toResponse(&n)
	}

	return &PaginatedNotificationsResponse{
		Data:       data,
		Pagination: pagination.NewPaginationInfo(params, total),
	}, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, ownerID string) (*UnreadCountResponse, error) {
	oid, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.CountUnread(ctx, oid)
	if err != nil {
		return nil, err
	}

	return &UnreadCountResponse{Count: count}, nil
}

func (s *Service) MarkAsRead(ctx context.Context, ownerID, notifID string) error {
	oid, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return err
	}
	nid, err := primitive.ObjectIDFromHex(notifID)
	if err != nil {
		return ErrInvalidNotificationID
	}
	return s.repo.MarkAsRead(ctx, oid, nid)
}

func (s *Service) MarkAllAsRead(ctx context.Context, ownerID string) error {
	oid, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return err
	}
	return s.repo.MarkAllAsRead(ctx, oid)
}

// --- Staff notification methods ---

func (s *Service) SendToStaff(ctx context.Context, dto *SendStaffDTO) error {
	userID, err := primitive.ObjectIDFromHex(dto.UserID)
	if err != nil {
		return err
	}
	tenantID, err := primitive.ObjectIDFromHex(dto.TenantID)
	if err != nil {
		return err
	}

	notif := &StaffNotification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TenantID:  tenantID,
		Type:      dto.Type,
		Title:     dto.Title,
		Body:      dto.Body,
		Data:      dto.Data,
		Read:      false,
		CreatedAt: time.Now(),
	}

	return s.staffRepo.CreateStaff(ctx, notif)
}

func (s *Service) GetForUser(ctx context.Context, userID string, params pagination.Params) (*PaginatedStaffNotificationsResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	items, total, err := s.staffRepo.FindByUser(ctx, uid, params)
	if err != nil {
		return nil, err
	}

	data := make([]StaffNotificationResponse, len(items))
	for i, n := range items {
		data[i] = toStaffResponse(&n)
	}

	return &PaginatedStaffNotificationsResponse{
		Data:       data,
		Pagination: pagination.NewPaginationInfo(params, total),
	}, nil
}

func (s *Service) GetUnreadCountStaff(ctx context.Context, userID string) (*UnreadCountResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	count, err := s.staffRepo.CountUnreadStaff(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &UnreadCountResponse{Count: count}, nil
}

func (s *Service) MarkStaffAsRead(ctx context.Context, userID, notifID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	nid, err := primitive.ObjectIDFromHex(notifID)
	if err != nil {
		return ErrInvalidNotificationID
	}
	return s.staffRepo.MarkStaffAsRead(ctx, uid, nid)
}

func (s *Service) MarkAllStaffAsRead(ctx context.Context, userID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return s.staffRepo.MarkAllStaffAsRead(ctx, uid)
}
