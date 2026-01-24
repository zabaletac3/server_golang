package tenant

import (
	"time"
)

// CreateTenantDTO request para crear tenant
// @name CreateTenantDto
type CreateTenantDTO struct {
	// ID del propietario (referencia a users)
	OwnerID string `json:"owner_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	// Nombre del tenant
	Name string `json:"name" binding:"required,min=2" example:"Mi Empresa"`
}

// UpdateTenantDTO request para actualizar tenant
// @name UpdateTenantDto
// Nombre del tenant
type UpdateTenantDTO struct {
	Name string `json:"name" example:"Jane Doe"`
}

// UpdateStatusTenantDTO request para actualizar estado del tenant
// @name UpdateStatusTenantDto
// Estado del tenant
type UpdateStatusTenantDTO struct {
	Status TenantStatus `json:"status" binding:"required" example:"active"`
}

// TenantResponse respuesta de tenant
// @name TenantResponse
type TenantResponse struct {
	// ID del tenant
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// ID del propietario
	OwnerID string `json:"owner_id" example:"507f1f77bcf86cd799439011"`
	// Nombre del tenant
	Name string `json:"name" example:"Mi Empresa"`
	// Estado del tenant
	Status    TenantStatus `json:"status" example:"active"`
	CreatedAt time.Time    `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time    `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ToResponse convierte un Tenant a TenantResponse
func ToResponse(t *Tenant) *TenantResponse {
	return &TenantResponse{
		ID:        t.ID.Hex(),
		OwnerID:   t.OwnerID.Hex(),
		Name:      t.Name,
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// ToResponseList convierte una lista de Tenant a lista de TenantResponse
func ToResponseList(tenants []Tenant) []*TenantResponse {
	responses := make([]*TenantResponse, len(tenants))
	for i, t := range tenants {
		responses[i] = ToResponse(&t)
	}
	return responses
}