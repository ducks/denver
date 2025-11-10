package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadActive(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Create .denver directory
	denverDir := filepath.Join(tmpDir, ".denver")
	if err := os.MkdirAll(denverDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create and save active environment
	activeEnv := &ActiveEnvironment{
		Name:      "test-env",
		RailsPID:  12345,
		EmberPID:  67890,
		RailsPort: 3000,
		EmberPort: 4200,
	}

	if err := SaveActive(activeEnv); err != nil {
		t.Fatalf("SaveActive failed: %v", err)
	}

	// Load it back
	loaded, err := LoadActive()
	if err != nil {
		t.Fatalf("LoadActive failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadActive returned nil")
	}

	// Verify contents
	if loaded.Name != activeEnv.Name {
		t.Errorf("Expected name '%s', got '%s'", activeEnv.Name, loaded.Name)
	}

	if loaded.RailsPID != activeEnv.RailsPID {
		t.Errorf("Expected Rails PID %d, got %d", activeEnv.RailsPID, loaded.RailsPID)
	}

	if loaded.EmberPID != activeEnv.EmberPID {
		t.Errorf("Expected Ember PID %d, got %d", activeEnv.EmberPID, loaded.EmberPID)
	}

	if loaded.RailsPort != activeEnv.RailsPort {
		t.Errorf("Expected Rails port %d, got %d", activeEnv.RailsPort, loaded.RailsPort)
	}

	if loaded.EmberPort != activeEnv.EmberPort {
		t.Errorf("Expected Ember port %d, got %d", activeEnv.EmberPort, loaded.EmberPort)
	}
}

func TestLoadActiveNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Create .denver directory
	denverDir := filepath.Join(tmpDir, ".denver")
	if err := os.MkdirAll(denverDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Load when no active file exists
	loaded, err := LoadActive()
	if err != nil {
		t.Fatalf("LoadActive failed: %v", err)
	}

	if loaded != nil {
		t.Error("Expected nil for non-existent active file, got non-nil")
	}
}

func TestClearActive(t *testing.T) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	// Create .denver directory
	denverDir := filepath.Join(tmpDir, ".denver")
	if err := os.MkdirAll(denverDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Save active environment
	activeEnv := &ActiveEnvironment{
		Name:      "test-env",
		RailsPID:  12345,
		EmberPID:  67890,
		RailsPort: 3000,
		EmberPort: 4200,
	}

	if err := SaveActive(activeEnv); err != nil {
		t.Fatalf("SaveActive failed: %v", err)
	}

	// Clear it
	if err := ClearActive(); err != nil {
		t.Fatalf("ClearActive failed: %v", err)
	}

	// Verify it's gone
	loaded, err := LoadActive()
	if err != nil {
		t.Fatalf("LoadActive failed: %v", err)
	}

	if loaded != nil {
		t.Error("Expected nil after ClearActive, got non-nil")
	}
}
