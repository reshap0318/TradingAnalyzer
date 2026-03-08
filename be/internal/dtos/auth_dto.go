package dtos

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Name    string `json:"name,omitempty"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	Username string `json:"username" binding:"required"`
}

// LogoutResponse represents a logout response
type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}