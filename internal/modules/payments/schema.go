package payments

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payment struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	
	// Relaciones
	TenantID primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	PlanID   primitive.ObjectID `bson:"plan_id,omitempty" json:"plan_id,omitempty"`
	
	// Monto (en centavos para evitar problemas de redondeo)
	Amount   int64  `bson:"amount" json:"amount"`
	Currency string `bson:"currency" json:"currency"`
	
	// Método y Estado
	PaymentMethod string        `bson:"payment_method" json:"payment_method"` // wompi, stripe, bank_transfer
	Status        PaymentStatus `bson:"status" json:"status"`
	
	// Transacción Externa
	ExternalTransactionID string `bson:"external_transaction_id,omitempty" json:"external_transaction_id,omitempty"`
	
	// Descripción
	Concept string `bson:"concept" json:"concept"`
	
	// Periodo cubierto (para suscripciones)
	PeriodStart *time.Time `bson:"period_start,omitempty" json:"period_start,omitempty"`
	PeriodEnd   *time.Time `bson:"period_end,omitempty" json:"period_end,omitempty"`
	
	// Auditoría
	ProcessedAt   *time.Time `bson:"processed_at,omitempty" json:"processed_at,omitempty"`
	FailureReason string     `bson:"failure_reason,omitempty" json:"failure_reason,omitempty"`
	
	// Metadata del provider
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	
	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
