package internal

import (
	"fmt"
	"os/exec"
)

// ExecResult captures the result of a command execution.
type ExecResult struct {
	Stdout string
	Stderr string
	Status int
}

// Execution represents a single command execution.
type Execution struct {
	Command    string
	Dir        string
	ExecResult *ExecResult
	Error      error
}

// Contains checks if the execution output contains the specified text.
func (e *Execution) Contains(text string) bool {
	if e.ExecResult == nil {
		return false
	}
	return containsString(e.ExecResult.Stdout, text) ||
		containsString(e.ExecResult.Stderr, text)
}

// UserMessage represents a message sent to the user.
type UserMessage struct {
	Message string
	Color   string
}

// Task remembers all commands that were executed and their results.
// Also remembers user output.
type Task struct {
	History []interface{} // Can contain *Execution or *UserMessage
}

// NewTask creates a new Task instance.
func NewTask() *Task {
	return &Task{
		History: make([]interface{}, 0),
	}
}

// MostRecentUserMessage returns the most recent UserMessage in history.
func (t *Task) MostRecentUserMessage() *UserMessage {
	for i := len(t.History) - 1; i >= 0; i-- {
		if msg, ok := t.History[i].(*UserMessage); ok {
			return msg
		}
	}
	return nil
}

// ExecResultExecution returns the most recent Execution in history.
func (t *Task) ExecResultExecution() *Execution {
	for i := len(t.History) - 1; i >= 0; i-- {
		if exec, ok := t.History[i].(*Execution); ok {
			return exec
		}
	}
	return nil
}

// MessageUser adds a user message to the history if verbosity allows.
func (t *Task) MessageUser(logLevel int, message string, color string) {
	if GetVerbosity() < logLevel {
		return
	}

	Log(logLevel, message, color)
	t.History = append(t.History, &UserMessage{
		Message: message,
		Color:   color,
	})
}

// Perform executes a command and records the execution in history.
func (t *Task) Perform(command, dir string) {
	execution := &Execution{
		Command: command,
		Dir:     dir,
	}

	result, err := RunCommand(command, dir)
	execution.ExecResult = result
	execution.Error = err

	t.History = append(t.History, execution)
}

// RunCommand executes a shell command in a specified directory.
func RunCommand(command, dir string) (*ExecResult, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir

	stdout, err := cmd.Output()
	result := &ExecResult{
		Stdout: string(stdout),
		Status: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Stderr = string(exitErr.Stderr)
			result.Status = exitErr.ExitCode()
		}
		return result, err
	}

	return result, nil
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		fmt.Sprint(s) != "" &&
		fmt.Sprint(substr) != "" &&
		(s == substr || len(s) >= len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
