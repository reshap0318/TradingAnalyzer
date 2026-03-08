package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/reshap/trading-bot/internal/helpers"
)

const authConfigPath = "internal/config/auth.yml"

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Session  string `yaml:"session"`
}

type AuthConfig struct {
	Users []User `yaml:"users"`
}

func updatePasswords(config *AuthConfig, targetUser string, length int) error {
	for i := range config.Users {
		// If targetUser is specified, only update that user
		if targetUser != "" && config.Users[i].Username != targetUser {
			continue
		}

		password, err := helpers.GenerateRandomString(length, "")
		if err != nil {
			return fmt.Errorf("failed to generate password for %s: %w", config.Users[i].Username, err)
		}

		oldPassword := config.Users[i].Password
		config.Users[i].Password = password
		config.Users[i].Session = "" // Clear session when password changes

		fmt.Printf("✓ %s: %s -> %s\n", config.Users[i].Username, oldPassword, password)
	}

	return nil
}

func main() {
	defaultLength := 16
	userFlag := flag.String("user", "", "Specific user to update (default: all users)")
	lengthFlag := flag.Int("length", defaultLength, "Password length")
	configFlag := flag.String("config", authConfigPath, "Path to auth.yml")

	flag.Parse()

	config, err := helpers.LoadYAML[AuthConfig](*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check if target user exists
	if *userFlag != "" {
		found := false
		for _, user := range config.Users {
			if user.Username == *userFlag {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Error: user '%s' not found\n", *userFlag)
			os.Exit(1)
		}
	}

	fmt.Printf("Updating passwords (length=%d)...\n", *lengthFlag)
	if *userFlag != "" {
		fmt.Printf("Target user: %s\n\n", *userFlag)
	} else {
		fmt.Println("Target: all users")
	}

	if err := updatePasswords(config, *userFlag, *lengthFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating passwords: %v\n", err)
		os.Exit(1)
	}

	// Get absolute path for display
	absPath, _ := filepath.Abs(*configFlag)
	if err := helpers.SaveYAML(*configFlag, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Saved to %s\n", absPath)
}
