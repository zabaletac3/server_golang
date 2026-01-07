package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eren_dev/go_server/internal/app"
	"github.com/eren_dev/go_server/internal/app/health"
	"github.com/eren_dev/go_server/internal/app/lifecycle"
	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/platform/logger"
)

func main() {
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		logger.Default().Error(
			context.Background(),
			"invalid_configuration",
			"error", err,
		)
		os.Exit(1)
	}

	log := logger.NewSlogLogger(cfg.Env)
	
	logger.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	workers := lifecycle.NewWorkers()

	server := app.NewServer(cfg)

	logger.Default().Info(context.Background(),
	"server_running",
	"port", cfg.Port,
	"env", cfg.Env,
)
	
	go func (){
		if err := server.Start(); err != nil {
			logger.Default().Error(context.Background(), "server_error", "error", err)
		}
	}()

	// Server listening then set to ready
	health.SetReady(true)

	<-ctx.Done()

	// Server stopped then set to not ready
	shutdowner := lifecycle.NewShutdowner(
		server,
		workers,
		10*time.Second,
	)
	
	shutdowner.Shutdown(context.Background())
}