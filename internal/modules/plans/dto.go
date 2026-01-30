package plans

import "time"

// CreatePlanDTO request para crear plan
type CreatePlanDTO struct {
	Name           string   `json:"name" binding:"required" example:"Pro Plan"`
	Description    string   `json:"description,omitempty" example:"Plan profesional con todas las funcionalidades"`
	MonthlyPrice   float64  `json:"monthly_price" binding:"required,min=0" example:"49.00"`
	AnnualPrice    float64  `json:"annual_price" binding:"required,min=0" example:"490.00"`
	Currency       string   `json:"currency" binding:"required" example:"USD"`
	MaxUsers       int      `json:"max_users" binding:"required,min=1" example:"10"`
	MaxBranches    int      `json:"max_branches" binding:"required,min=1" example:"3"`
	StorageLimitGB int      `json:"storage_limit_gb" binding:"required,min=1" example:"10"`
	Features       []string `json:"features,omitempty" example:"Gestión de pacientes,Historial clínico,Facturación"`
	IsVisible      bool     `json:"is_visible" example:"true"`
}

// UpdatePlanDTO request para actualizar plan
type UpdatePlanDTO struct {
	Name           string   `json:"name,omitempty" example:"Pro Plan"`
	Description    string   `json:"description,omitempty" example:"Plan profesional"`
	MonthlyPrice   float64  `json:"monthly_price,omitempty" example:"49.00"`
	AnnualPrice    float64  `json:"annual_price,omitempty" example:"490.00"`
	MaxUsers       int      `json:"max_users,omitempty" example:"10"`
	MaxBranches    int      `json:"max_branches,omitempty" example:"3"`
	StorageLimitGB int      `json:"storage_limit_gb,omitempty" example:"10"`
	Features       []string `json:"features,omitempty"`
	IsVisible      *bool    `json:"is_visible,omitempty"`
}

// PlanResponse respuesta de plan
type PlanResponse struct {
	ID             string    `json:"id" example:"507f1f77bcf86cd799439011"`
	Name           string    `json:"name" example:"Pro Plan"`
	Description    string    `json:"description" example:"Plan profesional"`
	MonthlyPrice   float64   `json:"monthly_price" example:"49.00"`
	AnnualPrice    float64   `json:"annual_price" example:"490.00"`
	Currency       string    `json:"currency" example:"USD"`
	MaxUsers       int       `json:"max_users" example:"10"`
	MaxBranches    int       `json:"max_branches" example:"3"`
	StorageLimitGB int       `json:"storage_limit_gb" example:"10"`
	Features       []string  `json:"features" example:"Gestión de pacientes,Historial clínico"`
	IsVisible      bool      `json:"is_visible" example:"true"`
	CreatedAt      time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ToResponse convierte un Plan a PlanResponse
func ToResponse(p *Plan) *PlanResponse {
	return &PlanResponse{
		ID:             p.ID.Hex(),
		Name:           p.Name,
		Description:    p.Description,
		MonthlyPrice:   p.MonthlyPrice,
		AnnualPrice:    p.AnnualPrice,
		Currency:       p.Currency,
		MaxUsers:       p.MaxUsers,
		MaxBranches:    p.MaxBranches,
		StorageLimitGB: p.StorageLimitGB,
		Features:       p.Features,
		IsVisible:      p.IsVisible,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// ToResponseList convierte una lista de Plan a lista de PlanResponse
func ToResponseList(plans []Plan) []*PlanResponse {
	responses := make([]*PlanResponse, len(plans))
	for i, p := range plans {
		responses[i] = ToResponse(&p)
	}
	return responses
}
