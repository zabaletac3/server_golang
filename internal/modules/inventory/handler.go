package inventory

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eren_dev/go_server/internal/shared/auth"
	sharedMiddleware "github.com/eren_dev/go_server/internal/shared/middleware"
	"github.com/eren_dev/go_server/internal/shared/pagination"
	"github.com/eren_dev/go_server/internal/shared/validation"
)

// Handler handles HTTP requests for inventory
type Handler struct {
	service *Service
}

// NewHandler creates a new inventory handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ==================== PRODUCTS ====================

// CreateProduct creates a new product
// @Summary Create product
// @Description Create a new product in the inventory
// @Tags inventory
// @Accept json
// @Produce json
// @Param product body CreateProductDTO true "Product data"
// @Success 201 {object} ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/products [post]
func (h *Handler) CreateProduct(c *gin.Context) (any, error) {
	var dto CreateProductDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	product, err := h.service.CreateProduct(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return product.ToResponse(), nil
}

// GetProduct gets a product by ID
// @Summary Get product
// @Description Get product details by ID
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "product ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	product, err := h.service.GetProduct(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return product.ToResponse(), nil
}

// ListProducts lists products with filters
// @Summary List products
// @Description Get a paginated list of products with optional filters
// @Tags inventory
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param active query bool false "Filter by active status"
// @Param low_stock query bool false "Filter low stock products"
// @Param expiring query bool false "Filter expiring products"
// @Param expired query bool false "Filter expired products"
// @Param search query string false "Search by name, SKU, barcode"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/products [get]
func (h *Handler) ListProducts(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := ProductListFilters{
		Category:     c.Query("category"),
		Search:       c.Query("search"),
		LowStock:     c.Query("low_stock") == "true",
		ExpiringSoon: c.Query("expiring") == "true",
		Expired:      c.Query("expired") == "true",
	}

	if active := c.Query("active"); active != "" {
		activeBool := active == "true"
		filters.Active = &activeBool
	}

	products, total, err := h.service.ListProducts(c.Request.Context(), filters, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]ProductResponse, len(products))
	for i, p := range products {
		data[i] = *p.ToResponse()
	}

	return gin.H{
		"data": data,
		"pagination": gin.H{
			"skip":        params.Skip,
			"limit":       params.Limit,
			"total":       total,
			"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
		},
	}, nil
}

// UpdateProduct updates a product
// @Summary Update product
// @Description Update product details
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body UpdateProductDTO true "Updated product data"
// @Success 200 {object} ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "product ID is required")
	}

	var dto UpdateProductDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	product, err := h.service.UpdateProduct(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return product.ToResponse(), nil
}

// DeleteProduct deletes a product
// @Summary Delete product
// @Description Soft delete a product (only if stock = 0)
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "product ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteProduct(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Product deleted successfully"}, nil
}

// StockIn adds stock to a product
// @Summary Add stock
// @Description Add stock to a product (purchase, return, adjustment)
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param stock body StockInDTO true "Stock in data"
// @Success 200 {object} StockMovementResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/products/{id}/stock-in [post]
func (h *Handler) StockIn(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "product ID is required")
	}

	var dto StockInDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)
	userIDStr := auth.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrValidation("user_id", "invalid user ID format")
	}

	movement, err := h.service.StockIn(c.Request.Context(), id, &dto, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return movement.ToResponse(), nil
}

// StockOut deducts stock from a product
// @Summary Deduct stock
// @Description Deduct stock from a product (sale, treatment, expired, etc.)
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param stock body StockOutDTO true "Stock out data"
// @Success 200 {object} StockMovementResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/products/{id}/stock-out [post]
func (h *Handler) StockOut(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "product ID is required")
	}

	var dto StockOutDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)
	userIDStr := auth.GetUserID(c)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, ErrValidation("user_id", "invalid user ID format")
	}

	movement, err := h.service.StockOut(c.Request.Context(), id, &dto, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return movement.ToResponse(), nil
}

// GetStockMovements lists stock movements
// @Summary List stock movements
// @Description Get a paginated list of stock movements
// @Tags inventory
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param product_id query string false "Filter by product ID"
// @Param type query string false "Filter by type (in, out, adjustment, expired)"
// @Param reason query string false "Filter by reason"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/stock-movements [get]
func (h *Handler) GetStockMovements(c *gin.Context) (any, error) {
	params := pagination.FromContext(c)
	tenantID := sharedMiddleware.GetTenantID(c)

	filters := StockMovementListFilters{
		ProductID:   c.Query("product_id"),
		Type:        c.Query("type"),
		Reason:      c.Query("reason"),
		DateFrom:    c.Query("date_from"),
		DateTo:      c.Query("date_to"),
		ReferenceID: c.Query("reference_id"),
	}

	movements, total, err := h.service.GetStockMovements(c.Request.Context(), filters, tenantID, params)
	if err != nil {
		return nil, err
	}

	data := make([]StockMovementResponse, len(movements))
	for i, m := range movements {
		data[i] = *m.ToResponse()
	}

	return gin.H{
		"data": data,
		"pagination": gin.H{
			"skip":        params.Skip,
			"limit":       params.Limit,
			"total":       total,
			"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
		},
	}, nil
}

// GetLowStockProducts gets products with low stock
// @Summary Get low stock products
// @Description Get products with stock below minimum threshold
// @Tags inventory
// @Accept json
// @Produce json
// @Success 200 {object} []ProductResponse
// @Security BearerAuth
// @Router /api/products/low-stock [get]
func (h *Handler) GetLowStockProducts(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	products, err := h.service.GetLowStockProducts(c.Request.Context(), tenantID)
	if err != nil {
		return nil, err
	}

	data := make([]ProductResponse, len(products))
	for i, p := range products {
		data[i] = *p.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// GetExpiringProducts gets products expiring soon
// @Summary Get expiring products
// @Description Get products expiring within 30 days
// @Tags inventory
// @Accept json
// @Produce json
// @Success 200 {object} []ProductResponse
// @Security BearerAuth
// @Router /api/products/expiring [get]
func (h *Handler) GetExpiringProducts(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	products, err := h.service.GetExpiringProducts(c.Request.Context(), tenantID, 30)
	if err != nil {
		return nil, err
	}

	data := make([]ProductResponse, len(products))
	for i, p := range products {
		data[i] = *p.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// GetProductAlerts gets all product alerts
// @Summary Get product alerts
// @Description Get all product alerts (low stock, expiring, expired)
// @Tags inventory
// @Accept json
// @Produce json
// @Success 200 {object} []ProductAlertResponse
// @Security BearerAuth
// @Router /api/products/alerts [get]
func (h *Handler) GetProductAlerts(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	alerts, err := h.service.GetProductAlerts(c.Request.Context(), tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"data": alerts}, nil
}

// ==================== CATEGORIES ====================

// CreateCategory creates a new category
// @Summary Create category
// @Description Create a new product category
// @Tags inventory
// @Accept json
// @Produce json
// @Param category body CreateCategoryDTO true "Category data"
// @Success 201 {object} CategoryResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/categories [post]
func (h *Handler) CreateCategory(c *gin.Context) (any, error) {
	var dto CreateCategoryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	category, err := h.service.CreateCategory(c.Request.Context(), &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return category.ToResponse(), nil
}

// GetCategory gets a category by ID
// @Summary Get category
// @Description Get category details by ID
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} CategoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/categories/{id} [get]
func (h *Handler) GetCategory(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "category ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	category, err := h.service.GetCategory(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return category.ToResponse(), nil
}

// ListCategories lists all categories
// @Summary List categories
// @Description Get all categories for the tenant
// @Tags inventory
// @Accept json
// @Produce json
// @Success 200 {object} []CategoryResponse
// @Security BearerAuth
// @Router /api/categories [get]
func (h *Handler) ListCategories(c *gin.Context) (any, error) {
	tenantID := sharedMiddleware.GetTenantID(c)

	categories, err := h.service.ListCategories(c.Request.Context(), tenantID)
	if err != nil {
		return nil, err
	}

	data := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		data[i] = *cat.ToResponse()
	}

	return gin.H{"data": data}, nil
}

// UpdateCategory updates a category
// @Summary Update category
// @Description Update category details
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body UpdateCategoryDTO true "Updated category data"
// @Success 200 {object} CategoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/categories/{id} [put]
func (h *Handler) UpdateCategory(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "category ID is required")
	}

	var dto UpdateCategoryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		return nil, validation.Validate(err)
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	category, err := h.service.UpdateCategory(c.Request.Context(), id, &dto, tenantID)
	if err != nil {
		return nil, err
	}

	return category.ToResponse(), nil
}

// DeleteCategory deletes a category
// @Summary Delete category
// @Description Soft delete a category
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/categories/{id} [delete]
func (h *Handler) DeleteCategory(c *gin.Context) (any, error) {
	id := c.Param("id")
	if id == "" {
		return nil, ErrValidation("id", "category ID is required")
	}

	tenantID := sharedMiddleware.GetTenantID(c)

	err := h.service.DeleteCategory(c.Request.Context(), id, tenantID)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Category deleted successfully"}, nil
}
