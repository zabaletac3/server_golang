package owners

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

type OwnerRepository interface {
	Create(ctx context.Context, dto *CreateOwnerDTO) (*Owner, error)
	FindByID(ctx context.Context, id string) (*Owner, error)
	FindByEmail(ctx context.Context, email string) (*Owner, error)
	FindAll(ctx context.Context, params pagination.Params) ([]*Owner, int64, error)
	Update(ctx context.Context, id string, dto *UpdateOwnerDTO) (*Owner, error)
	Delete(ctx context.Context, id string) error
	AddPushToken(ctx context.Context, id string, token PushToken) error
	RemovePushToken(ctx context.Context, id string, token string) error
	AddTenantID(ctx context.Context, id string, tenantID primitive.ObjectID) error
}

type ownerRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *database.MongoDB) OwnerRepository {
	return &ownerRepository{
		collection: db.Collection("owners"),
	}
}

func (r *ownerRepository) Create(ctx context.Context, dto *CreateOwnerDTO) (*Owner, error) {
	now := time.Now()
	owner := &Owner{
		ID:         primitive.NewObjectID(),
		Name:       dto.Name,
		Email:      dto.Email,
		Phone:      dto.Phone,
		Password:   dto.Password,
		PushTokens: []PushToken{},
		TenantIds:  []primitive.ObjectID{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := r.collection.InsertOne(ctx, owner)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return owner, nil
}

func (r *ownerRepository) FindByID(ctx context.Context, id string) (*Owner, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidOwnerID
	}

	var owner Owner
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "deleted_at": nil}).Decode(&owner)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrOwnerNotFound
		}
		return nil, err
	}

	return &owner, nil
}

func (r *ownerRepository) FindByEmail(ctx context.Context, email string) (*Owner, error) {
	var owner Owner
	err := r.collection.FindOne(ctx, bson.M{"email": email, "deleted_at": nil}).Decode(&owner)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrOwnerNotFound
		}
		return nil, err
	}

	return &owner, nil
}

func (r *ownerRepository) FindAll(ctx context.Context, params pagination.Params) ([]*Owner, int64, error) {
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

	var owners []*Owner
	if err := cursor.All(ctx, &owners); err != nil {
		return nil, 0, err
	}

	return owners, total, nil
}

func (r *ownerRepository) Update(ctx context.Context, id string, dto *UpdateOwnerDTO) (*Owner, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidOwnerID
	}

	set := bson.M{"updated_at": time.Now()}
	if dto.Name != "" {
		set["name"] = dto.Name
	}
	if dto.Phone != "" {
		set["phone"] = dto.Phone
	}
	if dto.AvatarURL != "" {
		set["avatar_url"] = dto.AvatarURL
	}
	if dto.Address != "" {
		set["address"] = dto.Address
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
		return nil, ErrOwnerNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *ownerRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidOwnerID
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": time.Now()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrOwnerNotFound
	}

	return nil
}

func (r *ownerRepository) AddPushToken(ctx context.Context, id string, token PushToken) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidOwnerID
	}

	// Remove existing entry for the same token string, then add the new one
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deleted_at": nil},
		bson.D{
			{Key: "$pull", Value: bson.M{"push_tokens": bson.M{"token": token.Token}}},
		},
	)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deleted_at": nil},
		bson.D{
			{Key: "$push", Value: bson.M{"push_tokens": token}},
			{Key: "$set", Value: bson.M{"updated_at": time.Now()}},
		},
	)
	return err
}

func (r *ownerRepository) RemovePushToken(ctx context.Context, id string, token string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidOwnerID
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deleted_at": nil},
		bson.D{
			{Key: "$pull", Value: bson.M{"push_tokens": bson.M{"token": token}}},
			{Key: "$set", Value: bson.M{"updated_at": time.Now()}},
		},
	)
	return err
}

func (r *ownerRepository) AddTenantID(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidOwnerID
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deleted_at": nil},
		bson.D{
			{Key: "$addToSet", Value: bson.M{"tenant_ids": tenantID}},
			{Key: "$set", Value: bson.M{"updated_at": time.Now()}},
		},
	)
	return err
}
