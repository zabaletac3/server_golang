package roles

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/permissions"
	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type Service struct {
	repo           RoleRepository
	permissionRepo permissions.PermissionRepository
	resourceRepo   resources.ResourceRepository
}

func NewService(repo RoleRepository, permissionRepo permissions.PermissionRepository, resourceRepo resources.ResourceRepository) *Service {
	return &Service{
		repo:           repo,
		permissionRepo: permissionRepo,
		resourceRepo:   resourceRepo,
	}
}

func (s *Service) Create(ctx context.Context, dto *CreateRoleDTO) (*RoleResponse, error) {
	role, err := s.repo.Create(ctx, dto)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, role)
}

func (s *Service) FindAll(ctx context.Context, params pagination.Params) (*PaginatedRolesResponse, error) {
	roleList, total, err := s.repo.FindAll(ctx, params)
	if err != nil {
		return nil, err
	}

	responses, err := s.toResponseList(ctx, roleList)
	if err != nil {
		return nil, err
	}

	paginationInfo := pagination.NewPaginationInfo(params, total)
	return &PaginatedRolesResponse{
		Data:       responses,
		Pagination: &paginationInfo,
	}, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, role)
}

func (s *Service) Update(ctx context.Context, id string, dto *UpdateRoleDTO) (*RoleResponse, error) {
	role, err := s.repo.Update(ctx, id, dto)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, role)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// toResponse puebla permisos y recursos de un rol con queries batch
func (s *Service) toResponse(ctx context.Context, role *Role) (*RoleResponse, error) {
	permRefs, resourceMap, err := s.fetchPopulated(ctx, role.PermissionsIds, role.ResourcesIds)
	if err != nil {
		return nil, err
	}

	resourceRefs := make([]ResourceRef, 0, len(role.ResourcesIds))
	for _, rid := range role.ResourcesIds {
		if ref, ok := resourceMap[rid.Hex()]; ok {
			resourceRefs = append(resourceRefs, ref)
		}
	}

	return &RoleResponse{
		ID:          role.ID.Hex(),
		TenantId:    role.TenantId.Hex(),
		Name:        role.Name,
		Description: role.Description,
		Permissions: permRefs,
		Resources:   resourceRefs,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

// toResponseList puebla una lista de roles usando queries batch
func (s *Service) toResponseList(ctx context.Context, roleList []*Role) ([]*RoleResponse, error) {
	if len(roleList) == 0 {
		return []*RoleResponse{}, nil
	}

	// Recolectar todos los IDs únicos de permisos y recursos
	allPermIDs := collectUniqueIDs(func() [][]primitive.ObjectID {
		ids := make([][]primitive.ObjectID, len(roleList))
		for i, r := range roleList {
			ids[i] = r.PermissionsIds
		}
		return ids
	}())

	allResourceIDs := collectUniqueIDs(func() [][]primitive.ObjectID {
		ids := make([][]primitive.ObjectID, len(roleList))
		for i, r := range roleList {
			ids[i] = r.ResourcesIds
		}
		return ids
	}())

	// Fetch permisos en batch
	permList, err := s.permissionRepo.FindByIDs(ctx, allPermIDs)
	if err != nil {
		return nil, err
	}

	// Recolectar resource IDs desde permisos también
	for _, p := range permList {
		allResourceIDs = appendIfMissing(allResourceIDs, p.ResourceId)
	}

	// Fetch recursos en batch
	resourceList, err := s.resourceRepo.FindByIDs(ctx, allResourceIDs)
	if err != nil {
		return nil, err
	}

	// Indexar recursos
	resourceMap := make(map[string]ResourceRef, len(resourceList))
	for _, r := range resourceList {
		resourceMap[r.ID.Hex()] = ResourceRef{ID: r.ID.Hex(), Name: r.Name, Description: r.Description}
	}

	// Indexar permisos
	permMap := make(map[string]PermissionRef, len(permList))
	for _, p := range permList {
		permMap[p.ID.Hex()] = PermissionRef{
			ID:       p.ID.Hex(),
			Action:   string(p.Action),
			Resource: resourceMap[p.ResourceId.Hex()],
		}
	}

	result := make([]*RoleResponse, len(roleList))
	for i, role := range roleList {
		permRefs := make([]PermissionRef, 0, len(role.PermissionsIds))
		for _, pid := range role.PermissionsIds {
			if ref, ok := permMap[pid.Hex()]; ok {
				permRefs = append(permRefs, ref)
			}
		}

		resourceRefs := make([]ResourceRef, 0, len(role.ResourcesIds))
		for _, rid := range role.ResourcesIds {
			if ref, ok := resourceMap[rid.Hex()]; ok {
				resourceRefs = append(resourceRefs, ref)
			}
		}

		result[i] = &RoleResponse{
			ID:          role.ID.Hex(),
			TenantId:    role.TenantId.Hex(),
			Name:        role.Name,
			Description: role.Description,
			Permissions: permRefs,
			Resources:   resourceRefs,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}
	}

	return result, nil
}

// fetchPopulated es un helper que carga permisos poblados y el mapa de recursos para un rol
func (s *Service) fetchPopulated(ctx context.Context, permIDs, resourceIDs []primitive.ObjectID) ([]PermissionRef, map[string]ResourceRef, error) {
	permList, err := s.permissionRepo.FindByIDs(ctx, permIDs)
	if err != nil {
		return nil, nil, err
	}

	// Unir resource IDs de los permisos con los resource IDs directos del rol
	allResourceIDs := make([]primitive.ObjectID, 0, len(resourceIDs))
	allResourceIDs = append(allResourceIDs, resourceIDs...)
	for _, p := range permList {
		allResourceIDs = appendIfMissing(allResourceIDs, p.ResourceId)
	}

	resourceList, err := s.resourceRepo.FindByIDs(ctx, allResourceIDs)
	if err != nil {
		return nil, nil, err
	}

	resourceMap := make(map[string]ResourceRef, len(resourceList))
	for _, r := range resourceList {
		resourceMap[r.ID.Hex()] = ResourceRef{ID: r.ID.Hex(), Name: r.Name, Description: r.Description}
	}

	permRefs := make([]PermissionRef, 0, len(permList))
	for _, p := range permList {
		permRefs = append(permRefs, PermissionRef{
			ID:       p.ID.Hex(),
			Action:   string(p.Action),
			Resource: resourceMap[p.ResourceId.Hex()],
		})
	}

	return permRefs, resourceMap, nil
}

// collectUniqueIDs aplana y deduplica varios slices de ObjectID
func collectUniqueIDs(slices [][]primitive.ObjectID) []primitive.ObjectID {
	seen := make(map[primitive.ObjectID]struct{})
	result := make([]primitive.ObjectID, 0)
	for _, s := range slices {
		for _, id := range s {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				result = append(result, id)
			}
		}
	}
	return result
}

// appendIfMissing agrega un ObjectID al slice solo si no está ya presente
func appendIfMissing(ids []primitive.ObjectID, id primitive.ObjectID) []primitive.ObjectID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}
