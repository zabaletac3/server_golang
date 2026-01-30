package plans

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlanService struct {
	repo PlanRepository
}

func NewPlanService(repo PlanRepository) *PlanService {
	return &PlanService{repo: repo}
}

func (s *PlanService) Create(ctx context.Context, dto *CreatePlanDTO) (*PlanResponse, error) {
	now := time.Now()
	
	plan := &Plan{
		ID:             primitive.NewObjectID(),
		Name:           dto.Name,
		Description:    dto.Description,
		MonthlyPrice:   dto.MonthlyPrice,
		AnnualPrice:    dto.AnnualPrice,
		Currency:       dto.Currency,
		MaxUsers:       dto.MaxUsers,
		MaxBranches:    dto.MaxBranches,
		StorageLimitGB: dto.StorageLimitGB,
		Features:       dto.Features,
		IsVisible:      dto.IsVisible,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, err
	}

	return ToResponse(plan), nil
}

func (s *PlanService) FindByID(ctx context.Context, id string) (*PlanResponse, error) {
	plan, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(plan), nil
}

func (s *PlanService) FindAll(ctx context.Context) ([]*PlanResponse, error) {
	plans, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return ToResponseList(plans), nil
}

func (s *PlanService) FindVisible(ctx context.Context) ([]*PlanResponse, error) {
	plans, err := s.repo.FindVisible(ctx)
	if err != nil {
		return nil, err
	}
	return ToResponseList(plans), nil
}

func (s *PlanService) Update(ctx context.Context, id string, dto *UpdatePlanDTO) (*PlanResponse, error) {
	plan, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Actualizar campos si están presentes
	if dto.Name != "" {
		plan.Name = dto.Name
	}
	if dto.Description != "" {
		plan.Description = dto.Description
	}
	if dto.MonthlyPrice > 0 {
		plan.MonthlyPrice = dto.MonthlyPrice
	}
	if dto.AnnualPrice > 0 {
		plan.AnnualPrice = dto.AnnualPrice
	}
	if dto.MaxUsers > 0 {
		plan.MaxUsers = dto.MaxUsers
	}
	if dto.MaxBranches > 0 {
		plan.MaxBranches = dto.MaxBranches
	}
	if dto.StorageLimitGB > 0 {
		plan.StorageLimitGB = dto.StorageLimitGB
	}
	if dto.Features != nil {
		plan.Features = dto.Features
	}
	if dto.IsVisible != nil {
		plan.IsVisible = *dto.IsVisible
	}

	if err := s.repo.Update(ctx, plan); err != nil {
		return nil, err
	}

	return ToResponse(plan), nil
}

func (s *PlanService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
