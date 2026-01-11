package lifecycle

import (
	"context"
	"time"

	"github.com/eren_dev/go_server/internal/modules/health"
	"github.com/eren_dev/go_server/internal/platform/logger"
)

type Shutdowner struct {
	server  HTTPServer
	workers *Workers
	timeout time.Duration
}

type HTTPServer interface {
	Shutdown(ctx context.Context) error
}

func NewShutdowner(server HTTPServer, workers *Workers, timeout time.Duration) *Shutdowner {
	return &Shutdowner{
		server:  server,
		workers: workers,
		timeout: timeout,
	}
}

func (s *Shutdowner) Shutdown(ctx context.Context) {
	log := logger.Default()

	log.Info(ctx, "shutdown_started")

	health.SetReady(false)
	log.Info(ctx, "readiness_disabled")

	shutdownCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Error(ctx, "http_shutdown_error", "error", err)
	} else {
		log.Info(ctx, "http_server_stopped")
	}

	done := make(chan struct{})
	go func() {
		s.workers.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info(ctx, "workers_stopped")
	case <-shutdownCtx.Done():
		log.Warn(ctx, "workers_shutdown_timeout")
	}

	log.Info(ctx, "shutdown_completed")
}
