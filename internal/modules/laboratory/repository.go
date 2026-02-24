package laboratory

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

// LabOrderRepository defines the interface for lab order data access
type LabOrderRepository interface {
	// Lab Order CRUD
	Create(ctx context.Context, order *LabOrder) error
	FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*LabOrder, error)
	FindByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID, params pagination.Params) ([]LabOrder, int64, error)
	FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters LabOrderListFilters, params pagination.Params) ([]LabOrder, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Status updates
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status LabOrderStatus, tenantID primitive.ObjectID) error

	// Alerts
	FindOverdueOrders(ctx context.Context, tenantID primitive.ObjectID, turnaroundDays int) ([]LabOrder, error)
	FindReadyForPickup(ctx context.Context, tenantID primitive.ObjectID) ([]LabOrder, error)

	// Lab Test Catalog CRUD
	CreateLabTest(ctx context.Context, test *LabTest) error
	FindLabTestByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*LabTest, error)
	FindLabTests(ctx context.Context, tenantID primitive.ObjectID, filters LabTestListFilters) ([]LabTest, error)
	UpdateLabTest(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	DeleteLabTest(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Indexes
	EnsureIndexes(ctx context.Context) error
}

type labOrderRepository struct {
	ordersCollection *mongo.Collection
	testsCollection  *mongo.Collection
}

// NewLabOrderRepository creates a new lab order repository
func NewLabOrderRepository(db *database.MongoDB) LabOrderRepository {
	return &labOrderRepository{
		ordersCollection: db.Collection("lab_orders"),
		testsCollection:  db.Collection("lab_tests"),
	}
}

// Lab Order methods

func (r *labOrderRepository) Create(ctx context.Context, order *LabOrder) error {
	result, err := r.ordersCollection.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	order.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *labOrderRepository) FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*LabOrder, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var order LabOrder
	err := r.ordersCollection.FindOne(ctx, filter).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrLabOrderNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (r *labOrderRepository) FindByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID, params pagination.Params) ([]LabOrder, int64, error) {
	filter := bson.M{
		"patient_id": patientID,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	// Count total
	total, err := r.ordersCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"order_date", -1}})

	cursor, err := r.ordersCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []LabOrder
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *labOrderRepository) FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters LabOrderListFilters, params pagination.Params) ([]LabOrder, int64, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	if filters.PatientID != "" {
		if patientID, err := primitive.ObjectIDFromHex(filters.PatientID); err == nil {
			filter["patient_id"] = patientID
		}
	}

	if filters.VeterinarianID != "" {
		if vetID, err := primitive.ObjectIDFromHex(filters.VeterinarianID); err == nil {
			filter["veterinarian_id"] = vetID
		}
	}

	if filters.Status != "" {
		filter["status"] = filters.Status
	}

	if filters.TestType != "" {
		filter["test_type"] = filters.TestType
	}

	if filters.LabID != "" {
		filter["lab_id"] = filters.LabID
	}

	if filters.DateFrom != "" || filters.DateTo != "" {
		dateFilter := bson.M{}
		if filters.DateFrom != "" {
			if df, err := time.Parse(time.RFC3339, filters.DateFrom); err == nil {
				dateFilter["$gte"] = df
			}
		}
		if filters.DateTo != "" {
			if dt, err := time.Parse(time.RFC3339, filters.DateTo); err == nil {
				dateFilter["$lte"] = dt
			}
		}
		if len(dateFilter) > 0 {
			filter["order_date"] = dateFilter
		}
	}

	// Count total
	total, err := r.ordersCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"order_date", -1}})

	cursor, err := r.ordersCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []LabOrder
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *labOrderRepository) Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.ordersCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrLabOrderNotFound
	}

	return nil
}

func (r *labOrderRepository) Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
	filter := bson.M{
		"_id":       id,
		"tenant_id": tenantID,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := r.ordersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrLabOrderNotFound
	}

	return nil
}

func (r *labOrderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status LabOrderStatus, tenantID primitive.ObjectID) error {
	updates := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}

	// Set status-specific dates
	switch status {
	case LabOrderStatusCollected:
		now := time.Now()
		updates["collection_date"] = now
	case LabOrderStatusProcessed:
		now := time.Now()
		updates["result_date"] = now
	}

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.ordersCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrLabOrderNotFound
	}

	return nil
}

func (r *labOrderRepository) FindOverdueOrders(ctx context.Context, tenantID primitive.ObjectID, turnaroundDays int) ([]LabOrder, error) {
	// Find orders that are not processed and are past their turnaround time
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"status": bson.M{
			"$nin": []string{string(LabOrderStatusProcessed)},
		},
	}

	cursor, err := r.ordersCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []LabOrder
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	// Filter orders that are actually overdue
	overdueOrders := make([]LabOrder, 0)
	for _, o := range orders {
		if o.IsOverdue(turnaroundDays) {
			overdueOrders = append(overdueOrders, o)
		}
	}

	return overdueOrders, nil
}

func (r *labOrderRepository) FindReadyForPickup(ctx context.Context, tenantID primitive.ObjectID) ([]LabOrder, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"status":     LabOrderStatusReceived,
	}

	cursor, err := r.ordersCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []LabOrder
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

// Lab Test methods

func (r *labOrderRepository) CreateLabTest(ctx context.Context, test *LabTest) error {
	result, err := r.testsCollection.InsertOne(ctx, test)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrLabTestNameExists
		}
		return err
	}
	test.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *labOrderRepository) FindLabTestByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*LabTest, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var test LabTest
	err := r.testsCollection.FindOne(ctx, filter).Decode(&test)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrLabTestNotFound
		}
		return nil, err
	}

	return &test, nil
}

func (r *labOrderRepository) FindLabTests(ctx context.Context, tenantID primitive.ObjectID, filters LabTestListFilters) ([]LabTest, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	if filters.Category != "" {
		filter["category"] = filters.Category
	}

	if filters.Active != nil {
		filter["active"] = *filters.Active
	}

	if filters.Search != "" {
		filter["name"] = bson.M{"$regex": filters.Search, "$options": "i"}
	}

	cursor, err := r.testsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tests []LabTest
	if err := cursor.All(ctx, &tests); err != nil {
		return nil, err
	}

	return tests, nil
}

func (r *labOrderRepository) UpdateLabTest(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.testsCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrLabTestNameExists
		}
		return err
	}

	if result.MatchedCount == 0 {
		return ErrLabTestNotFound
	}

	return nil
}

func (r *labOrderRepository) DeleteLabTest(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
	filter := bson.M{
		"_id":       id,
		"tenant_id": tenantID,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := r.testsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrLabTestNotFound
	}

	return nil
}

// EnsureIndexes creates required indexes for the laboratory collections
func (r *labOrderRepository) EnsureIndexes(ctx context.Context) error {
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
	_, err := r.ordersCollection.Indexes().CreateMany(ctx, ordersIndexes, opts)
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

	_, err = r.testsCollection.Indexes().CreateMany(ctx, testsIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
