package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ActiveEnvironment represents the currently running environment
type ActiveEnvironment struct {
	Name      string `json:"name"`
	RailsPID  int    `json:"rails_pid"`
	EmberPID  int    `json:"ember_pid"`
	RailsPort int    `json:"rails_port"`
	EmberPort int    `json:"ember_port"`
}

// GetActivePath returns the path to the active environment file
func GetActivePath() (string, error) {
	denverDir, err := GetDenverDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(denverDir, "active"), nil
}

// LoadActive loads the active environment info
func LoadActive() (*ActiveEnvironment, error) {
	activePath, err := GetActivePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No active environment
		}
		return nil, fmt.Errorf("failed to read active file: %w", err)
	}

	var active ActiveEnvironment
	if err := json.Unmarshal(data, &active); err != nil {
		return nil, fmt.Errorf("failed to parse active file: %w", err)
	}

	return &active, nil
}

// SaveActive saves the active environment info
func SaveActive(active *ActiveEnvironment) error {
	activePath, err := GetActivePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(active)
	if err != nil {
		return fmt.Errorf("failed to marshal active environment: %w", err)
	}

	if err := os.WriteFile(activePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write active file: %w", err)
	}

	return nil
}

// ClearActive removes the active environment file
func ClearActive() error {
	activePath, err := GetActivePath()
	if err != nil {
		return err
	}

	if err := os.Remove(activePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove active file: %w", err)
	}

	return nil
}
