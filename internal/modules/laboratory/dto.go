package laboratory

// CreateLabOrderDTO represents the request to create a lab order
type CreateLabOrderDTO struct {
	PatientID      string  `json:"patient_id" binding:"required"`
	VeterinarianID string  `json:"veterinarian_id" binding:"required"`
	LabID          string  `json:"lab_id"`
	TestType       string  `json:"test_type" binding:"required,oneof=blood urine biopsy stool skin ear other"`
	Notes          string  `json:"notes" max:"500"`
	Cost           float64 `json:"cost" binding:"omitempty,min=0"`
}

// UpdateLabOrderDTO represents the request to update a lab order
type UpdateLabOrderDTO struct {
	LabID     string  `json:"lab_id"`
	TestType  string  `json:"test_type" oneof=blood urine biopsy stool skin ear other"`
	Notes     string  `json:"notes" max:"500"`
	Cost      float64 `json:"cost" binding:"omitempty,min=0"`
}

// UpdateLabOrderStatusDTO represents the request to update lab order status
type UpdateLabOrderStatusDTO struct {
	Status       string `json:"status" binding:"required,oneof=pending collected sent received processed"`
	CollectionDate string `json:"collection_date"` // RFC3339
	Notes        string `json:"notes" max:"500"`
}

// UploadLabResultDTO represents the request to upload a lab result
type UploadLabResultDTO struct {
	ResultFileID string `json:"result_file_id" binding:"required"`
	Notes        string `json:"notes" max:"500"`
}

// CreateLabTestDTO represents the request to create a lab test catalog entry
type CreateLabTestDTO struct {
	Name           string  `json:"name" binding:"required,min=2,max=100"`
	Description    string  `json:"description" max:"500"`
	Category       string  `json:"category" binding:"required"`
	Price          float64 `json:"price" binding:"required,min=0"`
	TurnaroundTime int     `json:"turnaround_time" binding:"required,min=1"`
	Active         bool    `json:"active"`
}

// UpdateLabTestDTO represents the request to update a lab test catalog entry
type UpdateLabTestDTO struct {
	Name           string  `json:"name" max:"100"`
	Description    string  `json:"description" max:"500"`
	Category       string  `json:"category"`
	Price          float64 `json:"price" binding:"omitempty,min=0"`
	TurnaroundTime int     `json:"turnaround_time" binding:"omitempty,min=1"`
	Active         bool    `json:"active"`
}

// LabOrderListFilters represents filters for listing lab orders
type LabOrderListFilters struct {
	PatientID      string
	VeterinarianID string
	Status         string
	TestType       string
	LabID          string
	DateFrom       string // RFC3339
	DateTo         string // RFC3339
	Overdue        bool
}

// LabTestListFilters represents filters for listing lab tests
type LabTestListFilters struct {
	Category string
	Active   *bool
	Search   string // Search by name
}
