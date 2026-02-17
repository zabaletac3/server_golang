package owners

import (
	"context"
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type PaginatedOwnersResponse struct {
	Data       []*OwnerResponse       `json:"data"`
	Pagination pagination.PaginationInfo `json:"pagination"`
}

type Service struct {
	repo OwnerRepository
}

func NewService(repo OwnerRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMe(ctx context.Context, ownerID string) (*OwnerResponse, error) {
	owner, err := s.repo.FindByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return ToResponse(owner), nil
}

func (s *Service) UpdateMe(ctx context.Context, ownerID string, dto *UpdateOwnerDTO) (*OwnerResponse, error) {
	owner, err := s.repo.Update(ctx, ownerID, dto)
	if err != nil {
		return nil, err
	}
	return ToResponse(owner), nil
}

func (s *Service) AddPushToken(ctx context.Context, ownerID string, dto *RegisterPushTokenDTO) error {
	token := PushToken{
		Token:     dto.Token,
		Platform:  dto.Platform,
		Active:    true,
		UpdatedAt: time.Now(),
	}
	return s.repo.AddPushToken(ctx, ownerID, token)
}

func (s *Service) RemovePushToken(ctx context.Context, ownerID string, token string) error {
	return s.repo.RemovePushToken(ctx, ownerID, token)
}

// FindAll is for admin panel usage (staff with RBAC)
func (s *Service) FindAll(ctx context.Context, params pagination.Params) (*PaginatedOwnersResponse, error) {
	owners, total, err := s.repo.FindAll(ctx, params)
	if err != nil {
		return nil, err
	}

	data := make([]*OwnerResponse, len(owners))
	for i, o := range owners {
		data[i] = ToResponse(o)
	}

	return &PaginatedOwnersResponse{
		Data:       data,
		Pagination: pagination.NewPaginationInfo(params, total),
	}, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*OwnerResponse, error) {
	owner, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(owner), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
