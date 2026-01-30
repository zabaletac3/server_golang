package plans

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eren_dev/go_server/internal/shared/database"
)

type PlanRepository interface {
	Create(ctx context.Context, plan *Plan) error
	FindByID(ctx context.Context, id string) (*Plan, error)
	FindAll(ctx context.Context) ([]Plan, error)
	FindVisible(ctx context.Context) ([]Plan, error)
	Update(ctx context.Context, plan *Plan) error
	Delete(ctx context.Context, id string) error
}

type planRepository struct {
	collection *mongo.Collection
}

func NewPlanRepository(db *database.MongoDB) PlanRepository {
	return &planRepository{
		collection: db.Collection("plans"),
	}
}

func (r *planRepository) Create(ctx context.Context, plan *Plan) error {
	_, err := r.collection.InsertOne(ctx, plan)
	return err
}

func (r *planRepository) FindByID(ctx context.Context, id string) (*Plan, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidPlanID
	}

	var plan Plan
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&plan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) FindAll(ctx context.Context) ([]Plan, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []Plan
	if err = cursor.All(ctx, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *planRepository) FindVisible(ctx context.Context) ([]Plan, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"is_visible": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []Plan
	if err = cursor.All(ctx, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *planRepository) Update(ctx context.Context, plan *Plan) error {
	plan.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": plan.ID}
	update := bson.M{"$set": plan}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrPlanNotFound
	}
	
	return nil
}

func (r *planRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidPlanID
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return ErrPlanNotFound
	}
	
	return nil
}
