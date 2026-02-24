package inventory

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// EnsureIndexes creates required indexes for the inventory collections
func EnsureIndexes(ctx context.Context, db *database.MongoDB) error {
	// Products indexes
	productsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"sku", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"barcode", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{"name", 1}},
		},
		{
			Keys: bson.D{{"category", 1}},
		},
		{
			Keys: bson.D{{"expiration_date", 1}},
		},
		{
			Keys: bson.D{{"active", 1}},
		},
		{
			Keys: bson.D{{"stock", 1}, {"min_stock", 1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	productsCollection := db.Collection("products")
	_, err := productsCollection.Indexes().CreateMany(ctx, productsIndexes, opts)
	if err != nil {
		return err
	}

	// Categories indexes
	categoriesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"name", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}

	categoriesCollection := db.Collection("product_categories")
	_, err = categoriesCollection.Indexes().CreateMany(ctx, categoriesIndexes, opts)
	if err != nil {
		return err
	}

	// Stock movements indexes
	movementsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"product_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"type", 1}, {"created_at", -1}},
		},
	}

	movementsCollection := db.Collection("stock_movements")
	_, err = movementsCollection.Indexes().CreateMany(ctx, movementsIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
