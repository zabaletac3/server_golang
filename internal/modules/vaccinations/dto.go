package vaccinations

import (
	"time"
)

// CreateVaccinationDTO represents the request to create a vaccination
type CreateVaccinationDTO struct {
	PatientID       string `json:"patient_id" binding:"required"`
	VeterinarianID  string `json:"veterinarian_id" binding:"required"`
	VaccineName     string `json:"vaccine_name" binding:"required,min=1,max=100"`
	Manufacturer    string `json:"manufacturer" max:"100"`
	LotNumber       string `json:"lot_number" max:"50"`
	ApplicationDate string `json:"application_date" binding:"required"` // RFC3339
	NextDueDate     string `json:"next_due_date"` // RFC3339
	Notes           string `json:"notes" max:"500"`
}

// ParseApplicationDate parses the ApplicationDate string to time.Time
func (d *CreateVaccinationDTO) ParseApplicationDate() (time.Time, error) {
	return time.Parse(time.RFC3339, d.ApplicationDate)
}

// ParseNextDueDate parses the NextDueDate string to time.Time
func (d *CreateVaccinationDTO) ParseNextDueDate() (*time.Time, error) {
	if d.NextDueDate == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, d.NextDueDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateVaccinationDTO represents the request to update a vaccination
type UpdateVaccinationDTO struct {
	VaccineName     string `json:"vaccine_name" max:"100"`
	Manufacturer    string `json:"manufacturer" max:"100"`
	LotNumber       string `json:"lot_number" max:"50"`
	NextDueDate     string `json:"next_due_date"`
	Notes           string `json:"notes" max:"500"`
}

// UpdateVaccinationStatusDTO represents the request to update vaccination status
type UpdateVaccinationStatusDTO struct {
	Status string `json:"status" binding:"required,oneof=applied due overdue"`
}

// CreateVaccineDTO represents the request to create a vaccine catalog entry
type CreateVaccineDTO struct {
	Name           string   `json:"name" binding:"required,min=2,max=100"`
	Description    string   `json:"description" max:"500"`
	Manufacturer   string   `json:"manufacturer" max:"100"`
	DoseNumber     string   `json:"dose_number" binding:"required,oneof=first second booster"`
	ValidityMonths int      `json:"validity_months" binding:"required,min=1"`
	TargetSpecies  []string `json:"target_species" binding:"required"`
	Active         bool     `json:"active"`
}

// UpdateVaccineDTO represents the request to update a vaccine catalog entry
type UpdateVaccineDTO struct {
	Name           string   `json:"name" max:"100"`
	Description    string   `json:"description" max:"500"`
	Manufacturer   string   `json:"manufacturer" max:"100"`
	DoseNumber     string   `json:"dose_number" oneof=first second booster"`
	ValidityMonths int      `json:"validity_months" binding:"omitempty,min=1"`
	TargetSpecies  []string `json:"target_species"`
	Active         bool     `json:"active"`
}

// VaccinationListFilters represents filters for listing vaccinations
type VaccinationListFilters struct {
	PatientID      string
	VeterinarianID string
	Status         string
	VaccineName    string
	DateFrom       string // RFC3339
	DateTo         string // RFC3339
	DueSoon        bool   // Due within 30 days
	Overdue        bool   // Already overdue
}

// VaccineListFilters represents filters for listing vaccines
type VaccineListFilters struct {
	DoseNumber    string
	TargetSpecies string
	Active        *bool
	Search        string // Search by name, manufacturer
}

// VaccinationCertificateDTO represents a request to generate a certificate
type VaccinationCertificateDTO struct {
	VaccinationID string `json:"vaccination_id" binding:"required"`
}
