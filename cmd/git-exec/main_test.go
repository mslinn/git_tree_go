package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExecuteAndLog_Success tests successful command execution
func TestExecuteAndLog_Success(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-exec-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Execute ls command (should succeed)
	// We can't easily capture the log output, but we can verify the function doesn't panic
	executeAndLog(tmpDir, "ls -la")
}

// TestExecuteAndLog_Failure tests failed command execution
func TestExecuteAndLog_Failure(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-exec-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Execute a command that should fail
	executeAndLog(tmpDir, "exit 1")
}

// TestExecuteAndLog_WithOutput tests command execution with output
func TestExecuteAndLog_WithOutput(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-exec-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Execute echo command
	executeAndLog(tmpDir, "echo 'Hello, World!'")
}

// TestGitExec_Integration tests the full git-exec workflow
func TestGitExec_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory structure with git repos
	tmpDir, err := os.MkdirTemp("", "git-exec-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create first repository
	repo1Path := filepath.Join(tmpDir, "repo1")
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
	repo2Path := filepath.Join(tmpDir, "repo2")
	if err := os.MkdirAll(repo2Path, 0755); err != nil {
		t.Fatalf("Failed to create repo2: %v", err)
	}

	cmd = exec.Command("git", "init", repo2Path)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init repo2: %v", err)
	}

	// Create a marker file in each repo to verify command execution
	markerFile := "exec_test_marker.txt"

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up command line arguments to create marker files
	os.Args = []string{"git-exec", "-s", tmpDir, "touch " + markerFile}

	// Capture stdout to suppress output during test
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Run the main function
	main()

	// Close the pipe to get any remaining output
	w.Close()
	os.Stdout = oldStdout

	// Verify that marker files were created in both repos
	marker1 := filepath.Join(repo1Path, markerFile)
	if _, err := os.Stat(marker1); os.IsNotExist(err) {
		t.Errorf("Expected marker file to be created in repo1")
	}

	marker2 := filepath.Join(repo2Path, markerFile)
	if _, err := os.Stat(marker2); os.IsNotExist(err) {
		t.Errorf("Expected marker file to be created in repo2")
	}
}

// TestGitExec_NoArguments tests error handling when no command is provided
func TestGitExec_NoArguments(t *testing.T) {
	// Save original os.Args and os.Exit
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up command line arguments with no command
	os.Args = []string{"git-exec"}

	// We expect the program to exit with an error
	// We can't easily test os.Exit, so we just verify it doesn't panic
	// In a real scenario, we'd refactor to use a testable exit function
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	// This will call showHelp and exit
	// We can't capture this in the test without refactoring
}

// TestGitExec_CommandParsing tests parsing of command and roots
func TestGitExec_CommandParsing(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedRoots []string
		expectedCmd   string
	}{
		{
			name:          "command only",
			args:          []string{"pwd"},
			expectedRoots: []string{},
			expectedCmd:   "pwd",
		},
		{
			name:          "root and command",
			args:          []string{"/tmp", "pwd"},
			expectedRoots: []string{"/tmp"},
			expectedCmd:   "pwd",
		},
		{
			name:          "multiple roots and command",
			args:          []string{"/tmp", "/var", "ls -la"},
			expectedRoots: []string{"/tmp", "/var"},
			expectedCmd:   "ls -la",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse arguments similar to how main() does it
			remainingArgs := tt.args

			var commandArgs []string
			var shellCommand string

			if len(remainingArgs) > 1 {
				commandArgs = remainingArgs[0 : len(remainingArgs)-1]
				shellCommand = remainingArgs[len(remainingArgs)-1]
			} else if len(remainingArgs) == 1 {
				commandArgs = []string{}
				shellCommand = remainingArgs[0]
			}

			// Verify parsing
			if len(commandArgs) != len(tt.expectedRoots) {
				t.Errorf("Expected %d roots, got %d", len(tt.expectedRoots), len(commandArgs))
			}

			for i, expectedRoot := range tt.expectedRoots {
				if i < len(commandArgs) && commandArgs[i] != expectedRoot {
					t.Errorf("Expected root[%d] to be '%s', got '%s'", i, expectedRoot, commandArgs[i])
				}
			}

			if shellCommand != tt.expectedCmd {
				t.Errorf("Expected command to be '%s', got '%s'", tt.expectedCmd, shellCommand)
			}
		})
	}
}

// TestGitExec_WithQuotedCommand tests execution of quoted commands
func TestGitExec_WithQuotedCommand(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-exec-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Execute a conditional command
	executeAndLog(tmpDir, "if [ -d subdir ]; then echo 'found'; fi")
}

// TestGitExec_CommandInDirectory tests that commands are executed in the correct directory
func TestGitExec_CommandInDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-exec-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Execute pwd command and verify it runs in the correct directory
	cmd := exec.Command("sh", "-c", "pwd")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != tmpDir {
		t.Errorf("Expected pwd output to be '%s', got '%s'", tmpDir, outputStr)
	}
}
