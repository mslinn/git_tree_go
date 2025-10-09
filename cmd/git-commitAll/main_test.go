package main

import (
  "os"
  "os/exec"
  "path/filepath"
  "strings"
  "testing"
  "time"
)

// TestCommitAll_Integration tests the full commit and push workflow with a real repository
func TestCommitAll_Integration(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory for our test
  tmpDir, err := os.MkdirTemp("", "git-commitAll-test-*")
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
  repoPath := filepath.Join(tmpDir, "real_repo")
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

  // Create a change to be committed by the command
  newFilePath := filepath.Join(repoPath, "new_file.txt")
  if err := os.WriteFile(newFilePath, []byte("This is a new file."), 0644); err != nil {
    t.Fatalf("Failed to write new file: %v", err)
  }

  // Save original os.Args
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Set up command line arguments
  commitMessage := "Integration test commit"
  os.Args = []string{"git-commitAll", "-s", "-m", commitMessage, tmpDir}

  // Run the main function in a goroutine with timeout
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

  // Verify that the new commit exists
  cmd = exec.Command("git", "-C", repoPath, "log", "-1", "--pretty=%B")
  output, err := cmd.Output()
  if err != nil {
    t.Fatalf("Failed to get git log: %v", err)
  }

  logOutput := strings.TrimSpace(string(output))
  if logOutput != commitMessage {
    t.Errorf("Expected commit message to be '%s', got '%s'", commitMessage, logOutput)
  }

  // Verify the file was committed
  cmd = exec.Command("git", "-C", repoPath, "ls-tree", "-r", "HEAD", "--name-only")
  output, err = cmd.Output()
  if err != nil {
    t.Fatalf("Failed to list files: %v", err)
  }

  files := strings.Split(strings.TrimSpace(string(output)), "\n")
  foundNewFile := false
  for _, file := range files {
    if file == "new_file.txt" {
      foundNewFile = true
      break
    }
  }

  if !foundNewFile {
    t.Error("Expected new_file.txt to be committed")
  }
}

// TestRepoHasChanges tests the repoHasChanges function
func TestRepoHasChanges(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Initialize a git repository
  cmd := exec.Command("git", "init", tmpDir)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Configure git user
  cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.name: %v", err)
  }

  cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.email: %v", err)
  }

  // Test with no changes
  ctx := newTestContext()
  hasChanges := repoHasChanges(ctx, tmpDir)
  if hasChanges {
    t.Error("Expected no changes in empty repository")
  }

  // Create a file
  testFile := filepath.Join(tmpDir, "test.txt")
  if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
    t.Fatalf("Failed to write test file: %v", err)
  }

  // Test with untracked changes
  ctx = newTestContext()
  hasChanges = repoHasChanges(ctx, tmpDir)
  if !hasChanges {
    t.Error("Expected changes with untracked file")
  }

  // Add and commit the file
  cmd = exec.Command("git", "-C", tmpDir, "add", ".")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to git add: %v", err)
  }

  cmd = exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit")
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to commit: %v", err)
  }

  // Test with no changes after commit
  ctx = newTestContext()
  hasChanges = repoHasChanges(ctx, tmpDir)
  if hasChanges {
    t.Error("Expected no changes after commit")
  }

  // Modify the file
  if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
    t.Fatalf("Failed to modify file: %v", err)
  }

  // Test with modified file
  ctx = newTestContext()
  hasChanges = repoHasChanges(ctx, tmpDir)
  if !hasChanges {
    t.Error("Expected changes with modified file")
  }
}

// TestRepoHasStagedChanges tests the repoHasStagedChanges function
func TestRepoHasStagedChanges(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Initialize a git repository
  cmd := exec.Command("git", "init", tmpDir)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Configure git user
  cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.name: %v", err)
  }

  cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to configure user.email: %v", err)
  }

  // Test with no staged changes
  ctx := newTestContext()
  hasStagedChanges := repoHasStagedChanges(ctx, tmpDir)
  if hasStagedChanges {
    t.Error("Expected no staged changes in empty repository")
  }

  // Create and stage a file
  testFile := filepath.Join(tmpDir, "test.txt")
  if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
    t.Fatalf("Failed to write test file: %v", err)
  }

  cmd = exec.Command("git", "-C", tmpDir, "add", "test.txt")
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to git add: %v", err)
  }

  // Test with staged changes
  ctx = newTestContext()
  hasStagedChanges = repoHasStagedChanges(ctx, tmpDir)
  if !hasStagedChanges {
    t.Error("Expected staged changes after git add")
  }
}

// Helper function to create a test context with timeout
func newTestContext() *testContext {
  return &testContext{}
}

type testContext struct{}

func (tc *testContext) Deadline() (deadline time.Time, ok bool) {
  return time.Now().Add(5 * time.Second), true
}

func (tc *testContext) Done() <-chan struct{} {
  done := make(chan struct{})
  return done
}

func (tc *testContext) Err() error {
  return nil
}

func (tc *testContext) Value(key interface{}) interface{} {
  return nil
}
