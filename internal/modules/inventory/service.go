package inventory

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/modules/notifications"
	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// NotificationSender defines the interface for sending notifications
type NotificationSender interface {
	SendToStaff(ctx context.Context, dto *notifications.SendStaffDTO) error
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*users.User, error)
}

// Service provides business logic for inventory
type Service struct {
	repo            ProductRepository
	userRepo        UserRepository
	notificationSvc NotificationSender
}

// NewService creates a new inventory service
func NewService(repo ProductRepository, userRepo UserRepository, notificationSvc NotificationSender) *Service {
	return &Service{
		repo:            repo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
	}
}

// CreateProduct creates a new product
func (s *Service) CreateProduct(ctx context.Context, dto *CreateProductDTO, tenantID primitive.ObjectID) (*Product, error) {
	// Validate category if provided
	var categoryID primitive.ObjectID
	if dto.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(dto.CategoryID)
		if err != nil {
			return nil, ErrValidation("category_id", "invalid category ID format")
		}
		_, err = s.repo.FindCategoryByID(ctx, catID, tenantID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		categoryID = catID
	}

	// Validate supplier if provided
	var supplierID primitive.ObjectID
	if dto.SupplierID != "" {
		supID, err := primitive.ObjectIDFromHex(dto.SupplierID)
		if err != nil {
			return nil, ErrValidation("supplier_id", "invalid supplier ID format")
		}
		supplierID = supID
	}

	// Validate sale price >= purchase price
	if dto.SalePrice < dto.PurchasePrice {
		return nil, ErrSalePriceTooLow
	}

	// Validate expiration date
	var expirationDate *time.Time
	if dto.ExpirationDate != "" {
		expDate, err := time.Parse(time.RFC3339, dto.ExpirationDate)
		if err != nil {
			return nil, ErrValidation("expiration_date", "invalid date format, use RFC3339")
		}
		expirationDate = &expDate
	}

	// Create product
	now := time.Now()
	product := &Product{
		ID:             primitive.NewObjectID(),
		TenantID:       tenantID,
		CategoryID:     categoryID,
		Name:           dto.Name,
		Description:    dto.Description,
		SKU:            dto.SKU,
		Barcode:        dto.Barcode,
		Category:       ProductCategory(dto.Category),
		Unit:           ProductUnit(dto.Unit),
		PurchasePrice:  dto.PurchasePrice,
		SalePrice:      dto.SalePrice,
		Stock:          dto.Stock,
		MinStock:       dto.MinStock,
		ExpirationDate: expirationDate,
		SupplierID:     supplierID,
		Active:         dto.Active,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// GetProduct gets a product by ID
func (s *Service) GetProduct(ctx context.Context, id string, tenantID primitive.ObjectID) (*Product, error) {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid product ID format")
	}

	product, err := s.repo.FindByID(ctx, productID, tenantID)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// ListProducts lists products with filters
func (s *Service) ListProducts(ctx context.Context, filters ProductListFilters, tenantID primitive.ObjectID, params pagination.Params) ([]Product, int64, error) {
	return s.repo.FindByFilters(ctx, tenantID, filters, params)
}

// UpdateProduct updates a product
func (s *Service) UpdateProduct(ctx context.Context, id string, dto *UpdateProductDTO, tenantID primitive.ObjectID) (*Product, error) {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid product ID format")
	}

	product, err := s.repo.FindByID(ctx, productID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}

	// Validate and update category if provided
	if dto.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(dto.CategoryID)
		if err != nil {
			return nil, ErrValidation("category_id", "invalid category ID format")
		}
		_, err = s.repo.FindCategoryByID(ctx, catID, tenantID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		updates["category_id"] = catID
	}

	if dto.Name != "" {
		updates["name"] = dto.Name
	}

	if dto.Description != "" {
		updates["description"] = dto.Description
	}

	if dto.SKU != "" {
		updates["sku"] = dto.SKU
	}

	if dto.Barcode != "" {
		updates["barcode"] = dto.Barcode
	}

	if dto.Category != "" {
		updates["category"] = dto.Category
	}

	if dto.Unit != "" {
		updates["unit"] = dto.Unit
	}

	if dto.PurchasePrice > 0 {
		updates["purchase_price"] = dto.PurchasePrice
	}

	if dto.SalePrice > 0 {
		if dto.PurchasePrice > 0 && dto.SalePrice < dto.PurchasePrice {
			return nil, ErrSalePriceTooLow
		}
		// Also check against existing purchase price if not updating it
		if dto.PurchasePrice == 0 && dto.SalePrice < product.PurchasePrice {
			return nil, ErrSalePriceTooLow
		}
		updates["sale_price"] = dto.SalePrice
	}

	if dto.MinStock > 0 {
		updates["min_stock"] = dto.MinStock
	}

	if dto.ExpirationDate != "" {
		expDate, err := time.Parse(time.RFC3339, dto.ExpirationDate)
		if err != nil {
			return nil, ErrValidation("expiration_date", "invalid date format, use RFC3339")
		}
		updates["expiration_date"] = expDate
	} else if dto.ExpirationDate == "" && dto.ExpirationDate != product.ExpirationDate.String() {
		// Allow clearing expiration date
		updates["expiration_date"] = nil
	}

	if dto.SupplierID != "" {
		supID, err := primitive.ObjectIDFromHex(dto.SupplierID)
		if err != nil {
			return nil, ErrValidation("supplier_id", "invalid supplier ID format")
		}
		updates["supplier_id"] = supID
	}

	updates["active"] = dto.Active

	if err := s.repo.Update(ctx, productID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedProduct, err := s.repo.FindByID(ctx, productID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedProduct, nil
}

// DeleteProduct soft deletes a product
func (s *Service) DeleteProduct(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid product ID format")
	}

	return s.repo.Delete(ctx, productID, tenantID)
}

// StockIn adds stock to a product
func (s *Service) StockIn(ctx context.Context, id string, dto *StockInDTO, tenantID primitive.ObjectID, userID primitive.ObjectID) (*StockMovement, error) {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid product ID format")
	}

	product, err := s.repo.FindByID(ctx, productID, tenantID)
	if err != nil {
		return nil, err
	}

	// Validate movement type
	movementType := StockMovementType(dto.Reason)
	if !IsValidStockMovementType(string(movementType)) {
		return nil, ErrValidation("reason", "invalid stock movement reason")
	}

	// Update stock
	stockBefore := product.Stock
	if err := s.repo.UpdateStock(ctx, productID, dto.Quantity, tenantID); err != nil {
		return nil, err
	}

	// Get updated product to get new stock
	updatedProduct, err := s.repo.FindByID(ctx, productID, tenantID)
	if err != nil {
		return nil, err
	}
	stockAfter := updatedProduct.Stock

	// Create stock movement record
	var referenceID primitive.ObjectID
	if dto.ReferenceID != "" {
		if refID, err := primitive.ObjectIDFromHex(dto.ReferenceID); err == nil {
			referenceID = refID
		}
	}

	movement := &StockMovement{
		ID:          primitive.NewObjectID(),
		TenantID:    tenantID,
		ProductID:   productID,
		Type:        StockMovementIn,
		Reason:      StockMovementReason(dto.Reason),
		Quantity:    dto.Quantity,
		StockBefore: stockBefore,
		StockAfter:  stockAfter,
		ReferenceID: referenceID,
		UserID:      userID,
		Notes:       dto.Notes,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateStockMovement(ctx, movement); err != nil {
		return nil, err
	}

	return movement, nil
}

// StockOut deducts stock from a product
func (s *Service) StockOut(ctx context.Context, id string, dto *StockOutDTO, tenantID primitive.ObjectID, userID primitive.ObjectID) (*StockMovement, error) {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid product ID format")
	}

	product, err := s.repo.FindByID(ctx, productID, tenantID)
	if err != nil {
		return nil, err
	}

	// Check if there's enough stock
	if product.Stock < dto.Quantity {
		return nil, ErrInsufficientStock
	}

	// Validate movement type
	if !IsValidStockMovementReason(dto.Reason) {
		return nil, ErrValidation("reason", "invalid stock movement reason")
	}

	// Update stock (negative quantity to deduct)
	stockBefore := product.Stock
	if err := s.repo.UpdateStock(ctx, productID, -dto.Quantity, tenantID); err != nil {
		return nil, err
	}

	// Get updated product to get new stock
	updatedProduct, err := s.repo.FindByID(ctx, productID, tenantID)
	if err != nil {
		return nil, err
	}
	stockAfter := updatedProduct.Stock

	// Create stock movement record
	var referenceID primitive.ObjectID
	if dto.ReferenceID != "" {
		if refID, err := primitive.ObjectIDFromHex(dto.ReferenceID); err == nil {
			referenceID = refID
		}
	}

	movement := &StockMovement{
		ID:          primitive.NewObjectID(),
		TenantID:    tenantID,
		ProductID:   productID,
		Type:        StockMovementOut,
		Reason:      StockMovementReason(dto.Reason),
		Quantity:    dto.Quantity,
		StockBefore: stockBefore,
		StockAfter:  stockAfter,
		ReferenceID: referenceID,
		UserID:      userID,
		Notes:       dto.Notes,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateStockMovement(ctx, movement); err != nil {
		return nil, err
	}

	return movement, nil
}

// GetStockMovements lists stock movements with filters
func (s *Service) GetStockMovements(ctx context.Context, filters StockMovementListFilters, tenantID primitive.ObjectID, params pagination.Params) ([]StockMovement, int64, error) {
	return s.repo.FindStockMovements(ctx, filters, tenantID, params)
}

// GetLowStockProducts gets products with low stock
func (s *Service) GetLowStockProducts(ctx context.Context, tenantID primitive.ObjectID) ([]Product, error) {
	return s.repo.FindLowStockProducts(ctx, tenantID)
}

// GetExpiringProducts gets products expiring soon
func (s *Service) GetExpiringProducts(ctx context.Context, tenantID primitive.ObjectID, days int) ([]Product, error) {
	return s.repo.FindExpiringProducts(ctx, tenantID, days)
}

// GetExpiredProducts gets expired products
func (s *Service) GetExpiredProducts(ctx context.Context, tenantID primitive.ObjectID) ([]Product, error) {
	return s.repo.FindExpiredProducts(ctx, tenantID)
}

// GetProductAlerts gets all product alerts (low stock, expiring, expired)
func (s *Service) GetProductAlerts(ctx context.Context, tenantID primitive.ObjectID) ([]ProductAlertResponse, error) {
	alerts := make([]ProductAlertResponse, 0)

	// Low stock alerts
	lowStockProducts, err := s.GetLowStockProducts(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	for _, p := range lowStockProducts {
		alert := ProductAlertResponse{
			ProductID:       p.ID.Hex(),
			ProductName:     p.Name,
			AlertType:       "low_stock",
			CurrentStock:    p.Stock,
			MinStock:        p.MinStock,
			StockDifference: p.Stock - p.MinStock,
		}
		alerts = append(alerts, alert)
	}

	// Expiring products alerts
	expiringProducts, err := s.GetExpiringProducts(ctx, tenantID, 30)
	if err != nil {
		return nil, err
	}

	for _, p := range expiringProducts {
		daysUntil := int(p.ExpirationDate.Sub(time.Now()).Hours() / 24)
		alert := ProductAlertResponse{
			ProductID:       p.ID.Hex(),
			ProductName:     p.Name,
			AlertType:       "expiring",
			CurrentStock:    p.Stock,
			MinStock:        p.MinStock,
			ExpirationDate:  p.ExpirationDate,
			DaysUntilExpiry: &daysUntil,
		}
		alerts = append(alerts, alert)
	}

	// Expired products alerts
	expiredProducts, err := s.GetExpiredProducts(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	for _, p := range expiredProducts {
		daysSince := int(time.Since(*p.ExpirationDate).Hours() / 24)
		alert := ProductAlertResponse{
			ProductID:       p.ID.Hex(),
			ProductName:     p.Name,
			AlertType:       "expired",
			CurrentStock:    p.Stock,
			MinStock:        p.MinStock,
			ExpirationDate:  p.ExpirationDate,
			DaysUntilExpiry: &daysSince,
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// SendLowStockAlerts sends notifications for low stock products
func (s *Service) SendLowStockAlerts(ctx context.Context, tenantID primitive.ObjectID) error {
	products, err := s.GetLowStockProducts(ctx, tenantID)
	if err != nil {
		return err
	}

	for _, p := range products {
		s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
			UserID:   primitive.NilObjectID.Hex(), // Broadcast to all staff
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeStaffSystemAlert,
			Title:    "Stock Bajo - " + p.Name,
			Body:     "El producto " + p.Name + " tiene stock bajo: " + string(rune(p.Stock)) + " unidades (mínimo: " + string(rune(p.MinStock)) + ")",
			Data: map[string]string{
				"product_id": p.ID.Hex(),
				"product_name": p.Name,
				"current_stock": string(rune(p.Stock)),
				"min_stock": string(rune(p.MinStock)),
			},
		})
	}

	return nil
}

// SendExpiringAlerts sends notifications for expiring products
func (s *Service) SendExpiringAlerts(ctx context.Context, tenantID primitive.ObjectID) error {
	products, err := s.GetExpiringProducts(ctx, tenantID, 30)
	if err != nil {
		return err
	}

	for _, p := range products {
		daysUntil := int(p.ExpirationDate.Sub(time.Now()).Hours() / 24)
		s.notificationSvc.SendToStaff(ctx, &notifications.SendStaffDTO{
			UserID:   primitive.NilObjectID.Hex(),
			TenantID: tenantID.Hex(),
			Type:     notifications.TypeStaffSystemAlert,
			Title:    "Producto por Vencer - " + p.Name,
			Body:     "El producto " + p.Name + " vence en " + string(rune(daysUntil)) + " días",
			Data: map[string]string{
				"product_id": p.ID.Hex(),
				"product_name": p.Name,
				"days_until_expiry": string(rune(daysUntil)),
			},
		})
	}

	return nil
}

// CreateCategory creates a new category
func (s *Service) CreateCategory(ctx context.Context, dto *CreateCategoryDTO, tenantID primitive.ObjectID) (*Category, error) {
	// Validate parent category if provided
	var parentID primitive.ObjectID
	if dto.ParentID != "" {
		pID, err := primitive.ObjectIDFromHex(dto.ParentID)
		if err != nil {
			return nil, ErrValidation("parent_id", "invalid parent category ID format")
		}
		_, err = s.repo.FindCategoryByID(ctx, pID, tenantID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		parentID = pID
	}

	category := &Category{
		ID:          primitive.NewObjectID(),
		TenantID:    tenantID,
		Name:        dto.Name,
		Description: dto.Description,
		ParentID:    parentID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory gets a category by ID
func (s *Service) GetCategory(ctx context.Context, id string, tenantID primitive.ObjectID) (*Category, error) {
	categoryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid category ID format")
	}

	category, err := s.repo.FindCategoryByID(ctx, categoryID, tenantID)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// ListCategories lists all categories for a tenant
func (s *Service) ListCategories(ctx context.Context, tenantID primitive.ObjectID) ([]Category, error) {
	return s.repo.FindCategories(ctx, tenantID)
}

// UpdateCategory updates a category
func (s *Service) UpdateCategory(ctx context.Context, id string, dto *UpdateCategoryDTO, tenantID primitive.ObjectID) (*Category, error) {
	categoryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrValidation("id", "invalid category ID format")
	}

	_, err = s.repo.FindCategoryByID(ctx, categoryID, tenantID)
	if err != nil {
		return nil, err
	}

	updates := bson.M{}

	if dto.Name != "" {
		updates["name"] = dto.Name
	}

	if dto.Description != "" {
		updates["description"] = dto.Description
	}

	if dto.ParentID != "" {
		pID, err := primitive.ObjectIDFromHex(dto.ParentID)
		if err != nil {
			return nil, ErrValidation("parent_id", "invalid parent category ID format")
		}
		_, err = s.repo.FindCategoryByID(ctx, pID, tenantID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		updates["parent_id"] = pID
	}

	if err := s.repo.UpdateCategory(ctx, categoryID, updates, tenantID); err != nil {
		return nil, err
	}

	updatedCategory, err := s.repo.FindCategoryByID(ctx, categoryID, tenantID)
	if err != nil {
		return nil, err
	}

	return updatedCategory, nil
}

// DeleteCategory soft deletes a category
func (s *Service) DeleteCategory(ctx context.Context, id string, tenantID primitive.ObjectID) error {
	categoryID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrValidation("id", "invalid category ID format")
	}

	return s.repo.DeleteCategory(ctx, categoryID, tenantID)
}
