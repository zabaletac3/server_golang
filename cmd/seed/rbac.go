package main

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type resourceSeed struct {
	name        string
	description string
}

var vetResources = []resourceSeed{
	{"dashboard", "Panel principal con resumen de actividad"},
	{"appointments", "Agenda y gestión de citas veterinarias"},
	{"patients", "Pacientes (mascotas) registradas en la clínica"},
	{"species", "Especies animales (tags con deduplicación)"},
	{"owners", "Propietarios y contactos de las mascotas"},
	{"medical-records", "Historias clínicas y expedientes médicos"},
	{"vaccines", "Registro y control de vacunación"},
	{"prescriptions", "Recetas médicas y tratamientos"},
	{"inventory", "Inventario de medicamentos e insumos"},
	{"billing", "Facturación, pagos y cobros"},
	{"reports", "Reportes y estadísticas del negocio"},
	{"users", "Usuarios del sistema"},
	{"roles", "Roles y permisos de acceso"},
}

type permEntry struct {
	resource string
	action   string
}

var allActions = []string{"get", "post", "put", "patch", "delete"}

func allResourcePermissions(resourceNames []string) []permEntry {
	perms := make([]permEntry, 0, len(resourceNames)*len(allActions))
	for _, r := range resourceNames {
		for _, a := range allActions {
			perms = append(perms, permEntry{r, a})
		}
	}
	return perms
}

var veterinarianPermissions = []permEntry{
	{"dashboard", "get"},
	{"appointments", "get"}, {"appointments", "post"}, {"appointments", "patch"},
	{"patients", "get"}, {"patients", "post"}, {"patients", "put"}, {"patients", "patch"},
	{"species", "get"}, {"species", "post"},
	{"owners", "get"}, {"owners", "post"}, {"owners", "patch"},
	{"medical-records", "get"}, {"medical-records", "post"}, {"medical-records", "put"}, {"medical-records", "patch"}, {"medical-records", "delete"},
	{"vaccines", "get"}, {"vaccines", "post"}, {"vaccines", "patch"}, {"vaccines", "delete"},
	{"prescriptions", "get"}, {"prescriptions", "post"}, {"prescriptions", "patch"}, {"prescriptions", "delete"},
	{"inventory", "get"},
	{"billing", "get"},
}

var receptionistPermissions = []permEntry{
	{"dashboard", "get"},
	{"appointments", "get"}, {"appointments", "post"}, {"appointments", "patch"}, {"appointments", "delete"},
	{"patients", "get"}, {"patients", "post"}, {"patients", "patch"},
	{"species", "get"}, {"species", "post"},
	{"owners", "get"}, {"owners", "post"}, {"owners", "patch"},
	{"billing", "get"}, {"billing", "post"}, {"billing", "patch"},
	{"prescriptions", "get"},
}

var assistantPermissions = []permEntry{
	{"dashboard", "get"},
	{"appointments", "get"}, {"appointments", "patch"},
	{"patients", "get"},
	{"species", "get"},
	{"owners", "get"},
	{"medical-records", "get"},
	{"vaccines", "get"}, {"vaccines", "post"}, {"vaccines", "patch"},
	{"inventory", "get"}, {"inventory", "post"}, {"inventory", "patch"},
}

var accountantPermissions = []permEntry{
	{"dashboard", "get"},
	{"billing", "get"}, {"billing", "post"}, {"billing", "put"}, {"billing", "patch"},
	{"reports", "get"},
	{"inventory", "get"},
}

type roleSeed struct {
	name        string
	description string
	perms       []permEntry
}

// SeedRBAC crea recursos, permisos y roles para cada tenant.
// Retorna un mapa tenantIDHex -> roleName -> roleID.
// Si ya existen datos, carga los roles existentes sin recrear nada.
func (s *SeedService) SeedRBAC(ctx context.Context, tenantIDs []primitive.ObjectID) (map[string]map[string]primitive.ObjectID, error) {
	s.logger.Info("seeding RBAC: resources, permissions and roles...")

	rolesByTenant := make(map[string]map[string]primitive.ObjectID)

	// Idempotencia: si ya existen recursos, cargar roles y retornar sin crear nada
	existing, _, err := s.resourceRepo.FindAll(ctx, pagination.Params{Skip: 0, Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		s.logger.Info("RBAC already seeded, loading existing roles...")
		allRoles, _, err := s.roleRepo.FindAll(ctx, pagination.Params{Skip: 0, Limit: 200})
		if err != nil {
			return nil, err
		}
		for _, role := range allRoles {
			key := role.TenantId.Hex()
			if rolesByTenant[key] == nil {
				rolesByTenant[key] = make(map[string]primitive.ObjectID)
			}
			rolesByTenant[key][role.Name] = role.ID
		}
		return rolesByTenant, nil
	}

	resourceNames := make([]string, len(vetResources))
	for i, r := range vetResources {
		resourceNames[i] = r.name
	}

	roleDefs := []roleSeed{
		{
			name:        "admin",
			description: "Administrador: acceso total al sistema",
			perms:       allResourcePermissions(resourceNames),
		},
		{
			name:        "veterinarian",
			description: "Médico veterinario: gestión clínica completa",
			perms:       veterinarianPermissions,
		},
		{
			name:        "receptionist",
			description: "Recepcionista: agenda, clientes y facturación básica",
			perms:       receptionistPermissions,
		},
		{
			name:        "assistant",
			description: "Auxiliar veterinario: soporte en consulta e inventario",
			perms:       assistantPermissions,
		},
		{
			name:        "accountant",
			description: "Contador: facturación y reportes financieros",
			perms:       accountantPermissions,
		},
	}

	for _, tenantID := range tenantIDs {
		s.logger.Info("seeding RBAC for tenant", "tenant_id", tenantID.Hex())

		resourceMap, err := s.seedResources(ctx, tenantID)
		if err != nil {
			return nil, err
		}

		permMap, err := s.seedPermissions(ctx, tenantID, resourceMap)
		if err != nil {
			return nil, err
		}

		tenantRoleMap, err := s.seedRoles(ctx, tenantID, roleDefs, permMap, resourceMap)
		if err != nil {
			return nil, err
		}

		rolesByTenant[tenantID.Hex()] = tenantRoleMap
		s.logger.Info("RBAC seeded for tenant", "tenant_id", tenantID.Hex(), "roles", len(tenantRoleMap))
	}

	return rolesByTenant, nil
}

func (s *SeedService) seedResources(ctx context.Context, tenantID primitive.ObjectID) (map[string]primitive.ObjectID, error) {
	resourceMap := make(map[string]primitive.ObjectID, len(vetResources))

	for _, r := range vetResources {
		created, err := s.resourceRepo.Create(ctx, &resources.CreateResourceDTO{
			TenantId:    tenantID.Hex(),
			Name:        r.name,
			Description: r.description,
		})
		if err != nil {
			s.logger.Error("failed to create resource", "name", r.name, "error", err)
			return nil, err
		}
		resourceMap[r.name] = created.ID
	}

	s.logger.Info("resources created", "tenant_id", tenantID.Hex(), "count", len(resourceMap))
	return resourceMap, nil
}

func (s *SeedService) seedPermissions(ctx context.Context, tenantID primitive.ObjectID, resourceMap map[string]primitive.ObjectID) (map[string]primitive.ObjectID, error) {
	permMap := make(map[string]primitive.ObjectID, len(vetResources)*len(allActions))

	for _, res := range vetResources {
		resourceID, ok := resourceMap[res.name]
		if !ok {
			continue
		}

		for _, action := range allActions {
			created, err := s.permissionRepo.Create(ctx, &permissions.CreatePermissionDTO{
				TenantId:   tenantID.Hex(),
				ResourceId: resourceID.Hex(),
				Action:     action,
			})
			if err != nil {
				s.logger.Error("failed to create permission", "resource", res.name, "action", action, "error", err)
				return nil, err
			}
			permMap[res.name+":"+action] = created.ID
		}
	}

	s.logger.Info("permissions created", "tenant_id", tenantID.Hex(), "count", len(permMap))
	return permMap, nil
}

// seedRoles crea los roles de un tenant y retorna roleName -> roleID
func (s *SeedService) seedRoles(
	ctx context.Context,
	tenantID primitive.ObjectID,
	roleDefs []roleSeed,
	permMap map[string]primitive.ObjectID,
	resourceMap map[string]primitive.ObjectID,
) (map[string]primitive.ObjectID, error) {
	roleMap := make(map[string]primitive.ObjectID, len(roleDefs))

	for _, rd := range roleDefs {
		permIDSet := make(map[primitive.ObjectID]struct{})
		resourceIDSet := make(map[primitive.ObjectID]struct{})

		for _, pe := range rd.perms {
			if pid, ok := permMap[pe.resource+":"+pe.action]; ok {
				permIDSet[pid] = struct{}{}
			}
			if rid, ok := resourceMap[pe.resource]; ok {
				resourceIDSet[rid] = struct{}{}
			}
		}

		permIDs := make([]string, 0, len(permIDSet))
		for id := range permIDSet {
			permIDs = append(permIDs, id.Hex())
		}

		resourceIDs := make([]string, 0, len(resourceIDSet))
		for id := range resourceIDSet {
			resourceIDs = append(resourceIDs, id.Hex())
		}

		created, err := s.roleRepo.Create(ctx, &roles.CreateRoleDTO{
			TenantId:       tenantID.Hex(),
			Name:           rd.name,
			Description:    rd.description,
			PermissionsIds: permIDs,
			ResourcesIds:   resourceIDs,
		})
		if err != nil {
			if !errors.Is(err, roles.ErrRoleNameExists) {
				s.logger.Error("failed to create role", "name", rd.name, "tenant_id", tenantID.Hex(), "error", err)
				return nil, err
			}
			// Role already exists — find it and reuse its ID
			allRoles, _, ferr := s.roleRepo.FindAll(ctx, pagination.Params{Skip: 0, Limit: 200})
			if ferr != nil {
				return nil, ferr
			}
			var found *roles.Role
			for _, r := range allRoles {
				if r.TenantId == tenantID && r.Name == rd.name {
					found = r
					break
				}
			}
			if found == nil {
				return nil, fmt.Errorf("role %q not found for tenant %s after conflict", rd.name, tenantID.Hex())
			}
			roleMap[rd.name] = found.ID
			s.logger.Info("role already exists, reusing", "name", rd.name, "tenant_id", tenantID.Hex())
			continue
		}

		roleMap[rd.name] = created.ID
		s.logger.Info("role created", "name", rd.name, "tenant_id", tenantID.Hex(), "permissions", len(permIDs))
	}

	return roleMap, nil
}
