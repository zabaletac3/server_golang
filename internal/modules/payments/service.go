package payments

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentService struct {
	repo PaymentRepository
}

func NewPaymentService(repo PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) Create(ctx context.Context, dto *CreatePaymentDTO) (*PaymentResponse, error) {
	tenantID, err := primitive.ObjectIDFromHex(dto.TenantID)
	if err != nil {
		return nil, ErrInvalidTenantID
	}

	var planID primitive.ObjectID
	if dto.PlanID != "" {
		planID, err = primitive.ObjectIDFromHex(dto.PlanID)
		if err != nil {
			return nil, fmt.Errorf("invalid plan_id: %w", err)
		}
	}

	now := time.Now()
	payment := &Payment{
		ID:                    primitive.NewObjectID(),
		TenantID:              tenantID,
		PlanID:                planID,
		Amount:                dto.Amount,
		Currency:              dto.Currency,
		PaymentMethod:         dto.PaymentMethod,
		Status:                dto.Status,
		ExternalTransactionID: dto.ExternalTransactionID,
		Concept:               dto.Concept,
		PeriodStart:           dto.PeriodStart,
		PeriodEnd:             dto.PeriodEnd,
		ProcessedAt:           dto.ProcessedAt,
		FailureReason:         dto.FailureReason,
		Metadata:              dto.Metadata,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return ToResponse(payment), nil
}

func (s *PaymentService) FindByID(ctx context.Context, id string) (*PaymentResponse, error) {
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(payment), nil
}

func (s *PaymentService) FindByTenantID(ctx context.Context, tenantID string, limit int) ([]*PaymentResponse, error) {
	if limit <= 0 {
		limit = 50 // Default
	}
	if limit > 100 {
		limit = 100 // Max
	}

	payments, err := s.repo.FindByTenantID(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}
	return ToResponseList(payments), nil
}

func (s *PaymentService) UpdateStatus(ctx context.Context, id string, status PaymentStatus, processedAt *time.Time, failureReason string) error {
	return s.repo.UpdateStatus(ctx, id, status, processedAt, failureReason)
}
