package main

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/modules/users"
)

type SeedService struct {
	userRepo       users.UserRepository
	planRepo       plans.PlanRepository
	resourceRepo   resources.ResourceRepository
	permissionRepo permissions.PermissionRepository
	roleRepo       roles.RoleRepository
	logger         *slog.Logger
}

func NewSeedService(
	userRepo users.UserRepository,
	planRepo plans.PlanRepository,
	resourceRepo resources.ResourceRepository,
	permissionRepo permissions.PermissionRepository,
	roleRepo roles.RoleRepository,
	logger *slog.Logger,
) *SeedService {
	return &SeedService{
		userRepo:       userRepo,
		planRepo:       planRepo,
		resourceRepo:   resourceRepo,
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		logger:         logger,
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

func (s *SeedService) SeedPlans(ctx context.Context) error {
	initialPlans := []plans.Plan{
		{
			Name:           "Básico",
			Description:    "Plan ideal para pequeños consultorios y veterinarios independientes.",
			MonthlyPrice:   0, // Gratis por ahora o bajo costo
			AnnualPrice:    0,
			Currency:       "COP",
			MaxUsers:       1,
			MaxBranches:    1,
			StorageLimitGB: 1,
			Features: []string{
				"Gestión de pacientes básica",
				"Historias clínicas ilimitadas",
				"Agenda de citas",
			},
			IsVisible: true,
		},
		{
			Name:           "Pro",
			Description:    "Para clínicas en crecimiento que necesitan más control y usuarios.",
			MonthlyPrice:   4900000, // $49.000 COP (en centavos) -> Ajustar según lógica de precios
			AnnualPrice:    49000000, // $490.000 COP
			Currency:       "COP",
			MaxUsers:       5,
			MaxBranches:    1,
			StorageLimitGB: 10,
			Features: []string{
				"Todo lo del plan Básico",
				"Múltiples usuarios",
				"Recordatorios por WhatsApp",
				"Facturación electrónica básica",
				"Reportes financieros",
			},
			IsVisible: true,
		},
		{
			Name:           "Empresarial",
			Description:    "Solución completa para hospitales veterinarios y cadenas.",
			MonthlyPrice:   14900000, // $149.000 COP
			AnnualPrice:    149000000, // $1.490.000 COP
			Currency:       "COP",
			MaxUsers:       20,
			MaxBranches:    3,
			StorageLimitGB: 50,
			Features: []string{
				"Todo lo del plan Pro",
				"Múltiples sedes",
				"Roles y permisos avanzados",
				"API de integración",
				"Soporte prioritario 24/7",
			},
			IsVisible: true,
		},
	}

	// Verificar si ya existen planes, si no, crearlos
	existingPlans, err := s.planRepo.FindAll(ctx)
	if err != nil {
		s.logger.Error("failed to list plans", "error", err)
		return err
	}

	if len(existingPlans) > 0 {
		s.logger.Info("plans already exist, skipping seed")
		return nil
	}

	for _, p := range initialPlans {
		p.ID = primitive.NewObjectID()
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()

		if err := s.planRepo.Create(ctx, &p); err != nil {
			s.logger.Error("failed to create plan", "name", p.Name, "error", err)
			return err
		}
		s.logger.Info("plan created successfully", "name", p.Name)
	}

	return nil
}

// RunSeeds ejecuta todos los seeds
func (s *SeedService) RunSeeds(ctx context.Context) error {
	s.logger.Info("starting database seeding...")

	if err := s.SeedSuperAdmins(ctx); err != nil {
		return err
	}

	if err := s.SeedPlans(ctx); err != nil {
		return err
	}

	if err := s.SeedRBAC(ctx); err != nil {
		return err
	}

	s.logger.Info("database seeding completed successfully")
	return nil
}
