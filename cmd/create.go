package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	_ = createCmd.MarkFlagRequired("profile")
}

func createEnvironment(name string, profileName string, branch string, pluginOverrides []string) error {
	fmt.Printf("Creating environment '%s' with profile '%s'\n", name, profileName)

	// Load profile
	profile, err := config.LoadProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, name)
	discourseDir := filepath.Join(envDir, "discourse")
	pluginsDir := filepath.Join(envDir, "plugins")

	// Check if environment already exists
	if _, err := os.Stat(envDir); err == nil {
		return fmt.Errorf("environment '%s' already exists at %s", name, envDir)
	}

	// Create environment directory
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	// Ensure bare repo exists
	bareRepo, err := config.GetBareRepoPath()
	if err != nil {
		return err
	}

	if err := ensureBareRepo(bareRepo); err != nil {
		return fmt.Errorf("failed to setup bare repository: %w", err)
	}

	// Create worktree for this environment
	// Use environment name as branch name, base it on specified branch (or main)
	fmt.Printf("Creating discourse worktree at %s...\n", discourseDir)
	if err := createWorktree(bareRepo, discourseDir, name, branch); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	fmt.Println("✓ Discourse worktree created successfully")

	// Build plugin list (profile + command-line additions)
	plugins := buildPluginList(profile.Plugins, pluginOverrides)

	if len(plugins) > 0 {
		fmt.Printf("\nCloning %d plugin(s)...\n", len(plugins))

		// Create plugins directory
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugins directory: %w", err)
		}

		// Clone each plugin
		for _, plugin := range plugins {
			if err := clonePlugin(pluginsDir, discourseDir, plugin); err != nil {
				return fmt.Errorf("failed to clone plugin %s: %w", plugin.Name, err)
			}
		}
	}

	// TODO: Generate devcontainer.json
	// TODO: Create .denver.yml

	fmt.Printf("\nEnvironment '%s' created successfully!\n", name)
	fmt.Printf("Location: %s\n", envDir)

	return nil
}

func buildPluginList(profilePlugins []config.PluginConfig, pluginFlags []string) []config.PluginConfig {
	// Start with profile plugins
	plugins := make([]config.PluginConfig, len(profilePlugins))
	copy(plugins, profilePlugins)

	// Add plugins from command-line flags
	for _, flag := range pluginFlags {
		// For now, just parse plugin name (format: "discourse-yaks" or "discourse-yaks:branch")
		// Branch support will come with `denver plugin` command
		parts := strings.SplitN(flag, ":", 2)
		pluginName := parts[0]

		// Check if plugin already exists in profile
		exists := false
		for _, p := range plugins {
			if p.Name == pluginName {
				exists = true
				break
			}
		}

		if !exists {
			// Add new plugin (assume discourse org for now)
			plugins = append(plugins, config.PluginConfig{
				Name:   pluginName,
				Repo:   "discourse/" + pluginName,
				Branch: "", // Default branch
			})
		}
	}

	return plugins
}

func clonePlugin(pluginsDir string, discourseDir string, plugin config.PluginConfig) error {
	pluginPath := filepath.Join(pluginsDir, plugin.Name)
	symlinkPath := filepath.Join(discourseDir, "plugins", plugin.Name)

	fmt.Printf("  Cloning %s...\n", plugin.Name)

	// Build clone args
	repoURL := fmt.Sprintf("https://github.com/%s.git", plugin.Repo)
	cloneArgs := []string{"clone", repoURL, pluginPath}

	if plugin.Branch != "" {
		cloneArgs = append(cloneArgs, "--branch", plugin.Branch)
	}

	// Clone the plugin
	cmd := exec.Command("git", cloneArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// Create symlink
	if err := os.Symlink(pluginPath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("  ✓ %s cloned and linked\n", plugin.Name)

	return nil
}

func ensureBareRepo(bareRepoPath string) error {
	// Check if bare repo already exists
	if _, err := os.Stat(bareRepoPath); err == nil {
		fmt.Println("✓ Using existing bare repository")
		return nil
	}

	fmt.Printf("Cloning discourse bare repository (one-time setup)...\n")
	fmt.Println("This may take a few minutes...")

	// Clone as bare repo
	cmd := exec.Command("git", "clone", "--bare", "https://github.com/discourse/discourse.git", bareRepoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone bare repository: %w", err)
	}

	fmt.Println("✓ Bare repository cloned successfully")

	return nil
}

func createWorktree(bareRepoPath string, worktreePath string, envName string, baseBranch string) error {
	// Default base branch to main if not specified
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Create new branch with environment name, based on specified branch
	// git worktree add <path> -b <new-branch> <start-point>
	args := []string{"-C", bareRepoPath, "worktree", "add", worktreePath, "-b", envName, baseBranch}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
