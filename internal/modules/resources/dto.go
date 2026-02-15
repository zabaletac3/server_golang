package resources

import (
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// CreateResourceDTO request para crear un recurso
// @name CreateResourceDTO
type CreateResourceDTO struct {
	// ID del tenant al que pertenece
	TenantId string `json:"tenant_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	// Nombre del recurso (debe coincidir con el segmento de ruta, ej: appointments)
	Name string `json:"name" binding:"required,min=2" example:"appointments"`
	// Descripción del recurso
	Description string `json:"description" example:"Módulo de citas veterinarias"`
}

// UpdateResourceDTO request para actualizar un recurso
// @name UpdateResourceDTO
type UpdateResourceDTO struct {
	// Nombre del recurso
	Name string `json:"name" example:"appointments"`
	// Descripción del recurso
	Description string `json:"description" example:"Módulo de citas veterinarias"`
}

// ResourceResponse respuesta de recurso
// @name ResourceResponse
type ResourceResponse struct {
	// ID único del recurso
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// ID del tenant
	TenantId string `json:"tenant_id" example:"507f1f77bcf86cd799439011"`
	// Nombre del recurso
	Name string `json:"name" example:"appointments"`
	// Descripción del recurso
	Description string `json:"description" example:"Módulo de citas veterinarias"`
	// Fecha de creación
	CreatedAt time.Time `json:"created_at"`
	// Fecha de actualización
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginatedResourcesResponse respuesta paginada de recursos
// @name PaginatedResourcesResponse
type PaginatedResourcesResponse struct {
	// Lista de recursos
	Data []*ResourceResponse `json:"data"`
	// Información de paginación
	Pagination *pagination.PaginationInfo `json:"pagination"`
}

func ToResponse(r *Resource) *ResourceResponse {
	return &ResourceResponse{
		ID:          r.ID.Hex(),
		TenantId:    r.TenantId.Hex(),
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func ToResponseList(resources []*Resource) []*ResourceResponse {
	result := make([]*ResourceResponse, len(resources))
	for i, r := range resources {
		result[i] = ToResponse(r)
	}
	return result
}
