package resources

import (
	"context"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type Service struct {
	repo ResourceRepository
}

func NewService(repo ResourceRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, dto *CreateResourceDTO) (*ResourceResponse, error) {
	resource, err := s.repo.Create(ctx, dto)
	if err != nil {
		return nil, err
	}
	return ToResponse(resource), nil
}

func (s *Service) FindAll(ctx context.Context, params pagination.Params) (*PaginatedResourcesResponse, error) {
	resources, total, err := s.repo.FindAll(ctx, params)
	if err != nil {
		return nil, err
	}

	paginationInfo := pagination.NewPaginationInfo(params, total)
	return &PaginatedResourcesResponse{
		Data:       ToResponseList(resources),
		Pagination: &paginationInfo,
	}, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*ResourceResponse, error) {
	resource, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(resource), nil
}

func (s *Service) Update(ctx context.Context, id string, dto *UpdateResourceDTO) (*ResourceResponse, error) {
	resource, err := s.repo.Update(ctx, id, dto)
	if err != nil {
		return nil, err
	}
	return ToResponse(resource), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
