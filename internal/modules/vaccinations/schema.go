package vaccinations

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VaccinationStatus represents the status of a vaccination
type VaccinationStatus string

const (
	VaccinationStatusApplied   VaccinationStatus = "applied"
	VaccinationStatusDue       VaccinationStatus = "due"
	VaccinationStatusOverdue   VaccinationStatus = "overdue"
)

// IsValidVaccinationStatus checks if the status is valid
func IsValidVaccinationStatus(s string) bool {
	switch VaccinationStatus(s) {
	case VaccinationStatusApplied, VaccinationStatusDue, VaccinationStatusOverdue:
		return true
	}
	return false
}

// VaccineDoseType represents the dose type of a vaccine
type VaccineDoseType string

const (
	VaccineDoseFirst   VaccineDoseType = "first"
	VaccineDoseSecond  VaccineDoseType = "second"
	VaccineDoseBooster VaccineDoseType = "booster"
)

// IsValidVaccineDoseType checks if the dose type is valid
func IsValidVaccineDoseType(t string) bool {
	switch VaccineDoseType(t) {
	case VaccineDoseFirst, VaccineDoseSecond, VaccineDoseBooster:
		return true
	}
	return false
}

// Vaccination represents a vaccination record for a patient
type Vaccination struct {
	ID              primitive.ObjectID  `bson:"_id" json:"id"`
	TenantID        primitive.ObjectID  `bson:"tenant_id" json:"tenant_id"`
	PatientID       primitive.ObjectID  `bson:"patient_id" json:"patient_id"`
	OwnerID         primitive.ObjectID  `bson:"owner_id" json:"owner_id"`
	VeterinarianID  primitive.ObjectID  `bson:"veterinarian_id" json:"veterinarian_id"`
	VaccineName     string              `bson:"vaccine_name" json:"vaccine_name"`
	Manufacturer    string              `bson:"manufacturer,omitempty" json:"manufacturer,omitempty"`
	LotNumber       string              `bson:"lot_number,omitempty" json:"lot_number,omitempty"`
	ApplicationDate time.Time           `bson:"application_date" json:"application_date"`
	NextDueDate     *time.Time          `bson:"next_due_date,omitempty" json:"next_due_date,omitempty"`
	Status          VaccinationStatus   `bson:"status" json:"status"`
	CertificateNumber string            `bson:"certificate_number,omitempty" json:"certificate_number,omitempty"`
	Notes           string              `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time          `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts Vaccination to VaccinationResponse
func (v *Vaccination) ToResponse() *VaccinationResponse {
	resp := &VaccinationResponse{
		ID:              v.ID.Hex(),
		TenantID:        v.TenantID.Hex(),
		PatientID:       v.PatientID.Hex(),
		OwnerID:         v.OwnerID.Hex(),
		VeterinarianID:  v.VeterinarianID.Hex(),
		VaccineName:     v.VaccineName,
		Manufacturer:    v.Manufacturer,
		LotNumber:       v.LotNumber,
		ApplicationDate: v.ApplicationDate,
		Status:          string(v.Status),
		CertificateNumber: v.CertificateNumber,
		Notes:           v.Notes,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}

	if v.NextDueDate != nil {
		resp.NextDueDate = v.NextDueDate.Format(time.RFC3339)
	}

	return resp
}

// IsDue checks if the vaccination is due within the next 30 days
func (v *Vaccination) IsDue() bool {
	if v.NextDueDate == nil {
		return false
	}
	now := time.Now()
	dueThreshold := now.AddDate(0, 0, 30)
	return v.NextDueDate.After(now) && v.NextDueDate.Before(dueThreshold)
}

// IsOverdue checks if the vaccination is overdue
func (v *Vaccination) IsOverdue() bool {
	if v.NextDueDate == nil {
		return false
	}
	return v.NextDueDate.Before(time.Now())
}

// DaysUntilDue returns the number of days until the vaccination is due
func (v *Vaccination) DaysUntilDue() int {
	if v.NextDueDate == nil {
		return -1
	}
	hours := v.NextDueDate.Sub(time.Now()).Hours()
	return int(hours / 24)
}

// VaccinationResponse represents a vaccination in API responses
type VaccinationResponse struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	PatientID         string    `json:"patient_id"`
	OwnerID           string    `json:"owner_id"`
	VeterinarianID    string    `json:"veterinarian_id"`
	VaccineName       string    `json:"vaccine_name"`
	Manufacturer      string    `json:"manufacturer,omitempty"`
	LotNumber         string    `json:"lot_number,omitempty"`
	ApplicationDate   time.Time `json:"application_date"`
	NextDueDate       string    `json:"next_due_date,omitempty"`
	Status            string    `json:"status"`
	CertificateNumber string    `json:"certificate_number,omitempty"`
	Notes             string    `json:"notes,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Vaccine represents a vaccine in the catalog
type Vaccine struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	TenantID       primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description,omitempty" json:"description,omitempty"`
	Manufacturer   string             `bson:"manufacturer,omitempty" json:"manufacturer,omitempty"`
	DoseNumber     VaccineDoseType    `bson:"dose_number" json:"dose_number"`
	ValidityMonths int                `bson:"validity_months" json:"validity_months"`
	TargetSpecies  []string           `bson:"target_species" json:"target_species"`
	Active         bool               `bson:"active" json:"active"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts Vaccine to VaccineResponse
func (v *Vaccine) ToResponse() *VaccineResponse {
	return &VaccineResponse{
		ID:             v.ID.Hex(),
		TenantID:       v.TenantID.Hex(),
		Name:           v.Name,
		Description:    v.Description,
		Manufacturer:   v.Manufacturer,
		DoseNumber:     string(v.DoseNumber),
		ValidityMonths: v.ValidityMonths,
		TargetSpecies:  v.TargetSpecies,
		Active:         v.Active,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

// VaccineResponse represents a vaccine in API responses
type VaccineResponse struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Manufacturer   string    `json:"manufacturer,omitempty"`
	DoseNumber     string    `json:"dose_number"`
	ValidityMonths int       `json:"validity_months"`
	TargetSpecies  []string  `json:"target_species"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// VaccinationAlert represents a vaccination alert
type VaccinationAlert struct {
	VaccinationID   string     `json:"vaccination_id"`
	PatientID       string     `json:"patient_id"`
	PatientName     string     `json:"patient_name"`
	OwnerID         string     `json:"owner_id"`
	OwnerName       string     `json:"owner_name"`
	VaccineName     string     `json:"vaccine_name"`
	AlertType       string     `json:"alert_type"` // due_7days, due_24h, overdue
	NextDueDate     *time.Time `json:"next_due_date,omitempty"`
	DaysUntilDue    int        `json:"days_until_due,omitempty"`
	DaysOverdue     int        `json:"days_overdue,omitempty"`
}
