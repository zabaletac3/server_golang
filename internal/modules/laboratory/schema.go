package laboratory

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LabOrderStatus represents the status of a lab order
type LabOrderStatus string

const (
	LabOrderStatusPending   LabOrderStatus = "pending"
	LabOrderStatusCollected LabOrderStatus = "collected"
	LabOrderStatusSent      LabOrderStatus = "sent"
	LabOrderStatusReceived  LabOrderStatus = "received"
	LabOrderStatusProcessed LabOrderStatus = "processed"
)

// IsValidLabOrderStatus checks if the status is valid
func IsValidLabOrderStatus(s string) bool {
	switch LabOrderStatus(s) {
	case LabOrderStatusPending, LabOrderStatusCollected,
		LabOrderStatusSent, LabOrderStatusReceived, LabOrderStatusProcessed:
		return true
	}
	return false
}

// LabTestType represents the type of lab test
type LabTestType string

const (
	LabTestTypeBlood      LabTestType = "blood"
	LabTestTypeUrine      LabTestType = "urine"
	LabTestTypeBiopsy     LabTestType = "biopsy"
	LabTestTypeStool      LabTestType = "stool"
	LabTestTypeSkin       LabTestType = "skin"
	LabTestTypeEar        LabTestType = "ear"
	LabTestTypeOther      LabTestType = "other"
)

// IsValidLabTestType checks if the test type is valid
func IsValidLabTestType(t string) bool {
	switch LabTestType(t) {
	case LabTestTypeBlood, LabTestTypeUrine, LabTestTypeBiopsy,
		LabTestTypeStool, LabTestTypeSkin, LabTestTypeEar, LabTestTypeOther:
		return true
	}
	return false
}

// LabOrder represents a laboratory order
type LabOrder struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	TenantID      primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	PatientID     primitive.ObjectID `bson:"patient_id" json:"patient_id"`
	OwnerID       primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	VeterinarianID primitive.ObjectID `bson:"veterinarian_id" json:"veterinarian_id"`
	OrderDate     time.Time          `bson:"order_date" json:"order_date"`
	CollectionDate *time.Time         `bson:"collection_date,omitempty" json:"collection_date,omitempty"`
	ResultDate    *time.Time         `bson:"result_date,omitempty" json:"result_date,omitempty"`
	LabID         string             `bson:"lab_id,omitempty" json:"lab_id,omitempty"` // External lab name
	TestType      LabTestType        `bson:"test_type" json:"test_type"`
	Status        LabOrderStatus     `bson:"status" json:"status"`
	ResultFileID  string             `bson:"result_file_id,omitempty" json:"result_file_id,omitempty"` // Resource ID
	Notes         string             `bson:"notes,omitempty" json:"notes,omitempty"`
	Cost          float64            `bson:"cost,omitempty" json:"cost,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts LabOrder to LabOrderResponse
func (o *LabOrder) ToResponse() *LabOrderResponse {
	resp := &LabOrderResponse{
		ID:             o.ID.Hex(),
		TenantID:       o.TenantID.Hex(),
		PatientID:      o.PatientID.Hex(),
		OwnerID:        o.OwnerID.Hex(),
		VeterinarianID: o.VeterinarianID.Hex(),
		OrderDate:      o.OrderDate,
		LabID:          o.LabID,
		TestType:       string(o.TestType),
		Status:         string(o.Status),
		ResultFileID:   o.ResultFileID,
		Notes:          o.Notes,
		Cost:           o.Cost,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}

	if o.CollectionDate != nil {
		resp.CollectionDate = o.CollectionDate.Format(time.RFC3339)
	}

	if o.ResultDate != nil {
		resp.ResultDate = o.ResultDate.Format(time.RFC3339)
	}

	return resp
}

// IsOverdue checks if the order is overdue based on turnaround time
func (o *LabOrder) IsOverdue(turnaroundDays int) bool {
	if o.Status == LabOrderStatusProcessed {
		return false
	}
	overdueThreshold := o.OrderDate.AddDate(0, 0, turnaroundDays)
	return time.Now().After(overdueThreshold)
}

// DaysSinceOrder returns the number of days since the order was created
func (o *LabOrder) DaysSinceOrder() int {
	hours := time.Since(o.OrderDate).Hours()
	return int(hours / 24)
}

// LabOrderResponse represents a lab order in API responses
type LabOrderResponse struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	PatientID      string    `json:"patient_id"`
	OwnerID        string    `json:"owner_id"`
	VeterinarianID string    `json:"veterinarian_id"`
	OrderDate      time.Time `json:"order_date"`
	CollectionDate string    `json:"collection_date,omitempty"`
	ResultDate     string    `json:"result_date,omitempty"`
	LabID          string    `json:"lab_id,omitempty"`
	TestType       string    `json:"test_type"`
	Status         string    `json:"status"`
	ResultFileID   string    `json:"result_file_id,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	Cost           float64   `json:"cost,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LabTest represents a lab test in the catalog
type LabTest struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	TenantID       primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description,omitempty" json:"description,omitempty"`
	Category       string             `bson:"category" json:"category"` // hematology, biochemistry, urinalysis, etc.
	Price          float64            `bson:"price" json:"price"`
	TurnaroundTime int                `bson:"turnaround_time" json:"turnaround_time"` // days
	Active         bool               `bson:"active" json:"active"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts LabTest to LabTestResponse
func (t *LabTest) ToResponse() *LabTestResponse {
	return &LabTestResponse{
		ID:             t.ID.Hex(),
		TenantID:       t.TenantID.Hex(),
		Name:           t.Name,
		Description:    t.Description,
		Category:       t.Category,
		Price:          t.Price,
		TurnaroundTime: t.TurnaroundTime,
		Active:         t.Active,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

// LabTestResponse represents a lab test in API responses
type LabTestResponse struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Category       string    `json:"category"`
	Price          float64   `json:"price"`
	TurnaroundTime int       `json:"turnaround_time"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LabOrderAlert represents a lab order alert
type LabOrderAlert struct {
	OrderID        string     `json:"order_id"`
	PatientID      string     `json:"patient_id"`
	PatientName    string     `json:"patient_name"`
	OwnerID        string     `json:"owner_id"`
	OwnerName      string     `json:"owner_name"`
	TestType       string     `json:"test_type"`
	AlertType      string     `json:"alert_type"` // overdue, result_ready
	OrderDate      time.Time  `json:"order_date"`
	DaysSinceOrder int        `json:"days_since_order"`
	ResultFileID   string     `json:"result_file_id,omitempty"`
}
