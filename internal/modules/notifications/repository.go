package notifications

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	FindByOwner(ctx context.Context, ownerID primitive.ObjectID, params pagination.Params) ([]Notification, int64, error)
	CountUnread(ctx context.Context, ownerID primitive.ObjectID) (int64, error)
	MarkAsRead(ctx context.Context, ownerID, notifID primitive.ObjectID) error
	MarkAllAsRead(ctx context.Context, ownerID primitive.ObjectID) error
	MarkPushSent(ctx context.Context, id primitive.ObjectID) error
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository(db *database.MongoDB) Repository {
	return &repository{
		collection: db.Collection("notifications"),
	}
}

func (r *repository) Create(ctx context.Context, n *Notification) error {
	_, err := r.collection.InsertOne(ctx, n)
	return err
}

func (r *repository) FindByOwner(ctx context.Context, ownerID primitive.ObjectID, params pagination.Params) ([]Notification, int64, error) {
	filter := bson.M{"owner_id": ownerID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(params.Skip).
		SetLimit(params.Limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []Notification
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *repository) CountUnread(ctx context.Context, ownerID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"owner_id": ownerID,
		"read":     false,
	})
}

func (r *repository) MarkAsRead(ctx context.Context, ownerID, notifID primitive.ObjectID) error {
	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": notifID, "owner_id": ownerID},
		bson.M{"$set": bson.M{"read": true, "read_at": now}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (r *repository) MarkAllAsRead(ctx context.Context, ownerID primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"owner_id": ownerID, "read": false},
		bson.M{"$set": bson.M{"read": true, "read_at": now}},
	)
	return err
}

func (r *repository) MarkPushSent(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"push_sent": true, "push_sent_at": now}},
	)
	return err
}

// --- Staff repository ---

type StaffRepository interface {
	CreateStaff(ctx context.Context, n *StaffNotification) error
	FindByUser(ctx context.Context, userID primitive.ObjectID, params pagination.Params) ([]StaffNotification, int64, error)
	CountUnreadStaff(ctx context.Context, userID primitive.ObjectID) (int64, error)
	MarkStaffAsRead(ctx context.Context, userID, notifID primitive.ObjectID) error
	MarkAllStaffAsRead(ctx context.Context, userID primitive.ObjectID) error
}

type staffRepository struct {
	collection *mongo.Collection
}

func NewStaffRepository(db *database.MongoDB) StaffRepository {
	return &staffRepository{
		collection: db.Collection("staff_notifications"),
	}
}

func (r *staffRepository) CreateStaff(ctx context.Context, n *StaffNotification) error {
	_, err := r.collection.InsertOne(ctx, n)
	return err
}

func (r *staffRepository) FindByUser(ctx context.Context, userID primitive.ObjectID, params pagination.Params) ([]StaffNotification, int64, error) {
	filter := bson.M{"user_id": userID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(params.Skip).
		SetLimit(params.Limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []StaffNotification
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *staffRepository) CountUnreadStaff(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"read":    false,
	})
}

func (r *staffRepository) MarkStaffAsRead(ctx context.Context, userID, notifID primitive.ObjectID) error {
	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": notifID, "user_id": userID},
		bson.M{"$set": bson.M{"read": true, "read_at": now}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (r *staffRepository) MarkAllStaffAsRead(ctx context.Context, userID primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"user_id": userID, "read": false},
		bson.M{"$set": bson.M{"read": true, "read_at": now}},
	)
	return err
}
