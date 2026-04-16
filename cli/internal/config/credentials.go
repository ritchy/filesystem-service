// Package config manages persistent CLI configuration and credentials.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials holds the Cognito JWT tokens persisted between CLI invocations.
type Credentials struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

// configDir returns the path to ~/.filesystem, creating it if necessary.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".filesystem"), nil
}

// SaveCredentials writes credentials to ~/.filesystem/credentials.json
// with permissions restricted to the current user (0600).
func SaveCredentials(creds *Credentials) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode credentials: %w", err)
	}

	credPath := filepath.Join(dir, "credentials.json")
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", credPath, err)
	}

	return nil
}

// LoadCredentials reads credentials from ~/.filesystem/credentials.json.
// Returns an error (suitable for display) when the file does not exist.
func LoadCredentials() (*Credentials, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	credPath := filepath.Join(dir, "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials not found – run 'fs login'")
		}
		return nil, fmt.Errorf("failed to read %s: %w", credPath, err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	if creds.IDToken == "" {
		return nil, fmt.Errorf("incomplete credentials – run 'fs login' again")
	}

	return &creds, nil
}
