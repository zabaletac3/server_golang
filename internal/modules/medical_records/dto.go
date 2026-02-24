package medical_records

import (
	"time"
)

// CreateMedicalRecordDTO represents the request to create a medical record
type CreateMedicalRecordDTO struct {
	PatientID      string       `json:"patient_id" binding:"required"`
	VeterinarianID string       `json:"veterinarian_id" binding:"required"`
	AppointmentID  string       `json:"appointment_id"`
	Type           string       `json:"type" binding:"required,oneof=consultation emergency surgery checkup vaccination"`
	ChiefComplaint string       `json:"chief_complaint" binding:"required,min=1,max=500"`
	Diagnosis      string       `json:"diagnosis" max:"2000"`
	Symptoms       string       `json:"symptoms" max:"1000"`
	Weight         float64      `json:"weight" binding:"omitempty,min=0"`
	Temperature    float64      `json:"temperature" binding:"omitempty,min=30,max=45"`
	Treatment      string       `json:"treatment" max:"2000"`
	Medications    []MedicationDTO `json:"medications"`
	EvolutionNotes string       `json:"evolution_notes" max:"2000"`
	AttachmentIDs  []string     `json:"attachment_ids"`
	NextVisitDate  string       `json:"next_visit_date"` // RFC3339
}

// MedicationDTO represents a medication in DTOs
type MedicationDTO struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Dose     string `json:"dose" binding:"required,min=1,max=100"`
	Frequency string `json:"frequency" binding:"required,min=1,max=100"`
	Duration string `json:"duration" binding:"required,min=1,max=100"`
}

// ToMedications converts MedicationDTO slice to Medication slice
func (d *CreateMedicalRecordDTO) ToMedications() []Medication {
	meds := make([]Medication, len(d.Medications))
	for i, m := range d.Medications {
		meds[i] = Medication{
			Name:     m.Name,
			Dose:     m.Dose,
			Frequency: m.Frequency,
			Duration: m.Duration,
		}
	}
	return meds
}

// ParseNextVisitDate parses the NextVisitDate string to time.Time
func (d *CreateMedicalRecordDTO) ParseNextVisitDate() (*time.Time, error) {
	if d.NextVisitDate == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, d.NextVisitDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateMedicalRecordDTO represents the request to update a medical record
type UpdateMedicalRecordDTO struct {
	Diagnosis      string       `json:"diagnosis" max:"2000"`
	Symptoms       string       `json:"symptoms" max:"1000"`
	Weight         float64      `json:"weight" binding:"omitempty,min=0"`
	Temperature    float64      `json:"temperature" binding:"omitempty,min=30,max=45"`
	Treatment      string       `json:"treatment" max:"2000"`
	Medications    []MedicationDTO `json:"medications"`
	EvolutionNotes string       `json:"evolution_notes" max:"2000"`
	AttachmentIDs  []string     `json:"attachment_ids"`
	NextVisitDate  string       `json:"next_visit_date"` // RFC3339
}

// CreateAllergyDTO represents the request to create an allergy
type CreateAllergyDTO struct {
	PatientID   string `json:"patient_id" binding:"required"`
	Allergen    string `json:"allergen" binding:"required,min=1,max=100"`
	Severity    string `json:"severity" binding:"required,oneof=mild moderate severe"`
	Description string `json:"description" max:"500"`
}

// UpdateAllergyDTO represents the request to update an allergy
type UpdateAllergyDTO struct {
	Allergen    string `json:"allergen" max:"100"`
	Severity    string `json:"severity" oneof=mild moderate severe"`
	Description string `json:"description" max:"500"`
}

// CreateMedicalHistoryDTO represents the request to create/update medical history
type CreateMedicalHistoryDTO struct {
	PatientID         string   `json:"patient_id" binding:"required"`
	ChronicConditions []string `json:"chronic_conditions"`
	PreviousSurgeries []string `json:"previous_surgeries"`
	RiskFactors       []string `json:"risk_factors"`
	BloodType         string   `json:"blood_type" max:"10"`
}

// UpdateMedicalHistoryDTO represents the request to update medical history
type UpdateMedicalHistoryDTO struct {
	ChronicConditions []string `json:"chronic_conditions"`
	PreviousSurgeries []string `json:"previous_surgeries"`
	RiskFactors       []string `json:"risk_factors"`
	BloodType         string   `json:"blood_type" max:"10"`
}

// AddEvolutionNoteDTO represents the request to add an evolution note
type AddEvolutionNoteDTO struct {
	Note           string `json:"note" binding:"required,min=1,max=2000"`
	VeterinarianID string `json:"veterinarian_id" binding:"required"`
}

// MedicalRecordListFilters represents filters for listing medical records
type MedicalRecordListFilters struct {
	PatientID      string
	VeterinarianID string
	Type           string
	DateFrom       string // RFC3339
	DateTo         string // RFC3339
	HasAttachments bool
}

// TimelineFilters represents filters for timeline queries
type TimelineFilters struct {
	DateFrom   string // RFC3339
	DateTo     string // RFC3339
	RecordType string // consultation, vaccination, surgery, etc.
	Limit      int
	Skip       int
}
