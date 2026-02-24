package vaccinations

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// EnsureIndexes creates required indexes for the vaccinations collections
func EnsureIndexes(ctx context.Context, db *database.MongoDB) error {
	// Vaccinations indexes
	vaccinationsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys: bson.D{{"patient_id", 1}, {"application_date", -1}},
		},
		{
			Keys: bson.D{{"tenant_id", 1}, {"status", 1}, {"next_due_date", 1}},
		},
		{
			Keys: bson.D{{"next_due_date", 1}},
		},
		{
			Keys: bson.D{{"veterinarian_id", 1}, {"application_date", -1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	vaccinationsCollection := db.Collection("vaccinations")
	_, err := vaccinationsCollection.Indexes().CreateMany(ctx, vaccinationsIndexes, opts)
	if err != nil {
		return err
	}

	// Vaccines indexes
	vaccinesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"name", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{"dose_number", 1}, {"active", 1}},
		},
	}

	vaccinesCollection := db.Collection("vaccines")
	_, err = vaccinesCollection.Indexes().CreateMany(ctx, vaccinesIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
