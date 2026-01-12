package users

import (
	"time"

	"github.com/eren_dev/go_server/internal/shared/pagination"
)

// CreateUserDTO request para crear usuario
// @name CreateUserDTO
type CreateUserDTO struct {
	// Nombre del usuario
	Name string `json:"name" binding:"required,min=2" example:"John Doe"`
	// Email del usuario
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	// Contraseña (mínimo 6 caracteres)
	Password string `json:"password" binding:"required,min=6" example:"secret123"`
}

// UpdateUserDTO request para actualizar usuario
// @name UpdateUserDTO
type UpdateUserDTO struct {
	// Nombre del usuario
	Name string `json:"name" example:"Jane Doe"`
	// Email del usuario
	Email string `json:"email" binding:"omitempty,email" example:"jane@example.com"`
}

// UserResponse respuesta de usuario
// @name UserResponse
type UserResponse struct {
	// ID único del usuario
	ID string `json:"id" example:"507f1f77bcf86cd799439011"`
	// Nombre del usuario
	Name string `json:"name" example:"John Doe"`
	// Email del usuario
	Email string `json:"email" example:"john@example.com"`
	// Fecha de creación
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	// Fecha de actualización
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// PaginatedUsersResponse respuesta paginada de usuarios
// @name PaginatedUsersResponse
type PaginatedUsersResponse struct {
	// Lista de usuarios
	Data []*UserResponse `json:"data"`
	// Información de paginación
	Pagination *pagination.PaginationInfo `json:"pagination"`
}

func ToResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:        u.ID.Hex(),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ToResponseList(users []*User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, u := range users {
		result[i] = ToResponse(u)
	}
	return result
}
