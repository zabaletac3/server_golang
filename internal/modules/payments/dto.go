package payments

import "time"

// CreatePaymentDTO request para crear pago
type CreatePaymentDTO struct {
	TenantID              string                 `json:"tenant_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	Amount                int64                  `json:"amount" binding:"required,min=1" example:"4900"`
	Currency              string                 `json:"currency" binding:"required" example:"USD"`
	PaymentMethod         string                 `json:"payment_method" binding:"required" example:"wompi"`
	Status                PaymentStatus          `json:"status" binding:"required" example:"completed"`
	ExternalTransactionID string                 `json:"external_transaction_id,omitempty" example:"wompi_tx_123456"`
	Concept               string                 `json:"concept" binding:"required" example:"Pago mensual - Plan Pro"`
	PeriodStart           *time.Time             `json:"period_start,omitempty" example:"2024-01-01T00:00:00Z"`
	PeriodEnd             *time.Time             `json:"period_end,omitempty" example:"2024-02-01T00:00:00Z"`
	ProcessedAt           *time.Time             `json:"processed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	FailureReason         string                 `json:"failure_reason,omitempty" example:"Fondos insuficientes"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentResponse respuesta de pago
type PaymentResponse struct {
	ID                    string                 `json:"id" example:"507f1f77bcf86cd799439011"`
	TenantID              string                 `json:"tenant_id" example:"507f1f77bcf86cd799439011"`
	Amount                int64                  `json:"amount" example:"4900"`
	Currency              string                 `json:"currency" example:"USD"`
	PaymentMethod         string                 `json:"payment_method" example:"wompi"`
	Status                PaymentStatus          `json:"status" example:"completed"`
	ExternalTransactionID string                 `json:"external_transaction_id,omitempty" example:"wompi_tx_123456"`
	Concept               string                 `json:"concept" example:"Pago mensual - Plan Pro"`
	PeriodStart           *time.Time             `json:"period_start,omitempty" example:"2024-01-01T00:00:00Z"`
	PeriodEnd             *time.Time             `json:"period_end,omitempty" example:"2024-02-01T00:00:00Z"`
	ProcessedAt           *time.Time             `json:"processed_at,omitempty" example:"2024-01-15T10:30:00Z"`
	FailureReason         string                 `json:"failure_reason,omitempty" example:"Fondos insuficientes"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt             time.Time              `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt             time.Time              `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ToResponse convierte un Payment a PaymentResponse
func ToResponse(p *Payment) *PaymentResponse {
	return &PaymentResponse{
		ID:                    p.ID.Hex(),
		TenantID:              p.TenantID.Hex(),
		Amount:                p.Amount,
		Currency:              p.Currency,
		PaymentMethod:         p.PaymentMethod,
		Status:                p.Status,
		ExternalTransactionID: p.ExternalTransactionID,
		Concept:               p.Concept,
		PeriodStart:           p.PeriodStart,
		PeriodEnd:             p.PeriodEnd,
		ProcessedAt:           p.ProcessedAt,
		FailureReason:         p.FailureReason,
		Metadata:              p.Metadata,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}

// ToResponseList convierte una lista de Payment a lista de PaymentResponse
func ToResponseList(payments []Payment) []*PaymentResponse {
	responses := make([]*PaymentResponse, len(payments))
	for i, p := range payments {
		responses[i] = ToResponse(&p)
	}
	return responses
}
