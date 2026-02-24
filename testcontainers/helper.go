package testcontainers

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestMongoDB holds the MongoDB container and client for integration tests
type TestMongoDB struct {
	Container *mongodb.MongoDBContainer
	Client    *mongo.Client
	Database  string
	URI       string
}

// SetupTestContainer starts a MongoDB container for integration tests
// Returns the container, client, and a cleanup function
func SetupTestContainer(ctx context.Context) (*TestMongoDB, func(), error) {
	// Start MongoDB container
	mongoContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:7"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start MongoDB container: %w", err)
	}

	// Get connection URI
	uri, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		cleanupContainer(ctx, mongoContainer)
		return nil, nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Create MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		cleanupContainer(ctx, mongoContainer)
		return nil, nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Connect with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Connect(connectCtx); err != nil {
		cleanupContainer(ctx, mongoContainer)
		return nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(connectCtx, nil); err != nil {
		cleanupContainer(ctx, mongoContainer)
		return nil, nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := "test_vetsify"

	testDB := &TestMongoDB{
		Container: mongoContainer,
		Client:    client,
		Database:  database,
		URI:       uri,
	}

	// Cleanup function
	cleanup := func() {
		cleanupContainer(ctx, mongoContainer)
		if client != nil {
			if err := client.Disconnect(ctx); err != nil {
				fmt.Printf("failed to disconnect MongoDB client: %v\n", err)
			}
		}
	}

	return testDB, cleanup, nil
}

// cleanupContainer terminates the container and handles errors
func cleanupContainer(ctx context.Context, container testcontainers.Container) {
	if container == nil {
		return
	}
	if err := container.Terminate(ctx); err != nil {
		fmt.Printf("failed to terminate container: %v\n", err)
	}
}

// SetupTestDatabase creates a test database with required collections and indexes
func SetupTestDatabase(ctx context.Context, testDB *TestMongoDB) error {
	db := testDB.Client.Database(testDB.Database)

	// Create collections with indexes
	collections := []struct {
		Name    string
		Indexes []mongo.IndexModel
	}{
		{
			Name: "tenants",
			Indexes: []mongo.IndexModel{
				{
					Keys: map[string]interface{}{"subscription.external_subscription_id": 1},
					Options: options.Index().SetUnique(true).SetSparse(true),
				},
				{
					Keys:    map[string]interface{}{"domain": 1},
					Options: options.Index().SetUnique(true).SetSparse(true),
				},
				{
					Keys: map[string]interface{}{"owner_id": 1},
				},
				{
					Keys: map[string]interface{}{"status": 1},
				},
				{
					Keys:    map[string]interface{}{"deleted_at": 1},
					Options: options.Index().SetSparse(true),
				},
			},
		},
		{
			Name: "appointments",
			Indexes: []mongo.IndexModel{
				{
					Keys: map[string]interface{}{
						"tenant_id":    1,
						"status":       1,
						"scheduled_at": 1,
					},
				},
				{
					Keys: map[string]interface{}{
						"veterinarian_id": 1,
						"scheduled_at":    1,
					},
					Options: options.Index().SetUnique(true).SetPartialFilterExpression(
						map[string]interface{}{"veterinarian_id": map[string]interface{}{"$exists": true}},
					),
				},
				{
					Keys: map[string]interface{}{
						"tenant_id":  1,
						"owner_id":   1,
						"scheduled_at": -1,
					},
				},
				{
					Keys: map[string]interface{}{"deleted_at": 1},
					Options: options.Index().SetSparse(true),
				},
			},
		},
		{
			Name: "appointment_status_transitions",
			Indexes: []mongo.IndexModel{
				{
					Keys: map[string]interface{}{
						"appointment_id": 1,
						"created_at":     -1,
					},
				},
			},
		},
		{
			Name: "users",
			Indexes: []mongo.IndexModel{
				{
					Keys:    map[string]interface{}{"email": 1},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys:    map[string]interface{}{"deleted_at": 1},
					Options: options.Index().SetSparse(true),
				},
			},
		},
		{
			Name: "roles",
			Indexes: []mongo.IndexModel{
				{
					Keys: map[string]interface{}{"tenant_id": 1, "name": 1},
					Options: options.Index().SetUnique(true),
				},
			},
		},
		{
			Name: "permissions",
			Indexes: []mongo.IndexModel{
				{
					Keys: map[string]interface{}{"tenant_id": 1, "resource_id": 1, "action": 1},
					Options: options.Index().SetUnique(true),
				},
			},
		},
		{
			Name: "resources",
			Indexes: []mongo.IndexModel{
				{
					Keys: map[string]interface{}{"tenant_id": 1, "name": 1},
					Options: options.Index().SetUnique(true),
				},
			},
		},
	}

	for _, coll := range collections {
		collection := db.Collection(coll.Name)
		if len(coll.Indexes) > 0 {
			opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
			_, err := collection.Indexes().CreateMany(ctx, coll.Indexes, opts)
			if err != nil {
				return fmt.Errorf("failed to create indexes for %s: %w", coll.Name, err)
			}
		}
	}

	return nil
}

// ClearDatabase removes all documents from all collections (for test isolation)
func ClearDatabase(ctx context.Context, client *mongo.Client, database string) error {
	db := client.Database(database)
	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return err
	}

	for _, coll := range collections {
		if _, err := db.Collection(coll).DeleteMany(ctx, map[string]interface{}{}); err != nil {
			return err
		}
	}

	return nil
}
