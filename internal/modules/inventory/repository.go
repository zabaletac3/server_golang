package inventory

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	// Product CRUD
	Create(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Product, error)
	FindBySKU(ctx context.Context, sku string, tenantID primitive.ObjectID) (*Product, error)
	FindByBarcode(ctx context.Context, barcode string, tenantID primitive.ObjectID) (*Product, error)
	FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters ProductListFilters, params pagination.Params) ([]Product, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Stock operations
	UpdateStock(ctx context.Context, id primitive.ObjectID, quantity int, tenantID primitive.ObjectID) error

	// Alerts
	FindLowStockProducts(ctx context.Context, tenantID primitive.ObjectID) ([]Product, error)
	FindExpiringProducts(ctx context.Context, tenantID primitive.ObjectID, days int) ([]Product, error)
	FindExpiredProducts(ctx context.Context, tenantID primitive.ObjectID) ([]Product, error)

	// Stock Movement
	CreateStockMovement(ctx context.Context, movement *StockMovement) error
	FindStockMovements(ctx context.Context, filters StockMovementListFilters, tenantID primitive.ObjectID, params pagination.Params) ([]StockMovement, int64, error)

	// Category CRUD
	CreateCategory(ctx context.Context, category *Category) error
	FindCategoryByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Category, error)
	FindCategories(ctx context.Context, tenantID primitive.ObjectID) ([]Category, error)
	UpdateCategory(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error
	DeleteCategory(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error

	// Indexes
	EnsureIndexes(ctx context.Context) error
}

type productRepository struct {
	productsCollection  *mongo.Collection
	categoriesCollection *mongo.Collection
	movementsCollection *mongo.Collection
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *database.MongoDB) ProductRepository {
	return &productRepository{
		productsCollection:   db.Collection("products"),
		categoriesCollection: db.Collection("product_categories"),
		movementsCollection:  db.Collection("stock_movements"),
	}
}

// Product methods

func (r *productRepository) Create(ctx context.Context, product *Product) error {
	result, err := r.productsCollection.InsertOne(ctx, product)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrSKUAlreadyExists
		}
		return err
	}
	product.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *productRepository) FindByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Product, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var product Product
	err := r.productsCollection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) FindBySKU(ctx context.Context, sku string, tenantID primitive.ObjectID) (*Product, error) {
	filter := bson.M{
		"sku":        sku,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var product Product
	err := r.productsCollection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) FindByBarcode(ctx context.Context, barcode string, tenantID primitive.ObjectID) (*Product, error) {
	filter := bson.M{
		"barcode":    barcode,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var product Product
	err := r.productsCollection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) FindByFilters(ctx context.Context, tenantID primitive.ObjectID, filters ProductListFilters, params pagination.Params) ([]Product, int64, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	// Category filter
	if filters.CategoryID != "" {
		if catID, err := primitive.ObjectIDFromHex(filters.CategoryID); err == nil {
			filter["category_id"] = catID
		}
	}

	if filters.Category != "" {
		filter["category"] = filters.Category
	}

	// Active filter
	if filters.Active != nil {
		filter["active"] = *filters.Active
	}

	// Low stock filter
	if filters.LowStock {
		filter["$expr"] = bson.M{"$lte": []string{"$stock", "$min_stock"}}
	}

	// Expiring soon filter (within 30 days)
	if filters.ExpiringSoon {
		thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)
		filter["expiration_date"] = bson.M{
			"$lte": thirtyDaysFromNow,
			"$gte": time.Now(),
		}
	}

	// Expired filter
	if filters.Expired {
		filter["expiration_date"] = bson.M{"$lt": time.Now()}
	}

	// Search filter
	if filters.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"sku": bson.M{"$regex": filters.Search, "$options": "i"}},
			{"barcode": bson.M{"$regex": filters.Search, "$options": "i"}},
		}
	}

	// Count total
	total, err := r.productsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"name", 1}})

	cursor, err := r.productsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var products []Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Update(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.productsCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrSKUAlreadyExists
		}
		return err
	}

	if result.MatchedCount == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *productRepository) Delete(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
	// First check if product has stock
	product, err := r.FindByID(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if product.Stock > 0 {
		return ErrCannotDeleteWithStock
	}

	filter := bson.M{
		"_id":       id,
		"tenant_id": tenantID,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := r.productsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *productRepository) UpdateStock(ctx context.Context, id primitive.ObjectID, quantity int, tenantID primitive.ObjectID) error {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	update := bson.M{
		"$inc": bson.M{"stock": quantity},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.productsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *productRepository) FindLowStockProducts(ctx context.Context, tenantID primitive.ObjectID) ([]Product, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"active":     true,
		"$expr":      bson.M{"$lte": []string{"$stock", "$min_stock"}},
	}

	cursor, err := r.productsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepository) FindExpiringProducts(ctx context.Context, tenantID primitive.ObjectID, days int) ([]Product, error) {
	expiryThreshold := time.Now().AddDate(0, 0, days)

	filter := bson.M{
		"tenant_id":       tenantID,
		"deleted_at":      nil,
		"active":          true,
		"expiration_date": bson.M{"$lte": expiryThreshold, "$gte": time.Now()},
	}

	cursor, err := r.productsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepository) FindExpiredProducts(ctx context.Context, tenantID primitive.ObjectID) ([]Product, error) {
	filter := bson.M{
		"tenant_id":       tenantID,
		"deleted_at":      nil,
		"active":          true,
		"expiration_date": bson.M{"$lt": time.Now()},
	}

	cursor, err := r.productsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// Stock Movement methods

func (r *productRepository) CreateStockMovement(ctx context.Context, movement *StockMovement) error {
	result, err := r.movementsCollection.InsertOne(ctx, movement)
	if err != nil {
		return err
	}
	movement.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *productRepository) FindStockMovements(ctx context.Context, filters StockMovementListFilters, tenantID primitive.ObjectID, params pagination.Params) ([]StockMovement, int64, error) {
	filter := bson.M{
		"tenant_id": tenantID,
	}

	if filters.ProductID != "" {
		if productID, err := primitive.ObjectIDFromHex(filters.ProductID); err == nil {
			filter["product_id"] = productID
		}
	}

	if filters.Type != "" {
		filter["type"] = filters.Type
	}

	if filters.Reason != "" {
		filter["reason"] = filters.Reason
	}

	if filters.DateFrom != "" || filters.DateTo != "" {
		dateFilter := bson.M{}
		if filters.DateFrom != "" {
			if df, err := time.Parse(time.RFC3339, filters.DateFrom); err == nil {
				dateFilter["$gte"] = df
			}
		}
		if filters.DateTo != "" {
			if dt, err := time.Parse(time.RFC3339, filters.DateTo); err == nil {
				dateFilter["$lte"] = dt
			}
		}
		if len(dateFilter) > 0 {
			filter["created_at"] = dateFilter
		}
	}

	if filters.ReferenceID != "" {
		if refID, err := primitive.ObjectIDFromHex(filters.ReferenceID); err == nil {
			filter["reference_id"] = refID
		}
	}

	// Count total
	total, err := r.movementsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.movementsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var movements []StockMovement
	if err := cursor.All(ctx, &movements); err != nil {
		return nil, 0, err
	}

	return movements, total, nil
}

// Category methods

func (r *productRepository) CreateCategory(ctx context.Context, category *Category) error {
	result, err := r.categoriesCollection.InsertOne(ctx, category)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrCategoryNameExists
		}
		return err
	}
	category.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *productRepository) FindCategoryByID(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) (*Category, error) {
	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	var category Category
	err := r.categoriesCollection.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	return &category, nil
}

func (r *productRepository) FindCategories(ctx context.Context, tenantID primitive.ObjectID) ([]Category, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	cursor, err := r.categoriesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *productRepository) UpdateCategory(ctx context.Context, id primitive.ObjectID, updates bson.M, tenantID primitive.ObjectID) error {
	updates["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        id,
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	result, err := r.categoriesCollection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrCategoryNameExists
		}
		return err
	}

	if result.MatchedCount == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *productRepository) DeleteCategory(ctx context.Context, id primitive.ObjectID, tenantID primitive.ObjectID) error {
	filter := bson.M{
		"_id":       id,
		"tenant_id": tenantID,
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	result, err := r.categoriesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// EnsureIndexes creates required indexes for the inventory collections
func (r *productRepository) EnsureIndexes(ctx context.Context) error {
	// Products indexes
	productsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"sku", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"barcode", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{"name", 1}},
		},
		{
			Keys: bson.D{{"category", 1}},
		},
		{
			Keys: bson.D{{"expiration_date", 1}},
		},
		{
			Keys: bson.D{{"active", 1}},
		},
		{
			Keys: bson.D{{"stock", 1}, {"min_stock", 1}},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := r.productsCollection.Indexes().CreateMany(ctx, productsIndexes, opts)
	if err != nil {
		return err
	}

	// Categories indexes
	categoriesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"deleted_at", 1}},
		},
		{
			Keys:    bson.D{{"tenant_id", 1}, {"name", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}

	_, err = r.categoriesCollection.Indexes().CreateMany(ctx, categoriesIndexes, opts)
	if err != nil {
		return err
	}

	// Stock movements indexes
	movementsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"tenant_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"product_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"type", 1}, {"created_at", -1}},
		},
	}

	_, err = r.movementsCollection.Indexes().CreateMany(ctx, movementsIndexes, opts)
	if err != nil {
		return err
	}

	return nil
}
