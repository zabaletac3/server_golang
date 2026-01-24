package users

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/eren_dev/go_server/internal/shared/database"
	"github.com/eren_dev/go_server/internal/shared/pagination"
)

type UserRepository interface {
	Create(ctx context.Context, dto *CreateUserDTO) (*User, error)
	CreateWithPassword(ctx context.Context, name, email, hashedPassword string) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	FindAll(ctx context.Context, params pagination.Params) ([]*User, int64, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id string, dto *UpdateUserDTO) (*User, error)
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *database.MongoDB) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) Create(ctx context.Context, dto *CreateUserDTO) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		Name:      dto.Name,
		Email:     dto.Email,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (r *userRepository) CreateWithPassword(ctx context.Context, name, email, hashedPassword string) (*User, error) {
	now := time.Now()
	user := &User{
		Name:      name,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (r *userRepository) FindAll(ctx context.Context, params pagination.Params) ([]*User, int64, error) {
	filter := bson.M{"deleted_at": nil}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip(params.Skip).
		SetLimit(params.Limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var user User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, id string, dto *UpdateUserDTO) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if dto.Name != "" {
		update["$set"].(bson.M)["name"] = dto.Name
	}
	if dto.Email != "" {
		update["$set"].(bson.M)["email"] = dto.Email
	}

	var user User
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		update,
		nil,
	).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidUserID
	}

	now := time.Now()
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"deleted_at": now}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}
