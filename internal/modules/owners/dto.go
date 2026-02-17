package owners

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Input DTOs ---

type UpdateOwnerDTO struct {
	Name      string `json:"name"       example:"Juan Pérez"`
	Phone     string `json:"phone"      example:"+57 300 123 4567"`
	AvatarURL string `json:"avatar_url" example:"https://cdn.example.com/avatar.jpg"`
	Address   string `json:"address"    example:"Calle 10 # 20-30, Medellín"`
}

type RegisterPushTokenDTO struct {
	Token    string `json:"token"    binding:"required" example:"fcm-token-abc123"`
	Platform string `json:"platform" binding:"required,oneof=ios android" example:"android"`
}

type RemovePushTokenDTO struct {
	Token string `json:"token" binding:"required" example:"fcm-token-abc123"`
}

// --- Response DTOs ---

type PushTokenResponse struct {
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	Active    bool      `json:"active"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OwnerResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Email      string              `json:"email"`
	Phone      string              `json:"phone"`
	AvatarURL  string              `json:"avatar_url,omitempty"`
	Address    string              `json:"address,omitempty"`
	TenantIds  []string            `json:"tenant_ids"`
	PushTokens []PushTokenResponse `json:"push_tokens"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

func ToResponse(o *Owner) *OwnerResponse {
	tenantIDs := make([]string, len(o.TenantIds))
	for i, id := range o.TenantIds {
		tenantIDs[i] = id.Hex()
	}

	pushTokens := make([]PushTokenResponse, len(o.PushTokens))
	for i, pt := range o.PushTokens {
		pushTokens[i] = PushTokenResponse{
			Token:     pt.Token,
			Platform:  pt.Platform,
			Active:    pt.Active,
			UpdatedAt: pt.UpdatedAt,
		}
	}

	return &OwnerResponse{
		ID:         o.ID.Hex(),
		Name:       o.Name,
		Email:      o.Email,
		Phone:      o.Phone,
		AvatarURL:  o.AvatarURL,
		Address:    o.Address,
		TenantIds:  tenantIDs,
		PushTokens: pushTokens,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

// CreateOwnerDTO is used internally (registration flow uses mobile_auth)
type CreateOwnerDTO struct {
	Name     string
	Email    string
	Phone    string
	Password string // already hashed
}

// Internal use only
type ownerDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	TenantIds []primitive.ObjectID
}
