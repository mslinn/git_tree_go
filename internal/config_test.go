package internal

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfig_NewConfig_Defaults tests default configuration values
func TestConfig_NewConfig_Defaults(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Clear any environment variables that might affect the test
	os.Unsetenv("GIT_TREE_GIT_TIMEOUT")
	os.Unsetenv("GIT_TREE_VERBOSITY")
	os.Unsetenv("GIT_TREE_DEFAULT_ROOTS")

	config := NewConfig()

	if config.GitTimeout != 300 {
		t.Errorf("Expected default git_timeout to be 300, got %d", config.GitTimeout)
	}

	if config.Verbosity != LogNormal {
		t.Errorf("Expected default verbosity to be LogNormal (%d), got %d", LogNormal, config.Verbosity)
	}

	expectedRoots := []string{"$sites", "$sitesUbuntu", "$work"}
	if len(config.DefaultRoots) != len(expectedRoots) {
		t.Errorf("Expected default_roots length to be %d, got %d", len(expectedRoots), len(config.DefaultRoots))
	}

	for i, root := range expectedRoots {
		if i >= len(config.DefaultRoots) || config.DefaultRoots[i] != root {
			t.Errorf("Expected default_roots[%d] to be '%s', got '%s'", i, root, config.DefaultRoots[i])
		}
	}
}

// TestConfig_EnvironmentVariables tests that environment variables override defaults
func TestConfig_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("GIT_TREE_GIT_TIMEOUT", "600")
	os.Setenv("GIT_TREE_VERBOSITY", "3")
	os.Setenv("GIT_TREE_DEFAULT_ROOTS", "root1 root2 root3")
	defer func() {
		os.Unsetenv("GIT_TREE_GIT_TIMEOUT")
		os.Unsetenv("GIT_TREE_VERBOSITY")
		os.Unsetenv("GIT_TREE_DEFAULT_ROOTS")
	}()

	config := NewConfig()

	if config.GitTimeout != 600 {
		t.Errorf("Expected git_timeout to be 600, got %d", config.GitTimeout)
	}

	if config.Verbosity != 3 {
		t.Errorf("Expected verbosity to be 3, got %d", config.Verbosity)
	}

	expectedRoots := []string{"root1", "root2", "root3"}
	if len(config.DefaultRoots) != len(expectedRoots) {
		t.Errorf("Expected default_roots length to be %d, got %d", len(expectedRoots), len(config.DefaultRoots))
	}

	for i, root := range expectedRoots {
		if i >= len(config.DefaultRoots) || config.DefaultRoots[i] != root {
			t.Errorf("Expected default_roots[%d] to be '%s', got '%s'", i, root, config.DefaultRoots[i])
		}
	}
}

// TestConfig_InvalidEnvironmentVariables tests that invalid env vars are ignored
func TestConfig_InvalidEnvironmentVariables(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Set invalid environment variables
	os.Setenv("GIT_TREE_GIT_TIMEOUT", "not-a-number")
	os.Setenv("GIT_TREE_VERBOSITY", "also-not-a-number")
	defer func() {
		os.Unsetenv("GIT_TREE_GIT_TIMEOUT")
		os.Unsetenv("GIT_TREE_VERBOSITY")
	}()

	config := NewConfig()

	// Should fall back to defaults when env vars are invalid
	if config.GitTimeout != 300 {
		t.Errorf("Expected git_timeout to fall back to 300, got %d", config.GitTimeout)
	}

	if config.Verbosity != LogNormal {
		t.Errorf("Expected verbosity to fall back to LogNormal (%d), got %d", LogNormal, config.Verbosity)
	}
}

// TestConfig_SaveToFile tests saving configuration to a file
func TestConfig_SaveToFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	config := &Config{
		GitTimeout:   600,
		Verbosity:    3,
		DefaultRoots: []string{"test1", "test2"},
	}

	err = config.SaveToFile()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check that the file was created
	configPath := filepath.Join(tmpDir, ".treeconfig.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to be created at %s", configPath)
	}

	// Load the config back and verify
	loadedConfig := NewConfig()

	if loadedConfig.GitTimeout != 600 {
		t.Errorf("Expected loaded git_timeout to be 600, got %d", loadedConfig.GitTimeout)
	}

	if loadedConfig.Verbosity != 3 {
		t.Errorf("Expected loaded verbosity to be 3, got %d", loadedConfig.Verbosity)
	}

	expectedRoots := []string{"test1", "test2"}
	if len(loadedConfig.DefaultRoots) != len(expectedRoots) {
		t.Errorf("Expected loaded default_roots length to be %d, got %d", len(expectedRoots), len(loadedConfig.DefaultRoots))
	}

	for i, root := range expectedRoots {
		if i >= len(loadedConfig.DefaultRoots) || loadedConfig.DefaultRoots[i] != root {
			t.Errorf("Expected loaded default_roots[%d] to be '%s', got '%s'", i, root, loadedConfig.DefaultRoots[i])
		}
	}
}

// TestConfig_LoadFromFile tests loading configuration from a file
func TestConfig_LoadFromFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create a config file
	configPath := filepath.Join(tmpDir, ".treeconfig.yml")
	configContent := `git_timeout: 450
verbosity: 2
default_roots:
  - custom_root1
  - custom_root2
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load the config
	config := NewConfig()

	if config.GitTimeout != 450 {
		t.Errorf("Expected git_timeout to be 450, got %d", config.GitTimeout)
	}

	if config.Verbosity != 2 {
		t.Errorf("Expected verbosity to be 2, got %d", config.Verbosity)
	}

	expectedRoots := []string{"custom_root1", "custom_root2"}
	if len(config.DefaultRoots) != len(expectedRoots) {
		t.Errorf("Expected default_roots length to be %d, got %d", len(expectedRoots), len(config.DefaultRoots))
	}

	for i, root := range expectedRoots {
		if i >= len(config.DefaultRoots) || config.DefaultRoots[i] != root {
			t.Errorf("Expected default_roots[%d] to be '%s', got '%s'", i, root, config.DefaultRoots[i])
		}
	}
}

// TestConfig_EnvironmentOverridesFile tests that environment variables override file config
func TestConfig_EnvironmentOverridesFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create a config file
	configPath := filepath.Join(tmpDir, ".treeconfig.yml")
	configContent := `git_timeout: 450
verbosity: 2
default_roots:
  - file_root1
  - file_root2
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variables that should override the file
	os.Setenv("GIT_TREE_GIT_TIMEOUT", "800")
	os.Setenv("GIT_TREE_VERBOSITY", "4")
	defer func() {
		os.Unsetenv("GIT_TREE_GIT_TIMEOUT")
		os.Unsetenv("GIT_TREE_VERBOSITY")
	}()

	// Load the config
	config := NewConfig()

	// Environment variables should override file values
	if config.GitTimeout != 800 {
		t.Errorf("Expected git_timeout to be 800 (from env), got %d", config.GitTimeout)
	}

	if config.Verbosity != 4 {
		t.Errorf("Expected verbosity to be 4 (from env), got %d", config.Verbosity)
	}

	// default_roots should come from file (no env override)
	expectedRoots := []string{"file_root1", "file_root2"}
	if len(config.DefaultRoots) != len(expectedRoots) {
		t.Errorf("Expected default_roots length to be %d, got %d", len(expectedRoots), len(config.DefaultRoots))
	}

	for i, root := range expectedRoots {
		if i >= len(config.DefaultRoots) || config.DefaultRoots[i] != root {
			t.Errorf("Expected default_roots[%d] to be '%s', got '%s'", i, root, config.DefaultRoots[i])
		}
	}
}

// TestConfig_NonexistentConfigFile tests behavior when config file doesn't exist
func TestConfig_NonexistentConfigFile(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Clear environment variables
	os.Unsetenv("GIT_TREE_GIT_TIMEOUT")
	os.Unsetenv("GIT_TREE_VERBOSITY")
	os.Unsetenv("GIT_TREE_DEFAULT_ROOTS")

	// Load the config (should use defaults)
	config := NewConfig()

	// Should use default values
	if config.GitTimeout != 300 {
		t.Errorf("Expected git_timeout to be 300 (default), got %d", config.GitTimeout)
	}

	if config.Verbosity != LogNormal {
		t.Errorf("Expected verbosity to be LogNormal (%d), got %d", LogNormal, config.Verbosity)
	}
}
