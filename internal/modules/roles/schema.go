package roles

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty"`
	TenantID      primitive.ObjectID   `bson:"tenant_id"`
	Name          string               `bson:"name"`
	Description   string               `bson:"description"`
	PermissionIDs []primitive.ObjectID `bson:"permission_ids"`
	CreatedAt     time.Time            `bson:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at"`
	DeletedAt     *time.Time           `bson:"deleted_at,omitempty"`
}