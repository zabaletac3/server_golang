package plans

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Plan struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	
	// Pricing
	MonthlyPrice float64 `bson:"monthly_price" json:"monthly_price"`
	AnnualPrice  float64 `bson:"annual_price" json:"annual_price"`
	Currency     string  `bson:"currency" json:"currency"`
	
	// Límites
	MaxUsers       int `bson:"max_users" json:"max_users"`
	MaxBranches    int `bson:"max_branches" json:"max_branches"`
	StorageLimitGB int `bson:"storage_limit_gb" json:"storage_limit_gb"`
	
	// Features
	Features []string `bson:"features,omitempty" json:"features,omitempty"`
	
	// Visibilidad
	IsVisible bool `bson:"is_visible" json:"is_visible"`
	
	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
