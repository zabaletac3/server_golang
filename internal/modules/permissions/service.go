package permissions

import (
	"context"
)

type PermissionService struct {
	repo PermissionRepository
}

func NewPermissionService(repo PermissionRepository) *PermissionService {
	return &PermissionService{repo: repo}
}

func (s *PermissionService) Create(ctx context.Context, dto *CreatePermissionDTO) (*PermissionResponse, error) {
	// Validar resource
	if !dto.Resource.IsValid() {
		return nil, ErrInvalidResource
	}

	// Validar action
	if !dto.Action.IsValid() {
		return nil, ErrInvalidAction
	}

	// Verificar si ya existe
	existing, err := s.repo.FindByResourceAction(ctx, dto.Resource, dto.Action)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPermissionExists
	}

	permission := &Permission{
		Resource: dto.Resource,
		Action:   dto.Action,
	}

	if err := s.repo.Create(ctx, permission); err != nil {
		return nil, err
	}
	return ToResponse(permission), nil
}

func (s *PermissionService) FindByID(ctx context.Context, id string) (*PermissionResponse, error) {
	permission, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(permission), nil
}

func (s *PermissionService) FindAll(ctx context.Context) ([]*PermissionResponse, error) {
	permissions, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return ToResponseList(permissions), nil
}

func (s *PermissionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetAvailableOptions retorna todos los recursos y acciones disponibles
func (s *PermissionService) GetAvailableOptions(ctx context.Context) *AvailableOptionsResponse {
	return &AvailableOptionsResponse{
		Resources: AllResources(),
		Actions:   AllActions(),
	}
}
