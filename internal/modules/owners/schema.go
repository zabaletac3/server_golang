package owners

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PushToken struct {
	Token     string    `bson:"token"      json:"token"`
	Platform  string    `bson:"platform"   json:"platform"` // "ios" | "android"
	Active    bool      `bson:"active"     json:"active"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type Owner struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty"`
	Name       string               `bson:"name"`
	Email      string               `bson:"email"`
	Phone      string               `bson:"phone"`
	Password   string               `bson:"password,omitempty"`
	AvatarURL  string               `bson:"avatar_url,omitempty"`
	Address    string               `bson:"address,omitempty"`
	PushTokens []PushToken          `bson:"push_tokens"`
	TenantIds  []primitive.ObjectID `bson:"tenant_ids"`
	CreatedAt  time.Time            `bson:"created_at"`
	UpdatedAt  time.Time            `bson:"updated_at"`
	DeletedAt  *time.Time           `bson:"deleted_at,omitempty"`
}
