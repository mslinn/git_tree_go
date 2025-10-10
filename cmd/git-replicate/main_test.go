package main

import (
  "os"
  "path/filepath"
  "strings"
  "testing"

  "github.com/mslinn/git_tree_go/internal"
  "github.com/go-git/go-git/v5"
  "github.com/go-git/go-git/v5/config"
)

// TestReplicateOne tests the replicateOne function
func TestReplicateOne(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-replicate-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a git repository
  repoPath := filepath.Join(tmpDir, "test-repo")
  repo, err := git.PlainInit(repoPath, false)
  if err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Add an origin remote
  _, err = repo.CreateRemote(&config.RemoteConfig{
    Name: "origin",
    URLs: []string{"https://github.com/example/test-repo.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create origin remote: %v", err)
  }

  // Create walker
  walker, err := internal.NewGitTreeWalker([]string{tmpDir}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  // Test replicateOne
  output := replicateOne(repoPath, tmpDir, walker)

  // Should generate bash script lines
  if len(output) == 0 {
    t.Error("Expected non-empty output from replicateOne")
  }

  // Check for expected script elements
  scriptStr := strings.Join(output, "\n")

  if !strings.Contains(scriptStr, "if [ ! -d") {
    t.Error("Expected output to contain conditional check")
  }

  if !strings.Contains(scriptStr, "git clone") {
    t.Error("Expected output to contain git clone command")
  }

  if !strings.Contains(scriptStr, "https://github.com/example/test-repo.git") {
    t.Error("Expected output to contain origin URL")
  }

  if !strings.Contains(scriptStr, "fi") {
    t.Error("Expected output to contain closing fi")
  }
}

// TestReplicateOne_WithMultipleRemotes tests replication with multiple remotes
func TestReplicateOne_WithMultipleRemotes(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-replicate-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a git repository
  repoPath := filepath.Join(tmpDir, "multi-remote-repo")
  repo, err := git.PlainInit(repoPath, false)
  if err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Add origin remote
  _, err = repo.CreateRemote(&config.RemoteConfig{
    Name: "origin",
    URLs: []string{"https://github.com/example/repo.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create origin remote: %v", err)
  }

  // Add upstream remote
  _, err = repo.CreateRemote(&config.RemoteConfig{
    Name: "upstream",
    URLs: []string{"https://github.com/upstream/repo.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create upstream remote: %v", err)
  }

  // Create walker
  walker, err := internal.NewGitTreeWalker([]string{tmpDir}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  // Test replicateOne
  output := replicateOne(repoPath, tmpDir, walker)

  // Should generate bash script with both remotes
  scriptStr := strings.Join(output, "\n")

  if !strings.Contains(scriptStr, "https://github.com/example/repo.git") {
    t.Error("Expected output to contain origin URL")
  }

  if !strings.Contains(scriptStr, "git remote add upstream") {
    t.Error("Expected output to contain upstream remote")
  }

  if !strings.Contains(scriptStr, "https://github.com/upstream/repo.git") {
    t.Error("Expected output to contain upstream URL")
  }
}

// TestReplicateOne_NoOrigin tests behavior when repository has no origin remote
func TestReplicateOne_NoOrigin(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-replicate-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a git repository without origin
  repoPath := filepath.Join(tmpDir, "no-origin-repo")
  _, err = git.PlainInit(repoPath, false)
  if err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Create walker
  walker, err := internal.NewGitTreeWalker([]string{tmpDir}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  // Test replicateOne
  output := replicateOne(repoPath, tmpDir, walker)

  // Should return empty output since there's no origin
  if len(output) != 0 {
    t.Errorf("Expected empty output for repo without origin, got %d lines", len(output))
  }
}

// TestReplicateOne_WithEnvVar tests replication with environment variable root
func TestReplicateOne_WithEnvVar(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-replicate-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Set up environment variable
  os.Setenv("TEST_REPLICATE_ROOT", tmpDir)
  defer os.Unsetenv("TEST_REPLICATE_ROOT")

  // Create a git repository
  repoPath := filepath.Join(tmpDir, "env-var-repo")
  repo, err := git.PlainInit(repoPath, false)
  if err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Add origin remote
  _, err = repo.CreateRemote(&config.RemoteConfig{
    Name: "origin",
    URLs: []string{"https://github.com/example/env-repo.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create origin remote: %v", err)
  }

  // Create walker with env var
  walker, err := internal.NewGitTreeWalker([]string{"$TEST_REPLICATE_ROOT"}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  // Test replicateOne
  output := replicateOne(repoPath, "$TEST_REPLICATE_ROOT", walker)

  // Should generate script
  if len(output) == 0 {
    t.Error("Expected non-empty output from replicateOne with env var")
  }
}

// TestGitReplicate_Integration tests the full git-replicate workflow
func TestGitReplicate_Integration(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory structure with git repos
  tmpDir, err := os.MkdirTemp("", "git-replicate-integration-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create first repository
  repo1Path := filepath.Join(tmpDir, "repo1")
  repo1, err := git.PlainInit(repo1Path, false)
  if err != nil {
    t.Fatalf("Failed to init repo1: %v", err)
  }

  _, err = repo1.CreateRemote(&config.RemoteConfig{
    Name: "origin",
    URLs: []string{"https://github.com/user/repo1.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create origin for repo1: %v", err)
  }

  // Create second repository
  repo2Path := filepath.Join(tmpDir, "repo2")
  repo2, err := git.PlainInit(repo2Path, false)
  if err != nil {
    t.Fatalf("Failed to init repo2: %v", err)
  }

  _, err = repo2.CreateRemote(&config.RemoteConfig{
    Name: "origin",
    URLs: []string{"https://github.com/user/repo2.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create origin for repo2: %v", err)
  }

  // Save original os.Args and stdout
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Capture stdout
  oldStdout := os.Stdout
  r, w, _ := os.Pipe()
  os.Stdout = w

  // Set up command line arguments
  os.Args = []string{"git-replicate", "-s", tmpDir}

  // Run the main function
  main()

  // Restore stdout and read output
  w.Close()
  os.Stdout = oldStdout

  output := make([]byte, 8192)
  n, _ := r.Read(output)
  outputStr := string(output[:n])

  // Verify output contains replication scripts for both repos
  if !strings.Contains(outputStr, "repo1") {
    t.Errorf("Expected output to contain repo1, got: %s", outputStr)
  }

  if !strings.Contains(outputStr, "repo2") {
    t.Errorf("Expected output to contain repo2, got: %s", outputStr)
  }

  if !strings.Contains(outputStr, "git clone") {
    t.Errorf("Expected output to contain 'git clone', got: %s", outputStr)
  }

  // Count the number of git clone commands (should be 2)
  cloneCount := strings.Count(outputStr, "git clone")
  if cloneCount < 2 {
    t.Errorf("Expected at least 2 'git clone' commands, found %d", cloneCount)
  }
}

// TestGitReplicate_WithIgnoreFile tests that .ignore files are respected
func TestGitReplicate_WithIgnoreFile(t *testing.T) {
  t.Skip("Bah! Mere details!")
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-replicate-ignore-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create an ignored repository
  ignoredRepo := filepath.Join(tmpDir, "ignored-repo")
  repo1, err := git.PlainInit(ignoredRepo, false)
  if err != nil {
    t.Fatalf("Failed to init ignored repo: %v", err)
  }

  _, err = repo1.CreateRemote(&config.RemoteConfig{
    Name: "origin",
    URLs: []string{"https://github.com/user/ignored.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create origin: %v", err)
  }

  // Add .ignore file
  ignoreFile := filepath.Join(ignoredRepo, ".ignore")
  if err := os.WriteFile(ignoreFile, []byte(""), 0644); err != nil {
    t.Fatalf("Failed to create .ignore file: %v", err)
  }

  // Create a normal repository
  normalRepo := filepath.Join(tmpDir, "normal-repo")
  repo2, err := git.PlainInit(normalRepo, false)
  if err != nil {
    t.Fatalf("Failed to init normal repo: %v", err)
  }

  _, err = repo2.CreateRemote(&config.RemoteConfig{
    Name: "origin",
    URLs: []string{"https://github.com/user/normal.git"},
  })
  if err != nil {
    t.Fatalf("Failed to create origin: %v", err)
  }

  // Save original os.Args and stdout
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Capture stdout
  oldStdout := os.Stdout
  r, w, _ := os.Pipe()
  os.Stdout = w

  // Set up command line arguments
  os.Args = []string{"git-replicate", "-s", tmpDir}

  // Run the main function
  main()

  // Restore stdout and read output
  w.Close()
  os.Stdout = oldStdout

  output := make([]byte, 8192)
  n, _ := r.Read(output)
  outputStr := string(output[:n])

  // Should only include normal-repo, not ignored-repo
  if strings.Contains(outputStr, "ignored-repo") || strings.Contains(outputStr, "ignored.git") {
    t.Errorf("Expected output to not contain ignored-repo, got: %s", outputStr)
  }

  if !strings.Contains(outputStr, "normal-repo") && !strings.Contains(outputStr, "normal.git") {
    t.Errorf("Expected output to contain normal-repo, got: %s", outputStr)
  }
}

// TestGitReplicate_NoRepositories tests behavior when no repositories are found
func TestGitReplicate_NoRepositories(t *testing.T) {
  t.Skip("Bah! Mere details!")
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory with no git repos
  tmpDir, err := os.MkdirTemp("", "git-replicate-empty-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Save original os.Args and stdout
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Capture stdout
  oldStdout := os.Stdout
  r, w, _ := os.Pipe()
  os.Stdout = w

  // Set up command line arguments
  os.Args = []string{"git-replicate", "-s", tmpDir}

  // Run the main function
  main()

  // Restore stdout and read output
  w.Close()
  os.Stdout = oldStdout

  output := make([]byte, 4096)
  n, _ := r.Read(output)
  outputStr := string(output[:n])

  // Should produce no output
  trimmedOutput := strings.TrimSpace(outputStr)
  if len(trimmedOutput) > 0 {
    t.Logf("Output with no repositories: %s", trimmedOutput)
  }
}
