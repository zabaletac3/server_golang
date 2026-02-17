package patients

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

// --- Species Repository ---

type SpeciesRepository interface {
	Create(ctx context.Context, s *Species) error
	FindByNormalizedName(ctx context.Context, tenantID primitive.ObjectID, normalized string) (*Species, error)
	FindAllByTenant(ctx context.Context, tenantID primitive.ObjectID) ([]Species, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*Species, error)
}

type speciesRepository struct {
	collection *mongo.Collection
}

func NewSpeciesRepository(db *database.MongoDB) SpeciesRepository {
	return &speciesRepository{
		collection: db.Collection("species"),
	}
}

func (r *speciesRepository) Create(ctx context.Context, s *Species) error {
	_, err := r.collection.InsertOne(ctx, s)
	return err
}

func (r *speciesRepository) FindByNormalizedName(ctx context.Context, tenantID primitive.ObjectID, normalized string) (*Species, error) {
	var s Species
	err := r.collection.FindOne(ctx, bson.M{
		"tenant_id":       tenantID,
		"normalized_name": normalized,
		"deleted_at":      nil,
	}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *speciesRepository) FindAllByTenant(ctx context.Context, tenantID primitive.ObjectID) ([]Species, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []Species
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *speciesRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*Species, error) {
	var s Species
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrSpeciesNotFound
		}
		return nil, err
	}
	return &s, nil
}

// --- Patient Repository ---

type PatientRepository interface {
	Create(ctx context.Context, p *Patient) error
	FindAll(ctx context.Context, tenantID primitive.ObjectID, params pagination.Params) ([]Patient, int64, error)
	FindByID(ctx context.Context, tenantID primitive.ObjectID, id string) (*Patient, error)
	FindByOwner(ctx context.Context, tenantID primitive.ObjectID, ownerID primitive.ObjectID, params pagination.Params) ([]Patient, int64, error)
	Update(ctx context.Context, tenantID primitive.ObjectID, id string, dto *UpdatePatientDTO) (*Patient, error)
	Delete(ctx context.Context, tenantID primitive.ObjectID, id string) error
}

type patientRepository struct {
	collection *mongo.Collection
}

func NewPatientRepository(db *database.MongoDB) PatientRepository {
	return &patientRepository{
		collection: db.Collection("patients"),
	}
}

func (r *patientRepository) Create(ctx context.Context, p *Patient) error {
	_, err := r.collection.InsertOne(ctx, p)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrMicrochipExists
		}
		return err
	}
	return nil
}

func (r *patientRepository) FindAll(ctx context.Context, tenantID primitive.ObjectID, params pagination.Params) ([]Patient, int64, error) {
	filter := bson.M{"tenant_id": tenantID, "deleted_at": nil}

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

	var results []Patient
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *patientRepository) FindByID(ctx context.Context, tenantID primitive.ObjectID, id string) (*Patient, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidPatientID
	}

	var p Patient
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        oid,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPatientNotFound
		}
		return nil, err
	}

	return &p, nil
}

func (r *patientRepository) FindByOwner(ctx context.Context, tenantID primitive.ObjectID, ownerID primitive.ObjectID, params pagination.Params) ([]Patient, int64, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"owner_id":   ownerID,
		"deleted_at": nil,
	}

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

	var results []Patient
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *patientRepository) Update(ctx context.Context, tenantID primitive.ObjectID, id string, dto *UpdatePatientDTO) (*Patient, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidPatientID
	}

	set := bson.M{"updated_at": time.Now()}

	if dto.Name != "" {
		set["name"] = dto.Name
	}
	if dto.SpeciesID != "" {
		speciesOID, err := primitive.ObjectIDFromHex(dto.SpeciesID)
		if err != nil {
			return nil, ErrInvalidSpeciesID
		}
		set["species_id"] = speciesOID
	}
	if dto.Breed != "" {
		set["breed"] = dto.Breed
	}
	if dto.Color != "" {
		set["color"] = dto.Color
	}
	if dto.BirthDate != nil {
		set["birth_date"] = dto.BirthDate
	}
	if dto.Gender != "" {
		set["gender"] = dto.Gender
	}
	if dto.Weight > 0 {
		set["weight"] = dto.Weight
	}
	if dto.Microchip != "" {
		set["microchip"] = dto.Microchip
	}
	if dto.Sterilized != nil {
		set["sterilized"] = *dto.Sterilized
	}
	if dto.AvatarURL != "" {
		set["avatar_url"] = dto.AvatarURL
	}
	if dto.Notes != "" {
		set["notes"] = dto.Notes
	}
	if dto.Active != nil {
		set["active"] = *dto.Active
	}

	filter := bson.M{"_id": oid, "tenant_id": tenantID, "deleted_at": nil}

	result, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": set})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrMicrochipExists
		}
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, ErrPatientNotFound
	}

	return r.FindByID(ctx, tenantID, id)
}

func (r *patientRepository) Delete(ctx context.Context, tenantID primitive.ObjectID, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidPatientID
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": oid, "tenant_id": tenantID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": time.Now()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrPatientNotFound
	}

	return nil
}
