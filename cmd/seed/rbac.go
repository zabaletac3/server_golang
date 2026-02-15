package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/modules/roles"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// demoTenantID es el tenant de referencia para el seed de demostración.
// En producción cada tenant tendrá sus propios recursos, permisos y roles.
var demoTenantID = func() primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex("65f0000000000000000001aa")
	return id
}()

// ---------------------------------------------------------------------------
// Recursos de una veterinaria
// ---------------------------------------------------------------------------

type resourceSeed struct {
	name        string
	description string
}

var vetResources = []resourceSeed{
	{"dashboard", "Panel principal con resumen de actividad"},
	{"appointments", "Agenda y gestión de citas veterinarias"},
	{"patients", "Pacientes (mascotas) registradas en la clínica"},
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

// ---------------------------------------------------------------------------
// Permisos por rol — cada entrada es { recurso, acción }
// ---------------------------------------------------------------------------

type permEntry struct {
	resource string
	action   string
}

func allPermissions(resourceNames []string) []permEntry {
	allActions := []string{"get", "post", "put", "patch", "delete"}
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
	{"owners", "get"}, {"owners", "post"}, {"owners", "patch"},
	{"billing", "get"}, {"billing", "post"}, {"billing", "patch"},
	{"prescriptions", "get"},
}

var assistantPermissions = []permEntry{
	{"dashboard", "get"},
	{"appointments", "get"}, {"appointments", "patch"},
	{"patients", "get"},
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

// ---------------------------------------------------------------------------
// Definición de roles
// ---------------------------------------------------------------------------

type roleSeed struct {
	name        string
	description string
	perms       []permEntry
}

// ---------------------------------------------------------------------------
// SeedRBAC — método principal (idempotente)
// ---------------------------------------------------------------------------

func (s *SeedService) SeedRBAC(ctx context.Context) error {
	s.logger.Info("seeding RBAC: resources, permissions and roles...")

	// Idempotencia: si ya existen recursos, saltar todo el seed
	existing, _, err := s.resourceRepo.FindAll(ctx, pagination.Params{Skip: 0, Limit: 1})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		s.logger.Info("RBAC resources already exist, skipping seed")
		return nil
	}

	// 1. Crear recursos → mapa name → ObjectID
	resourceMap, err := s.seedResources(ctx)
	if err != nil {
		return err
	}

	// 2. Crear todos los permisos posibles (12 recursos × 5 acciones)
	//    y construir mapa "resource:action" → ObjectID
	permMap, err := s.seedPermissions(ctx, resourceMap)
	if err != nil {
		return err
	}

	// 3. Crear roles con sus permisos y recursos asignados
	resourceNames := make([]string, len(vetResources))
	for i, r := range vetResources {
		resourceNames[i] = r.name
	}

	roleDefs := []roleSeed{
		{
			name:        "admin",
			description: "Administrador de clínica: acceso total al sistema",
			perms:       allPermissions(resourceNames),
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

	return s.seedRoles(ctx, roleDefs, permMap, resourceMap)
}

// seedResources inserta todos los recursos y devuelve un mapa name → ObjectID
func (s *SeedService) seedResources(ctx context.Context) (map[string]primitive.ObjectID, error) {
	resourceMap := make(map[string]primitive.ObjectID, len(vetResources))

	for _, r := range vetResources {
		created, err := s.resourceRepo.Create(ctx, &resources.CreateResourceDTO{
			TenantId:    demoTenantID.Hex(),
			Name:        r.name,
			Description: r.description,
		})
		if err != nil {
			s.logger.Error("failed to create resource", "name", r.name, "error", err)
			return nil, err
		}
		resourceMap[r.name] = created.ID
		s.logger.Info("resource created", "name", r.name)
	}

	return resourceMap, nil
}

// seedPermissions crea todas las combinaciones (recurso × acción) y devuelve
// un mapa "resource:action" → ObjectID del permiso creado
func (s *SeedService) seedPermissions(ctx context.Context, resourceMap map[string]primitive.ObjectID) (map[string]primitive.ObjectID, error) {
	allActions := []string{"get", "post", "put", "patch", "delete"}
	permMap := make(map[string]primitive.ObjectID, len(vetResources)*len(allActions))

	for _, res := range vetResources {
		resourceID, ok := resourceMap[res.name]
		if !ok {
			continue
		}

		for _, action := range allActions {
			created, err := s.permissionRepo.Create(ctx, &permissions.CreatePermissionDTO{
				TenantId:   demoTenantID.Hex(),
				ResourceId: resourceID.Hex(),
				Action:     action,
			})
			if err != nil {
				s.logger.Error("failed to create permission", "resource", res.name, "action", action, "error", err)
				return nil, err
			}

			key := res.name + ":" + action
			permMap[key] = created.ID
		}
	}

	s.logger.Info("permissions created", "total", len(permMap))
	return permMap, nil
}

// seedRoles crea cada rol asignando los permisos y recursos que le corresponden
func (s *SeedService) seedRoles(ctx context.Context, roleDefs []roleSeed, permMap map[string]primitive.ObjectID, resourceMap map[string]primitive.ObjectID) error {
	for _, rd := range roleDefs {
		// Deduplicar permission IDs y resource IDs de este rol
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

		_, err := s.roleRepo.Create(ctx, &roles.CreateRoleDTO{
			TenantId:       demoTenantID.Hex(),
			Name:           rd.name,
			Description:    rd.description,
			PermissionsIds: permIDs,
			ResourcesIds:   resourceIDs,
		})
		if err != nil {
			s.logger.Error("failed to create role", "name", rd.name, "error", err)
			return err
		}

		s.logger.Info("role created",
			"name", rd.name,
			"permissions", len(permIDs),
			"resources", len(resourceIDs),
		)
	}

	return nil
}
