package resources

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resource struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	TenantId    primitive.ObjectID `bson:"tenant_id"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty"`
}
