package patients

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderUnknown Gender = "unknown"
)

// Species is a tenant-scoped tag with trigram-based deduplication.
type Species struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	TenantID       primitive.ObjectID `bson:"tenant_id"`
	Name           string             `bson:"name"`
	NormalizedName string             `bson:"normalized_name"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	DeletedAt      *time.Time         `bson:"deleted_at,omitempty"`
}

// Patient represents a veterinary patient (animal).
type Patient struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	TenantID   primitive.ObjectID `bson:"tenant_id"`
	OwnerID    primitive.ObjectID `bson:"owner_id"`
	SpeciesID  primitive.ObjectID `bson:"species_id"`
	Name       string             `bson:"name"`
	Breed      string             `bson:"breed,omitempty"`
	Color      string             `bson:"color,omitempty"`
	BirthDate  *time.Time         `bson:"birth_date,omitempty"`
	Gender     Gender             `bson:"gender"`
	Weight     float64            `bson:"weight"`
	Microchip  string             `bson:"microchip,omitempty"`
	Sterilized bool               `bson:"sterilized"`
	AvatarURL  string             `bson:"avatar_url,omitempty"`
	Notes      string             `bson:"notes,omitempty"`
	Active     bool               `bson:"active"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
	DeletedAt  *time.Time         `bson:"deleted_at,omitempty"`
}
