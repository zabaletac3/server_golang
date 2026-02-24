package vaccinations

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

// VaccinationRepository defines the interface for vaccination data access
type VaccinationRepository interface {
	// Vaccination CRUD
	Create(ctx context.Context, vaccination *Vaccination) error
	FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Vaccination, error)
	FindByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID, params pagination.Params) ([]Vaccination, int64, error)
	FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters VaccinationListFilters, params pagination.Params) ([]Vaccination, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Status updates
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status VaccinationStatus, tenantID primitive.ObjectID) error

	// Alerts
	FindDueVaccinations(ctx context.Context, tenantID primitive.ObjectID, days int) ([]Vaccination, error)
	FindOverdueVaccinations(ctx context.Context, tenantID primitive.ObjectID) ([]Vaccination, error)

	// Vaccine Catalog CRUD
	CreateVaccine(ctx context.Context, vaccine *Vaccine) error
	FindVaccineByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Vaccine, error)
	FindVaccines(ctx context.Context, tenantID primitive.ObjectID, filters VaccineListFilters) ([]Vaccine, error)
	UpdateVaccine(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	DeleteVaccine(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Indexes
	EnsureIndexes(ctx context.Context) error
}

type vaccinationRepository struct {
	vaccinationsCollection *mongo.Collection
	vaccinesCollection     *mongo.Collection
}

// NewVaccinationRepository creates a new vaccination repository
func NewVaccinationRepository(db *database.MongoDB) VaccinationRepository {
	return &vaccinationRepository{
		vaccinationsCollection: db.Collection("vaccinations"),
		vaccinesCollection:     db.Collection("vaccines"),
	}
}

// Vaccination methods

func (r *vaccinationRepository) Create(ctx context.Context, vaccination *Vaccination) error {
	result, err := r.vaccinationsCollection.InsertOne(ctx, vaccination)
	if err != nil {
		return err
	}
	vaccination.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *vaccinationRepository) FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Vaccination, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var vaccination Vaccination
	err := r.vaccinationsCollection.FindOne(ctx, filter).Decode(&vaccination)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrVaccinationNotFound
		}
		return nil, err
	}

	return &vaccination, nil
}

func (r *vaccinationRepository) FindByPatient(ctx context.Context, patientID, tenantID primitive.ObjectID, params pagination.Params) ([]Vaccination, int64, error) {
	filter := bson.M{
		"patient_id": patientID,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	// Count total
	total, err := r.vaccinationsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"application_date", -1}})

	cursor, err := r.vaccinationsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var vaccinations []Vaccination
	if err := cursor.All(ctx, &vaccinations); err != nil {
		return nil, 0, err
	}

	return vaccinations, total, nil
}

func (r *vaccinationRepository) FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters VaccinationListFilters, params pagination.Params) ([]Vaccination, int64, error) {
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

	if filters.VaccineName != "" {
		filter["vaccine_name"] = bson.M{"$regex": filters.VaccineName, "$options": "i"}
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
			filter["application_date"] = dateFilter
		}
	}

	if filters.DueSoon {
		now := time.Now()
		dueThreshold := now.AddDate(0, 0, 30)
		filter["next_due_date"] = bson.M{
			"$gte": now,
			"$lte": dueThreshold,
		}
	}

	if filters.Overdue {
		filter["next_due_date"] = bson.M{"$lt": time.Now()}
		filter["status"] = VaccinationStatusOverdue
	}

	// Count total
	total, err := r.vaccinationsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"application_date", -1}})

	cursor, err := r.vaccinationsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var vaccinations []Vaccination
	if err := cursor.All(ctx, &vaccinations); err != nil {
		return nil, 0, err
	}

	return vaccinations, total, nil
}

func (r *vaccinationRepository) Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.vaccinationsCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrVaccinationNotFound
	}

	return nil
}

func (r *vaccinationRepository) Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
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

	result, err := r.vaccinationsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrVaccinationNotFound
	}

	return nil
}

func (r *vaccinationRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status VaccinationStatus, tenantID primitive.ObjectID) error {
	updates := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.vaccinationsCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrVaccinationNotFound
	}

	return nil
}

func (r *vaccinationRepository) FindDueVaccinations(ctx context.Context, tenantID primitive.ObjectID, days int) ([]Vaccination, error) {
	now := time.Now()
	dueThreshold := now.AddDate(0, 0, days)

	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"next_due_date": bson.M{
			"$gte": now,
			"$lte": dueThreshold,
		},
	}

	cursor, err := r.vaccinationsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vaccinations []Vaccination
	if err := cursor.All(ctx, &vaccinations); err != nil {
		return nil, err
	}

	return vaccinations, nil
}

func (r *vaccinationRepository) FindOverdueVaccinations(ctx context.Context, tenantID primitive.ObjectID) ([]Vaccination, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"next_due_date": bson.M{"$lt": time.Now()},
	}

	cursor, err := r.vaccinationsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vaccinations []Vaccination
	if err := cursor.All(ctx, &vaccinations); err != nil {
		return nil, err
	}

	return vaccinations, nil
}

// Vaccine methods

func (r *vaccinationRepository) CreateVaccine(ctx context.Context, vaccine *Vaccine) error {
	result, err := r.vaccinesCollection.InsertOne(ctx, vaccine)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrVaccineNameExists
		}
		return err
	}
	vaccine.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *vaccinationRepository) FindVaccineByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Vaccine, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var vaccine Vaccine
	err := r.vaccinesCollection.FindOne(ctx, filter).Decode(&vaccine)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrVaccineNotFound
		}
		return nil, err
	}

	return &vaccine, nil
}

func (r *vaccinationRepository) FindVaccines(ctx context.Context, tenantID primitive.ObjectID, filters VaccineListFilters) ([]Vaccine, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	if filters.DoseNumber != "" {
		filter["dose_number"] = filters.DoseNumber
	}

	if filters.TargetSpecies != "" {
		filter["target_species"] = filters.TargetSpecies
	}

	if filters.Active != nil {
		filter["active"] = *filters.Active
	}

	if filters.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"manufacturer": bson.M{"$regex": filters.Search, "$options": "i"}},
		}
	}

	cursor, err := r.vaccinesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vaccines []Vaccine
	if err := cursor.All(ctx, &vaccines); err != nil {
		return nil, err
	}

	return vaccines, nil
}

func (r *vaccinationRepository) UpdateVaccine(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.vaccinesCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrVaccineNameExists
		}
		return err
	}

	if result.MatchedCount == 0 {
		return ErrVaccineNotFound
	}

	return nil
}

func (r *vaccinationRepository) DeleteVaccine(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
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

	result, err := r.vaccinesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrVaccineNotFound
	}

	return nil
}

// EnsureIndexes creates required indexes for the vaccinations collections
func (r *vaccinationRepository) EnsureIndexes(ctx context.Context) error {
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
	_, err := r.vaccinationsCollection.Indexes().CreateMany(ctx, vaccinationsIndexes, opts)
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

	_, err = r.vaccinesCollection.Indexes().CreateMany(ctx, vaccinesIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
