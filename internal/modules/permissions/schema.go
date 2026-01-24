package permissions

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Permission struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Resource  Resource           `bson:"resource"`
	Action    Action             `bson:"action"`
	CreatedAt time.Time          `bson:"created_at"`
}
