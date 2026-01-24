package roles

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/permissions"
)

type RoleService struct {
	repo     RoleRepository
	permRepo permissions.PermissionRepository
}

func NewRoleService(repo RoleRepository, permRepo permissions.PermissionRepository) *RoleService {
	return &RoleService{
		repo:     repo,
		permRepo: permRepo,
	}
}

func (s *RoleService) Create(ctx context.Context, dto *CreateRoleDTO) (*RoleSimpleResponse, error) {
	tenantID, err := primitive.ObjectIDFromHex(dto.TenantID)
	if err != nil {
		return nil, ErrInvalidRoleID
	}

	// Verificar si ya existe un rol con ese nombre en el tenant
	existing, err := s.repo.FindByName(ctx, dto.TenantID, dto.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRoleExists
	}

	// Convertir permission IDs
	permIDs := make([]primitive.ObjectID, len(dto.PermissionIDs))
	for i, id := range dto.PermissionIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, permissions.ErrInvalidPermissionID
		}
		permIDs[i] = objID
	}

	role := &Role{
		TenantID:      tenantID,
		Name:          dto.Name,
		Description:   dto.Description,
		PermissionIDs: permIDs,
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return ToSimpleResponse(role), nil
}

func (s *RoleService) FindByID(ctx context.Context, id string) (*RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Obtener permisos expandidos
	perms, err := s.permRepo.FindByIDs(ctx, role.PermissionIDs)
	if err != nil {
		return nil, err
	}

	return &RoleResponse{
		ID:          role.ID.Hex(),
		TenantID:    role.TenantID.Hex(),
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions.ToResponseList(perms),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

func (s *RoleService) FindByTenantID(ctx context.Context, tenantID string) ([]*RoleSimpleResponse, error) {
	roles, err := s.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return ToSimpleResponseList(roles), nil
}

func (s *RoleService) Update(ctx context.Context, id string, dto *UpdateRoleDTO) (*RoleSimpleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name != "" {
		role.Name = dto.Name
	}
	if dto.Description != "" {
		role.Description = dto.Description
	}
	if dto.PermissionIDs != nil {
		permIDs := make([]primitive.ObjectID, len(dto.PermissionIDs))
		for i, pid := range dto.PermissionIDs {
			objID, err := primitive.ObjectIDFromHex(pid)
			if err != nil {
				return nil, permissions.ErrInvalidPermissionID
			}
			permIDs[i] = objID
		}
		role.PermissionIDs = permIDs
	}

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}
	return ToSimpleResponse(role), nil
}

func (s *RoleService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
