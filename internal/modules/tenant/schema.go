package tenant

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TenantSubscription información de suscripción embebida
type TenantSubscription struct {
	PlanID                 primitive.ObjectID `bson:"plan_id,omitempty" json:"plan_id,omitempty"`
	PaymentProvider        string             `bson:"payment_provider,omitempty" json:"payment_provider,omitempty"` // wompi, stripe
	ExternalSubscriptionID string             `bson:"external_subscription_id,omitempty" json:"external_subscription_id,omitempty"`
	BillingStatus          string             `bson:"billing_status" json:"billing_status"` // trial, active, past_due, canceled
	TrialEndsAt            *time.Time         `bson:"trial_ends_at,omitempty" json:"trial_ends_at,omitempty"`
	SubscriptionEndsAt     *time.Time         `bson:"subscription_ends_at,omitempty" json:"subscription_ends_at,omitempty"`
	MRR                    float64            `bson:"mrr" json:"mrr"` // Monthly Recurring Revenue
}

// TenantUsage límites y uso actual
type TenantUsage struct {
	UsersCount     int       `bson:"users_count" json:"users_count"`
	UsersLimit     int       `bson:"users_limit" json:"users_limit"`
	StorageUsedMB  int       `bson:"storage_used_mb" json:"storage_used_mb"`
	StorageLimitMB int       `bson:"storage_limit_mb" json:"storage_limit_mb"`
	LastResetDate  time.Time `bson:"last_reset_date" json:"last_reset_date"`
}

type Tenant struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID              primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	
	// Información Básica
	Name                 string `bson:"name" json:"name"`
	CommercialName       string `bson:"commercial_name" json:"commercial_name"`
	IdentificationNumber string `bson:"identification_number" json:"identification_number"`
	Industry             string `bson:"industry" json:"industry"` // veterinary, health, retail
	
	// Contacto
	Email          string `bson:"email" json:"email"`
	Phone          string `bson:"phone" json:"phone"`
	SecondaryPhone string `bson:"secondary_phone,omitempty" json:"secondary_phone,omitempty"`
	Address        string `bson:"address" json:"address"`
	Country        string `bson:"country" json:"country"`
	
	// Técnico
	Domain   string `bson:"domain" json:"domain"`
	TimeZone string `bson:"time_zone" json:"time_zone"`
	Currency string `bson:"currency" json:"currency"`
	Logo     string `bson:"logo,omitempty" json:"logo,omitempty"`
	
	// Suscripción (embebido)
	Subscription TenantSubscription `bson:"subscription" json:"subscription"`
	
	// Uso (embebido)
	Usage TenantUsage `bson:"usage" json:"usage"`
	
	// Estado
	Status    TenantStatus `bson:"status" json:"status"`
	CreatedAt time.Time    `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time    `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time   `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}