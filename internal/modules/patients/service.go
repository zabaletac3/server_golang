package patients

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type PatientService struct {
	repo           PatientRepository
	speciesService *SpeciesService
}

func NewService(repo PatientRepository, speciesService *SpeciesService) *PatientService {
	return &PatientService{
		repo:           repo,
		speciesService: speciesService,
	}
}

func (s *PatientService) Create(ctx context.Context, tenantID primitive.ObjectID, dto *CreatePatientDTO) (*PatientResponse, error) {
	ownerID, err := primitive.ObjectIDFromHex(dto.OwnerID)
	if err != nil {
		return nil, ErrInvalidOwnerID
	}
	speciesID, err := primitive.ObjectIDFromHex(dto.SpeciesID)
	if err != nil {
		return nil, ErrInvalidSpeciesID
	}

	// Validate species exists
	if _, err := s.speciesService.repo.FindByID(ctx, speciesID); err != nil {
		return nil, err
	}

	now := time.Now()
	patient := &Patient{
		ID:         primitive.NewObjectID(),
		TenantID:   tenantID,
		OwnerID:    ownerID,
		SpeciesID:  speciesID,
		Name:       dto.Name,
		Breed:      dto.Breed,
		Color:      dto.Color,
		BirthDate:  dto.BirthDate,
		Gender:     dto.Gender,
		Weight:     dto.Weight,
		Microchip:  dto.Microchip,
		Sterilized: dto.Sterilized,
		AvatarURL:  dto.AvatarURL,
		Notes:      dto.Notes,
		Active:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, patient); err != nil {
		return nil, err
	}

	resp := toPatientResponse(patient)
	return &resp, nil
}

func (s *PatientService) FindAll(ctx context.Context, tenantID primitive.ObjectID, params pagination.Params) (*PaginatedPatientsResponse, error) {
	items, total, err := s.repo.FindAll(ctx, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]PatientResponse, len(items))
	for i, p := range items {
		data[i] = toPatientResponse(&p)
	}

	return &PaginatedPatientsResponse{
		Data:       data,
		Pagination: pagination.NewPaginationInfo(params, total),
	}, nil
}

func (s *PatientService) FindByID(ctx context.Context, tenantID primitive.ObjectID, id string) (*PatientResponse, error) {
	p, err := s.repo.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	resp := toPatientResponse(p)
	return &resp, nil
}

func (s *PatientService) FindByOwner(ctx context.Context, tenantID primitive.ObjectID, ownerID primitive.ObjectID, params pagination.Params) (*PaginatedPatientsResponse, error) {
	items, total, err := s.repo.FindByOwner(ctx, tenantID, ownerID, params)
	if err != nil {
		return nil, err
	}

	data := make([]PatientResponse, len(items))
	for i, p := range items {
		data[i] = toPatientResponse(&p)
	}

	return &PaginatedPatientsResponse{
		Data:       data,
		Pagination: pagination.NewPaginationInfo(params, total),
	}, nil
}

func (s *PatientService) Update(ctx context.Context, tenantID primitive.ObjectID, id string, dto *UpdatePatientDTO) (*PatientResponse, error) {
	// Validate species if being updated
	if dto.SpeciesID != "" {
		speciesOID, err := primitive.ObjectIDFromHex(dto.SpeciesID)
		if err != nil {
			return nil, ErrInvalidSpeciesID
		}
		if _, err := s.speciesService.repo.FindByID(ctx, speciesOID); err != nil {
			return nil, err
		}
	}

	p, err := s.repo.Update(ctx, tenantID, id, dto)
	if err != nil {
		return nil, err
	}
	resp := toPatientResponse(p)
	return &resp, nil
}

func (s *PatientService) Delete(ctx context.Context, tenantID primitive.ObjectID, id string) error {
	return s.repo.Delete(ctx, tenantID, id)
}
