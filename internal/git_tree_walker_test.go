package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGitTreeWalker_Initialization tests walker initialization
func TestGitTreeWalker_Initialization(t *testing.T) {
	walker, err := NewGitTreeWalker([]string{"/tmp"}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	if walker.Config == nil {
		t.Error("Expected config to be initialized")
	}

	if walker.RootMap == nil {
		t.Error("Expected root map to be initialized")
	}

	if walker.DisplayRoots == nil {
		t.Error("Expected display roots to be initialized")
	}
}

// TestGitTreeWalker_DetermineRoots_NoArgs tests that default roots are used when no args provided
func TestGitTreeWalker_DetermineRoots_NoArgs(t *testing.T) {
	walker, err := NewGitTreeWalker([]string{}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Should use default roots from config
	if len(walker.DisplayRoots) == 0 {
		t.Error("Expected display roots to be populated from config")
	}

	// Display roots should match config default roots
	expectedRoots := walker.Config.DefaultRoots
	if len(walker.DisplayRoots) != len(expectedRoots) {
		t.Errorf("Expected %d display roots, got %d", len(expectedRoots), len(walker.DisplayRoots))
	}
}

// TestGitTreeWalker_DetermineRoots_WithArgs tests root determination with explicit args
func TestGitTreeWalker_DetermineRoots_WithArgs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	walker, err := NewGitTreeWalker([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Display roots should match the args
	if len(walker.DisplayRoots) != 1 {
		t.Errorf("Expected 1 display root, got %d", len(walker.DisplayRoots))
	}

	if walker.DisplayRoots[0] != tmpDir {
		t.Errorf("Expected display root to be '%s', got '%s'", tmpDir, walker.DisplayRoots[0])
	}

	// Root map should contain the absolute path
	absTmpDir, _ := filepath.Abs(tmpDir)
	if paths, ok := walker.RootMap[tmpDir]; !ok || len(paths) != 1 {
		t.Errorf("Expected root map to contain '%s'", tmpDir)
	} else if paths[0] != absTmpDir {
		t.Errorf("Expected root map path to be '%s', got '%s'", absTmpDir, paths[0])
	}
}

// TestGitTreeWalker_DetermineRoots_WithEnvVar tests environment variable expansion
func TestGitTreeWalker_DetermineRoots_WithEnvVar(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set an environment variable
	os.Setenv("TEST_WORK", tmpDir)
	defer os.Unsetenv("TEST_WORK")

	walker, err := NewGitTreeWalker([]string{"$TEST_WORK"}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Display roots should show the variable name
	if len(walker.DisplayRoots) != 1 {
		t.Errorf("Expected 1 display root, got %d", len(walker.DisplayRoots))
	}

	if walker.DisplayRoots[0] != "$TEST_WORK" {
		t.Errorf("Expected display root to be '$TEST_WORK', got '%s'", walker.DisplayRoots[0])
	}

	// Root map should contain the expanded path
	absTmpDir, _ := filepath.Abs(tmpDir)
	if paths, ok := walker.RootMap["$TEST_WORK"]; !ok || len(paths) != 1 {
		t.Errorf("Expected root map to contain '$TEST_WORK'")
	} else if paths[0] != absTmpDir {
		t.Errorf("Expected root map path to be '%s', got '%s'", absTmpDir, paths[0])
	}
}

// TestGitTreeWalker_DetermineRoots_UndefinedEnvVar tests error on undefined env var
func TestGitTreeWalker_DetermineRoots_UndefinedEnvVar(t *testing.T) {
	os.Unsetenv("UNDEFINED_VAR")

	_, err := NewGitTreeWalker([]string{"$UNDEFINED_VAR"}, false)
	if err == nil {
		t.Error("Expected error for undefined environment variable")
	}

	if !strings.Contains(err.Error(), "undefined") {
		t.Errorf("Expected error message to contain 'undefined', got: %v", err)
	}
}

// TestGitTreeWalker_DetermineRoots_MultipleEnvVars tests multiple environment variables
func TestGitTreeWalker_DetermineRoots_MultipleEnvVars(t *testing.T) {
	tmpDir1, err := os.MkdirTemp("", "git-tree-test1")
	if err != nil {
		t.Fatalf("Failed to create temp dir 1: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "git-tree-test2")
	if err != nil {
		t.Fatalf("Failed to create temp dir 2: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	os.Setenv("TEST_WORK", tmpDir1)
	os.Setenv("TEST_SITES", tmpDir2)
	defer func() {
		os.Unsetenv("TEST_WORK")
		os.Unsetenv("TEST_SITES")
	}()

	walker, err := NewGitTreeWalker([]string{"$TEST_WORK", "$TEST_SITES"}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Check display roots
	expectedDisplayRoots := []string{"$TEST_WORK", "$TEST_SITES"}
	if len(walker.DisplayRoots) != len(expectedDisplayRoots) {
		t.Errorf("Expected %d display roots, got %d", len(expectedDisplayRoots), len(walker.DisplayRoots))
	}

	// Check root map
	if len(walker.RootMap) != 2 {
		t.Errorf("Expected 2 entries in root map, got %d", len(walker.RootMap))
	}

	absTmpDir1, _ := filepath.Abs(tmpDir1)
	absTmpDir2, _ := filepath.Abs(tmpDir2)

	if paths, ok := walker.RootMap["$TEST_WORK"]; !ok || len(paths) != 1 || paths[0] != absTmpDir1 {
		t.Errorf("Expected root map to contain '$TEST_WORK' -> '%s'", absTmpDir1)
	}

	if paths, ok := walker.RootMap["$TEST_SITES"]; !ok || len(paths) != 1 || paths[0] != absTmpDir2 {
		t.Errorf("Expected root map to contain '$TEST_SITES' -> '%s'", absTmpDir2)
	}
}

// TestGitTreeWalker_AbbreviatePath tests path abbreviation
func TestGitTreeWalker_AbbreviatePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("TEST_ROOT", tmpDir)
	defer os.Unsetenv("TEST_ROOT")

	walker, err := NewGitTreeWalker([]string{"$TEST_ROOT"}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Create a subdirectory path
	subDir := filepath.Join(tmpDir, "subdir", "project")

	// Abbreviate the path
	abbreviated := walker.AbbreviatePath(subDir)

	// Should replace the tmpDir with $TEST_ROOT
	expected := filepath.Join("$TEST_ROOT", "subdir", "project")
	if abbreviated != expected {
		t.Errorf("Expected abbreviated path to be '%s', got '%s'", expected, abbreviated)
	}
}

// TestGitTreeWalker_AbbreviatePath_NoMatch tests abbreviation with no matching root
func TestGitTreeWalker_AbbreviatePath_NoMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	walker, err := NewGitTreeWalker([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Try to abbreviate a path that doesn't match any root
	unrelatedPath := "/some/other/path"
	abbreviated := walker.AbbreviatePath(unrelatedPath)

	// Should return the path unchanged
	if abbreviated != unrelatedPath {
		t.Errorf("Expected path to be unchanged, got '%s'", abbreviated)
	}
}

// TestGitTreeWalker_FindGitRepos tests finding git repositories
func TestGitTreeWalker_FindGitRepos(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock git repo
	repo1 := filepath.Join(tmpDir, "repo1")
	os.MkdirAll(filepath.Join(repo1, ".git"), 0755)

	// Create a nested git repo
	repo2 := filepath.Join(tmpDir, "subdir", "repo2")
	os.MkdirAll(filepath.Join(repo2, ".git"), 0755)

	// Create a non-git directory
	nonRepo := filepath.Join(tmpDir, "not-a-repo")
	os.MkdirAll(nonRepo, 0755)

	walker, err := NewGitTreeWalker([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Find all repos
	foundRepos := []string{}
	walker.FindAndProcessRepos(func(dir, rootArg string) {
		foundRepos = append(foundRepos, dir)
	})

	// Should find both repos
	if len(foundRepos) != 2 {
		t.Errorf("Expected to find 2 repos, found %d: %v", len(foundRepos), foundRepos)
	}

	// Check that both repos were found
	foundRepo1 := false
	foundRepo2 := false
	for _, repo := range foundRepos {
		if repo == repo1 {
			foundRepo1 = true
		}
		if repo == repo2 {
			foundRepo2 = true
		}
	}

	if !foundRepo1 {
		t.Errorf("Expected to find repo1 at '%s'", repo1)
	}

	if !foundRepo2 {
		t.Errorf("Expected to find repo2 at '%s'", repo2)
	}
}

// TestGitTreeWalker_FindGitRepos_WithIgnoreFile tests that .ignore files are respected
func TestGitTreeWalker_FindGitRepos_WithIgnoreFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a repo that should be ignored
	ignoredRepo := filepath.Join(tmpDir, "ignored-repo")
	os.MkdirAll(filepath.Join(ignoredRepo, ".git"), 0755)
	os.WriteFile(filepath.Join(ignoredRepo, ".ignore"), []byte(""), 0644)

	// Create a normal repo
	normalRepo := filepath.Join(tmpDir, "normal-repo")
	os.MkdirAll(filepath.Join(normalRepo, ".git"), 0755)

	walker, err := NewGitTreeWalker([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Find all repos
	foundRepos := []string{}
	walker.FindAndProcessRepos(func(dir, rootArg string) {
		foundRepos = append(foundRepos, dir)
	})

	// Should only find the normal repo
	if len(foundRepos) != 1 {
		t.Errorf("Expected to find 1 repo, found %d: %v", len(foundRepos), foundRepos)
	}

	if len(foundRepos) > 0 && foundRepos[0] != normalRepo {
		t.Errorf("Expected to find normal repo at '%s', got '%s'", normalRepo, foundRepos[0])
	}
}

// TestGitTreeWalker_SerialFlag tests serial mode flag
func TestGitTreeWalker_SerialFlag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	walker, err := NewGitTreeWalker([]string{tmpDir}, true)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	if !walker.Serial {
		t.Error("Expected serial flag to be true")
	}
}

// TestGitTreeWalker_Process tests the Process function
func TestGitTreeWalker_Process(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a couple of git repos
	repo1 := filepath.Join(tmpDir, "repo1")
	os.MkdirAll(filepath.Join(repo1, ".git"), 0755)

	repo2 := filepath.Join(tmpDir, "repo2")
	os.MkdirAll(filepath.Join(repo2, ".git"), 0755)

	walker, err := NewGitTreeWalker([]string{tmpDir}, false)
	if err != nil {
		t.Fatalf("Failed to create walker: %v", err)
	}

	// Process repos and collect them
	processedRepos := []string{}
	walker.Process(func(dir string, threadID int, w *GitTreeWalker) {
		processedRepos = append(processedRepos, dir)
	})

	// Should have processed both repos
	if len(processedRepos) != 2 {
		t.Errorf("Expected to process 2 repos, processed %d", len(processedRepos))
	}
}
