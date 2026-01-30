package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/plans"
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

	// Crear repositorios y servicios
	userRepo := users.NewRepository(db)
	planRepo := plans.NewPlanRepository(db)
	seedService := NewSeedService(userRepo, planRepo, logger)

	// Ejecutar seeds
	ctx := context.Background()
	if err := seedService.RunSeeds(ctx); err != nil {
		log.Fatal("failed to run seeds:", err)
	}

	logger.Info("seeding completed!")
}
