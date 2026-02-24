package appointments

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// EnsureIndexes creates required indexes for the appointments collection
func EnsureIndexes(ctx context.Context, db *database.MongoDB) error {
	collection := db.Collection("appointments")

	indexes := []mongo.IndexModel{
		// Compound index for tenant-scoped queries with status and date
		{
			Keys: bson.D{
				{"tenant_id", 1},
				{"status", 1},
				{"scheduled_at", 1},
			},
		},
		// Unique compound index to prevent double-booking veterinarians
		{
			Keys: bson.D{
				{"veterinarian_id", 1},
				{"scheduled_at", 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(
				bson.D{{"veterinarian_id", bson.D{{"$exists", true}}}},
			),
		},
		// Index for owner-based queries
		{
			Keys: bson.D{
				{"tenant_id", 1},
				{"owner_id", 1},
				{"scheduled_at", -1},
			},
		},
		// Index for patient-based queries
		{
			Keys: bson.D{
				{"tenant_id", 1},
				{"patient_id", 1},
			},
		},
		// Index for veterinarian-based queries
		{
			Keys: bson.D{
				{"tenant_id", 1},
				{"veterinarian_id", 1},
				{"scheduled_at", -1},
			},
		},
		// Index for soft delete filtering
		{
			Keys:    bson.D{{"deleted_at", 1}},
			Options: options.Index().SetSparse(true),
		},
		// Index for status transitions lookup
		{
			Keys: bson.D{{"appointment_id", 1}, {"created_at", -1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		return fmt.Errorf("failed to create appointment indexes: %w", err)
	}

	// Create indexes for transition collection
	transitionCollection := db.Collection("appointment_status_transitions")
	transitionIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"appointment_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"tenant_id", 1}, {"created_at", -1}},
		},
	}

	_, err = transitionCollection.Indexes().CreateMany(ctx, transitionIndexes, opts)
	if err != nil {
		return fmt.Errorf("failed to create transition indexes: %w", err)
	}

	return nil
}
