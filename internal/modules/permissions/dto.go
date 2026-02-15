package permissions

import (
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// CreatePermissionDTO request para crear un permiso
// @name CreatePermissionDTO
type CreatePermissionDTO struct {
	// ID del tenant al que pertenece
	TenantId string `json:"tenant_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	// ID del recurso asociado
	ResourceId string `json:"resource_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	// Acción permitida: get, post, put, patch, delete
	Action string `json:"action" binding:"required" example:"get"`
}

// UpdatePermissionDTO request para actualizar un permiso
// @name UpdatePermissionDTO
type UpdatePermissionDTO struct {
	// Acción permitida: get, post, put, patch, delete
	Action string `json:"action" example:"post"`
}

// ResourceRef referencia simplificada de un recurso dentro de un permiso
type ResourceRef struct {
	// ID del recurso
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// Nombre del recurso
	Name string `json:"name" example:"appointments"`
}

// PermissionResponse respuesta de permiso con recurso poblado
// @name PermissionResponse
type PermissionResponse struct {
	// ID único del permiso
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// ID del tenant
	TenantId string `json:"tenant_id" example:"507f1f77bcf86cd799439011"`
	// Recurso asociado (poblado)
	Resource ResourceRef `json:"resource"`
	// Acción permitida
	Action Action `json:"action" example:"get"`
	// Fecha de creación
	CreatedAt time.Time `json:"created_at"`
	// Fecha de actualización
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginatedPermissionsResponse respuesta paginada de permisos
// @name PaginatedPermissionsResponse
type PaginatedPermissionsResponse struct {
	// Lista de permisos
	Data []*PermissionResponse `json:"data"`
	// Información de paginación
	Pagination *pagination.PaginationInfo `json:"pagination"`
}

func ToResponse(p *Permission, resource ResourceRef) *PermissionResponse {
	return &PermissionResponse{
		ID:        p.ID.Hex(),
		TenantId:  p.TenantId.Hex(),
		Resource:  resource,
		Action:    p.Action,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
