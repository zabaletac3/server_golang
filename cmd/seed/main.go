package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/database"
)

func main() {
	// Cargar variables de entorno
	_ = godotenv.Load(".env")

	// Cargar configuración
	cfg := config.Load()

	// Conectar a MongoDB
	db, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close(context.Background())

	// Crear logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Crear repositorios
	userRepo := users.NewRepository(db)
	planRepo := plans.NewPlanRepository(db)
	tenantRepo := tenant.NewTenantRepository(db)
	resourceRepo := resources.NewRepository(db)
	permissionRepo := permissions.NewRepository(db)
	roleRepo := roles.NewRepository(db)

	seedService := NewSeedService(db, userRepo, planRepo, tenantRepo, resourceRepo, permissionRepo, roleRepo, logger)

	// Ejecutar seeds
	ctx := context.Background()
	if err := seedService.RunSeeds(ctx); err != nil {
		log.Fatal("failed to run seeds:", err)
	}

	logger.Info("seeding completed!")
}
