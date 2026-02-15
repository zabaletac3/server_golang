package permissions

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type PermissionRepository interface {
	Create(ctx context.Context, dto *CreatePermissionDTO) (*Permission, error)
	FindAll(ctx context.Context, params pagination.Params) ([]*Permission, int64, error)
	FindByID(ctx context.Context, id string) (*Permission, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*Permission, error)
	FindByIDsAndAction(ctx context.Context, ids []primitive.ObjectID, action Action) ([]*Permission, error)
	Update(ctx context.Context, id string, dto *UpdatePermissionDTO) (*Permission, error)
	Delete(ctx context.Context, id string) error
}

type permissionRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *database.MongoDB) PermissionRepository {
	return &permissionRepository{
		collection: db.Collection("permissions"),
	}
}

func (r *permissionRepository) Create(ctx context.Context, dto *CreatePermissionDTO) (*Permission, error) {
	tenantID, err := primitive.ObjectIDFromHex(dto.TenantId)
	if err != nil {
		return nil, ErrInvalidPermissionID
	}

	resourceID, err := primitive.ObjectIDFromHex(dto.ResourceId)
	if err != nil {
		return nil, ErrInvalidPermissionID
	}

	action := Action(dto.Action)
	if !ValidActions[action] {
		return nil, ErrInvalidAction
	}

	now := time.Now()
	permission := &Permission{
		TenantId:   tenantID,
		ResourceId: resourceID,
		Action:     action,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result, err := r.collection.InsertOne(ctx, permission)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrPermissionExists
		}
		return nil, err
	}

	permission.ID = result.InsertedID.(primitive.ObjectID)
	return permission, nil
}

func (r *permissionRepository) FindAll(ctx context.Context, params pagination.Params) ([]*Permission, int64, error) {
	filter := bson.M{"deleted_at": nil}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(params.Skip).
		SetLimit(params.Limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var permissions []*Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

func (r *permissionRepository) FindByID(ctx context.Context, id string) (*Permission, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidPermissionID
	}

	var permission Permission
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "deleted_at": nil}).Decode(&permission)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}

	return &permission, nil
}

func (r *permissionRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*Permission, error) {
	if len(ids) == 0 {
		return []*Permission{}, nil
	}

	filter := bson.M{
		"_id":        bson.M{"$in": ids},
		"deleted_at": nil,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []*Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// FindByIDsAndAction busca permisos por IDs y filtra por acción.
// Usado por el middleware RBAC para verificar autorización eficientemente.
func (r *permissionRepository) FindByIDsAndAction(ctx context.Context, ids []primitive.ObjectID, action Action) ([]*Permission, error) {
	if len(ids) == 0 {
		return []*Permission{}, nil
	}

	filter := bson.M{
		"_id":        bson.M{"$in": ids},
		"action":     action,
		"deleted_at": nil,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []*Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *permissionRepository) Update(ctx context.Context, id string, dto *UpdatePermissionDTO) (*Permission, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidPermissionID
	}

	action := Action(dto.Action)
	if dto.Action != "" && !ValidActions[action] {
		return nil, ErrInvalidAction
	}

	set := bson.M{"updated_at": time.Now()}
	if dto.Action != "" {
		set["action"] = action
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deleted_at": nil},
		bson.M{"$set": set},
	)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, ErrPermissionNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidPermissionID
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrPermissionNotFound
	}

	return nil
}
