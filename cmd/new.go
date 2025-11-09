package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var newCmd = &cobra.Command{
	Use:   "new profile <name>",
	Short: "Create a new profile interactively",
	Long: `Create a new profile by answering prompts for name, description, plugins, and themes.

The profile will be saved to ~/.denver/profiles/<name>.yml and can be used
with 'denver create' to set up new environments.

Examples:
  denver new profile epic
  denver new profile my-plugins`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "profile" {
			return fmt.Errorf("only 'profile' is supported (got: %s)", args[0])
		}
		profileName := args[1]
		return createProfile(profileName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func createProfile(filename string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Creating new profile: %s\n\n", filename)

	// Get profile name
	fmt.Print("Profile name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		name = filename
	}

	// Get description
	fmt.Print("Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	// Build profile
	profile := config.Profile{
		Name:        name,
		Description: description,
		Plugins:     []config.PluginConfig{},
		Themes:      []config.ThemeConfig{},
		Settings:    make(map[string]any),
	}

	// Add plugins
	fmt.Print("\nAdd plugins? (y/n): ")
	addPlugins, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(addPlugins)) == "y" {
		fmt.Println("\nEnter plugin names (format: discourse/plugin-name or just plugin-name)")
		fmt.Println("Leave blank to finish")
		for {
			fmt.Print("Plugin: ")
			pluginInput, _ := reader.ReadString('\n')
			pluginInput = strings.TrimSpace(pluginInput)

			if pluginInput == "" {
				break
			}

			// Parse plugin input
			var repo string
			if strings.Contains(pluginInput, "/") {
				repo = pluginInput
			} else {
				repo = "discourse/" + pluginInput
			}

			// Extract name from repo
			parts := strings.Split(repo, "/")
			pluginName := parts[len(parts)-1]

			profile.Plugins = append(profile.Plugins, config.PluginConfig{
				Name: pluginName,
				Repo: repo,
			})
		}
	}

	// Add themes
	fmt.Print("\nAdd themes? (y/n): ")
	addThemes, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(addThemes)) == "y" {
		fmt.Println("\nEnter theme repos (format: discourse/theme-name)")
		fmt.Println("Leave blank to finish")
		for {
			fmt.Print("Theme repo: ")
			themeInput, _ := reader.ReadString('\n')
			themeInput = strings.TrimSpace(themeInput)

			if themeInput == "" {
				break
			}

			profile.Themes = append(profile.Themes, config.ThemeConfig{
				Repo: themeInput,
			})
		}
	}

	// Save profile
	denverDir, err := config.GetDenverDir()
	if err != nil {
		return err
	}

	profilesDir := filepath.Join(denverDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	profilePath := filepath.Join(profilesDir, filename+".yml")

	// Check if profile already exists
	if _, err := os.Stat(profilePath); err == nil {
		fmt.Printf("\nProfile '%s' already exists. Overwrite? (y/n): ", filename)
		overwrite, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(overwrite)) != "y" {
			return fmt.Errorf("cancelled")
		}
	}

	// Write YAML
	data, err := yaml.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	fmt.Printf("\n✓ Profile created: %s\n", profilePath)
	fmt.Printf("\nUse it with: denver create <name> --profile %s\n", filename)

	return nil
}
