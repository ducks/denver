package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy <name>",
	Short: "Destroy a Discourse environment",
	Long: `Destroy an existing Discourse development environment.

This will delete the entire environment directory including the cloned
discourse repository and all plugins. This action cannot be undone.

Examples:
  denver destroy test-env
  denver destroy yaks`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return destroyEnvironment(name)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}

func destroyEnvironment(name string) error {
	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, name)

	// Check if environment exists
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", name)
	}

	// Check if environment is currently running
	active, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("failed to check active environment: %w", err)
	}

	if active != nil && active.Name == name {
		return fmt.Errorf("cannot destroy running environment '%s'. Stop it first with: denver stop", name)
	}

	fmt.Printf("Destroying environment '%s'...\n", name)

	// Remove the worktree
	discourseDir := filepath.Join(envDir, "discourse")
	fmt.Println("Removing worktree...")
	removeWorktreeCmd := exec.Command("git", "worktree", "remove", discourseDir, "--force")
	bareRepoPath, err := config.GetBareRepoPath()
	if err != nil {
		return err
	}
	removeWorktreeCmd.Dir = bareRepoPath
	if err := removeWorktreeCmd.Run(); err != nil {
		fmt.Printf("Warning: failed to remove worktree: %v\n", err)
	}

	// Check if the branch has uncommitted changes or unique commits
	shouldDeleteBranch := true
	branchHasChanges, err := checkBranchHasChanges(name)
	if err != nil {
		fmt.Printf("Warning: couldn't check branch status: %v\n", err)
	} else if branchHasChanges {
		// Ask user if they want to delete the branch
		fmt.Printf("\nBranch '%s' has uncommitted changes or unique commits.\n", name)
		fmt.Print("Delete the branch anyway? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		shouldDeleteBranch = response == "y" || response == "yes"
	}

	// Delete the git branch
	if shouldDeleteBranch {
		fmt.Println("Deleting git branch...")
		deleteBranchCmd := exec.Command("git", "branch", "-D", name)
		deleteBranchCmd.Dir = bareRepoPath
		if err := deleteBranchCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to delete branch: %v\n", err)
		}
	} else {
		fmt.Printf("Keeping branch '%s'\n", name)
	}

	// Remove the entire environment directory
	fmt.Printf("Removing: %s\n", envDir)
	if err := os.RemoveAll(envDir); err != nil {
		return fmt.Errorf("failed to destroy environment: %w", err)
	}

	fmt.Printf("\n✓ Environment '%s' destroyed successfully\n", name)

	return nil
}

func checkBranchHasChanges(branchName string) (bool, error) {
	bareRepoPath, err := config.GetBareRepoPath()
	if err != nil {
		return false, err
	}

	// Check if branch has commits not in main
	cmd := exec.Command("git", "log", "main.."+branchName, "--oneline")
	cmd.Dir = bareRepoPath
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// If there's any output, the branch has unique commits
	return len(strings.TrimSpace(string(output))) > 0, nil
}
