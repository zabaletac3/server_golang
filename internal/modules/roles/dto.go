package roles

import (
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// CreateRoleDTO request para crear un rol
// @name CreateRoleDTO
type CreateRoleDTO struct {
	// ID del tenant al que pertenece
	TenantId string `json:"tenant_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	// Nombre del rol
	Name string `json:"name" binding:"required,min=2" example:"veterinary"`
	// Descripción del rol
	Description string `json:"description" example:"Rol para veterinarios"`
	// IDs de permisos asignados (pares recurso+acción)
	PermissionsIds []string `json:"permissions_ids" example:"507f1f77bcf86cd799439011"`
	// IDs de recursos accesibles por este rol
	ResourcesIds []string `json:"resources_ids" example:"507f1f77bcf86cd799439011"`
}

// UpdateRoleDTO request para actualizar un rol
// @name UpdateRoleDTO
type UpdateRoleDTO struct {
	// Nombre del rol
	Name string `json:"name" example:"veterinary"`
	// Descripción del rol
	Description string `json:"description" example:"Rol para veterinarios"`
	// IDs de permisos asignados
	PermissionsIds []string `json:"permissions_ids"`
	// IDs de recursos accesibles
	ResourcesIds []string `json:"resources_ids"`
}

// ResourceRef referencia simplificada de recurso en la respuesta de rol
type ResourceRef struct {
	// ID del recurso
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// Nombre del recurso
	Name string `json:"name" example:"appointments"`
	// Descripción
	Description string `json:"description" example:"Módulo de citas"`
}

// PermissionRef referencia poblada de permiso en la respuesta de rol
type PermissionRef struct {
	// ID del permiso
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// Recurso asociado
	Resource ResourceRef `json:"resource"`
	// Acción permitida
	Action string `json:"action" example:"get"`
}

// RoleResponse respuesta de rol con permisos y recursos poblados
// @name RoleResponse
type RoleResponse struct {
	// ID único del rol
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// ID del tenant
	TenantId string `json:"tenant_id" example:"507f1f77bcf86cd799439011"`
	// Nombre del rol
	Name string `json:"name" example:"veterinary"`
	// Descripción del rol
	Description string `json:"description" example:"Rol para veterinarios"`
	// Permisos poblados (recurso + acción)
	Permissions []PermissionRef `json:"permissions"`
	// Recursos accesibles poblados
	Resources []ResourceRef `json:"resources"`
	// Fecha de creación
	CreatedAt time.Time `json:"created_at"`
	// Fecha de actualización
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginatedRolesResponse respuesta paginada de roles
// @name PaginatedRolesResponse
type PaginatedRolesResponse struct {
	// Lista de roles
	Data []*RoleResponse `json:"data"`
	// Información de paginación
	Pagination *pagination.PaginationInfo `json:"pagination"`
}
