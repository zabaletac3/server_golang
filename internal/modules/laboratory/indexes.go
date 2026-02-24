package laboratory

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// EnsureIndexes creates required indexes for the laboratory collections
func EnsureIndexes(ctx context.Context, db *database.MongoDB) error {
	// Lab Orders indexes
	ordersIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys: bson.D{{"patient_id", 1}, {"order_date", -1}},
		},
		{
			Keys: bson.D{{"tenant_id", 1}, {"status", 1}, {"order_date", -1}},
		},
		{
			Keys: bson.D{{"veterinarian_id", 1}, {"order_date", -1}},
		},
		{
			Keys: bson.D{{"test_type", 1}, {"order_date", -1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	ordersCollection := db.Collection("lab_orders")
	_, err := ordersCollection.Indexes().CreateMany(ctx, ordersIndexes, opts)
	if err != nil {
		return err
	}

	// Lab Tests indexes
	testsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"name", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{"category", 1}, {"active", 1}},
		},
	}

	testsCollection := db.Collection("lab_tests")
	_, err = testsCollection.Indexes().CreateMany(ctx, testsIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
