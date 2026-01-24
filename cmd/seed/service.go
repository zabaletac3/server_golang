package main

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/eren_dev/go_server/internal/modules/users"
)

type SeedService struct {
	userRepo users.UserRepository
	logger   *slog.Logger
}

func NewSeedService(userRepo users.UserRepository, logger *slog.Logger) *SeedService {
	return &SeedService{
		userRepo: userRepo,
		logger:   logger,
	}
}

type SuperAdminData struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

func (s *SeedService) SeedSuperAdmins(ctx context.Context) error {
	superAdmins := []SuperAdminData{
		{
			Name:     "Carlos Rodríguez",
			Email:    "carlos.admin@vetsify.com",
			Phone:    "+57 300 123 4567",
			Password: "SuperAdmin123!",
		},
		{
			Name:     "María González",
			Email:    "maria.admin@vetsify.com",
			Phone:    "+57 301 234 5678",
			Password: "SuperAdmin123!",
		},
		{
			Name:     "Andrés Martínez",
			Email:    "andres.admin@vetsify.com",
			Phone:    "+57 302 345 6789",
			Password: "SuperAdmin123!",
		},
	}

	for _, admin := range superAdmins {
		// Verificar si el usuario ya existe
		existingUser, _ := s.userRepo.FindByEmail(ctx, admin.Email)
		if existingUser != nil {
			s.logger.Info("super admin already exists", "email", admin.Email)
			continue
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password", "email", admin.Email, "error", err)
			return err
		}

		// Crear usuario super admin
		now := time.Now()
		user := &users.User{
			ID:           primitive.NewObjectID(),
			Name:         admin.Name,
			Email:        admin.Email,
			Phone:        admin.Phone,
			Password:     string(hashedPassword),
			IsSuperAdmin: true,
			TenantIds:    []primitive.ObjectID{},
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		_, err = s.userRepo.CreateUser(ctx, user)
		if err != nil {
			s.logger.Error("failed to create super admin", "email", admin.Email, "error", err)
			return err
		}

		s.logger.Info("super admin created successfully", "email", admin.Email, "name", admin.Name)
	}

	return nil
}

// RunSeeds ejecuta todos los seeds
func (s *SeedService) RunSeeds(ctx context.Context) error {
	s.logger.Info("starting database seeding...")

	if err := s.SeedSuperAdmins(ctx); err != nil {
		return err
	}

	s.logger.Info("database seeding completed successfully")
	return nil
}
