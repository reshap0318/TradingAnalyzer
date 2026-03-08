package models

// User represents a user in the auth system
type User struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"-"`
	Name     string `yaml:"name" json:"name"`
	Session  string `yaml:"session" json:"session,omitempty"`
}

// AuthConfig represents the auth configuration
type AuthConfig struct {
	Users []User `yaml:"users" json:"users"`
}
