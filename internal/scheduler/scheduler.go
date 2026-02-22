package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/eren_dev/go_server/internal/app/lifecycle"
	"github.com/eren_dev/go_server/internal/modules/appointments"
	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/shared/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Scheduler struct {
	appointmentRepo appointments.AppointmentRepository
	notificationSvc *notifications.Service
	interval        time.Duration
	logger          *slog.Logger
	stopCh          chan struct{}
}

func New(db *database.MongoDB, notificationSvc *notifications.Service, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		appointmentRepo: appointments.NewAppointmentRepository(db),
		notificationSvc: notificationSvc,
		interval:        15 * time.Minute, // check every 15 minutes
		logger:          logger,
		stopCh:          make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context, workers *lifecycle.Workers) {
	workers.Add(1)
	go func() {
		defer workers.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.logger.Info("appointment scheduler started", "interval", s.interval)

		for {
			select {
			case <-ticker.C:
				s.processReminders(ctx)
				s.processAutoCancellations(ctx)
			case <-s.stopCh:
				s.logger.Info("appointment scheduler stopped")
				return
			case <-ctx.Done():
				s.logger.Info("appointment scheduler context cancelled")
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) processReminders(ctx context.Context) {
	// Buscar citas próximas en las siguientes 24h que estén confirmed
	upcoming, err := s.appointmentRepo.FindUpcoming(ctx, primitive.NilObjectID, 24)
	if err != nil {
		s.logger.Error("failed to find upcoming appointments", "error", err)
		return
	}

	now := time.Now()
	for _, appt := range upcoming {
		if appt.Status != "confirmed" {
			continue
		}

		timeUntil := appt.ScheduledAt.Sub(now)

		// Recordatorio 24h (entre 23h30m y 24h30m)
		if timeUntil >= 23*time.Hour+30*time.Minute && timeUntil <= 24*time.Hour+30*time.Minute {
			s.sendReminder(ctx, &appt, "24 horas")
		}

		// Recordatorio 2h (entre 1h30m y 2h30m)
		if timeUntil >= 1*time.Hour+30*time.Minute && timeUntil <= 2*time.Hour+30*time.Minute {
			s.sendReminder(ctx, &appt, "2 horas")
		}
	}
}

func (s *Scheduler) sendReminder(ctx context.Context, appt *appointments.Appointment, timeframe string) {
	s.notificationSvc.Send(ctx, &notifications.SendDTO{
		OwnerID:  appt.OwnerID.Hex(),
		TenantID: appt.TenantID.Hex(),
		Type:     notifications.TypeAppointmentReminder,
		Title:    "Recordatorio de cita",
		Body:     fmt.Sprintf("Tu cita es en %s (%s)", timeframe, appt.ScheduledAt.Format("02/01/2006 15:04")),
		Data:     map[string]string{"appointment_id": appt.ID.Hex()},
		SendPush: true,
	})
}

func (s *Scheduler) processAutoCancellations(ctx context.Context) {
	cutoff := time.Now().Add(-24 * time.Hour)

	unconfirmed, err := s.appointmentRepo.FindUnconfirmedBefore(ctx, cutoff)
	if err != nil {
		s.logger.Error("failed to find unconfirmed appointments", "error", err)
		return
	}

	now := time.Now()
	updates := bson.M{
		"status":        "cancelled",
		"cancelled_at":  now,
		"cancel_reason": "Auto-cancelada: no confirmada en 24 horas",
		"updated_at":    now,
	}

	for _, appt := range unconfirmed {
		if err := s.appointmentRepo.Update(ctx, appt.ID, updates, appt.TenantID); err != nil {
			s.logger.Error("failed to auto-cancel appointment", "id", appt.ID.Hex(), "error", err)
			continue
		}

		s.notificationSvc.Send(ctx, &notifications.SendDTO{
			OwnerID:  appt.OwnerID.Hex(),
			TenantID: appt.TenantID.Hex(),
			Type:     notifications.TypeAppointmentCancelled,
			Title:    "Cita cancelada automáticamente",
			Body:     fmt.Sprintf("La cita del %s fue cancelada por no ser confirmada en 24 horas", appt.ScheduledAt.Format("02/01/2006 15:04")),
			Data:     map[string]string{"appointment_id": appt.ID.Hex()},
			SendPush: true,
		})

		s.logger.Info("auto-cancelled unconfirmed appointment", "id", appt.ID.Hex())
	}
}
