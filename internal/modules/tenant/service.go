package tenant

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/config"
	"github.com/eren_dev/go_server/internal/modules/audit"
	"github.com/eren_dev/go_server/internal/modules/payments"
	"github.com/eren_dev/go_server/internal/modules/plans"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/platform/payment"
)

var ErrPlanNotVisible = errors.New("plan is not available for purchase")

type TenantService struct {
	repo           TenantRepository
	userRepo       users.UserRepository
	planRepo       plans.PlanRepository
	paymentService *payments.PaymentService
	paymentManager *payment.PaymentManager
	auditService   *audit.Service
	cfg            *config.Config
}

func NewTenantService(repo TenantRepository, userRepo users.UserRepository, planRepo plans.PlanRepository, paymentService *payments.PaymentService, paymentManager *payment.PaymentManager, auditService *audit.Service, cfg *config.Config) *TenantService {
	return &TenantService{
		repo:           repo,
		userRepo:       userRepo,
		planRepo:       planRepo,
		paymentService: paymentService,
		paymentManager: paymentManager,
		auditService:   auditService,
		cfg:            cfg,
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
	trialEndsAt := now.AddDate(0, 0, s.cfg.TenantTrialDays)

	tenant := &Tenant{
		ID:                   primitive.NewObjectID(),
		OwnerID:              ownerID,
		Name:                 dto.Name,
		CommercialName:       dto.CommercialName,
		IdentificationNumber: dto.IdentificationNumber,
		Industry:             dto.Industry,
		Email:                dto.Email,
		Phone:                dto.Phone,
		SecondaryPhone:       dto.SecondaryPhone,
		Address:              dto.Address,
		Country:              dto.Country,
		Domain:               dto.Domain,
		TimeZone:             dto.TimeZone,
		Currency:             dto.Currency,
		Logo:                 dto.Logo,
		Status:               Trial, // Inicia en trial

		// Inicializar suscripción con valores por defecto
		Subscription: TenantSubscription{
			BillingStatus: "trial",
			TrialEndsAt:   &trialEndsAt,
			MRR:           0,
		},

		// Inicializar uso con límites básicos (plan free)
		Usage: TenantUsage{
			UsersCount:     1, // El owner cuenta como primer usuario
			UsersLimit:     5,
			StorageUsedMB:  0,
			StorageLimitMB: 1000, // 1GB
			LastResetDate:  now,
		},

		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	// Audit log
	if s.auditService != nil {
		_ = s.auditService.LogTenantAction(ctx, tenant.ID, ownerID, audit.EventTenantCreated, "create", "Tenant created", map[string]interface{}{
			"name":           tenant.Name,
			"commercial_name": tenant.CommercialName,
			"email":          tenant.Email,
		})
	}

	return ToResponse(tenant), nil
}

func (s *TenantService) Subscribe(ctx context.Context, tenantID string, dto *SubscribeDTO) (*SubscribeResponse, error) {
	// 1. Obtener tenant
	tenant, err := s.repo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 2. Obtener plan
	plan, err := s.planRepo.FindByID(ctx, dto.PlanID)
	if err != nil {
		return nil, err
	}

	if !plan.IsVisible {
		return nil, ErrPlanNotVisible
	}

	// 3. Calcular monto según período de facturación
	var price float64
	switch dto.BillingPeriod {
	case "monthly":
		price = plan.MonthlyPrice
	case "annual":
		price = plan.AnnualPrice
	}
	amountInCents := int64(math.Round(price * 100))

	// 4. Crear payment link en Wompi
	subReq := &payment.SubscriptionRequest{
		TenantID:      tenantID,
		PlanID:        dto.PlanID,
		PlanName:      plan.Name,
		CustomerEmail: tenant.Email,
		CustomerName:  tenant.Name,
		BillingPeriod: dto.BillingPeriod,
		Amount:        amountInCents,
		Currency:      plan.Currency,
	}

	subResp, err := s.paymentManager.CreateSubscription(ctx, subReq, nil)
	if err != nil {
		return nil, fmt.Errorf("payment provider error: %w", err)
	}

	// 5. Crear registro de pago pendiente
	paymentResp, err := s.paymentService.Create(ctx, &payments.CreatePaymentDTO{
		TenantID:              tenantID,
		PlanID:                dto.PlanID,
		Amount:                amountInCents,
		Currency:              plan.Currency,
		PaymentMethod:         "wompi",
		Status:                payments.PaymentPending,
		ExternalTransactionID: subResp.SubscriptionID,
		Concept:               fmt.Sprintf("Suscripción %s - %s", plan.Name, dto.BillingPeriod),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// 6. Actualizar suscripción del tenant
	tenant.Subscription.PlanID = plan.ID
	tenant.Subscription.PaymentProvider = "wompi"
	tenant.Subscription.ExternalSubscriptionID = subResp.SubscriptionID
	tenant.Subscription.BillingStatus = "pending"
	if subResp.NextBillingAt != nil {
		tenant.Subscription.SubscriptionEndsAt = subResp.NextBillingAt
	}
	tenant.Subscription.MRR = plan.MonthlyPrice
	tenant.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	// Audit log
	if s.auditService != nil {
		_ = s.auditService.LogTenantAction(ctx, tenant.ID, tenant.OwnerID, audit.EventTenantSubscription, "subscribe", fmt.Sprintf("Subscription initiated for plan %s", plan.Name), map[string]interface{}{
			"plan_id":         plan.ID.Hex(),
			"plan_name":       plan.Name,
			"billing_period":  dto.BillingPeriod,
			"payment_id":      paymentResp.ID,
			"amount":          amountInCents,
		})
	}

	return &SubscribeResponse{
		PaymentLinkURL: subResp.PaymentLinkURL,
		PaymentID:      paymentResp.ID,
	}, nil
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

	// Actualizar campos si están presentes
	if dto.Name != "" {
		tenant.Name = dto.Name
	}
	if dto.CommercialName != "" {
		tenant.CommercialName = dto.CommercialName
	}
	if dto.IdentificationNumber != "" {
		tenant.IdentificationNumber = dto.IdentificationNumber
	}
	if dto.Industry != "" {
		tenant.Industry = dto.Industry
	}
	if dto.Email != "" {
		tenant.Email = dto.Email
	}
	if dto.Phone != "" {
		tenant.Phone = dto.Phone
	}
	if dto.SecondaryPhone != "" {
		tenant.SecondaryPhone = dto.SecondaryPhone
	}
	if dto.Address != "" {
		tenant.Address = dto.Address
	}
	if dto.Country != "" {
		tenant.Country = dto.Country
	}
	if dto.TimeZone != "" {
		tenant.TimeZone = dto.TimeZone
	}
	if dto.Currency != "" {
		tenant.Currency = dto.Currency
	}
	if dto.Logo != "" {
		tenant.Logo = dto.Logo
	}

	tenant.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return ToResponse(tenant), nil
}

func (s *TenantService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
