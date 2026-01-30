package tenant

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/platform/payment"
)

type TenantService struct {
	repo           TenantRepository
	userRepo       users.UserRepository
	paymentManager *payment.PaymentManager
}

func NewTenantService(repo TenantRepository, userRepo users.UserRepository, paymentManager *payment.PaymentManager) *TenantService {
	return &TenantService{
		repo:           repo,
		userRepo:       userRepo,
		paymentManager: paymentManager,
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
	trialEndsAt := now.Add(14 * 24 * time.Hour) // 14 días de trial
	
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