package permissions

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eren_dev/go_server/internal/shared/database"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *Permission) error
	FindByID(ctx context.Context, id string) (*Permission, error)
	FindAll(ctx context.Context) ([]Permission, error)
	FindByResourceAction(ctx context.Context, resource Resource, action Action) (*Permission, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]Permission, error)
	Delete(ctx context.Context, id string) error
}

type permissionRepository struct {
	collection *mongo.Collection
}

func NewPermissionRepository(db *database.MongoDB) PermissionRepository {
	return &permissionRepository{
		collection: db.Collection("permissions"),
	}
}

func (r *permissionRepository) Create(ctx context.Context, permission *Permission) error {
	if permission.ID.IsZero() {
		permission.ID = primitive.NewObjectID()
	}
	if permission.CreatedAt.IsZero() {
		permission.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, permission)
	return err
}

func (r *permissionRepository) FindByID(ctx context.Context, id string) (*Permission, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidPermissionID
	}

	var permission Permission
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&permission)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindAll(ctx context.Context) ([]Permission, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []Permission
	if err = cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) FindByResourceAction(ctx context.Context, resource Resource, action Action) (*Permission, error) {
	var permission Permission
	err := r.collection.FindOne(ctx, bson.M{
		"resource": resource,
		"action":   action,
	}).Decode(&permission)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]Permission, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"_id": bson.M{"$in": ids},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []Permission
	if err = cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidPermissionID
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrPermissionNotFound
	}
	return nil
}
