package audit

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// Repository defines the interface for audit log data access
type Repository interface {
	Create(ctx context.Context, event *AuditEvent) error
	FindByFilter(ctx context.Context, filter AuditFilter) ([]AuditEvent, int64, error)
	FindByResource(ctx context.Context, tenantID, resourceID primitive.ObjectID, resource string) ([]AuditEvent, error)
	FindByUser(ctx context.Context, tenantID, userID primitive.ObjectID) ([]AuditEvent, error)
	FindByEventType(ctx context.Context, tenantID primitive.ObjectID, eventType EventType) ([]AuditEvent, error)
}

type repository struct {
	collection *mongo.Collection
}

// NewRepository creates a new audit repository
func NewRepository(db *database.MongoDB) Repository {
	return &repository{
		collection: db.Collection("audit_logs"),
	}
}

func (r *repository) Create(ctx context.Context, event *AuditEvent) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *repository) FindByFilter(ctx context.Context, filter AuditFilter) ([]AuditEvent, int64, error) {
	bsonFilter := bson.M{"tenant_id": filter.TenantID}

	if filter.UserID != nil {
		bsonFilter["user_id"] = *filter.UserID
	}

	if filter.EventType != nil {
		bsonFilter["event_type"] = string(*filter.EventType)
	}

	if filter.Resource != nil {
		bsonFilter["resource"] = *filter.Resource
	}

	if filter.DateFrom != nil || filter.DateTo != nil {
		dateFilter := bson.M{}
		if filter.DateFrom != nil {
			dateFilter["$gte"] = *filter.DateFrom
		}
		if filter.DateTo != nil {
			dateFilter["$lte"] = *filter.DateTo
		}
		bsonFilter["created_at"] = dateFilter
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination
	skip := int64(filter.Skip)
	limit := int64(filter.Limit)
	if limit == 0 {
		limit = 50
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{"created_at", -1}}) // Newest first

	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var events []AuditEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *repository) FindByResource(ctx context.Context, tenantID, resourceID primitive.ObjectID, resource string) ([]AuditEvent, error) {
	filter := bson.M{
		"tenant_id":   tenantID,
		"resource":    resource,
		"resource_id": resourceID,
	}

	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []AuditEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *repository) FindByUser(ctx context.Context, tenantID, userID primitive.ObjectID) ([]AuditEvent, error) {
	filter := bson.M{
		"tenant_id": tenantID,
		"user_id":   userID,
	}

	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetLimit(100)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []AuditEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *repository) FindByEventType(ctx context.Context, tenantID primitive.ObjectID, eventType EventType) ([]AuditEvent, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"event_type": eventType,
	}

	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}).
		SetLimit(100)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []AuditEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// EnsureIndexes creates required indexes for the audit_logs collection
func EnsureIndexes(ctx context.Context, db *database.MongoDB) error {
	collection := db.Collection("audit_logs")

	indexes := []mongo.IndexModel{
		// Index for tenant-based queries
		{
			Keys: bson.D{{"tenant_id", 1}, {"created_at", -1}},
		},
		// Index for user-based queries
		{
			Keys: bson.D{{"tenant_id", 1}, {"user_id", 1}, {"created_at", -1}},
		},
		// Index for resource-based queries
		{
			Keys: bson.D{{"tenant_id", 1}, {"resource", 1}, {"resource_id", 1}},
		},
		// Index for event type queries
		{
			Keys: bson.D{{"tenant_id", 1}, {"event_type", 1}, {"created_at", -1}},
		},
		// TTL index for automatic cleanup (optional, 1 year retention)
		{
			Keys:    bson.D{{"created_at", 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(365 * 24 * 60 * 60)),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		return err
	}

	return nil
}
