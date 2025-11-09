package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var (
	profileFlag string
	branchFlag  string
	pluginFlags []string
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new Discourse environment",
	Long: `Create a new isolated Discourse development environment from a profile.

Examples:
  denver create yaks --profile base --plugin discourse-yaks:feature/new-stuff
  denver create test-epic --profile epic-games --branch my-pr
  denver create minimal --profile base`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return createEnvironment(name, profileFlag, branchFlag, pluginFlags)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to use for environment setup (required)")
	createCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "Discourse core branch to checkout (default: main)")
	createCmd.Flags().StringArrayVar(&pluginFlags, "plugin", []string{}, "Add or override plugin (format: name:branch, repeatable)")
	createCmd.MarkFlagRequired("profile")
}

func createEnvironment(name string, profileName string, branch string, pluginOverrides []string) error {
	fmt.Printf("Creating environment '%s' with profile '%s'\n", name, profileName)

	// Load profile
	profile, err := config.LoadProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}
	_ = profile // TODO: Use profile for plugin cloning

	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, name)
	discourseDir := filepath.Join(envDir, "discourse")

	// Check if environment already exists
	if _, err := os.Stat(envDir); err == nil {
		return fmt.Errorf("environment '%s' already exists at %s", name, envDir)
	}

	// Create environment directory
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	fmt.Printf("Cloning discourse to %s...\n", discourseDir)

	// Clone discourse
	if err := cloneDiscourse(discourseDir, branch); err != nil {
		return fmt.Errorf("failed to clone discourse: %w", err)
	}

	fmt.Println("✓ Discourse cloned successfully")

	// TODO: Clone plugins based on profile + overrides
	// TODO: Generate devcontainer.json
	// TODO: Create .denver.yml

	fmt.Printf("\nEnvironment '%s' created successfully!\n", name)
	fmt.Printf("Location: %s\n", envDir)

	return nil
}

func cloneDiscourse(destDir string, branch string) error {
	cloneArgs := []string{"clone", "https://github.com/discourse/discourse.git", destDir}

	if branch != "" {
		cloneArgs = append(cloneArgs, "--branch", branch)
	}

	cmd := exec.Command("git", cloneArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
