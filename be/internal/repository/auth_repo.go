package repository

import (
	"context"

	"github.com/reshap/trading-bot/internal/helpers"
	"github.com/reshap/trading-bot/internal/models"
)

const configPath = "internal/config/auth.yml"

// AuthRepository handles auth data operations
type AuthRepository struct {}

// NewAuthRepository creates a new AuthRepository
func NewAuthRepository() *AuthRepository {
	return &AuthRepository{}
}


// FindUserByUsername finds a user by username
func (r *AuthRepository) FindUserByUsername(ctx context.Context, username string) (*models.User, error) {
	config, err := r.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	for i := range config.Users {
		if config.Users[i].Username == username {
			return &config.Users[i], nil
		}
	}

	return nil, nil // User not found, return nil without error
}

// SaveUser saves user data (updates session)
func (r *AuthRepository) SaveUser(ctx context.Context, user *models.User) error {
	config, err := r.loadConfig(ctx)
	if err != nil {
		return err
	}

	for i := range config.Users {
		if config.Users[i].Username == user.Username {
			config.Users[i] = *user
			break
		}
	}

	return helpers.SaveYAML(configPath, config)
}

// ClearUserSession clears user session
func (r *AuthRepository) ClearUserSession(ctx context.Context, username string) error {
	config, err := r.loadConfig(ctx)
	if err != nil {
		return err
	}

	for i := range config.Users {
		if config.Users[i].Username == username {
			config.Users[i].Session = ""
			break
		}
	}

	return helpers.SaveYAML(configPath, config)
}

// loadConfig loads the auth configuration from YAML file
func (r *AuthRepository) loadConfig(ctx context.Context) (*models.AuthConfig, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return helpers.LoadYAML[models.AuthConfig](configPath)
	}
}
