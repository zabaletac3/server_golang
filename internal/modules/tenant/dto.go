package tenant

import (
	"time"
)

// CreateTenantDTO request para crear tenant
// @name CreateTenantDto
type CreateTenantDTO struct {
	// ID del propietario (referencia a users)
	OwnerID string `json:"owner_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	
	// Información Básica
	Name                 string `json:"name" binding:"required,min=2" example:"Clínica Vet Vida"`
	CommercialName       string `json:"commercial_name" binding:"required" example:"Vet Vida"`
	IdentificationNumber string `json:"identification_number" binding:"required" example:"900123456-7"`
	Industry             string `json:"industry" binding:"required" example:"veterinary"`
	
	// Contacto
	Email          string `json:"email" binding:"required,email" example:"contacto@vetvida.com"`
	Phone          string `json:"phone" binding:"required" example:"+57 300 123 4567"`
	SecondaryPhone string `json:"secondary_phone,omitempty" example:"+57 300 987 6543"`
	Address        string `json:"address" binding:"required" example:"Calle 123 #45-67"`
	Country        string `json:"country" binding:"required" example:"Colombia"`
	
	// Técnico
	Domain   string `json:"domain" binding:"required" example:"vetvida"`
	TimeZone string `json:"timezone" binding:"required" example:"America/Bogota"`
	Currency string `json:"currency" binding:"required" example:"COP"`
	Logo     string `json:"logo,omitempty" example:"https://example.com/logo.png"`
}

// UpdateTenantDTO request para actualizar tenant
// @name UpdateTenantDto
type UpdateTenantDTO struct {
	Name                 string `json:"name,omitempty" example:"Clínica Vet Vida"`
	CommercialName       string `json:"commercial_name,omitempty" example:"Vet Vida"`
	IdentificationNumber string `json:"identification_number,omitempty" example:"900123456-7"`
	Industry             string `json:"industry,omitempty" example:"veterinary"`
	Email                string `json:"email,omitempty" example:"contacto@vetvida.com"`
	Phone                string `json:"phone,omitempty" example:"+57 300 123 4567"`
	SecondaryPhone       string `json:"secondary_phone,omitempty" example:"+57 300 987 6543"`
	Address              string `json:"address,omitempty" example:"Calle 123 #45-67"`
	Country              string `json:"country,omitempty" example:"Colombia"`
	TimeZone             string `json:"timezone,omitempty" example:"America/Bogota"`
	Currency             string `json:"currency,omitempty" example:"COP"`
	Logo                 string `json:"logo,omitempty" example:"https://example.com/logo.png"`
}

// UpdateStatusTenantDTO request para actualizar estado del tenant
// @name UpdateStatusTenantDto
type UpdateStatusTenantDTO struct {
	Status TenantStatus `json:"status" binding:"required" example:"active"`
}

// TenantSubscriptionResponse respuesta de suscripción
type TenantSubscriptionResponse struct {
	PlanID                 string     `json:"plan_id,omitempty"`
	PaymentProvider        string     `json:"payment_provider,omitempty"`
	ExternalSubscriptionID string     `json:"external_subscription_id,omitempty"`
	BillingStatus          string     `json:"billing_status"`
	TrialEndsAt            *time.Time `json:"trial_ends_at,omitempty"`
	SubscriptionEndsAt     *time.Time `json:"subscription_ends_at,omitempty"`
	MRR                    float64    `json:"mrr"`
}

// TenantUsageResponse respuesta de uso
type TenantUsageResponse struct {
	UsersCount     int       `json:"users_count"`
	UsersLimit     int       `json:"users_limit"`
	StorageUsedMB  int       `json:"storage_used_mb"`
	StorageLimitMB int       `json:"storage_limit_mb"`
	LastResetDate  time.Time `json:"last_reset_date"`
}

// TenantResponse respuesta de tenant
// @name TenantResponse
type TenantResponse struct {
	ID                   string                      `json:"id" example:"507f1f77bcf86cd799439011"`
	OwnerID              string                      `json:"owner_id" example:"507f1f77bcf86cd799439011"`
	Name                 string                      `json:"name" example:"Clínica Vet Vida"`
	CommercialName       string                      `json:"commercial_name" example:"Vet Vida"`
	IdentificationNumber string                      `json:"identification_number" example:"900123456-7"`
	Industry             string                      `json:"industry" example:"veterinary"`
	Email                string                      `json:"email" example:"contacto@vetvida.com"`
	Phone                string                      `json:"phone" example:"+57 300 123 4567"`
	SecondaryPhone       string                      `json:"secondary_phone,omitempty" example:"+57 300 987 6543"`
	Address              string                      `json:"address" example:"Calle 123 #45-67"`
	Country              string                      `json:"country" example:"Colombia"`
	Domain               string                      `json:"domain" example:"vetvida"`
	TimeZone             string                      `json:"timezone" example:"America/Bogota"`
	Currency             string                      `json:"currency" example:"COP"`
	Logo                 string                      `json:"logo,omitempty" example:"https://example.com/logo.png"`
	Subscription         TenantSubscriptionResponse  `json:"subscription"`
	Usage                TenantUsageResponse         `json:"usage"`
	Status               TenantStatus                `json:"status" example:"active"`
	CreatedAt            time.Time                   `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt            time.Time                   `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ToResponse convierte un Tenant a TenantResponse
func ToResponse(t *Tenant) *TenantResponse {
	response := &TenantResponse{
		ID:                   t.ID.Hex(),
		OwnerID:              t.OwnerID.Hex(),
		Name:                 t.Name,
		CommercialName:       t.CommercialName,
		IdentificationNumber: t.IdentificationNumber,
		Industry:             t.Industry,
		Email:                t.Email,
		Phone:                t.Phone,
		SecondaryPhone:       t.SecondaryPhone,
		Address:              t.Address,
		Country:              t.Country,
		Domain:               t.Domain,
		TimeZone:             t.TimeZone,
		Currency:             t.Currency,
		Logo:                 t.Logo,
		Status:               t.Status,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
		Subscription: TenantSubscriptionResponse{
			BillingStatus:          t.Subscription.BillingStatus,
			PaymentProvider:        t.Subscription.PaymentProvider,
			ExternalSubscriptionID: t.Subscription.ExternalSubscriptionID,
			TrialEndsAt:            t.Subscription.TrialEndsAt,
			SubscriptionEndsAt:     t.Subscription.SubscriptionEndsAt,
			MRR:                    t.Subscription.MRR,
		},
		Usage: TenantUsageResponse{
			UsersCount:     t.Usage.UsersCount,
			UsersLimit:     t.Usage.UsersLimit,
			StorageUsedMB:  t.Usage.StorageUsedMB,
			StorageLimitMB: t.Usage.StorageLimitMB,
			LastResetDate:  t.Usage.LastResetDate,
		},
	}
	
	// Agregar PlanID solo si existe
	if !t.Subscription.PlanID.IsZero() {
		planID := t.Subscription.PlanID.Hex()
		response.Subscription.PlanID = planID
	}
	
	return response
}

// ToResponseList convierte una lista de Tenant a lista de TenantResponse
func ToResponseList(tenants []Tenant) []*TenantResponse {
	responses := make([]*TenantResponse, len(tenants))
	for i, t := range tenants {
		responses[i] = ToResponse(&t)
	}
	return responses
}