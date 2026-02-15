package resources

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

type ResourceRepository interface {
	Create(ctx context.Context, dto *CreateResourceDTO) (*Resource, error)
	FindAll(ctx context.Context, params pagination.Params) ([]*Resource, int64, error)
	FindByID(ctx context.Context, id string) (*Resource, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*Resource, error)
	FindByIDsAndName(ctx context.Context, ids []primitive.ObjectID, name string) (bool, error)
	Update(ctx context.Context, id string, dto *UpdateResourceDTO) (*Resource, error)
	Delete(ctx context.Context, id string) error
}

type resourceRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *database.MongoDB) ResourceRepository {
	return &resourceRepository{
		collection: db.Collection("resources"),
	}
}

func (r *resourceRepository) Create(ctx context.Context, dto *CreateResourceDTO) (*Resource, error) {
	tenantID, err := primitive.ObjectIDFromHex(dto.TenantId)
	if err != nil {
		return nil, ErrInvalidResourceID
	}

	now := time.Now()
	resource := &Resource{
		TenantId:    tenantID,
		Name:        dto.Name,
		Description: dto.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := r.collection.InsertOne(ctx, resource)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrResourceNameExists
		}
		return nil, err
	}

	resource.ID = result.InsertedID.(primitive.ObjectID)
	return resource, nil
}

func (r *resourceRepository) FindAll(ctx context.Context, params pagination.Params) ([]*Resource, int64, error) {
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

	var resources []*Resource
	if err := cursor.All(ctx, &resources); err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

func (r *resourceRepository) FindByID(ctx context.Context, id string) (*Resource, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidResourceID
	}

	var resource Resource
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "deleted_at": nil}).Decode(&resource)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return &resource, nil
}

func (r *resourceRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*Resource, error) {
	if len(ids) == 0 {
		return []*Resource{}, nil
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

	var resources []*Resource
	if err := cursor.All(ctx, &resources); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *resourceRepository) FindByIDsAndName(ctx context.Context, ids []primitive.ObjectID, name string) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	filter := bson.M{
		"_id":        bson.M{"$in": ids},
		"name":       name,
		"deleted_at": nil,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *resourceRepository) Update(ctx context.Context, id string, dto *UpdateResourceDTO) (*Resource, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidResourceID
	}

	set := bson.M{"updated_at": time.Now()}
	if dto.Name != "" {
		set["name"] = dto.Name
	}
	if dto.Description != "" {
		set["description"] = dto.Description
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
		return nil, ErrResourceNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *resourceRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidResourceID
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
		return ErrResourceNotFound
	}

	return nil
}
