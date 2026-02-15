package permissions

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Action string

const (
	ActionGet    Action = "get"
	ActionPost   Action = "post"
	ActionPut    Action = "put"
	ActionPatch  Action = "patch"
	ActionDelete Action = "delete"
)

var ValidActions = map[Action]bool{
	ActionGet:    true,
	ActionPost:   true,
	ActionPut:    true,
	ActionPatch:  true,
	ActionDelete: true,
}

type Permission struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	TenantId   primitive.ObjectID `bson:"tenant_id"`
	ResourceId primitive.ObjectID `bson:"resource_id"`
	Action     Action             `bson:"action"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
	DeletedAt  *time.Time         `bson:"deleted_at,omitempty"`
}
