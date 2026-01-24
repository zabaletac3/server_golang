package roles

import (
	"time"

	"github.com/eren_dev/go_server/internal/modules/permissions"
)

// CreateRoleDTO request para crear rol
// @name CreateRoleDTO
type CreateRoleDTO struct {
	TenantID      string   `json:"tenant_id" binding:"required" example:"507f1f77bcf86cd799439011"`
	Name          string   `json:"name" binding:"required,min=2" example:"Administrador"`
	Description   string   `json:"description" example:"Rol con acceso total"`
	PermissionIDs []string `json:"permission_ids" example:"507f1f77bcf86cd799439011,507f1f77bcf86cd799439012"`
}

// UpdateRoleDTO request para actualizar rol
// @name UpdateRoleDTO
type UpdateRoleDTO struct {
	Name          string   `json:"name" example:"Administrador"`
	Description   string   `json:"description" example:"Rol con acceso total"`
	PermissionIDs []string `json:"permission_ids"`
}

// RoleResponse respuesta de rol
// @name RoleResponse
type RoleResponse struct {
	ID          string                          `json:"id" example:"507f1f77bcf86cd799439011"`
	TenantID    string                          `json:"tenant_id" example:"507f1f77bcf86cd799439011"`
	Name        string                          `json:"name" example:"Administrador"`
	Description string                          `json:"description" example:"Rol con acceso total"`
	Permissions []*permissions.PermissionResponse `json:"permissions"`
	CreatedAt   time.Time                       `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time                       `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// RoleSimpleResponse respuesta simple de rol (sin permisos expandidos)
// @name RoleSimpleResponse
type RoleSimpleResponse struct {
	ID            string    `json:"id" example:"507f1f77bcf86cd799439011"`
	TenantID      string    `json:"tenant_id" example:"507f1f77bcf86cd799439011"`
	Name          string    `json:"name" example:"Administrador"`
	Description   string    `json:"description" example:"Rol con acceso total"`
	PermissionIDs []string  `json:"permission_ids"`
	CreatedAt     time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ToSimpleResponse convierte Role a RoleSimpleResponse
func ToSimpleResponse(r *Role) *RoleSimpleResponse {
	permIDs := make([]string, len(r.PermissionIDs))
	for i, id := range r.PermissionIDs {
		permIDs[i] = id.Hex()
	}

	return &RoleSimpleResponse{
		ID:            r.ID.Hex(),
		TenantID:      r.TenantID.Hex(),
		Name:          r.Name,
		Description:   r.Description,
		PermissionIDs: permIDs,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

// ToSimpleResponseList convierte lista de Role a lista de RoleSimpleResponse
func ToSimpleResponseList(roles []Role) []*RoleSimpleResponse {
	responses := make([]*RoleSimpleResponse, len(roles))
	for i, r := range roles {
		responses[i] = ToSimpleResponse(&r)
	}
	return responses
}
