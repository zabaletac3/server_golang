package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/appointments"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/platform/logger"
	"github.com/eren_dev/go_server/internal/shared/database"
)

// Migration script for creating MongoDB indexes
// Usage: go run cmd/migrate-indexes/main.go
func main() {
	_ = godotenv.Load(".env")

	cfg := config.Load()
	log := logger.NewSlogLogger(cfg.Env)
	logger.SetDefault(log)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.NewProvider(cfg)
	if err != nil {
		logger.Default().Error(ctx, "database_connection_failed", "error", err)
		os.Exit(1)
	}
	defer db.Close(ctx)

	logger.Default().Info(ctx, "database_connected", "database", cfg.MongoDatabase)

	success := true

	// Create tenant indexes
	logger.Default().Info(ctx, "creating_tenant_indexes")
	if err := tenant.EnsureIndexes(ctx, db); err != nil {
		logger.Default().Error(ctx, "tenant_indexes_creation_failed", "error", err)
		success = false
	} else {
		logger.Default().Info(ctx, "tenant_indexes_created")
	}

	// Create appointments indexes
	logger.Default().Info(ctx, "creating_appointments_indexes")
	if err := appointments.EnsureIndexes(ctx, db); err != nil {
		logger.Default().Error(ctx, "appointments_indexes_creation_failed", "error", err)
		success = false
	} else {
		logger.Default().Info(ctx, "appointments_indexes_created")
	}

	fmt.Println()
	if success {
		logger.Default().Info(ctx, "migration_completed_successfully")
		fmt.Println("✅ Index migration completed successfully!")
		os.Exit(0)
	} else {
		logger.Default().Error(ctx, "migration_completed_with_errors", "error", "some indexes failed to create")
		fmt.Println("❌ Index migration completed with errors. Check logs for details.")
		os.Exit(1)
	}
}
