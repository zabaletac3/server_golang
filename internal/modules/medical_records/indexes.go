package medical_records

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// EnsureIndexes creates required indexes for the medical records collections
func EnsureIndexes(ctx context.Context, db *database.MongoDB) error {
	// Medical Records indexes
	recordsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"patient_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys: bson.D{{"veterinarian_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"type", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"appointment_id", 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	recordsCollection := db.Collection("medical_records")
	_, err := recordsCollection.Indexes().CreateMany(ctx, recordsIndexes, opts)
	if err != nil {
		return err
	}

	// Allergies indexes
	allergiesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"patient_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys: bson.D{{"severity", 1}},
		},
	}

	allergiesCollection := db.Collection("allergies")
	_, err = allergiesCollection.Indexes().CreateMany(ctx, allergiesIndexes, opts)
	if err != nil {
		return err
	}

	// Medical History indexes
	historyIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"patient_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
	}

	historyCollection := db.Collection("medical_histories")
	_, err = historyCollection.Indexes().CreateMany(ctx, historyIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
