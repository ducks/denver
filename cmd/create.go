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
	baseFlag    string
	pluginFlags []string
	localFlags  []string
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new Discourse environment",
	Long: `Create a new isolated Discourse development environment from a profile.

The environment name will be used as the branch name in the discourse worktree.

Examples:
  denver create yaks --profile base --plugin discourse-yaks:feature/new-stuff
  denver create test-epic --profile epic-games --base fix/button-refactor
  denver create minimal --profile base
  denver create frndr --profile base --local discourse-frndr:~/dev/discourse-frndr`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return createEnvironment(name, profileFlag, baseFlag, pluginFlags, localFlags)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to use for environment setup (required)")
	createCmd.Flags().StringVar(&baseFlag, "base", "", "Base branch to branch from (default: main)")
	createCmd.Flags().StringArrayVar(&pluginFlags, "plugin", []string{}, "Add or override plugin (format: name:branch, repeatable)")
	createCmd.Flags().StringArrayVar(&localFlags, "local", []string{}, "Use local plugin path (format: name:path, repeatable)")
	_ = createCmd.MarkFlagRequired("profile")
}

func createEnvironment(name string, profileName string, baseBranch string, pluginOverrides []string, localOverrides []string) error {
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
	// Environment name becomes the branch name, branching from base (or main)
	fmt.Printf("Creating discourse worktree at %s...\n", discourseDir)
	if err := createWorktree(bareRepo, discourseDir, name, baseBranch); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	fmt.Println("✓ Discourse worktree created successfully")

	// Build plugin list (profile + command-line additions + local overrides)
	plugins := buildPluginList(profile.Plugins, pluginOverrides, localOverrides)

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

	// Save environment configuration
	envConfig := &config.EnvironmentConfig{
		Name:    name,
		Profile: profileName,
		Base:    baseBranch,
		Plugins: plugins,
	}

	if err := config.SaveEnvironmentConfig(envDir, envConfig); err != nil {
		return fmt.Errorf("failed to save environment config: %w", err)
	}

	// TODO: Generate devcontainer.json

	fmt.Printf("\nEnvironment '%s' created successfully!\n", name)
	fmt.Printf("Location: %s\n", envDir)

	return nil
}

func buildPluginList(profilePlugins []config.PluginConfig, pluginFlags []string, localFlags []string) []config.PluginConfig {
	// Start with profile plugins
	plugins := make([]config.PluginConfig, len(profilePlugins))
	copy(plugins, profilePlugins)

	// Add plugins from command-line flags
	for _, flag := range pluginFlags {
		// Parse plugin (format: "name", "owner/name", "name:branch", or "owner/name:branch")
		parts := strings.SplitN(flag, ":", 2)
		pluginSpec := parts[0]
		branch := ""
		if len(parts) > 1 {
			branch = parts[1]
		}

		// Determine plugin name and repo
		var pluginName, repo string
		if strings.Contains(pluginSpec, "/") {
			// Owner provided: "ducks/discourse-frndr"
			repo = pluginSpec
			repoParts := strings.Split(pluginSpec, "/")
			pluginName = repoParts[len(repoParts)-1]
		} else {
			// No owner: "discourse-yaks" - default to discourse org
			pluginName = pluginSpec
			repo = "discourse/" + pluginName
		}

		// Check if plugin already exists in profile
		exists := false
		for _, p := range plugins {
			if p.Name == pluginName {
				exists = true
				break
			}
		}

		if !exists {
			plugins = append(plugins, config.PluginConfig{
				Name:   pluginName,
				Repo:   repo,
				Branch: branch,
			})
		}
	}

	// Apply local path overrides
	for _, flag := range localFlags {
		// Parse local flag (format: "name:path")
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("Warning: invalid --local flag format '%s', expected 'name:path'\n", flag)
			continue
		}

		pluginName := parts[0]
		localPath := parts[1]

		// Find and update existing plugin or add new one
		found := false
		for i := range plugins {
			if plugins[i].Name == pluginName {
				plugins[i].Local = localPath
				found = true
				break
			}
		}

		if !found {
			// Add new plugin with only local path (no repo/branch needed)
			plugins = append(plugins, config.PluginConfig{
				Name:  pluginName,
				Local: localPath,
			})
		}
	}

	return plugins
}

func clonePlugin(pluginsDir string, discourseDir string, plugin config.PluginConfig) error {
	symlinkPath := filepath.Join(discourseDir, "plugins", plugin.Name)

	// If local path is specified, symlink directly to it
	if plugin.Local != "" {
		fmt.Printf("  Linking %s (local)...\n", plugin.Name)

		// Expand ~ to home directory
		localPath := plugin.Local
		if strings.HasPrefix(localPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			localPath = filepath.Join(homeDir, localPath[2:])
		}

		// Check that local path exists
		if _, err := os.Stat(localPath); err != nil {
			return fmt.Errorf("local plugin path does not exist: %s", localPath)
		}

		// Create symlink
		if err := os.Symlink(localPath, symlinkPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		fmt.Printf("  ✓ %s linked to %s\n", plugin.Name, localPath)
		return nil
	}

	// Otherwise, clone from GitHub
	pluginPath := filepath.Join(pluginsDir, plugin.Name)
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
