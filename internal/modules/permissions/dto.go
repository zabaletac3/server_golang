package permissions

import "time"

// CreatePermissionDTO request para crear permiso
// @name CreatePermissionDTO
type CreatePermissionDTO struct {
	Resource Resource `json:"resource" binding:"required" example:"users"`
	Action   Action   `json:"action" binding:"required" example:"read"`
}

// PermissionResponse respuesta de permiso
// @name PermissionResponse
type PermissionResponse struct {
	ID        string   `json:"id" example:"507f1f77bcf86cd799439011"`
	Resource  Resource `json:"resource" example:"users"`
	Action    Action   `json:"action" example:"read"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ToResponse convierte Permission a PermissionResponse
func ToResponse(p *Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:        p.ID.Hex(),
		Resource:  p.Resource,
		Action:    p.Action,
		CreatedAt: p.CreatedAt,
	}
}

// ToResponseList convierte lista de Permission a lista de PermissionResponse
func ToResponseList(permissions []Permission) []*PermissionResponse {
	responses := make([]*PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = ToResponse(&p)
	}
	return responses
}

// AvailableOptionsResponse respuesta con recursos y acciones disponibles
// @name AvailableOptionsResponse
type AvailableOptionsResponse struct {
	Resources []Resource `json:"resources"`
	Actions   []Action   `json:"actions"`
}
