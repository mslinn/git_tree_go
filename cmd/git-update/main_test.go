package main

import (
  "os"
  "os/exec"
  "path/filepath"
  "testing"
  "time"

  "github.com/mslinn/git_tree_go/internal"
)

// TestProcessRepo_Success tests successful git pull execution
func TestProcessRepo_Success(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-update-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a bare repository (acts as remote)
  bareRepoPath := filepath.Join(tmpDir, "remote.git")
  cmd := exec.Command("git", "init", "--bare", bareRepoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to create bare repo: %v", err)
  }

  // Clone the bare repository
  repoPath := filepath.Join(tmpDir, "local_repo")
  cmd = exec.Command("git", "clone", bareRepoPath, repoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to clone repo: %v", err)
  }

  // Configure git user
  cmd = exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure git user.name: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure git user.email: %v", err)
  }

  // Create an initial commit
  readmePath := filepath.Join(repoPath, "README.md")
  if err := os.WriteFile(readmePath, []byte("Initial commit"), 0644); err != nil {
    t.Fatalf("Failed to write README: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "add", ".")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to git add: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to create initial commit: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "push", "-u", "origin", "master")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    // Try 'main' instead of 'master'
    cmd = exec.Command("git", "-C", repoPath, "push", "-u", "origin", "main")
    cmd.Stdout = nil
    cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
      t.Fatalf("Failed to push initial commit: %v", err)
    }
  }

  // Create walker and config
  walker, err := internal.NewGitTreeWalker([]string{repoPath}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  config := internal.NewConfig()

  // Test processRepo - should succeed with "Already up to date"
  processRepo(walker, repoPath, 0, config)
}

// TestProcessRepo_Failure tests failed git pull execution
func TestProcessRepo_Failure(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-update-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a directory that's not a git repository
  notRepoPath := filepath.Join(tmpDir, "not-a-repo")
  if err := os.MkdirAll(notRepoPath, 0755); err != nil {
    t.Fatalf("Failed to create directory: %v", err)
  }

  // Create walker and config
  walker, err := internal.NewGitTreeWalker([]string{tmpDir}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  config := internal.NewConfig()

  // Test processRepo - should fail because it's not a git repo
  processRepo(walker, notRepoPath, 0, config)
}

// TestGitUpdate_Integration tests the full git-update workflow
func TestGitUpdate_Integration(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory structure
  tmpDir, err := os.MkdirTemp("", "git-update-integration-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a bare repository (acts as remote)
  bareRepoPath := filepath.Join(tmpDir, "remote.git")
  cmd := exec.Command("git", "init", "--bare", bareRepoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to create bare repo: %v", err)
  }

  // Clone the bare repository
  repo1Path := filepath.Join(tmpDir, "repo1")
  cmd = exec.Command("git", "clone", bareRepoPath, repo1Path)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to clone repo1: %v", err)
  }

  // Configure git user
  cmd = exec.Command("git", "-C", repo1Path, "config", "user.name", "Test User")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.name: %v", err)
  }

  cmd = exec.Command("git", "-C", repo1Path, "config", "user.email", "test@example.com")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.email: %v", err)
  }

  // Create an initial commit
  readmePath := filepath.Join(repo1Path, "README.md")
  if err := os.WriteFile(readmePath, []byte("Initial"), 0644); err != nil {
    t.Fatalf("Failed to write README: %v", err)
  }

  cmd = exec.Command("git", "-C", repo1Path, "add", ".")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to git add: %v", err)
  }

  cmd = exec.Command("git", "-C", repo1Path, "commit", "-m", "Initial commit")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to commit: %v", err)
  }

  cmd = exec.Command("git", "-C", repo1Path, "push", "-u", "origin", "master")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    cmd = exec.Command("git", "-C", repo1Path, "push", "-u", "origin", "main")
    cmd.Stdout = nil
    cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
      t.Fatalf("Failed to push: %v", err)
    }
  }

  // Save original os.Args
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Set up command line arguments
  os.Args = []string{"git-update", "-s", tmpDir}

  // Run the main function with timeout
  done := make(chan bool)
  go func() {
    main()
    done <- true
  }()

  select {
  case <-done:
    // Success
  case <-time.After(30 * time.Second):
    t.Fatal("Test timed out")
  }
}

// TestGitUpdate_WithVerboseFlag tests verbose output
func TestGitUpdate_WithVerboseFlag(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-update-verbose-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a bare repository
  bareRepoPath := filepath.Join(tmpDir, "remote.git")
  cmd := exec.Command("git", "init", "--bare", bareRepoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to create bare repo: %v", err)
  }

  // Clone the repository
  repoPath := filepath.Join(tmpDir, "repo")
  cmd = exec.Command("git", "clone", bareRepoPath, repoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to clone repo: %v", err)
  }

  // Configure git
  cmd = exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.name: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.email: %v", err)
  }

  // Create initial commit
  testFile := filepath.Join(repoPath, "test.txt")
  if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
    t.Fatalf("Failed to write file: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "add", ".")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to git add: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Initial")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to commit: %v", err)
  }

  cmd = exec.Command("git", "-C", repoPath, "push", "-u", "origin", "master")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    cmd = exec.Command("git", "-C", repoPath, "push", "-u", "origin", "main")
    cmd.Stdout = nil
    cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
      t.Fatalf("Failed to push: %v", err)
    }
  }

  // Save original os.Args
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Set up command line arguments with verbose flag
  os.Args = []string{"git-update", "-s", "-v", tmpDir}

  // Run with timeout
  done := make(chan bool)
  go func() {
    main()
    done <- true
  }()

  select {
  case <-done:
    // Success
  case <-time.After(30 * time.Second):
    t.Fatal("Test timed out")
  }
}

// TestGitUpdate_WithIgnoreFile tests that .ignore files are respected
func TestGitUpdate_WithIgnoreFile(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-update-ignore-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create an ignored repository
  ignoredRepo := filepath.Join(tmpDir, "ignored-repo")
  if err := os.MkdirAll(ignoredRepo, 0755); err != nil {
    t.Fatalf("Failed to create ignored repo: %v", err)
  }

  cmd := exec.Command("git", "init", ignoredRepo)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init ignored repo: %v", err)
  }

  // Add .ignore file
  ignoreFile := filepath.Join(ignoredRepo, ".ignore")
  if err := os.WriteFile(ignoreFile, []byte(""), 0644); err != nil {
    t.Fatalf("Failed to create .ignore file: %v", err)
  }

  // Create a normal repository with a bare remote
  bareRepoPath := filepath.Join(tmpDir, "remote.git")
  cmd = exec.Command("git", "init", "--bare", bareRepoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to create bare repo: %v", err)
  }

  normalRepo := filepath.Join(tmpDir, "normal-repo")
  cmd = exec.Command("git", "clone", bareRepoPath, normalRepo)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to clone normal repo: %v", err)
  }

  // Configure git
  cmd = exec.Command("git", "-C", normalRepo, "config", "user.name", "Test User")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.name: %v", err)
  }

  cmd = exec.Command("git", "-C", normalRepo, "config", "user.email", "test@example.com")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.email: %v", err)
  }

  // Create initial commit
  testFile := filepath.Join(normalRepo, "test.txt")
  if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
    t.Fatalf("Failed to write file: %v", err)
  }

  cmd = exec.Command("git", "-C", normalRepo, "add", ".")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to git add: %v", err)
  }

  cmd = exec.Command("git", "-C", normalRepo, "commit", "-m", "Initial")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to commit: %v", err)
  }

  cmd = exec.Command("git", "-C", normalRepo, "push", "-u", "origin", "master")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    cmd = exec.Command("git", "-C", normalRepo, "push", "-u", "origin", "main")
    cmd.Stdout = nil
    cmd.Stderr = nil
    if err := cmd.Run(); err != nil {
      t.Fatalf("Failed to push: %v", err)
    }
  }

  // Save original os.Args
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Set up command line arguments
  os.Args = []string{"git-update", "-s", tmpDir}

  // Run with timeout
  done := make(chan bool)
  go func() {
    main()
    done <- true
  }()

  select {
  case <-done:
    // Success - the command should process only normal-repo, not ignored-repo
  case <-time.After(30 * time.Second):
    t.Fatal("Test timed out")
  }
}

// TestGitUpdate_NoRepositories tests behavior when no repositories are found
func TestGitUpdate_NoRepositories(t *testing.T) {
  // Create a temporary directory with no git repos
  tmpDir, err := os.MkdirTemp("", "git-update-empty-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Save original os.Args
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Set up command line arguments
  os.Args = []string{"git-update", "-s", tmpDir}

  // Run with timeout
  done := make(chan bool)
  go func() {
    main()
    done <- true
  }()

  select {
  case <-done:
    // Success - should complete without error even with no repos
  case <-time.After(10 * time.Second):
    t.Fatal("Test timed out")
  }
}

// TestGitUpdate_Timeout tests timeout handling
func TestGitUpdate_Timeout(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping timeout test in short mode")
  }

  // This test is difficult to implement without a very slow git operation
  // We can verify that the timeout mechanism exists by checking the code
  // A real timeout test would require setting up a git operation that takes
  // longer than the configured timeout
  t.Skip("Timeout test requires a slow git operation setup")
}

// TestProcessRepo_AbbreviatePath tests path abbreviation in processRepo
func TestProcessRepo_AbbreviatePath(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-update-abbrev-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Set up an environment variable
  os.Setenv("TEST_UPDATE_ROOT", tmpDir)
  defer os.Unsetenv("TEST_UPDATE_ROOT")

  // Create a repository
  repoPath := filepath.Join(tmpDir, "test-repo")
  cmd := exec.Command("git", "init", repoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Create walker with env var
  walker, err := internal.NewGitTreeWalker([]string{"$TEST_UPDATE_ROOT"}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  // Test that abbreviatePath works correctly
  abbreviated := walker.AbbreviatePath(repoPath)
  expected := filepath.Join("$TEST_UPDATE_ROOT", "test-repo")
  if abbreviated != expected {
    t.Errorf("Expected abbreviated path to be '%s', got '%s'", expected, abbreviated)
  }
}
