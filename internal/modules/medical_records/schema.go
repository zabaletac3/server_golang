package medical_records

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MedicalRecordType represents the type of medical record
type MedicalRecordType string

const (
	MedicalRecordTypeConsultation MedicalRecordType = "consultation"
	MedicalRecordTypeEmergency    MedicalRecordType = "emergency"
	MedicalRecordTypeSurgery      MedicalRecordType = "surgery"
	MedicalRecordTypeCheckup      MedicalRecordType = "checkup"
	MedicalRecordTypeVaccination  MedicalRecordType = "vaccination"
)

// IsValidMedicalRecordType checks if the type is valid
func IsValidMedicalRecordType(t string) bool {
	switch MedicalRecordType(t) {
	case MedicalRecordTypeConsultation, MedicalRecordTypeEmergency,
		MedicalRecordTypeSurgery, MedicalRecordTypeCheckup, MedicalRecordTypeVaccination:
		return true
	}
	return false
}

// AllergySeverity represents the severity of an allergy
type AllergySeverity string

const (
	AllergySeverityMild     AllergySeverity = "mild"
	AllergySeverityModerate AllergySeverity = "moderate"
	AllergySeveritySevere   AllergySeverity = "severe"
)

// IsValidAllergySeverity checks if the severity is valid
func IsValidAllergySeverity(s string) bool {
	switch AllergySeverity(s) {
	case AllergySeverityMild, AllergySeverityModerate, AllergySeveritySevere:
		return true
	}
	return false
}

// Medication represents a medication in a medical record
type Medication struct {
	Name     string `bson:"name" json:"name"`
	Dose     string `bson:"dose" json:"dose"`
	Frequency string `bson:"frequency" json:"frequency"`
	Duration string `bson:"duration" json:"duration"`
}

// DailyProgress represents a daily progress note in hospitalization
type DailyProgress struct {
	Date           time.Time `bson:"date" json:"date"`
	Notes          string    `bson:"notes" json:"notes"`
	MedicationsGiven string  `bson:"medications_given" json:"medications_given"`
	FoodProvided   string    `bson:"food_provided" json:"food_provided"`
	VetID          primitive.ObjectID `bson:"vet_id" json:"vet_id"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
}

// MedicalRecord represents a clinical record entry
type MedicalRecord struct {
	ID             primitive.ObjectID  `bson:"_id" json:"id"`
	TenantID       primitive.ObjectID  `bson:"tenant_id" json:"tenant_id"`
	PatientID      primitive.ObjectID  `bson:"patient_id" json:"patient_id"`
	OwnerID        primitive.ObjectID  `bson:"owner_id" json:"owner_id"`
	VeterinarianID primitive.ObjectID  `bson:"veterinarian_id" json:"veterinarian_id"`
	AppointmentID  *primitive.ObjectID `bson:"appointment_id,omitempty" json:"appointment_id,omitempty"`
	Type           MedicalRecordType   `bson:"type" json:"type"`
	ChiefComplaint string              `bson:"chief_complaint" json:"chief_complaint"`
	Diagnosis      string              `bson:"diagnosis" json:"diagnosis"`
	Symptoms       string              `bson:"symptoms" json:"symptoms"`
	Weight         float64             `bson:"weight,omitempty" json:"weight,omitempty"`
	Temperature    float64             `bson:"temperature,omitempty" json:"temperature,omitempty"`
	Treatment      string              `bson:"treatment" json:"treatment"`
	Medications    []Medication        `bson:"medications" json:"medications"`
	EvolutionNotes string              `bson:"evolution_notes" json:"evolution_notes"`
	AttachmentIDs  []string            `bson:"attachment_ids,omitempty" json:"attachment_ids,omitempty"`
	NextVisitDate  *time.Time          `bson:"next_visit_date,omitempty" json:"next_visit_date,omitempty"`
	CreatedAt      time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time          `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts MedicalRecord to MedicalRecordResponse
func (m *MedicalRecord) ToResponse() *MedicalRecordResponse {
	resp := &MedicalRecordResponse{
		ID:             m.ID.Hex(),
		TenantID:       m.TenantID.Hex(),
		PatientID:      m.PatientID.Hex(),
		OwnerID:        m.OwnerID.Hex(),
		VeterinarianID: m.VeterinarianID.Hex(),
		Type:           string(m.Type),
		ChiefComplaint: m.ChiefComplaint,
		Diagnosis:      m.Diagnosis,
		Symptoms:       m.Symptoms,
		Weight:         m.Weight,
		Temperature:    m.Temperature,
		Treatment:      m.Treatment,
		Medications:    m.Medications,
		EvolutionNotes: m.EvolutionNotes,
		AttachmentIDs:  m.AttachmentIDs,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}

	if m.AppointmentID != nil {
		resp.AppointmentID = m.AppointmentID.Hex()
	}

	if m.NextVisitDate != nil {
		resp.NextVisitDate = m.NextVisitDate.Format(time.RFC3339)
	}

	return resp
}

// IsEditable checks if the record can be edited (within 24 hours)
func (m *MedicalRecord) IsEditable() bool {
	return time.Since(m.CreatedAt) < 24*time.Hour
}

// MedicalRecordResponse represents a medical record in API responses
type MedicalRecordResponse struct {
	ID             string       `json:"id"`
	TenantID       string       `json:"tenant_id"`
	PatientID      string       `json:"patient_id"`
	OwnerID        string       `json:"owner_id"`
	VeterinarianID string       `json:"veterinarian_id"`
	AppointmentID  string       `json:"appointment_id,omitempty"`
	Type           string       `json:"type"`
	ChiefComplaint string       `json:"chief_complaint"`
	Diagnosis      string       `json:"diagnosis"`
	Symptoms       string       `json:"symptoms"`
	Weight         float64      `json:"weight,omitempty"`
	Temperature    float64      `json:"temperature,omitempty"`
	Treatment      string       `json:"treatment"`
	Medications    []Medication `json:"medications"`
	EvolutionNotes string       `json:"evolution_notes"`
	AttachmentIDs  []string     `json:"attachment_ids,omitempty"`
	NextVisitDate  string       `json:"next_visit_date,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// Allergy represents a patient allergy
type Allergy struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	PatientID   primitive.ObjectID `bson:"patient_id" json:"patient_id"`
	Allergen    string             `bson:"allergen" json:"allergen"`
	Severity    AllergySeverity    `bson:"severity" json:"severity"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts Allergy to AllergyResponse
func (a *Allergy) ToResponse() *AllergyResponse {
	return &AllergyResponse{
		ID:          a.ID.Hex(),
		TenantID:    a.TenantID.Hex(),
		PatientID:   a.PatientID.Hex(),
		Allergen:    a.Allergen,
		Severity:    string(a.Severity),
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// AllergyResponse represents an allergy in API responses
type AllergyResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	PatientID   string    `json:"patient_id"`
	Allergen    string    `json:"allergen"`
	Severity    string    `json:"severity"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MedicalHistory represents a patient's medical history summary
type MedicalHistory struct {
	ID                primitive.ObjectID `bson:"_id" json:"id"`
	TenantID          primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	PatientID         primitive.ObjectID `bson:"patient_id" json:"patient_id"`
	ChronicConditions []string           `bson:"chronic_conditions,omitempty" json:"chronic_conditions,omitempty"`
	PreviousSurgeries []string           `bson:"previous_surgeries,omitempty" json:"previous_surgeries,omitempty"`
	RiskFactors       []string           `bson:"risk_factors,omitempty" json:"risk_factors,omitempty"`
	BloodType         string             `bson:"blood_type,omitempty" json:"blood_type,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt         *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts MedicalHistory to MedicalHistoryResponse
func (h *MedicalHistory) ToResponse() *MedicalHistoryResponse {
	return &MedicalHistoryResponse{
		ID:                h.ID.Hex(),
		TenantID:          h.TenantID.Hex(),
		PatientID:         h.PatientID.Hex(),
		ChronicConditions: h.ChronicConditions,
		PreviousSurgeries: h.PreviousSurgeries,
		RiskFactors:       h.RiskFactors,
		BloodType:         h.BloodType,
		CreatedAt:         h.CreatedAt,
		UpdatedAt:         h.UpdatedAt,
	}
}

// MedicalHistoryResponse represents a medical history summary in API responses
type MedicalHistoryResponse struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	PatientID         string    `json:"patient_id"`
	ChronicConditions []string  `json:"chronic_conditions,omitempty"`
	PreviousSurgeries []string  `json:"previous_surgeries,omitempty"`
	RiskFactors       []string  `json:"risk_factors,omitempty"`
	BloodType         string    `json:"blood_type,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TimelineEntry represents an entry in the medical timeline
type TimelineEntry struct {
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // consultation, vaccination, surgery, etc.
	Description string    `json:"description"`
	Veterinarian string   `json:"veterinarian"`
	RecordID    string    `json:"record_id,omitempty"`
}

// MedicalTimeline represents a patient's medical timeline
type MedicalTimeline struct {
	PatientID   string          `json:"patient_id"`
	PatientName string          `json:"patient_name"`
	Entries     []TimelineEntry `json:"entries"`
	TotalCount  int64           `json:"total_count"`
}
