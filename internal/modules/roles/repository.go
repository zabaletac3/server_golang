package roles

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

type RoleRepository interface {
	Create(ctx context.Context, dto *CreateRoleDTO) (*Role, error)
	FindAll(ctx context.Context, params pagination.Params) ([]*Role, int64, error)
	FindByID(ctx context.Context, id string) (*Role, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*Role, error)
	Update(ctx context.Context, id string, dto *UpdateRoleDTO) (*Role, error)
	Delete(ctx context.Context, id string) error
}

type roleRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *database.MongoDB) RoleRepository {
	return &roleRepository{
		collection: db.Collection("roles"),
	}
}

func (r *roleRepository) Create(ctx context.Context, dto *CreateRoleDTO) (*Role, error) {
	tenantID, err := primitive.ObjectIDFromHex(dto.TenantId)
	if err != nil {
		return nil, ErrInvalidRoleID
	}

	permissionsIds := parseObjectIDs(dto.PermissionsIds)
	resourcesIds := parseObjectIDs(dto.ResourcesIds)

	now := time.Now()
	role := &Role{
		TenantId:       tenantID,
		Name:           dto.Name,
		Description:    dto.Description,
		PermissionsIds: permissionsIds,
		ResourcesIds:   resourcesIds,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result, err := r.collection.InsertOne(ctx, role)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrRoleNameExists
		}
		return nil, err
	}

	role.ID = result.InsertedID.(primitive.ObjectID)
	return role, nil
}

func (r *roleRepository) FindAll(ctx context.Context, params pagination.Params) ([]*Role, int64, error) {
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

	var roles []*Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*Role, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidRoleID
	}

	var role Role
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "deleted_at": nil}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *roleRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*Role, error) {
	if len(ids) == 0 {
		return []*Role{}, nil
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

	var roles []*Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, id string, dto *UpdateRoleDTO) (*Role, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidRoleID
	}

	set := bson.M{"updated_at": time.Now()}
	if dto.Name != "" {
		set["name"] = dto.Name
	}
	if dto.Description != "" {
		set["description"] = dto.Description
	}
	if dto.PermissionsIds != nil {
		set["permissions_ids"] = parseObjectIDs(dto.PermissionsIds)
	}
	if dto.ResourcesIds != nil {
		set["resources_ids"] = parseObjectIDs(dto.ResourcesIds)
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
		return nil, ErrRoleNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidRoleID
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
		return ErrRoleNotFound
	}

	return nil
}

// parseObjectIDs convierte []string a []primitive.ObjectID ignorando IDs inválidos
func parseObjectIDs(ids []string) []primitive.ObjectID {
	result := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			result = append(result, oid)
		}
	}
	return result
}
