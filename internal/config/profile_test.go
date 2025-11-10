package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfile(t *testing.T) {
	// Create a temporary profile
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, ".denver", "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatal(err)
	}

	profileContent := `name: Test Profile
description: A test profile
plugins:
  - name: discourse-chat
    repo: discourse/discourse-chat
site_settings:
  title: "Test Site"
seed:
  admin: true
  sample_users: 5
  sample_topics: 10
`

	profilePath := filepath.Join(profilesDir, "test.yml")
	if err := os.WriteFile(profilePath, []byte(profileContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Load the profile
	profile, err := LoadProfile("test")
	if err != nil {
		t.Fatalf("LoadProfile failed: %v", err)
	}

	// Verify profile contents
	if profile.Name != "Test Profile" {
		t.Errorf("Expected name 'Test Profile', got '%s'", profile.Name)
	}

	if len(profile.Plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(profile.Plugins))
	}

	if profile.Plugins[0].Name != "discourse-chat" {
		t.Errorf("Expected plugin name 'discourse-chat', got '%s'", profile.Plugins[0].Name)
	}

	if profile.Settings["title"] != "Test Site" {
		t.Errorf("Expected site title 'Test Site', got '%v'", profile.Settings["title"])
	}

	if profile.Seed.Admin != true {
		t.Error("Expected admin to be true")
	}

	if profile.Seed.SampleUsers != 5 {
		t.Errorf("Expected 5 sample users, got %d", profile.Seed.SampleUsers)
	}
}

func TestLoadProfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Try to load non-existent profile
	_, err := LoadProfile("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent profile, got nil")
	}
}

func TestGetDenverDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	denverDir, err := GetDenverDir()
	if err != nil {
		t.Fatalf("GetDenverDir failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".denver")
	if denverDir != expected {
		t.Errorf("Expected denver dir '%s', got '%s'", expected, denverDir)
	}
}

func TestGetEnvironmentsDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	envsDir, err := GetEnvironmentsDir()
	if err != nil {
		t.Fatalf("GetEnvironmentsDir failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".denver", "environments")
	if envsDir != expected {
		t.Errorf("Expected environments dir '%s', got '%s'", expected, envsDir)
	}
}

func TestGetBareRepoPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	bareRepoPath, err := GetBareRepoPath()
	if err != nil {
		t.Fatalf("GetBareRepoPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".denver", "discourse.git")
	if bareRepoPath != expected {
		t.Errorf("Expected bare repo path '%s', got '%s'", expected, bareRepoPath)
	}
}
