package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check Denver prerequisites",
	Long: `Check that all required tools and services are available.

Denver requires:
- git (for repository management)
- Ruby and bundler (for Rails)
- Node.js and pnpm (for Ember)
- PostgreSQL (for database)
- Redis (for caching)

Examples:
  denver doctor`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkPrerequisites()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name    string
	ok      bool
	version string
	message string
}

func checkPrerequisites() error {
	fmt.Println("Checking Denver prerequisites...")

	checks := []checkResult{
		checkGit(),
		checkRuby(),
		checkBundler(),
		checkNode(),
		checkPnpm(),
		checkPostgres(),
		checkRedis(),
	}

	allOk := true
	for _, check := range checks {
		status := "✓"
		color := ""
		if !check.ok {
			status = "✗"
			allOk = false
		}

		versionInfo := ""
		if check.version != "" {
			versionInfo = fmt.Sprintf(" (%s)", check.version)
		}

		fmt.Printf("%s %s%s%s\n", status, check.name, versionInfo, color)

		if check.message != "" {
			fmt.Printf("  %s\n", check.message)
		}
	}

	fmt.Println()

	if allOk {
		fmt.Println("✓ All prerequisites met!")
		return nil
	}

	fmt.Println("✗ Some prerequisites are missing or not running")
	fmt.Println("\nFor Discourse setup instructions, see:")
	fmt.Println("https://meta.discourse.org/t/beginners-guide-to-install-discourse-for-development-using-docker/102009")

	return nil
}

func checkGit() checkResult {
	version, err := exec.Command("git", "--version").Output()
	if err != nil {
		return checkResult{
			name:    "git",
			ok:      false,
			message: "Install git: https://git-scm.com/downloads",
		}
	}

	versionStr := strings.TrimSpace(string(version))
	versionStr = strings.TrimPrefix(versionStr, "git version ")

	return checkResult{
		name:    "git",
		ok:      true,
		version: versionStr,
	}
}

func checkRuby() checkResult {
	version, err := exec.Command("ruby", "--version").Output()
	if err != nil {
		return checkResult{
			name:    "ruby",
			ok:      false,
			message: "Install Ruby 3.3+: https://www.ruby-lang.org/en/downloads/",
		}
	}

	versionStr := strings.TrimSpace(string(version))
	// Extract version number from "ruby 3.3.9 (2025-07-24 revision f5c772fc7c) [x86_64-linux]"
	parts := strings.Fields(versionStr)
	if len(parts) >= 2 {
		versionStr = parts[1]
	}

	return checkResult{
		name:    "ruby",
		ok:      true,
		version: versionStr,
	}
}

func checkBundler() checkResult {
	version, err := exec.Command("bundle", "--version").Output()
	if err != nil {
		return checkResult{
			name:    "bundler",
			ok:      false,
			message: "Install bundler: gem install bundler",
		}
	}

	versionStr := strings.TrimSpace(string(version))
	versionStr = strings.TrimPrefix(versionStr, "Bundler version ")

	return checkResult{
		name:    "bundler",
		ok:      true,
		version: versionStr,
	}
}

func checkNode() checkResult {
	version, err := exec.Command("node", "--version").Output()
	if err != nil {
		return checkResult{
			name:    "node",
			ok:      false,
			message: "Install Node.js 20+: https://nodejs.org/",
		}
	}

	versionStr := strings.TrimSpace(string(version))

	return checkResult{
		name:    "node",
		ok:      true,
		version: versionStr,
	}
}

func checkPnpm() checkResult {
	version, err := exec.Command("pnpm", "--version").Output()
	if err != nil {
		return checkResult{
			name:    "pnpm",
			ok:      false,
			message: "Install pnpm: npm install -g pnpm",
		}
	}

	versionStr := strings.TrimSpace(string(version))

	return checkResult{
		name:    "pnpm",
		ok:      true,
		version: versionStr,
	}
}

func checkPostgres() checkResult {
	// Try to connect to postgres
	err := exec.Command("psql", "-c", "SELECT version();", "-d", "postgres").Run()
	if err != nil {
		return checkResult{
			name:    "postgresql",
			ok:      false,
			message: "PostgreSQL is not running or not accessible. Start it with your db_start command or system service.",
		}
	}

	// Try to get version
	version, err := exec.Command("psql", "--version").Output()
	versionStr := ""
	if err == nil {
		versionStr = strings.TrimSpace(string(version))
		versionStr = strings.TrimPrefix(versionStr, "psql (PostgreSQL) ")
	}

	return checkResult{
		name:    "postgresql",
		ok:      true,
		version: versionStr,
	}
}

func checkRedis() checkResult {
	// Try to ping redis
	err := exec.Command("redis-cli", "ping").Run()
	if err != nil {
		return checkResult{
			name:    "redis",
			ok:      false,
			message: "Redis is not running. Start it with your db_start command or system service.",
		}
	}

	// Try to get version
	version, err := exec.Command("redis-cli", "--version").Output()
	versionStr := ""
	if err == nil {
		versionStr = strings.TrimSpace(string(version))
		versionStr = strings.TrimPrefix(versionStr, "redis-cli ")
	}

	return checkResult{
		name:    "redis",
		ok:      true,
		version: versionStr,
	}
}
