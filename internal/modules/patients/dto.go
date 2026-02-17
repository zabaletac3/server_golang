package patients

import (
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// --- Species DTOs ---

type ResolveSpeciesDTO struct {
	Name string `json:"name" binding:"required" example:"Canino"`
}

type SpeciesResponse struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Name           string    `json:"name"`
	NormalizedName string    `json:"normalized_name"`
	CreatedAt      time.Time `json:"created_at"`
}

type SpeciesConflictResponse struct {
	Message     string            `json:"message"`
	Suggestions []SpeciesResponse `json:"suggestions"`
}

func toSpeciesResponse(s *Species) SpeciesResponse {
	return SpeciesResponse{
		ID:             s.ID.Hex(),
		TenantID:       s.TenantID.Hex(),
		Name:           s.Name,
		NormalizedName: s.NormalizedName,
		CreatedAt:      s.CreatedAt,
	}
}

// --- Patient DTOs ---

type CreatePatientDTO struct {
	OwnerID    string     `json:"owner_id"   binding:"required"`
	SpeciesID  string     `json:"species_id" binding:"required"`
	Name       string     `json:"name"       binding:"required"                    example:"Max"`
	Breed      string     `json:"breed"                                            example:"Labrador"`
	Color      string     `json:"color"                                            example:"Dorado"`
	BirthDate  *time.Time `json:"birth_date"`
	Gender     Gender     `json:"gender"     binding:"required,oneof=male female unknown"`
	Weight     float64    `json:"weight"     binding:"required,gt=0"               example:"12.5"`
	Microchip  string     `json:"microchip"                                        example:"985112000123456"`
	Sterilized bool       `json:"sterilized"`
	AvatarURL  string     `json:"avatar_url"`
	Notes      string     `json:"notes"`
}

type UpdatePatientDTO struct {
	SpeciesID  string     `json:"species_id"`
	Name       string     `json:"name"`
	Breed      string     `json:"breed"`
	Color      string     `json:"color"`
	BirthDate  *time.Time `json:"birth_date"`
	Gender     Gender     `json:"gender"     binding:"omitempty,oneof=male female unknown"`
	Weight     float64    `json:"weight"     binding:"omitempty,gt=0"`
	Microchip  string     `json:"microchip"`
	Sterilized *bool      `json:"sterilized"`
	AvatarURL  string     `json:"avatar_url"`
	Notes      string     `json:"notes"`
	Active     *bool      `json:"active"`
}

type PatientResponse struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	OwnerID    string     `json:"owner_id"`
	SpeciesID  string     `json:"species_id"`
	Name       string     `json:"name"`
	Breed      string     `json:"breed,omitempty"`
	Color      string     `json:"color,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Gender     Gender     `json:"gender"`
	Weight     float64    `json:"weight"`
	Microchip  string     `json:"microchip,omitempty"`
	Sterilized bool       `json:"sterilized"`
	AvatarURL  string     `json:"avatar_url,omitempty"`
	Notes      string     `json:"notes,omitempty"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type PaginatedPatientsResponse struct {
	Data       []PatientResponse        `json:"data"`
	Pagination pagination.PaginationInfo `json:"pagination"`
}

func toPatientResponse(p *Patient) PatientResponse {
	return PatientResponse{
		ID:         p.ID.Hex(),
		TenantID:   p.TenantID.Hex(),
		OwnerID:    p.OwnerID.Hex(),
		SpeciesID:  p.SpeciesID.Hex(),
		Name:       p.Name,
		Breed:      p.Breed,
		Color:      p.Color,
		BirthDate:  p.BirthDate,
		Gender:     p.Gender,
		Weight:     p.Weight,
		Microchip:  p.Microchip,
		Sterilized: p.Sterilized,
		AvatarURL:  p.AvatarURL,
		Notes:      p.Notes,
		Active:     p.Active,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}
