package tenant

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// EnsureIndexes creates required indexes for the tenants collection
func EnsureIndexes(ctx context.Context, db *database.MongoDB) error {
	collection := db.Collection("tenants")

	indexes := []mongo.IndexModel{
		// Unique index on external subscription ID for fast lookup
		{
			Keys:    bson.D{{"subscription.external_subscription_id", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		// Unique index on domain for tenant identification
		{
			Keys:    bson.D{{"domain", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		// Index on owner_id for owner-based queries
		{
			Keys: bson.D{{"owner_id", 1}},
		},
		// Index on status for filtering active/trial tenants
		{
			Keys: bson.D{{"status", 1}},
		},
		// Compound index for soft delete filtering
		{
			Keys:    bson.D{{"deleted_at", 1}},
			Options: options.Index().SetSparse(true),
		},
		// Index on email for tenant email lookups
		{
			Keys: bson.D{{"email", 1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		return fmt.Errorf("failed to create tenant indexes: %w", err)
	}

	return nil
}
