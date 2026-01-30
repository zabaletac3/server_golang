package payments

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByTenantID(ctx context.Context, tenantID string, limit int) ([]Payment, error)
	UpdateStatus(ctx context.Context, id string, status PaymentStatus, processedAt *time.Time, failureReason string) error
}

type paymentRepository struct {
	collection *mongo.Collection
}

func NewPaymentRepository(db *database.MongoDB) PaymentRepository {
	return &paymentRepository{
		collection: db.Collection("payments"),
	}
}

func (r *paymentRepository) Create(ctx context.Context, payment *Payment) error {
	_, err := r.collection.InsertOne(ctx, payment)
	return err
}

func (r *paymentRepository) FindByID(ctx context.Context, id string) (*Payment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidPaymentID
	}

	var payment Payment
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByTenantID(ctx context.Context, tenantID string, limit int) ([]Payment, error) {
	objectID, err := primitive.ObjectIDFromHex(tenantID)
	if err != nil {
		return nil, ErrInvalidTenantID
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}). // Más recientes primero
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{"tenant_id": objectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payments []Payment
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status PaymentStatus, processedAt *time.Time, failureReason string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidPaymentID
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if processedAt != nil {
		update["$set"].(bson.M)["processed_at"] = processedAt
	}

	if failureReason != "" {
		update["$set"].(bson.M)["failure_reason"] = failureReason
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrPaymentNotFound
	}

	return nil
}
