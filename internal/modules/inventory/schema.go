package inventory

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductCategory represents the category of a product
type ProductCategory string

const (
	ProductCategoryMedicine  ProductCategory = "medicine"
	ProductCategorySupply    ProductCategory = "supply"
	ProductCategoryFood      ProductCategory = "food"
	ProductCategoryEquipment ProductCategory = "equipment"
)

// IsValidProductCategory checks if the category is valid
func IsValidProductCategory(c string) bool {
	switch ProductCategory(c) {
	case ProductCategoryMedicine, ProductCategorySupply,
		ProductCategoryFood, ProductCategoryEquipment:
		return true
	}
	return false
}

// ProductUnit represents the unit of measurement for a product
type ProductUnit string

const (
	ProductUnitTablet ProductUnit = "tablet"
	ProductUnitML     ProductUnit = "ml"
	ProductUnitPiece  ProductUnit = "piece"
	ProductUnitKG     ProductUnit = "kg"
	ProductUnitGram   ProductUnit = "gram"
	ProductUnitBox    ProductUnit = "box"
	ProductUnitBottle ProductUnit = "bottle"
)

// IsValidProductUnit checks if the unit is valid
func IsValidProductUnit(u string) bool {
	switch ProductUnit(u) {
	case ProductUnitTablet, ProductUnitML, ProductUnitPiece,
		ProductUnitKG, ProductUnitGram, ProductUnitBox, ProductUnitBottle:
		return true
	}
	return false
}

// StockMovementType represents the type of stock movement
type StockMovementType string

const (
	StockMovementIn        StockMovementType = "in"        // Purchase, return
	StockMovementOut       StockMovementType = "out"       // Sale, treatment
	StockMovementAdjustment StockMovementType = "adjustment" // Inventory adjustment
	StockMovementExpired   StockMovementType = "expired"   // Expired products
)

// IsValidStockMovementType checks if the movement type is valid
func IsValidStockMovementType(t string) bool {
	switch StockMovementType(t) {
	case StockMovementIn, StockMovementOut,
		StockMovementAdjustment, StockMovementExpired:
		return true
	}
	return false
}

// IsValidStockMovementReason checks if the movement reason is valid
func IsValidStockMovementReason(r string) bool {
	switch StockMovementReason(r) {
	case StockReasonPurchase, StockReasonSale, StockReasonTreatment,
		StockReasonAdjustment, StockReasonReturn, StockReasonExpired,
		StockReasonDamaged, StockReasonLost:
		return true
	}
	return false
}

// StockMovementReason represents the reason for a stock movement
type StockMovementReason string

const (
	StockReasonPurchase    StockMovementReason = "purchase"
	StockReasonSale        StockMovementReason = "sale"
	StockReasonTreatment   StockMovementReason = "treatment"
	StockReasonAdjustment  StockMovementReason = "adjustment"
	StockReasonReturn      StockMovementReason = "return"
	StockReasonExpired     StockMovementReason = "expired"
	StockReasonDamaged     StockMovementReason = "damaged"
	StockReasonLost        StockMovementReason = "lost"
)

// Product represents a product in the inventory
type Product struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	TenantID       primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	CategoryID     primitive.ObjectID `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description,omitempty" json:"description,omitempty"`
	SKU            string             `bson:"sku" json:"sku"`
	Barcode        string             `bson:"barcode,omitempty" json:"barcode,omitempty"`
	Category       ProductCategory    `bson:"category" json:"category"`
	Unit           ProductUnit        `bson:"unit" json:"unit"`
	PurchasePrice  float64            `bson:"purchase_price" json:"purchase_price"`
	SalePrice      float64            `bson:"sale_price" json:"sale_price"`
	Stock          int                `bson:"stock" json:"stock"`
	MinStock       int                `bson:"min_stock" json:"min_stock"`
	ExpirationDate *time.Time         `bson:"expiration_date,omitempty" json:"expiration_date,omitempty"`
	SupplierID     primitive.ObjectID `bson:"supplier_id,omitempty" json:"supplier_id,omitempty"`
	Active         bool               `bson:"active" json:"active"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts Product to ProductResponse
func (p *Product) ToResponse() *ProductResponse {
	resp := &ProductResponse{
		ID:            p.ID.Hex(),
		TenantID:      p.TenantID.Hex(),
		Name:          p.Name,
		Description:   p.Description,
		SKU:           p.SKU,
		Barcode:       p.Barcode,
		Category:      string(p.Category),
		Unit:          string(p.Unit),
		PurchasePrice: p.PurchasePrice,
		SalePrice:     p.SalePrice,
		Stock:         p.Stock,
		MinStock:      p.MinStock,
		Active:        p.Active,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}

	if p.CategoryID != primitive.NilObjectID {
		resp.CategoryID = p.CategoryID.Hex()
	}

	if p.ExpirationDate != nil {
		resp.ExpirationDate = p.ExpirationDate.Format(time.RFC3339)
	}

	if p.SupplierID != primitive.NilObjectID {
		resp.SupplierID = p.SupplierID.Hex()
	}

	return resp
}

// IsLowStock checks if the product is below minimum stock
func (p *Product) IsLowStock() bool {
	return p.Stock <= p.MinStock
}

// IsExpiringSoon checks if the product expires within the given days
func (p *Product) IsExpiringSoon(days int) bool {
	if p.ExpirationDate == nil {
		return false
	}
	expiryThreshold := time.Now().AddDate(0, 0, days)
	return p.ExpirationDate.Before(expiryThreshold)
}

// IsExpired checks if the product is already expired
func (p *Product) IsExpired() bool {
	if p.ExpirationDate == nil {
		return false
	}
	return p.ExpirationDate.Before(time.Now())
}

// ProductResponse represents a product in API responses
type ProductResponse struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	CategoryID     string    `json:"category_id,omitempty"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	SKU            string    `json:"sku"`
	Barcode        string    `json:"barcode,omitempty"`
	Category       string    `json:"category"`
	Unit           string    `json:"unit"`
	PurchasePrice  float64   `json:"purchase_price"`
	SalePrice      float64   `json:"sale_price"`
	Stock          int       `json:"stock"`
	MinStock       int       `json:"min_stock"`
	ExpirationDate string    `json:"expiration_date,omitempty"`
	SupplierID     string    `json:"supplier_id,omitempty"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Category represents a product category
type Category struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	ParentID    primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// ToResponse converts Category to CategoryResponse
func (c *Category) ToResponse() *CategoryResponse {
	resp := &CategoryResponse{
		ID:          c.ID.Hex(),
		TenantID:    c.TenantID.Hex(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.ParentID != primitive.NilObjectID {
		resp.ParentID = c.ParentID.Hex()
	}

	return resp
}

// CategoryResponse represents a category in API responses
type CategoryResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ParentID    string    `json:"parent_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StockMovement represents a stock movement record
type StockMovement struct {
	ID          primitive.ObjectID  `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID  `bson:"tenant_id" json:"tenant_id"`
	ProductID   primitive.ObjectID  `bson:"product_id" json:"product_id"`
	Type        StockMovementType   `bson:"type" json:"type"`
	Reason      StockMovementReason `bson:"reason" json:"reason"`
	Quantity    int                 `bson:"quantity" json:"quantity"`
	StockBefore int                 `bson:"stock_before" json:"stock_before"`
	StockAfter  int                 `bson:"stock_after" json:"stock_after"`
	ReferenceID primitive.ObjectID  `bson:"reference_id,omitempty" json:"reference_id,omitempty"` // AppointmentID, OrderID, etc.
	UserID      primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Notes       string              `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
}

// ToResponse converts StockMovement to StockMovementResponse
func (m *StockMovement) ToResponse() *StockMovementResponse {
	resp := &StockMovementResponse{
		ID:          m.ID.Hex(),
		TenantID:    m.TenantID.Hex(),
		ProductID:   m.ProductID.Hex(),
		Type:        string(m.Type),
		Reason:      string(m.Reason),
		Quantity:    m.Quantity,
		StockBefore: m.StockBefore,
		StockAfter:  m.StockAfter,
		UserID:      m.UserID.Hex(),
		Notes:       m.Notes,
		CreatedAt:   m.CreatedAt,
	}

	if m.ReferenceID != primitive.NilObjectID {
		resp.ReferenceID = m.ReferenceID.Hex()
	}

	return resp
}

// StockMovementResponse represents a stock movement in API responses
type StockMovementResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ProductID   string    `json:"product_id"`
	Type        string    `json:"type"`
	Reason      string    `json:"reason"`
	Quantity    int       `json:"quantity"`
	StockBefore int       `json:"stock_before"`
	StockAfter  int       `json:"stock_after"`
	ReferenceID string    `json:"reference_id,omitempty"`
	UserID      string    `json:"user_id"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// StockAlert represents a stock alert (low stock or expiring)
type StockAlert struct {
	ProductID   primitive.ObjectID `bson:"product_id" json:"product_id"`
	ProductName string             `bson:"product_name" json:"product_name"`
	AlertType   string             `bson:"alert_type" json:"alert_type"` // low_stock, expiring, expired
	CurrentStock int               `bson:"current_stock" json:"current_stock"`
	MinStock    int                `bson:"min_stock" json:"min_stock"`
	ExpirationDate *time.Time      `bson:"expiration_date,omitempty" json:"expiration_date,omitempty"`
	DaysUntilExpiry int            `bson:"days_until_expiry,omitempty" json:"days_until_expiry,omitempty"`
}
