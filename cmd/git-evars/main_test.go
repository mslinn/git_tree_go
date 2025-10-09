package main

import (
  "os"
  "os/exec"
  "path/filepath"
  "strings"
  "testing"

  "git-tree-go/internal"
)

// TestEnvVarName tests the envVarName function
func TestEnvVarName(t *testing.T) {
  tests := []struct {
    name     string
    input    string
    expected string
  }{
    {"simple name", "/path/to/myrepo", "myrepo"},
    {"with hyphens", "/path/to/my-repo", "my_repo"},
    {"with spaces", "/path/to/my repo", "my_repo"},
    {"www prefix", "/path/to/www.example.com", "example"},
    {"with extension", "/path/to/repo.git", "repo"},
    {"empty", "", ""},
    {"root", "/", ""},
    {"dot", ".", ""},
    {"complex name", "/home/user/www.github.com", "github"},
    {"underscore existing", "/path/to/my_repo", "my_repo"},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      result := envVarName(tt.input)
      if result != tt.expected {
        t.Errorf("envVarName(%q) = %q, expected %q", tt.input, result, tt.expected)
      }
    })
  }
}

// TestMakeEnvVarWithSubstitution tests environment variable generation with substitution
func TestMakeEnvVarWithSubstitution(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-evars-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Set up an environment variable
  os.Setenv("TEST_ROOT", tmpDir)
  defer os.Unsetenv("TEST_ROOT")

  // Create a subdirectory
  repoPath := filepath.Join(tmpDir, "test-repo")
  if err := os.MkdirAll(repoPath, 0755); err != nil {
    t.Fatalf("Failed to create repo path: %v", err)
  }

  // Create walker
  walker, err := internal.NewGitTreeWalker([]string{"$TEST_ROOT"}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  // Test makeEnvVarWithSubstitution
  result := makeEnvVarWithSubstitution(repoPath, "$TEST_ROOT", walker)

  // Should generate an export statement with variable substitution
  if !strings.HasPrefix(result, "export ") {
    t.Errorf("Expected result to start with 'export ', got: %s", result)
  }

  if !strings.Contains(result, "$TEST_ROOT") {
    t.Errorf("Expected result to contain '$TEST_ROOT', got: %s", result)
  }

  if !strings.Contains(result, "test_repo") {
    t.Errorf("Expected result to contain 'test_repo', got: %s", result)
  }
}

// TestMakeEnvVarWithSubstitution_DirectPath tests with a direct path (not an env var)
func TestMakeEnvVarWithSubstitution_DirectPath(t *testing.T) {
  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-evars-test-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a subdirectory
  repoPath := filepath.Join(tmpDir, "direct-repo")
  if err := os.MkdirAll(repoPath, 0755); err != nil {
    t.Fatalf("Failed to create repo path: %v", err)
  }

  // Create walker with direct path
  walker, err := internal.NewGitTreeWalker([]string{tmpDir}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  // Test makeEnvVarWithSubstitution with direct path
  result := makeEnvVarWithSubstitution(repoPath, tmpDir, walker)

  // Should generate an export statement
  if !strings.HasPrefix(result, "export ") {
    t.Errorf("Expected result to start with 'export ', got: %s", result)
  }
}

// TestGitEvars_Integration tests the full git-evars workflow
func TestGitEvars_Integration(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory structure with git repos
  tmpDir, err := os.MkdirTemp("", "git-evars-integration-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create first repository
  repo1Path := filepath.Join(tmpDir, "test-repo-1")
  if err := os.MkdirAll(repo1Path, 0755); err != nil {
    t.Fatalf("Failed to create repo1: %v", err)
  }

  cmd := exec.Command("git", "init", repo1Path)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init repo1: %v", err)
  }

  // Create second repository
  repo2Path := filepath.Join(tmpDir, "test-repo-2")
  if err := os.MkdirAll(repo2Path, 0755); err != nil {
    t.Fatalf("Failed to create repo2: %v", err)
  }

  cmd = exec.Command("git", "init", repo2Path)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init repo2: %v", err)
  }

  // Set up environment variable
  os.Setenv("TEST_EVARS_ROOT", tmpDir)
  defer os.Unsetenv("TEST_EVARS_ROOT")

  // Save original os.Args and stdout
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Capture stdout
  oldStdout := os.Stdout
  r, w, _ := os.Pipe()
  os.Stdout = w

  // Set up command line arguments
  os.Args = []string{"git-evars", "-s", "$TEST_EVARS_ROOT"}

  // Run the main function
  main()

  // Restore stdout and read output
  w.Close()
  os.Stdout = oldStdout

  output := make([]byte, 4096)
  n, _ := r.Read(output)
  outputStr := string(output[:n])

  // Verify output contains export statements
  if !strings.Contains(outputStr, "export ") {
    t.Errorf("Expected output to contain 'export ', got: %s", outputStr)
  }

  // Verify both repositories are included
  if !strings.Contains(outputStr, "test_repo_1") && !strings.Contains(outputStr, "repo_1") {
    t.Errorf("Expected output to contain reference to test-repo-1, got: %s", outputStr)
  }

  if !strings.Contains(outputStr, "test_repo_2") && !strings.Contains(outputStr, "repo_2") {
    t.Errorf("Expected output to contain reference to test-repo-2, got: %s", outputStr)
  }
}

// TestGitEvars_WithZoweeOption tests the --zowee optimization option
func TestGitEvars_WithZoweeOption(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory structure with git repos
  tmpDir, err := os.MkdirTemp("", "git-evars-zowee-*")
  if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
  }
  defer os.RemoveAll(tmpDir)

  // Create a repository
  repoPath := filepath.Join(tmpDir, "zowee-test")
  if err := os.MkdirAll(repoPath, 0755); err != nil {
    t.Fatalf("Failed to create repo: %v", err)
  }

  cmd := exec.Command("git", "init", repoPath)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init repo: %v", err)
  }

  // Save original os.Args and stdout
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Capture stdout
  oldStdout := os.Stdout
  r, w, _ := os.Pipe()
  os.Stdout = w

  // Set up command line arguments with --zowee flag
  os.Args = []string{"git-evars", "-s", "--zowee", tmpDir}

  // Run the main function
  main()

  // Restore stdout and read output
  w.Close()
  os.Stdout = oldStdout

  output := make([]byte, 4096)
  n, _ := r.Read(output)
  outputStr := string(output[:n])

  // Verify output contains export statement
  // The zowee optimizer should still produce valid output
  if len(strings.TrimSpace(outputStr)) > 0 {
    if !strings.Contains(outputStr, "export ") {
      t.Errorf("Expected zowee output to contain 'export ', got: %s", outputStr)
    }
  }
}

// TestGitEvars_NoRepositories tests behavior when no repositories are found
func TestGitEvars_NoRepositories(t *testing.T) {
  // Create a temporary directory with no git repos
  tmpDir, err := os.MkdirTemp("", "git-evars-empty-*")
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
  os.Args = []string{"git-evars", "-s", tmpDir}

  // Run the main function
  main()

  // Restore stdout and read output
  w.Close()
  os.Stdout = oldStdout

  output := make([]byte, 4096)
  n, _ := r.Read(output)
  outputStr := string(output[:n])

  // Should produce no output or minimal output
  trimmedOutput := strings.TrimSpace(outputStr)
  if len(trimmedOutput) > 0 {
    t.Logf("Output with no repositories: %s", trimmedOutput)
  }
}

// TestGitEvars_WithIgnoreFile tests that .ignore files are respected
func TestGitEvars_WithIgnoreFile(t *testing.T) {
  if testing.Short() {
    t.Skip("Skipping integration test in short mode")
  }

  // Create a temporary directory
  tmpDir, err := os.MkdirTemp("", "git-evars-ignore-*")
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

  // Create a normal repository
  normalRepo := filepath.Join(tmpDir, "normal-repo")
  if err := os.MkdirAll(normalRepo, 0755); err != nil {
    t.Fatalf("Failed to create normal repo: %v", err)
  }

  cmd = exec.Command("git", "init", normalRepo)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    t.Fatalf("Failed to init normal repo: %v", err)
  }

  // Save original os.Args and stdout
  oldArgs := os.Args
  defer func() { os.Args = oldArgs }()

  // Capture stdout
  oldStdout := os.Stdout
  r, w, _ := os.Pipe()
  os.Stdout = w

  // Set up command line arguments
  os.Args = []string{"git-evars", "-s", tmpDir}

  // Run the main function
  main()

  // Restore stdout and read output
  w.Close()
  os.Stdout = oldStdout

  output := make([]byte, 4096)
  n, _ := r.Read(output)
  outputStr := string(output[:n])

  // Should only include normal-repo, not ignored-repo
  if strings.Contains(outputStr, "ignored_repo") {
    t.Errorf("Expected output to not contain ignored-repo, got: %s", outputStr)
  }

  if len(strings.TrimSpace(outputStr)) > 0 {
    if !strings.Contains(outputStr, "normal_repo") && !strings.Contains(outputStr, "normal") {
      t.Errorf("Expected output to contain normal-repo, got: %s", outputStr)
    }
  }
}
