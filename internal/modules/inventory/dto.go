package inventory

import (
	"time"
)

// CreateProductDTO represents the request to create a product
type CreateProductDTO struct {
	CategoryID     string  `json:"category_id"`
	Name           string  `json:"name" binding:"required,min=3,max=100"`
	Description    string  `json:"description" max:"500"`
	SKU            string  `json:"sku" binding:"required,min=1,max=50"`
	Barcode        string  `json:"barcode" max:"50"`
	Category       string  `json:"category" binding:"required,oneof=medicine supply food equipment"`
	Unit           string  `json:"unit" binding:"required,oneof=tablet ml piece kg gram box bottle"`
	PurchasePrice  float64 `json:"purchase_price" binding:"required,min=0"`
	SalePrice      float64 `json:"sale_price" binding:"required,min=0"`
	Stock          int     `json:"stock" binding:"min=0"`
	MinStock       int     `json:"min_stock" binding:"min=0"`
	ExpirationDate string  `json:"expiration_date"` // RFC3339
	SupplierID     string  `json:"supplier_id"`
	Active         bool    `json:"active"`
}

// ParseExpirationDate parses the ExpirationDate string to time.Time
func (d *CreateProductDTO) ParseExpirationDate() (*time.Time, error) {
	if d.ExpirationDate == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, d.ExpirationDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateProductDTO represents the request to update a product
type UpdateProductDTO struct {
	CategoryID     string  `json:"category_id"`
	Name           string  `json:"name" max:"100"`
	Description    string  `json:"description" max:"500"`
	SKU            string  `json:"sku" max:"50"`
	Barcode        string  `json:"barcode" max:"50"`
	Category       string  `json:"category" oneof=medicine supply food equipment"`
	Unit           string  `json:"unit" oneof=tablet ml piece kg gram box bottle"`
	PurchasePrice  float64 `json:"purchase_price" binding:"omitempty,min=0"`
	SalePrice      float64 `json:"sale_price" binding:"omitempty,min=0"`
	MinStock       int     `json:"min_stock" binding:"omitempty,min=0"`
	ExpirationDate string  `json:"expiration_date"`
	SupplierID     string  `json:"supplier_id"`
	Active         bool    `json:"active"`
}

// StockInDTO represents the request to add stock
type StockInDTO struct {
	Quantity    int    `json:"quantity" binding:"required,min=1"`
	Reason      string `json:"reason" binding:"required,oneof=purchase return adjustment"`
	ReferenceID string `json:"reference_id"` // Purchase order ID, etc.
	Notes       string `json:"notes" max:"500"`
}

// StockOutDTO represents the request to deduct stock
type StockOutDTO struct {
	Quantity    int    `json:"quantity" binding:"required,min=1"`
	Reason      string `json:"reason" binding:"required,oneof=sale treatment adjustment expired damaged lost"`
	ReferenceID string `json:"reference_id"` // AppointmentID, OrderID, etc.
	Notes       string `json:"notes" max:"500"`
}

// CreateCategoryDTO represents the request to create a category
type CreateCategoryDTO struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" max:"500"`
	ParentID    string `json:"parent_id"`
}

// UpdateCategoryDTO represents the request to update a category
type UpdateCategoryDTO struct {
	Name        string `json:"name" max:"100"`
	Description string `json:"description" max:"500"`
	ParentID    string `json:"parent_id"`
}

// ProductListFilters represents filters for listing products
type ProductListFilters struct {
	CategoryID   string
	Category     string
	Active       *bool
	LowStock     bool
	ExpiringSoon bool
	Expired      bool
	Search       string // Search by name, SKU, barcode
}

// StockMovementListFilters represents filters for listing stock movements
type StockMovementListFilters struct {
	ProductID   string
	Type        string
	Reason      string
	DateFrom    string // RFC3339
	DateTo      string // RFC3339
	ReferenceID string
}

// ProductAlertResponse represents a product alert in API responses
type ProductAlertResponse struct {
	ProductID       string     `json:"product_id"`
	ProductName     string     `json:"product_name"`
	AlertType       string     `json:"alert_type"` // low_stock, expiring, expired
	CurrentStock    int        `json:"current_stock"`
	MinStock        int        `json:"min_stock"`
	StockDifference int        `json:"stock_difference"` // Negative if below min
	ExpirationDate  *time.Time `json:"expiration_date,omitempty"`
	DaysUntilExpiry *int       `json:"days_until_expiry,omitempty"`
}
