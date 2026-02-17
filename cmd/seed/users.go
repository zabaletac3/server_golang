package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/eren_dev/go_server/internal/modules/users"
)

type userSeed struct {
	id       primitive.ObjectID
	tenantID primitive.ObjectID
	name     string
	email    string
	phone    string
	password string
	roleName string // clave para buscar en el roleMap del tenant
}

// Tenant 1 - Clínica Veterinaria Vida Animal (Medellín)
var tenant1Users = []userSeed{
	{
		id:       owner1ID,
		tenantID: tenant1ID,
		name:     "Dr. Santiago Herrera",
		email:    "santiago.herrera@vidaanimal.com.co",
		phone:    "+57 300 111 0001",
		password: "Admin123!",
		roleName: "admin",
	},
	{
		id:       staffT1VetID,
		tenantID: tenant1ID,
		name:     "Dra. Laura Quintero",
		email:    "laura.quintero@vidaanimal.com.co",
		phone:    "+57 300 111 0002",
		password: "Vet123!",
		roleName: "veterinarian",
	},
	{
		id:       staffT1ReceptionistID,
		tenantID: tenant1ID,
		name:     "Camila Torres",
		email:    "camila.torres@vidaanimal.com.co",
		phone:    "+57 300 111 0003",
		password: "Recep123!",
		roleName: "receptionist",
	},
	{
		id:       staffT1AssistantID,
		tenantID: tenant1ID,
		name:     "Miguel Ángel López",
		email:    "miguel.lopez@vidaanimal.com.co",
		phone:    "+57 300 111 0004",
		password: "Asist123!",
		roleName: "assistant",
	},
	{
		id:       staffT1AccountantID,
		tenantID: tenant1ID,
		name:     "Roberto Castillo",
		email:    "roberto.castillo@vidaanimal.com.co",
		phone:    "+57 300 111 0005",
		password: "Conta123!",
		roleName: "accountant",
	},
}

// Tenant 2 - Centro Veterinario Patitas Felices (Bogotá)
var tenant2Users = []userSeed{
	{
		id:       owner2ID,
		tenantID: tenant2ID,
		name:     "Dra. Valentina Morales",
		email:    "valentina.morales@patitasfelices.com.co",
		phone:    "+57 310 222 0001",
		password: "Admin123!",
		roleName: "admin",
	},
	{
		id:       staffT2VetID,
		tenantID: tenant2ID,
		name:     "Dr. Andrés Ruiz",
		email:    "andres.ruiz@patitasfelices.com.co",
		phone:    "+57 310 222 0002",
		password: "Vet123!",
		roleName: "veterinarian",
	},
	{
		id:       staffT2ReceptionistID,
		tenantID: tenant2ID,
		name:     "Juliana Ramírez",
		email:    "juliana.ramirez@patitasfelices.com.co",
		phone:    "+57 310 222 0003",
		password: "Recep123!",
		roleName: "receptionist",
	},
	{
		id:       staffT2AssistantID,
		tenantID: tenant2ID,
		name:     "Felipe García",
		email:    "felipe.garcia@patitasfelices.com.co",
		phone:    "+57 310 222 0004",
		password: "Asist123!",
		roleName: "assistant",
	},
}

// SeedTenantUsers crea los usuarios de cada clínica con sus roles asignados.
// rolesByTenant: tenantIDHex -> roleName -> roleID
func (s *SeedService) SeedTenantUsers(ctx context.Context, rolesByTenant map[string]map[string]primitive.ObjectID) error {
	s.logger.Info("seeding tenant users...")

	allUserSeeds := append(tenant1Users, tenant2Users...)

	for _, us := range allUserSeeds {
		// Idempotencia: verificar si el usuario ya existe
		existing, _ := s.userRepo.FindByEmail(ctx, us.email)
		if existing != nil {
			s.logger.Info("user already exists, skipping", "email", us.email)
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(us.password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password", "email", us.email, "error", err)
			return err
		}

		// Obtener roleID del roleMap del tenant correspondiente
		roleMap := rolesByTenant[us.tenantID.Hex()]
		roleIDs := []primitive.ObjectID{}
		if roleMap != nil {
			if roleID, ok := roleMap[us.roleName]; ok {
				roleIDs = []primitive.ObjectID{roleID}
			}
		}

		now := time.Now()
		user := &users.User{
			ID:           us.id,
			Name:         us.name,
			Email:        us.email,
			Phone:        us.phone,
			Password:     string(hashedPassword),
			TenantIds:    []primitive.ObjectID{us.tenantID},
			RoleIds:      roleIDs,
			IsSuperAdmin: false,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		_, err = s.userRepo.CreateUser(ctx, user)
		if err != nil {
			s.logger.Error("failed to create user", "email", us.email, "error", err)
			return err
		}

		s.logger.Info("user created",
			"name", us.name,
			"email", us.email,
			"role", us.roleName,
			"tenant_id", us.tenantID.Hex(),
		)
	}

	return nil
}
