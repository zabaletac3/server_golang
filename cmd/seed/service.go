package main

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/modules/tenant"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/database"
)

type SeedService struct {
	db             *database.MongoDB
	userRepo       users.UserRepository
	planRepo       plans.PlanRepository
	tenantRepo     tenant.TenantRepository
	resourceRepo   resources.ResourceRepository
	permissionRepo permissions.PermissionRepository
	roleRepo       roles.RoleRepository
	logger         *slog.Logger
}

func NewSeedService(
	db *database.MongoDB,
	userRepo users.UserRepository,
	planRepo plans.PlanRepository,
	tenantRepo tenant.TenantRepository,
	resourceRepo resources.ResourceRepository,
	permissionRepo permissions.PermissionRepository,
	roleRepo roles.RoleRepository,
	logger *slog.Logger,
) *SeedService {
	return &SeedService{
		db:             db,
		userRepo:       userRepo,
		planRepo:       planRepo,
		tenantRepo:     tenantRepo,
		resourceRepo:   resourceRepo,
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		logger:         logger,
	}
}

// runMigrations elimina índices huérfanos y crea los correctos
func (s *SeedService) runMigrations(ctx context.Context) error {
	// 1. Eliminar índices obsoletos (globales o sin partial filter)
	type dropMigration struct {
		collection string
		index      string
	}
	drops := []dropMigration{
		{"plans", "slug_1"},
		{"roles", "name_1"},
		{"roles", "tenant_id_1_name_1"},      // drop non-partial version if it exists
		{"resources", "name_1"},
		{"resources", "tenant_id_1_name_1"},  // drop non-partial version if it exists
	}
	for _, m := range drops {
		col := s.db.Collection(m.collection)
		if _, err := col.Indexes().DropOne(ctx, m.index); err != nil {
			s.logger.Info("index not found or already dropped", "collection", m.collection, "index", m.index)
		} else {
			s.logger.Info("dropped orphaned index", "collection", m.collection, "index", m.index)
		}
	}

	// 2. Crear índices compuestos
	type createMigration struct {
		collection string
		keys       bson.D
		name       string
		unique     bool
		partial    bson.D // nil = no partial filter
	}
	creates := []createMigration{
		// Roles: unique active name per tenant
		{
			collection: "roles",
			keys:       bson.D{{Key: "tenant_id", Value: 1}, {Key: "name", Value: 1}},
			name:       "tenant_id_1_name_1_active",
			unique:     true,
			partial:    bson.D{{Key: "deleted_at", Value: bson.D{{Key: "$eq", Value: nil}}}},
		},
		// Resources: unique active name per tenant
		{
			collection: "resources",
			keys:       bson.D{{Key: "tenant_id", Value: 1}, {Key: "name", Value: 1}},
			name:       "tenant_id_1_name_1_active",
			unique:     true,
			partial:    bson.D{{Key: "deleted_at", Value: bson.D{{Key: "$eq", Value: nil}}}},
		},
		// Patients: query by tenant + soft delete
		{
			collection: "patients",
			keys:       bson.D{{Key: "tenant_id", Value: 1}, {Key: "deleted_at", Value: 1}},
			name:       "patients_tenant_deleted",
		},
		// Patients: query by tenant + owner
		{
			collection: "patients",
			keys:       bson.D{{Key: "tenant_id", Value: 1}, {Key: "owner_id", Value: 1}},
			name:       "patients_tenant_owner",
		},
		// Patients: unique microchip per tenant (only when microchip exists and is non-empty)
		{
			collection: "patients",
			keys:       bson.D{{Key: "tenant_id", Value: 1}, {Key: "microchip", Value: 1}},
			name:       "patients_tenant_microchip_unique",
			unique:     true,
			partial: bson.D{
				{Key: "microchip", Value: bson.D{{Key: "$exists", Value: true}, {Key: "$ne", Value: ""}}},
			},
		},
		// Species: unique normalized name per tenant (active only)
		{
			collection: "species",
			keys:       bson.D{{Key: "tenant_id", Value: 1}, {Key: "normalized_name", Value: 1}},
			name:       "species_tenant_normalized_name_active",
			unique:     true,
			partial:    bson.D{{Key: "deleted_at", Value: bson.D{{Key: "$eq", Value: nil}}}},
		},
	}
	for _, m := range creates {
		col := s.db.Collection(m.collection)
		opts := options.Index().SetName(m.name)
		if m.unique {
			opts.SetUnique(true)
		}
		if m.partial != nil {
			opts.SetPartialFilterExpression(m.partial)
		}
		_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    m.keys,
			Options: opts,
		})
		if err != nil {
			s.logger.Info("index already exists or could not be created", "collection", m.collection, "index", m.name)
		} else {
			s.logger.Info("created index", "collection", m.collection, "index", m.name)
		}
	}

	return nil
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
		existingUser, _ := s.userRepo.FindByEmail(ctx, admin.Email)
		if existingUser != nil {
			s.logger.Info("super admin already exists", "email", admin.Email)
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password", "email", admin.Email, "error", err)
			return err
		}

		now := time.Now()
		user := &users.User{
			ID:           primitive.NewObjectID(),
			Name:         admin.Name,
			Email:        admin.Email,
			Phone:        admin.Phone,
			Password:     string(hashedPassword),
			IsSuperAdmin: true,
			TenantIds:    []primitive.ObjectID{},
			RoleIds:      []primitive.ObjectID{},
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		_, err = s.userRepo.CreateUser(ctx, user)
		if err != nil {
			s.logger.Error("failed to create super admin", "email", admin.Email, "error", err)
			return err
		}

		s.logger.Info("super admin created", "email", admin.Email, "name", admin.Name)
	}

	return nil
}

func (s *SeedService) SeedPlans(ctx context.Context) error {
	initialPlans := []plans.Plan{
		{
			Name:           "Básico",
			Description:    "Plan ideal para pequeños consultorios y veterinarios independientes.",
			MonthlyPrice:   0,
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
			MonthlyPrice:   4900000,
			AnnualPrice:    49000000,
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
			MonthlyPrice:   14900000,
			AnnualPrice:    149000000,
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
		s.logger.Info("plan created", "name", p.Name)
	}

	return nil
}

// getPlanMap retorna un mapa nombre -> ObjectID de los planes existentes
func (s *SeedService) getPlanMap(ctx context.Context) (map[string]primitive.ObjectID, error) {
	allPlans, err := s.planRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	planMap := make(map[string]primitive.ObjectID, len(allPlans))
	for _, p := range allPlans {
		planMap[p.Name] = p.ID
	}
	return planMap, nil
}

func (s *SeedService) RunSeeds(ctx context.Context) error {
	s.logger.Info("starting database seeding...")

	// 0. Migraciones previas (limpiar índices huérfanos)
	if err := s.runMigrations(ctx); err != nil {
		return err
	}

	// 1. Super admins de plataforma
	if err := s.SeedSuperAdmins(ctx); err != nil {
		return err
	}

	// 2. Planes de suscripción
	if err := s.SeedPlans(ctx); err != nil {
		return err
	}

	// 3. RBAC por tenant (recursos, permisos y roles)
	tenantIDs := []primitive.ObjectID{tenant1ID, tenant2ID}
	rolesByTenant, err := s.SeedRBAC(ctx, tenantIDs)
	if err != nil {
		return err
	}

	// 4. Tenants (clínicas)
	planMap, err := s.getPlanMap(ctx)
	if err != nil {
		return err
	}
	if err := s.SeedTenants(ctx, planMap); err != nil {
		return err
	}

	// 5. Usuarios por tenant con sus roles asignados
	if err := s.SeedTenantUsers(ctx, rolesByTenant); err != nil {
		return err
	}

	s.logger.Info("database seeding completed successfully")
	return nil
}
