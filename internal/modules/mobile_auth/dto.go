package mobile_auth

// --- Input DTOs ---

type RegisterDTO struct {
	Name     string `json:"name"     binding:"required,min=2"  example:"Juan Pérez"`
	Email    string `json:"email"    binding:"required,email"  example:"juan@example.com"`
	Phone    string `json:"phone"                              example:"+57 300 123 4567"`
	Password string `json:"password" binding:"required,min=6"  example:"secret123"`
}

type LoginDTO struct {
	Email    string `json:"email"    binding:"required,email" example:"juan@example.com"`
	Password string `json:"password" binding:"required"        example:"secret123"`
}

type RefreshDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGci..."`
}

// --- Response DTOs ---

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type OwnerInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
