package medical_records

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

// MedicalRecordRepository defines the interface for medical record data access
type MedicalRecordRepository interface {
	// Medical Record CRUD
	Create(ctx context.Context, record *MedicalRecord) error
	FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*MedicalRecord, error)
	FindByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID, params pagination.Params) ([]MedicalRecord, int64, error)
	FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters MedicalRecordListFilters, params pagination.Params) ([]MedicalRecord, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Timeline
	FindTimeline(ctx context.Context, patientID, tenantID primitive.ObjectID, filters TimelineFilters) ([]TimelineEntry, int64, error)

	// Allergy CRUD
	CreateAllergy(ctx context.Context, allergy *Allergy) error
	FindAllergiesByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID) ([]Allergy, error)
	FindAllergyByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Allergy, error)
	UpdateAllergy(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	DeleteAllergy(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Medical History CRUD
	CreateHistory(ctx context.Context, history *MedicalHistory) error
	FindHistoryByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID) (*MedicalHistory, error)
	UpdateHistory(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	DeleteHistory(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Indexes
	EnsureIndexes(ctx context.Context) error
}

type medicalRecordRepository struct {
	recordsCollection   *mongo.Collection
	allergiesCollection *mongo.Collection
	historyCollection   *mongo.Collection
}

// NewMedicalRecordRepository creates a new medical record repository
func NewMedicalRecordRepository(db *database.MongoDB) MedicalRecordRepository {
	return &medicalRecordRepository{
		recordsCollection:   db.Collection("medical_records"),
		allergiesCollection: db.Collection("allergies"),
		historyCollection:   db.Collection("medical_histories"),
	}
}

// Medical Record methods

func (r *medicalRecordRepository) Create(ctx context.Context, record *MedicalRecord) error {
	result, err := r.recordsCollection.InsertOne(ctx, record)
	if err != nil {
		return err
	}
	record.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *medicalRecordRepository) FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*MedicalRecord, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var record MedicalRecord
	err := r.recordsCollection.FindOne(ctx, filter).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &record, nil
}

func (r *medicalRecordRepository) FindByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID, params pagination.Params) ([]MedicalRecord, int64, error) {
	filter := bson.M{
		"patient_id": patientID,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	// Count total
	total, err := r.recordsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.recordsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []MedicalRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *medicalRecordRepository) FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters MedicalRecordListFilters, params pagination.Params) ([]MedicalRecord, int64, error) {
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

	if filters.Type != "" {
		filter["type"] = filters.Type
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
			filter["created_at"] = dateFilter
		}
	}

	if filters.HasAttachments {
		filter["attachment_ids"] = bson.M{"$exists": true, "$ne": []string{}, "$not": bson.M{"$size": 0}}
	}

	// Count total
	total, err := r.recordsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.recordsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []MedicalRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *medicalRecordRepository) Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.recordsCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (r *medicalRecordRepository) Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
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

	result, err := r.recordsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (r *medicalRecordRepository) FindTimeline(ctx context.Context, patientID, tenantID primitive.ObjectID, filters TimelineFilters) ([]TimelineEntry, int64, error) {
	filter := bson.M{
		"patient_id": patientID,
		"tenant_id":  tenantID,
		"deleted_at": nil,
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
			filter["created_at"] = dateFilter
		}
	}

	if filters.RecordType != "" {
		filter["type"] = filters.RecordType
	}

	// Count total
	total, err := r.recordsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	limit := int64(50)
	skip := int64(0)
	if filters.Limit > 0 {
		limit = int64(filters.Limit)
	}
	if filters.Skip > 0 {
		skip = int64(filters.Skip)
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.recordsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []MedicalRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}

	// Convert to timeline entries
	entries := make([]TimelineEntry, len(records))
	for i, rec := range records {
		entries[i] = TimelineEntry{
			Date:        rec.CreatedAt,
			Type:        string(rec.Type),
			Description: rec.ChiefComplaint,
			RecordID:    rec.ID.Hex(),
		}
	}

	return entries, total, nil
}

// Allergy methods

func (r *medicalRecordRepository) CreateAllergy(ctx context.Context, allergy *Allergy) error {
	result, err := r.allergiesCollection.InsertOne(ctx, allergy)
	if err != nil {
		return err
	}
	allergy.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *medicalRecordRepository) FindAllergiesByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID) ([]Allergy, error) {
	filter := bson.M{
		"patient_id": patientID,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	cursor, err := r.allergiesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var allergies []Allergy
	if err := cursor.All(ctx, &allergies); err != nil {
		return nil, err
	}

	return allergies, nil
}

func (r *medicalRecordRepository) FindAllergyByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Allergy, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var allergy Allergy
	err := r.allergiesCollection.FindOne(ctx, filter).Decode(&allergy)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrAllergyNotFound
		}
		return nil, err
	}

	return &allergy, nil
}

func (r *medicalRecordRepository) UpdateAllergy(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.allergiesCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrAllergyNotFound
	}

	return nil
}

func (r *medicalRecordRepository) DeleteAllergy(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
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

	result, err := r.allergiesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrAllergyNotFound
	}

	return nil
}

// Medical History methods

func (r *medicalRecordRepository) CreateHistory(ctx context.Context, history *MedicalHistory) error {
	result, err := r.historyCollection.InsertOne(ctx, history)
	if err != nil {
		// Check for duplicate key error
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateHistory
		}
		return err
	}
	history.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *medicalRecordRepository) FindHistoryByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID) (*MedicalHistory, error) {
	filter := bson.M{
		"patient_id": patientID,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var history MedicalHistory
	err := r.historyCollection.FindOne(ctx, filter).Decode(&history)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrMedicalHistoryNotFound
		}
		return nil, err
	}

	return &history, nil
}

func (r *medicalRecordRepository) UpdateHistory(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.historyCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrMedicalHistoryNotFound
	}

	return nil
}

func (r *medicalRecordRepository) DeleteHistory(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
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

	result, err := r.historyCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrMedicalHistoryNotFound
	}

	return nil
}

// EnsureIndexes creates required indexes for the medical records collections
func (r *medicalRecordRepository) EnsureIndexes(ctx context.Context) error {
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
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := r.recordsCollection.Indexes().CreateMany(ctx, recordsIndexes, opts)
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
	}

	_, err = r.allergiesCollection.Indexes().CreateMany(ctx, allergiesIndexes, opts)
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

	_, err = r.historyCollection.Indexes().CreateMany(ctx, historyIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
