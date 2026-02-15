package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Phone     string             `bson:"phone"`
	Password  string             `bson:"password,omitempty"`
	TenantIds []primitive.ObjectID `bson:"tenant_ids"`
	RoleIds      []primitive.ObjectID `bson:"role_ids"`
	IsSuperAdmin bool             `bson:"is_super_admin"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt *time.Time          `bson:"deleted_at,omitempty"`
}
