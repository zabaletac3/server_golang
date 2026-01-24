package tenant

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eren_dev/go_server/internal/shared/database"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	FindByID(ctx context.Context, id string) (*Tenant, error)
	FindAll(ctx context.Context) ([]Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
}

type tenantRepository struct {
	collection *mongo.Collection
}

func NewTenantRepository(db *database.MongoDB) TenantRepository {
	return &tenantRepository{
		collection: db.Collection("tenants"),
	}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *Tenant) error {
	_, err := r.collection.InsertOne(ctx, tenant)
	return err
}

func (r *tenantRepository) FindByID(ctx context.Context, id string) (*Tenant, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidTenantID
	}

	var tenant Tenant
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&tenant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) FindAll(ctx context.Context) ([]Tenant, error) {
	filter := bson.M{
		"deleted_at": nil,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tenants []Tenant
	if err = cursor.All(ctx, &tenants); err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *Tenant) error {
	tenant.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": tenant.ID},
		bson.M{"$set": tenant},
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *tenantRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidTenantID
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"status":     Inactive,
			"deleted_at": now,
		}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrTenantNotFound
	}
	return nil
}