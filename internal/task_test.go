package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTask_Initialization tests task initialization
func TestTask_Initialization(t *testing.T) {
	task := NewTask()

	if task == nil {
		t.Fatal("Expected task to be initialized")
	}

	if task.History == nil {
		t.Error("Expected history to be initialized")
	}

	if len(task.History) != 0 {
		t.Errorf("Expected empty history, got %d items", len(task.History))
	}
}

// TestTask_MessageUser tests adding user messages to history
func TestTask_MessageUser(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	SetVerbosity(LogNormal)

	task := NewTask()
	message := "Test user message"

	task.MessageUser(LogNormal, message, ColorGreen)

	if len(task.History) != 1 {
		t.Fatalf("Expected 1 history item, got %d", len(task.History))
	}

	userMsg, ok := task.History[0].(*UserMessage)
	if !ok {
		t.Fatal("Expected history item to be UserMessage")
	}

	if userMsg.Message != message {
		t.Errorf("Expected message to be '%s', got '%s'", message, userMsg.Message)
	}

	if userMsg.Color != ColorGreen {
		t.Errorf("Expected color to be green, got '%s'", userMsg.Color)
	}
}

// TestTask_MessageUser_WithQuietVerbosity tests that quiet verbosity suppresses messages
func TestTask_MessageUser_WithQuietVerbosity(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	SetVerbosity(LogQuiet)

	task := NewTask()
	task.MessageUser(LogNormal, "This should not be added", "")

	if len(task.History) != 0 {
		t.Errorf("Expected empty history in quiet mode, got %d items", len(task.History))
	}
}

// TestTask_MessageUser_MultipleMessages tests adding multiple messages
func TestTask_MessageUser_MultipleMessages(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	SetVerbosity(LogNormal)

	task := NewTask()
	messages := []string{"Message 1", "Message 2", "Message 3"}

	for _, msg := range messages {
		task.MessageUser(LogNormal, msg, "")
	}

	if len(task.History) != len(messages) {
		t.Errorf("Expected %d history items, got %d", len(messages), len(task.History))
	}

	for i, msg := range messages {
		userMsg, ok := task.History[i].(*UserMessage)
		if !ok {
			t.Errorf("Expected history item %d to be UserMessage", i)
			continue
		}

		if userMsg.Message != msg {
			t.Errorf("Expected message %d to be '%s', got '%s'", i, msg, userMsg.Message)
		}
	}
}

// TestTask_MostRecentUserMessage tests retrieving the most recent user message
func TestTask_MostRecentUserMessage(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	SetVerbosity(LogNormal)

	task := NewTask()

	// Add multiple messages
	task.MessageUser(LogNormal, "First message", "")
	task.MessageUser(LogNormal, "Second message", "")
	task.MessageUser(LogNormal, "Third message", "")

	recent := task.MostRecentUserMessage()

	if recent == nil {
		t.Fatal("Expected to find a recent message")
	}

	if recent.Message != "Third message" {
		t.Errorf("Expected most recent message to be 'Third message', got '%s'", recent.Message)
	}
}

// TestTask_MostRecentUserMessage_NoMessages tests when there are no messages
func TestTask_MostRecentUserMessage_NoMessages(t *testing.T) {
	task := NewTask()

	recent := task.MostRecentUserMessage()

	if recent != nil {
		t.Error("Expected nil when there are no messages")
	}
}

// TestTask_MostRecentUserMessage_MixedHistory tests retrieving message from mixed history
func TestTask_MostRecentUserMessage_MixedHistory(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	SetVerbosity(LogNormal)

	task := NewTask()

	// Add a message
	task.MessageUser(LogNormal, "User message", "")

	// Add an execution
	task.History = append(task.History, &Execution{
		Command: "test command",
		Dir:     "/tmp",
	})

	// Add another message
	task.MessageUser(LogNormal, "Later message", "")

	recent := task.MostRecentUserMessage()

	if recent == nil {
		t.Fatal("Expected to find a recent message")
	}

	if recent.Message != "Later message" {
		t.Errorf("Expected most recent message to be 'Later message', got '%s'", recent.Message)
	}
}

// TestTask_Perform tests executing a command
func TestTask_Perform(t *testing.T) {
	task := NewTask()

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Execute a simple command
	task.Perform("echo hello", tmpDir)

	if len(task.History) != 1 {
		t.Fatalf("Expected 1 history item, got %d", len(task.History))
	}

	exec, ok := task.History[0].(*Execution)
	if !ok {
		t.Fatal("Expected history item to be Execution")
	}

	if exec.Command != "echo hello" {
		t.Errorf("Expected command to be 'echo hello', got '%s'", exec.Command)
	}

	if exec.Dir != tmpDir {
		t.Errorf("Expected dir to be '%s', got '%s'", tmpDir, exec.Dir)
	}

	if exec.ExecResult == nil {
		t.Fatal("Expected exec result to be set")
	}

	if !strings.Contains(exec.ExecResult.Stdout, "hello") {
		t.Errorf("Expected stdout to contain 'hello', got '%s'", exec.ExecResult.Stdout)
	}

	if exec.Error != nil {
		t.Errorf("Expected no error, got %v", exec.Error)
	}
}

// TestTask_Perform_WithError tests executing a command that fails
func TestTask_Perform_WithError(t *testing.T) {
	task := NewTask()

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Execute a command that will fail
	task.Perform("false", tmpDir)

	if len(task.History) != 1 {
		t.Fatalf("Expected 1 history item, got %d", len(task.History))
	}

	exec, ok := task.History[0].(*Execution)
	if !ok {
		t.Fatal("Expected history item to be Execution")
	}

	if exec.Error == nil {
		t.Error("Expected an error for failed command")
	}

	if exec.ExecResult == nil {
		t.Fatal("Expected exec result to be set even on error")
	}

	if exec.ExecResult.Status == 0 {
		t.Error("Expected non-zero exit status for failed command")
	}
}

// TestTask_ExecResultExecution tests retrieving the most recent execution
func TestTask_ExecResultExecution(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	SetVerbosity(LogNormal)

	task := NewTask()

	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Add a message
	task.MessageUser(LogNormal, "Message", "")

	// Add an execution
	task.Perform("echo first", tmpDir)

	// Add another message
	task.MessageUser(LogNormal, "Another message", "")

	// Add another execution
	task.Perform("echo second", tmpDir)

	recent := task.ExecResultExecution()

	if recent == nil {
		t.Fatal("Expected to find a recent execution")
	}

	if recent.Command != "echo second" {
		t.Errorf("Expected most recent command to be 'echo second', got '%s'", recent.Command)
	}
}

// TestTask_ExecResultExecution_NoExecutions tests when there are no executions
func TestTask_ExecResultExecution_NoExecutions(t *testing.T) {
	task := NewTask()

	recent := task.ExecResultExecution()

	if recent != nil {
		t.Error("Expected nil when there are no executions")
	}
}

// TestExecution_Contains tests the Contains method
func TestExecution_Contains(t *testing.T) {
	execution := &Execution{
		Command: "test",
		Dir:     "/tmp",
		ExecResult: &ExecResult{
			Stdout: "This is stdout output",
			Stderr: "This is stderr output",
			Status: 0,
		},
	}

	// Test stdout match
	if !execution.Contains("stdout") {
		t.Error("Expected Contains to find 'stdout' in output")
	}

	// Test stderr match
	if !execution.Contains("stderr") {
		t.Error("Expected Contains to find 'stderr' in output")
	}

	// Test no match
	if execution.Contains("notfound") {
		t.Error("Expected Contains to not find 'notfound' in output")
	}
}

// TestExecution_Contains_NilResult tests Contains with nil result
func TestExecution_Contains_NilResult(t *testing.T) {
	execution := &Execution{
		Command:    "test",
		Dir:        "/tmp",
		ExecResult: nil,
	}

	if execution.Contains("anything") {
		t.Error("Expected Contains to return false when ExecResult is nil")
	}
}

// TestRunCommand tests the RunCommand function directly
func TestRunCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result, err := RunCommand("echo test", tmpDir)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if !strings.Contains(result.Stdout, "test") {
		t.Errorf("Expected stdout to contain 'test', got '%s'", result.Stdout)
	}

	if result.Status != 0 {
		t.Errorf("Expected status to be 0, got %d", result.Status)
	}
}

// TestRunCommand_WithError tests RunCommand with a failing command
func TestRunCommand_WithError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result, err := RunCommand("exit 1", tmpDir)

	if err == nil {
		t.Error("Expected an error for failed command")
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil even on error")
	}

	if result.Status == 0 {
		t.Error("Expected non-zero status for failed command")
	}
}

// TestRunCommand_InSpecificDirectory tests that command runs in specified directory
func TestRunCommand_InSpecificDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file in the temp directory
	testFile := filepath.Join(tmpDir, "testfile.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Run a command that lists files
	result, err := RunCommand("ls testfile.txt", tmpDir)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !strings.Contains(result.Stdout, "testfile.txt") {
		t.Error("Expected to find testfile.txt, command did not run in correct directory")
	}
}

// TestContainsString tests the containsString helper function
func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "hello", "hello", true},
		{"substring at start", "hello world", "hello", true},
		{"substring at end", "hello world", "world", true},
		{"substring in middle", "hello world", "lo wo", true},
		{"not found", "hello world", "xyz", false},
		{"empty substring", "hello", "", false},
		{"empty string", "", "hello", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("containsString(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

// TestTask_ComplexWorkflow tests a complex workflow with mixed operations
func TestTask_ComplexWorkflow(t *testing.T) {
	// Save original verbosity
	originalVerbosity := GetVerbosity()
	defer SetVerbosity(originalVerbosity)

	SetVerbosity(LogNormal)

	task := NewTask()

	tmpDir, err := os.MkdirTemp("", "git-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Complex workflow
	task.MessageUser(LogNormal, "Starting task", ColorGreen)
	task.Perform("echo step1", tmpDir)
	task.MessageUser(LogNormal, "Step 1 complete", ColorGreen)
	task.Perform("echo step2", tmpDir)
	task.MessageUser(LogNormal, "Task complete", ColorGreen)

	// Verify history
	if len(task.History) != 5 {
		t.Errorf("Expected 5 history items, got %d", len(task.History))
	}

	// Verify most recent message
	recentMsg := task.MostRecentUserMessage()
	if recentMsg == nil || recentMsg.Message != "Task complete" {
		t.Error("Expected most recent message to be 'Task complete'")
	}

	// Verify most recent execution
	recentExec := task.ExecResultExecution()
	if recentExec == nil || recentExec.Command != "echo step2" {
		t.Error("Expected most recent execution to be 'echo step2'")
	}
}
