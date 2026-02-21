package appointments

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
)

// AppointmentRepository defines the interface for appointment data access
type AppointmentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, appointment *Appointment) error
	FindByID(ctx context.Context, id primitive.ObjectID, tenantIDs []primitive.ObjectID) (*Appointment, error)
	List(ctx context.Context, filters appointmentFilters, tenantIDs []primitive.ObjectID, page, limit int) ([]Appointment, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantIDs []primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID, tenantIDs []primitive.ObjectID) error

	// Business-specific methods
	FindByDateRange(ctx context.Context, from, to time.Time, tenantIDs []primitive.ObjectID) ([]Appointment, error)
	FindByPatient(ctx context.Context, patientID primitive.ObjectID, tenantIDs []primitive.ObjectID, page, limit int) ([]Appointment, int64, error)
	FindByOwner(ctx context.Context, ownerID primitive.ObjectID, tenantIDs []primitive.ObjectID, page, limit int) ([]Appointment, int64, error)
	FindByVeterinarian(ctx context.Context, vetID primitive.ObjectID, from, to time.Time, tenantIDs []primitive.ObjectID) ([]Appointment, error)
	CheckConflicts(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantIDs []primitive.ObjectID) (bool, error)

	// Status transitions
	CreateStatusTransition(ctx context.Context, transition *AppointmentStatusTransition) error
	GetStatusHistory(ctx context.Context, appointmentID primitive.ObjectID) ([]AppointmentStatusTransition, error)

	// Analytics and reporting
	CountByStatus(ctx context.Context, status string, tenantIDs []primitive.ObjectID) (int64, error)
	FindUpcoming(ctx context.Context, tenantIDs []primitive.ObjectID, hours int) ([]Appointment, error)
}

// appointmentRepository implements AppointmentRepository interface
type appointmentRepository struct {
	collection           *mongo.Collection
	transitionCollection *mongo.Collection
}

// NewAppointmentRepository creates a new appointment repository
func NewAppointmentRepository(db *database.MongoDB) AppointmentRepository {
	return &appointmentRepository{
		collection:           db.Database().Collection("appointments"),
		transitionCollection: db.Database().Collection("appointment_status_transitions"),
	}
}

// Create inserts a new appointment
func (r *appointmentRepository) Create(ctx context.Context, appointment *Appointment) error {
	result, err := r.collection.InsertOne(ctx, appointment)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAppointmentAlreadyExists
		}
		return err
	}

	appointment.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID finds an appointment by ID
func (r *appointmentRepository) FindByID(ctx context.Context, id primitive.ObjectID, tenantIDs []primitive.ObjectID) (*Appointment, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
	}

	var appointment Appointment
	err := r.collection.FindOne(ctx, filter).Decode(&appointment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	return &appointment, nil
}

// List returns appointments with filters and pagination
func (r *appointmentRepository) List(ctx context.Context, filters appointmentFilters, tenantIDs []primitive.ObjectID, page, limit int) ([]Appointment, int64, error) {
	filter := r.buildFilter(filters, tenantIDs)

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Build options
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "scheduled_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var appointments []Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}

// Update updates an appointment
func (r *appointmentRepository) Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantIDs []primitive.ObjectID) error {
	filter := bson.M{
		"_id":        id,
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
	}

	updates["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrAppointmentNotFound
	}

	return nil
}

// Delete soft deletes an appointment
func (r *appointmentRepository) Delete(ctx context.Context, id primitive.ObjectID, tenantIDs []primitive.ObjectID) error {
	filter := bson.M{
		"_id":        id,
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrAppointmentNotFound
	}

	return nil
}

// FindByDateRange finds appointments within a date range
func (r *appointmentRepository) FindByDateRange(ctx context.Context, from, to time.Time, tenantIDs []primitive.ObjectID) ([]Appointment, error) {
	filter := bson.M{
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
		"scheduled_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "scheduled_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments []Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, err
	}

	return appointments, nil
}

// FindByPatient finds appointments for a specific patient with pagination
func (r *appointmentRepository) FindByPatient(ctx context.Context, patientID primitive.ObjectID, tenantIDs []primitive.ObjectID, page, limit int) ([]Appointment, int64, error) {
	filter := bson.M{
		"patient_id": patientID,
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "scheduled_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var appointments []Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}

// FindByOwner finds appointments for a specific owner with pagination
func (r *appointmentRepository) FindByOwner(ctx context.Context, ownerID primitive.ObjectID, tenantIDs []primitive.ObjectID, page, limit int) ([]Appointment, int64, error) {
	filter := bson.M{
		"owner_id":   ownerID,
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "scheduled_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var appointments []Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}

// FindByVeterinarian finds appointments for a specific veterinarian within a date range
func (r *appointmentRepository) FindByVeterinarian(ctx context.Context, vetID primitive.ObjectID, from, to time.Time, tenantIDs []primitive.ObjectID) ([]Appointment, error) {
	filter := bson.M{
		"veterinarian_id": vetID,
		"tenant_ids":      bson.M{"$in": tenantIDs},
		"deleted_at":      nil,
		"scheduled_at": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "scheduled_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments []Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, err
	}

	return appointments, nil
}

// CheckConflicts checks if there are conflicting appointments
func (r *appointmentRepository) CheckConflicts(ctx context.Context, vetID primitive.ObjectID, scheduledAt time.Time, duration int, excludeID *primitive.ObjectID, tenantIDs []primitive.ObjectID) (bool, error) {
	endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)

	filter := bson.M{
		"veterinarian_id": vetID,
		"tenant_ids":      bson.M{"$in": tenantIDs},
		"deleted_at":      nil,
		"status": bson.M{"$nin": []string{
			AppointmentStatusCancelled,
			AppointmentStatusNoShow,
		}},
		"$or": []bson.M{
			// Appointment starts during our time slot
			{
				"scheduled_at": bson.M{
					"$gte": scheduledAt,
					"$lt":  endTime,
				},
			},
			// Appointment ends during our time slot
			{
				"$expr": bson.M{
					"$and": []bson.M{
						{
							"$lte": bson.A{
								"$scheduled_at",
								scheduledAt,
							},
						},
						{
							"$gt": bson.A{
								bson.M{
									"$add": bson.A{
										"$scheduled_at",
										bson.M{"$multiply": bson.A{"$duration", 60000}},
									},
								},
								scheduledAt,
							},
						},
					},
				},
			},
		},
	}

	// Exclude specific appointment if provided (for updates)
	if excludeID != nil {
		filter["_id"] = bson.M{"$ne": *excludeID}
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateStatusTransition creates a status transition record
func (r *appointmentRepository) CreateStatusTransition(ctx context.Context, transition *AppointmentStatusTransition) error {
	result, err := r.transitionCollection.InsertOne(ctx, transition)
	if err != nil {
		return err
	}

	transition.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetStatusHistory gets the status history for an appointment
func (r *appointmentRepository) GetStatusHistory(ctx context.Context, appointmentID primitive.ObjectID) ([]AppointmentStatusTransition, error) {
	filter := bson.M{
		"appointment_id": appointmentID,
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.transitionCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transitions []AppointmentStatusTransition
	if err = cursor.All(ctx, &transitions); err != nil {
		return nil, err
	}

	return transitions, nil
}

// CountByStatus counts appointments by status
func (r *appointmentRepository) CountByStatus(ctx context.Context, status string, tenantIDs []primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"status":     status,
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
	}

	return r.collection.CountDocuments(ctx, filter)
}

// FindUpcoming finds upcoming appointments within specified hours
func (r *appointmentRepository) FindUpcoming(ctx context.Context, tenantIDs []primitive.ObjectID, hours int) ([]Appointment, error) {
	now := time.Now()
	futureTime := now.Add(time.Duration(hours) * time.Hour)

	filter := bson.M{
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
		"status":     bson.M{"$in": []string{AppointmentStatusScheduled, AppointmentStatusConfirmed}},
		"scheduled_at": bson.M{
			"$gte": now,
			"$lte": futureTime,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "scheduled_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments []Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, err
	}

	return appointments, nil
}

// buildFilter constructs MongoDB filter from appointmentFilters
func (r *appointmentRepository) buildFilter(filters appointmentFilters, tenantIDs []primitive.ObjectID) bson.M {
	filter := bson.M{
		"tenant_ids": bson.M{"$in": tenantIDs},
		"deleted_at": nil,
	}

	if len(filters.Status) > 0 {
		filter["status"] = bson.M{"$in": filters.Status}
	}

	if len(filters.Type) > 0 {
		filter["type"] = bson.M{"$in": filters.Type}
	}

	if filters.VeterinarianID != nil {
		filter["veterinarian_id"] = *filters.VeterinarianID
	}

	if filters.PatientID != nil {
		filter["patient_id"] = *filters.PatientID
	}

	if filters.OwnerID != nil {
		filter["owner_id"] = *filters.OwnerID
	}

	if filters.Priority != nil {
		filter["priority"] = *filters.Priority
	}

	// Date range filter
	if filters.DateFrom != nil || filters.DateTo != nil {
		dateFilter := bson.M{}
		if filters.DateFrom != nil {
			dateFilter["$gte"] = *filters.DateFrom
		}
		if filters.DateTo != nil {
			dateFilter["$lte"] = *filters.DateTo
		}
		filter["scheduled_at"] = dateFilter
	}

	return filter
}
