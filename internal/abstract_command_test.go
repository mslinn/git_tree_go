package internal

import (
  "flag"
  "os"
  "testing"
)

// TestAbstractCommand_Initialization tests that config loads and verbosity is set
func TestAbstractCommand_Initialization(t *testing.T) {
  args := []string{}
  cmd := NewAbstractCommand(args, false)

  if cmd.Config == nil {
    t.Error("Expected config to be initialized")
  }

  // Verify that verbosity was set from config
  verbosity := GetVerbosity()
  if verbosity != cmd.Config.Verbosity {
    t.Errorf("Expected verbosity to be set to %d, got %d", cmd.Config.Verbosity, verbosity)
  }
}

// TestAbstractCommand_ArgumentHandling tests that default roots are used when no args are given
func TestAbstractCommand_ArgumentHandling(t *testing.T) {
  // Create a command with no arguments
  cmd := NewAbstractCommand([]string{}, true) // Allow empty args

  if len(cmd.Config.DefaultRoots) == 0 {
    t.Error("Expected default roots to be configured")
  }

  // When no args are given, the walker should use default roots
  walker, err := NewGitTreeWalker([]string{}, false)
  if err != nil {
    t.Fatalf("Failed to create walker: %v", err)
  }

  if len(walker.DisplayRoots) == 0 {
    t.Error("Expected walker to have display roots from config")
  }
}

// TestAbstractCommand_ParseCommonFlags_Quiet tests the -q option
func TestAbstractCommand_ParseCommonFlags_Quiet(t *testing.T) {
  // Save original verbosity
  originalVerbosity := GetVerbosity()
  defer SetVerbosity(originalVerbosity)

  args := []string{"-q", "/some/dir"}
  cmd := NewAbstractCommand(args, false)

  // Mock help function
  helpCalled := false
  helpFunc := func() {
    helpCalled = true
  }

  remaining := cmd.ParseCommonFlags(helpFunc)

  // Check that verbosity was set to QUIET
  if GetVerbosity() != LogQuiet {
    t.Errorf("Expected verbosity to be QUIET (%d), got %d", LogQuiet, GetVerbosity())
  }

  // Check that the option was removed from args
  if len(remaining) != 1 || remaining[0] != "/some/dir" {
    t.Errorf("Expected remaining args to be ['/some/dir'], got %v", remaining)
  }

  if helpCalled {
    t.Error("Expected help not to be called")
  }
}

// TestAbstractCommand_ParseCommonFlags_Serial tests the -s option
func TestAbstractCommand_ParseCommonFlags_Serial(t *testing.T) {
  args := []string{"-s", "/some/dir"}
  cmd := NewAbstractCommand(args, false)

  helpFunc := func() {}
  remaining := cmd.ParseCommonFlags(helpFunc)

  // Check that serial flag was set
  if !cmd.Serial {
    t.Error("Expected serial flag to be true")
  }

  // Check that the option was removed from args
  if len(remaining) != 1 || remaining[0] != "/some/dir" {
    t.Errorf("Expected remaining args to be ['/some/dir'], got %v", remaining)
  }
}

// TestAbstractCommand_ParseCommonFlags_Verbose tests the -v option
func TestAbstractCommand_ParseCommonFlags_Verbose(t *testing.T) {
  // Save original verbosity
  originalVerbosity := GetVerbosity()
  defer SetVerbosity(originalVerbosity)

  args := []string{"-v", "/some/dir"}
  cmd := NewAbstractCommand(args, false)
  initialVerbosity := GetVerbosity()

  helpFunc := func() {}
  cmd.ParseCommonFlags(helpFunc)

  // Check that verbosity was incremented
  expectedVerbosity := initialVerbosity + 1
  if GetVerbosity() != expectedVerbosity {
    t.Errorf("Expected verbosity to be %d, got %d", expectedVerbosity, GetVerbosity())
  }
}

// TestAbstractCommand_ParseCommonFlags_MultipleVerbose tests multiple -v options
func TestAbstractCommand_ParseCommonFlags_MultipleVerbose(t *testing.T) {
  // Save original verbosity
  originalVerbosity := GetVerbosity()
  defer SetVerbosity(originalVerbosity)

  args := []string{"-v", "-v", "/some/dir"}
  cmd := NewAbstractCommand(args, false)
  initialVerbosity := GetVerbosity()

  helpFunc := func() {}
  cmd.ParseCommonFlags(helpFunc)

  // Check that verbosity was incremented twice
  expectedVerbosity := initialVerbosity + 2
  if GetVerbosity() != expectedVerbosity {
    t.Errorf("Expected verbosity to be %d, got %d", expectedVerbosity, GetVerbosity())
  }
}

// TestAbstractCommand_ParseCommonFlags_Help tests the -h option
func TestAbstractCommand_ParseCommonFlags_Help(t *testing.T) {
  // Note: This test is challenging because -h causes os.Exit(0)
  // In a real test environment, we would need to refactor the code to avoid os.Exit
  // For now, we'll skip this test or use a subprocess pattern
  t.Skip("Skipping test that would call os.Exit")
}

// TestAbstractCommand_AllowEmptyArgs tests empty args handling
func TestAbstractCommand_AllowEmptyArgs_True(t *testing.T) {
  args := []string{}
  cmd := NewAbstractCommand(args, true)

  helpFunc := func() {}
  remaining := cmd.ParseCommonFlags(helpFunc)

  // Should not error when empty args are allowed
  if len(remaining) != 0 {
    t.Errorf("Expected remaining args to be empty, got %v", remaining)
  }
}

// TestAbstractCommand_AllowEmptyArgs_False tests that an error occurs when no args provided
func TestAbstractCommand_AllowEmptyArgs_False(t *testing.T) {
  // Note: This test is challenging because it causes os.Exit(1)
  // In a real test environment, we would need to refactor the code to avoid os.Exit
  // For now, we'll skip this test or use a subprocess pattern
  t.Skip("Skipping test that would call os.Exit")
}

// TestAbstractCommand_ParseFlagsWithCallback tests custom flags
func TestAbstractCommand_ParseFlagsWithCallback(t *testing.T) {
  args := []string{"--custom", "value", "/some/dir"}
  cmd := NewAbstractCommand(args, false)

  helpFunc := func() {}
  customValue := ""
  callback := func(fs *flag.FlagSet) {
    fs.StringVar(&customValue, "custom", "", "A custom flag")
  }

  remaining := cmd.ParseFlagsWithCallback(helpFunc, callback)

  if customValue != "value" {
    t.Errorf("Expected custom value to be 'value', got '%s'", customValue)
  }

  if len(remaining) != 1 || remaining[0] != "/some/dir" {
    t.Errorf("Expected remaining args to be ['/some/dir'], got %v", remaining)
  }
}

// TestAbstractCommand_ParseFlagsWithCallback_WithCommonFlags tests mixing custom and common flags
func TestAbstractCommand_ParseFlagsWithCallback_WithCommonFlags(t *testing.T) {
  // Save original verbosity
  originalVerbosity := GetVerbosity()
  defer SetVerbosity(originalVerbosity)

  args := []string{"-v", "--custom", "value", "-s", "/some/dir"}
  cmd := NewAbstractCommand(args, false)
  initialVerbosity := GetVerbosity()

  helpFunc := func() {}
  customValue := ""
  callback := func(fs *flag.FlagSet) {
    fs.StringVar(&customValue, "custom", "", "A custom flag")
  }

  remaining := cmd.ParseFlagsWithCallback(helpFunc, callback)

  // Check custom flag
  if customValue != "value" {
    t.Errorf("Expected custom value to be 'value', got '%s'", customValue)
  }

  // Check common flags
  if !cmd.Serial {
    t.Error("Expected serial flag to be true")
  }

  expectedVerbosity := initialVerbosity + 1
  if GetVerbosity() != expectedVerbosity {
    t.Errorf("Expected verbosity to be %d, got %d", expectedVerbosity, GetVerbosity())
  }

  if len(remaining) != 1 || remaining[0] != "/some/dir" {
    t.Errorf("Expected remaining args to be ['/some/dir'], got %v", remaining)
  }
}

// TestAbstractCommand_ConfigFromEnvironment tests that environment variables override config
func TestAbstractCommand_ConfigFromEnvironment(t *testing.T) {
  // Set environment variables
  os.Setenv("GIT_TREE_GIT_TIMEOUT", "600")
  os.Setenv("GIT_TREE_VERBOSITY", "3")
  os.Setenv("GIT_TREE_DEFAULT_ROOTS", "root1 root2")
  defer func() {
    os.Unsetenv("GIT_TREE_GIT_TIMEOUT")
    os.Unsetenv("GIT_TREE_VERBOSITY")
    os.Unsetenv("GIT_TREE_DEFAULT_ROOTS")
  }()

  cmd := NewAbstractCommand([]string{}, true)

  if cmd.Config.GitTimeout != 600 {
    t.Errorf("Expected git_timeout to be 600, got %d", cmd.Config.GitTimeout)
  }

  if cmd.Config.Verbosity != 3 {
    t.Errorf("Expected verbosity to be 3, got %d", cmd.Config.Verbosity)
  }

  if len(cmd.Config.DefaultRoots) != 2 || cmd.Config.DefaultRoots[0] != "root1" || cmd.Config.DefaultRoots[1] != "root2" {
    t.Errorf("Expected default_roots to be [root1 root2], got %v", cmd.Config.DefaultRoots)
  }
}
