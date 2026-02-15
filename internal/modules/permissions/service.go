package permissions

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/resources"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type Service struct {
	repo         PermissionRepository
	resourceRepo resources.ResourceRepository
}

func NewService(repo PermissionRepository, resourceRepo resources.ResourceRepository) *Service {
	return &Service{
		repo:         repo,
		resourceRepo: resourceRepo,
	}
}

func (s *Service) Create(ctx context.Context, dto *CreatePermissionDTO) (*PermissionResponse, error) {
	// Validar que el recurso existe
	_, err := s.resourceRepo.FindByID(ctx, dto.ResourceId)
	if err != nil {
		return nil, err
	}

	permission, err := s.repo.Create(ctx, dto)
	if err != nil {
		return nil, err
	}

	return s.toResponse(ctx, permission)
}

func (s *Service) FindAll(ctx context.Context, params pagination.Params) (*PaginatedPermissionsResponse, error) {
	perms, total, err := s.repo.FindAll(ctx, params)
	if err != nil {
		return nil, err
	}

	responses, err := s.toResponseList(ctx, perms)
	if err != nil {
		return nil, err
	}

	paginationInfo := pagination.NewPaginationInfo(params, total)
	return &PaginatedPermissionsResponse{
		Data:       responses,
		Pagination: &paginationInfo,
	}, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*PermissionResponse, error) {
	permission, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, permission)
}

func (s *Service) Update(ctx context.Context, id string, dto *UpdatePermissionDTO) (*PermissionResponse, error) {
	permission, err := s.repo.Update(ctx, id, dto)
	if err != nil {
		return nil, err
	}
	return s.toResponse(ctx, permission)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// toResponse puebla el recurso de un permiso en una sola query
func (s *Service) toResponse(ctx context.Context, p *Permission) (*PermissionResponse, error) {
	resource, err := s.resourceRepo.FindByID(ctx, p.ResourceId.Hex())
	if err != nil {
		return nil, err
	}
	ref := ResourceRef{ID: resource.ID.Hex(), Name: resource.Name}
	return ToResponse(p, ref), nil
}

// toResponseList puebla los recursos de una lista de permisos con una sola query batch
func (s *Service) toResponseList(ctx context.Context, perms []*Permission) ([]*PermissionResponse, error) {
	if len(perms) == 0 {
		return []*PermissionResponse{}, nil
	}

	// Recolectar IDs únicos de recursos
	resourceIDSet := make(map[primitive.ObjectID]struct{})
	for _, p := range perms {
		resourceIDSet[p.ResourceId] = struct{}{}
	}
	resourceIDs := make([]primitive.ObjectID, 0, len(resourceIDSet))
	for id := range resourceIDSet {
		resourceIDs = append(resourceIDs, id)
	}

	// Fetch todos los recursos en una sola query
	resourceList, err := s.resourceRepo.FindByIDs(ctx, resourceIDs)
	if err != nil {
		return nil, err
	}

	// Indexar por ID para lookup O(1)
	resourceMap := make(map[string]ResourceRef, len(resourceList))
	for _, r := range resourceList {
		resourceMap[r.ID.Hex()] = ResourceRef{ID: r.ID.Hex(), Name: r.Name}
	}

	result := make([]*PermissionResponse, len(perms))
	for i, p := range perms {
		ref := resourceMap[p.ResourceId.Hex()]
		result[i] = ToResponse(p, ref)
	}
	return result, nil
}
