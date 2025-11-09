package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Plugins     []PluginConfig  `yaml:"plugins"`
	Themes      []ThemeConfig   `yaml:"themes"`
	Settings    map[string]any  `yaml:"site_settings"`
	Seed        *SeedConfig     `yaml:"seed"`
}

type PluginConfig struct {
	Name    string `yaml:"name"`
	Repo    string `yaml:"repo"`
	Branch  string `yaml:"branch"`
	Private bool   `yaml:"private"`
}

type ThemeConfig struct {
	Repo   string `yaml:"repo"`
	Branch string `yaml:"branch"`
}

type SeedConfig struct {
	Admin        bool `yaml:"admin"`
	SampleUsers  int  `yaml:"sample_users"`
	SampleTopics int  `yaml:"sample_topics"`
}

// LoadProfile loads a profile from ~/.denver/profiles/<name>.yml
func LoadProfile(name string) (*Profile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	profilePath := filepath.Join(homeDir, ".denver", "profiles", name+".yml")

	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile %s: %w", name, err)
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile %s: %w", name, err)
	}

	return &profile, nil
}

// GetDenverDir returns the denver config directory (~/.denver)
func GetDenverDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".denver"), nil
}

// GetEnvironmentsDir returns the environments directory (~/.denver/environments)
func GetEnvironmentsDir() (string, error) {
	denverDir, err := GetDenverDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(denverDir, "environments"), nil
}
