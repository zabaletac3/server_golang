package roles

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eren_dev/go_server/internal/shared/database"
)

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, id string) (*Role, error)
	FindByTenantID(ctx context.Context, tenantID string) ([]Role, error)
	FindByName(ctx context.Context, tenantID, name string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
}

type roleRepository struct {
	collection *mongo.Collection
}

func NewRoleRepository(db *database.MongoDB) RoleRepository {
	return &roleRepository{
		collection: db.Collection("roles"),
	}
}

func (r *roleRepository) Create(ctx context.Context, role *Role) error {
	if role.ID.IsZero() {
		role.ID = primitive.NewObjectID()
	}
	now := time.Now()
	if role.CreatedAt.IsZero() {
		role.CreatedAt = now
	}
	if role.UpdatedAt.IsZero() {
		role.UpdatedAt = now
	}

	_, err := r.collection.InsertOne(ctx, role)
	return err
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*Role, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidRoleID
	}

	var role Role
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"deleted_at": nil,
	}).Decode(&role)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByTenantID(ctx context.Context, tenantID string) ([]Role, error) {
	objectID, err := primitive.ObjectIDFromHex(tenantID)
	if err != nil {
		return nil, ErrInvalidRoleID
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"tenant_id":  objectID,
		"deleted_at": nil,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []Role
	if err = cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) FindByName(ctx context.Context, tenantID, name string) (*Role, error) {
	objectID, err := primitive.ObjectIDFromHex(tenantID)
	if err != nil {
		return nil, ErrInvalidRoleID
	}

	var role Role
	err = r.collection.FindOne(ctx, bson.M{
		"tenant_id":  objectID,
		"name":       name,
		"deleted_at": nil,
	}).Decode(&role)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *Role) error {
	role.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": role.ID},
		bson.M{"$set": role},
	)
	return err
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidRoleID
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"deleted_at": now}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRoleNotFound
	}
	return nil
}
