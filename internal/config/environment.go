package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EnvironmentConfig represents the configuration for a specific environment
type EnvironmentConfig struct {
	Name    string         `yaml:"name"`
	Profile string         `yaml:"profile"`
	Base    string         `yaml:"base,omitempty"`
	Plugins []PluginConfig `yaml:"plugins,omitempty"`
}

// SaveEnvironmentConfig saves the environment configuration to .denver.yml
func SaveEnvironmentConfig(envDir string, config *EnvironmentConfig) error {
	configPath := filepath.Join(envDir, ".denver.yml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal environment config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment config: %w", err)
	}

	return nil
}

// LoadEnvironmentConfig loads the environment configuration from .denver.yml
func LoadEnvironmentConfig(envDir string) (*EnvironmentConfig, error) {
	configPath := filepath.Join(envDir, ".denver.yml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment config: %w", err)
	}

	var config EnvironmentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse environment config: %w", err)
	}

	return &config, nil
}
