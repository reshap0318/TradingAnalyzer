package service

import (
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/reshap/trading-bot/internal/dtos"
)

var mu sync.RWMutex

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}


// Login authenticates a user and generates a session token
func (s *Services) Login(ctx *gin.Context, username, password string) (*dtos.LoginResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	// Find user
	user, err := s.repo.Auth.FindUserByUsername(ctx, username)
	if err != nil {
		return &dtos.LoginResponse{
			Success: false,
			Message: "Failed to load user data",
		}, nil
	}

	if user == nil {
		return &dtos.LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// Verify password (plain text comparison)
	if user.Password != password {
		return &dtos.LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// Generate JWT token
	token, err := s.generateToken(user.Username, user.Name)
	if err != nil {
		log.Println(err)
		return &dtos.LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		}, nil
	}

	// Update session
	user.Session = token
	if err := s.repo.Auth.SaveUser(ctx, user); err != nil {
		return &dtos.LoginResponse{
			Success: false,
			Message: "Failed to save session",
		}, nil
	}

	return &dtos.LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		Name:    user.Name,
	}, nil
}

// Logout invalidates a user's session
func (s *Services) Logout(ctx *gin.Context, username string) (*dtos.LogoutResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	// Check if user exists
	user, err := s.repo.Auth.FindUserByUsername(ctx, username)
	if err != nil {
		return &dtos.LogoutResponse{
			Success: false,
			Message: "Failed to load user data",
		}, nil
	}

	if user == nil {
		return &dtos.LogoutResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	// Clear session
	if err := s.repo.Auth.ClearUserSession(ctx, username); err != nil {
		return &dtos.LogoutResponse{
			Success: false,
			Message: "Failed to save session",
		}, nil
	}

	return &dtos.LogoutResponse{
		Success: true,
		Message: "Logout successful",
	}, nil
}

// generateToken generates a JWT token for a user
func (s *Services) generateToken(username, name string) (string, error) {
	claims := Claims{
		Username: username,
		Name:     name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.cfg.APP.JWTSecret)
}

// ValidateToken validates a JWT token
func (s *Services) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.cfg.APP.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
