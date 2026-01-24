package tenant

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/users"
)

type TenantService struct {
	repo     TenantRepository
	userRepo users.UserRepository
}

func NewTenantService(repo TenantRepository, userRepo users.UserRepository) *TenantService {
	return &TenantService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *TenantService) Create(ctx context.Context, dto *CreateTenantDTO) (*TenantResponse, error) {
	// Validar que el owner existe
	ownerID, err := primitive.ObjectIDFromHex(dto.OwnerID)
	if err != nil {
		return nil, ErrInvalidOwnerID
	}

	_, err = s.userRepo.FindByID(ctx, dto.OwnerID)
	if err != nil {
		if err == users.ErrUserNotFound {
			return nil, ErrOwnerNotFound
		}
		return nil, err
	}

	now := time.Now()
	tenant := &Tenant{
		ID:        primitive.NewObjectID(),
		OwnerID:   ownerID,
		Name:      dto.Name,
		Status:    Active,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}
	return ToResponse(tenant), nil
}

func (s *TenantService) FindByID(ctx context.Context, id string) (*TenantResponse, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(tenant), nil
}

func (s *TenantService) FindAll(ctx context.Context) ([]*TenantResponse, error) {
	tenants, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return ToResponseList(tenants), nil
}

func (s *TenantService) Update(ctx context.Context, id string, dto *UpdateTenantDTO) (*TenantResponse, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name != "" {
		tenant.Name = dto.Name
	}

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return ToResponse(tenant), nil
}

func (s *TenantService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}