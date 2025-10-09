package internal

import (
	"testing"
)

// TestTrimToLevel tests trimming paths to a specific level
func TestTrimToLevel(t *testing.T) {
	paths := []string{
		"/root/sub3/sub1",
	}

	// Test level 1
	result := TrimToLevel(paths, 1)
	expected := []string{"/root"}
	if len(result) != len(expected) || result[0] != expected[0] {
		t.Errorf("TrimToLevel(level=1): expected %v, got %v", expected, result)
	}

	// Test level 2
	result = TrimToLevel(paths, 2)
	expected = []string{"/root/sub3"}
	if len(result) != len(expected) || result[0] != expected[0] {
		t.Errorf("TrimToLevel(level=2): expected %v, got %v", expected, result)
	}

	// Test level 3
	result = TrimToLevel(paths, 3)
	expected = []string{"/root/sub3/sub1"}
	if len(result) != len(expected) || result[0] != expected[0] {
		t.Errorf("TrimToLevel(level=3): expected %v, got %v", expected, result)
	}

	// Test level beyond path depth
	result = TrimToLevel(paths, 4)
	expected = []string{"/root/sub3/sub1"}
	if len(result) != len(expected) || result[0] != expected[0] {
		t.Errorf("TrimToLevel(level=4): expected %v, got %v", expected, result)
	}
}

// TestRoots_Level1_OnePathWithOneSlash tests finding level 1 root for one path with 1 slash
func TestRoots_Level1_OnePathWithOneSlash(t *testing.T) {
	paths := []string{
		"/root",
	}

	// Without allow_root_match
	result := Roots(paths, 1, false)
	expected := ""
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// With allow_root_match
	result = Roots(paths, 1, true)
	expected = "/"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestRoots_TwoPaths tests finding roots for two paths
func TestRoots_TwoPaths(t *testing.T) {
	paths := []string{
		"/root/sub1/sub2/blah",
		"/root/sub3/sub1",
	}
	result := Roots(paths, 1, false)
	expected := "/root"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestRoots_MultiplePaths tests finding level 1 root for multiple paths
func TestRoots_MultiplePaths(t *testing.T) {
	paths := []string{
		"/root/sub1/sub2/blah",
		"/root/sub1/sub2",
		"/root/sub1",
		"/root/sub3/sub1",
	}
	result := Roots(paths, 1, false)
	expected := "/root"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestCommonPrefix_EmptyPaths tests common prefix with empty paths
func TestCommonPrefix_EmptyPaths(t *testing.T) {
	paths := []string{}
	result := CommonPrefix(paths, false)
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestCommonPrefix_SinglePath tests common prefix with single path
func TestCommonPrefix_SinglePath(t *testing.T) {
	paths := []string{"/root/sub1/sub2"}
	result := CommonPrefix(paths, false)
	expected := "/root/sub1"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestCommonPrefix_SinglePathAtRoot tests common prefix with single path at root
func TestCommonPrefix_SinglePathAtRoot(t *testing.T) {
	paths := []string{"/root"}

	// Without allow_root_match
	result := CommonPrefix(paths, false)
	expected := ""
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// With allow_root_match
	result = CommonPrefix(paths, true)
	expected = "/"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestCommonPrefix_MultiplePaths tests common prefix with multiple paths
func TestCommonPrefix_MultiplePaths(t *testing.T) {
	paths := []string{
		"/root/sub1/sub2",
		"/root/sub1/sub3",
		"/root/sub1/sub4",
	}
	result := CommonPrefix(paths, false)
	expected := "/root/sub1"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestCommonPrefix_DifferentRoots tests common prefix with different roots
func TestCommonPrefix_DifferentRoots(t *testing.T) {
	paths := []string{
		"/root1/sub1",
		"/root2/sub1",
	}
	result := CommonPrefix(paths, false)
	expected := ""

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestCommonPrefix_SameRoot tests common prefix with same root
func TestCommonPrefix_SameRoot(t *testing.T) {
	paths := []string{
		"/root/sub1/sub2",
		"/root/sub3/sub4",
	}
	result := CommonPrefix(paths, false)
	expected := "/root"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestTrimToLevel_MultiplePaths tests trimming multiple paths
func TestTrimToLevel_MultiplePaths(t *testing.T) {
	paths := []string{
		"/root/sub1/sub2/blah",
		"/root/sub1/sub2",
		"/root/sub1",
		"/root/sub3/sub1",
	}

	result := TrimToLevel(paths, 1)
	expected := []string{"/root"}
	if len(result) != len(expected) || result[0] != expected[0] {
		t.Errorf("TrimToLevel(level=1): expected %v, got %v", expected, result)
	}

	result = TrimToLevel(paths, 2)
	expected = []string{"/root/sub1", "/root/sub3"}
	if len(result) != len(expected) {
		t.Errorf("TrimToLevel(level=2): expected length %d, got %d", len(expected), len(result))
	}
	for i, exp := range expected {
		if i >= len(result) || result[i] != exp {
			t.Errorf("TrimToLevel(level=2): expected result[%d] to be '%s', got '%s'", i, exp, result[i])
		}
	}

	result = TrimToLevel(paths, 3)
	expected = []string{"/root/sub1/sub1", "/root/sub1/sub2", "/root/sub3/sub1"}
	if len(result) != len(expected) {
		t.Errorf("TrimToLevel(level=3): expected length %d, got %d", len(expected), len(result))
	}
}

// TestTrimToLevel_Uniqueness tests that trimmed paths are unique
func TestTrimToLevel_Uniqueness(t *testing.T) {
	paths := []string{
		"/root/sub1/a",
		"/root/sub1/b",
		"/root/sub1/c",
	}

	result := TrimToLevel(paths, 2)
	expected := []string{"/root/sub1"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d unique paths, got %d", len(expected), len(result))
	}

	if result[0] != expected[0] {
		t.Errorf("Expected '%s', got '%s'", expected[0], result[0])
	}
}

// TestEnsureEndsWith tests ensuring a string ends with a suffix
func TestEnsureEndsWith(t *testing.T) {
	tests := []struct {
		str      string
		suffix   string
		expected string
	}{
		{"/path/to/dir", "/", "/path/to/dir/"},
		{"/path/to/dir/", "/", "/path/to/dir/"},
		{"hello", "world", "helloworld"},
		{"helloworld", "world", "helloworld"},
	}

	for _, test := range tests {
		result := EnsureEndsWith(test.str, test.suffix)
		if result != test.expected {
			t.Errorf("EnsureEndsWith('%s', '%s'): expected '%s', got '%s'", test.str, test.suffix, test.expected, result)
		}
	}
}

// TestExpandEnv tests environment variable expansion
func TestExpandEnv(t *testing.T) {
	// Set test environment variables
	t.Setenv("TEST_VAR", "test_value")
	t.Setenv("HOME", "/home/user")

	tests := []struct {
		input    string
		expected string
	}{
		{"$TEST_VAR", "test_value"},
		{"${TEST_VAR}", "test_value"},
		{"%TEST_VAR%", "test_value"},
		{"$HOME/documents", "/home/user/documents"},
		{"${HOME}/documents", "/home/user/documents"},
		{"%HOME%/documents", "/home/user/documents"},
		{"no variables", "no variables"},
		{"$UNDEFINED_VAR", ""}, // Undefined variables expand to empty string
	}

	for _, test := range tests {
		result := ExpandEnv(test.input)
		if result != test.expected {
			t.Errorf("ExpandEnv('%s'): expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}

// TestTrimToLevel_EmptyPaths tests trimming empty paths
func TestTrimToLevel_EmptyPaths(t *testing.T) {
	paths := []string{}
	result := TrimToLevel(paths, 1)

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}
}

// TestRoots_EmptyPaths tests roots with empty paths
func TestRoots_EmptyPaths(t *testing.T) {
	paths := []string{}

	result := Roots(paths, 1, false)
	expected := ""
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	result = Roots(paths, 1, true)
	expected = "/"
	if result != expected {
		t.Errorf("Expected '%s' with allow_root_match, got '%s'", expected, result)
	}
}

// TestRoots_DeeperLevel tests finding roots at deeper levels
func TestRoots_DeeperLevel(t *testing.T) {
	paths := []string{
		"/root/sub1/sub2/sub3",
		"/root/sub1/sub2/sub4",
	}

	result := Roots(paths, 1, false)
	expected := "/root/sub1/sub2"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	result = Roots(paths, 2, false)
	expected = "/root/sub1"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestCommonPrefix_Sorted tests that common prefix works with unsorted paths
func TestCommonPrefix_Sorted(t *testing.T) {
	paths := []string{
		"/root/sub3",
		"/root/sub1",
		"/root/sub2",
	}
	result := CommonPrefix(paths, false)
	expected := "/root"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestTrimToLevel_Sorted tests that trimmed paths are sorted
func TestTrimToLevel_Sorted(t *testing.T) {
	paths := []string{
		"/root/sub3/a",
		"/root/sub1/b",
		"/root/sub2/c",
	}

	result := TrimToLevel(paths, 2)
	expected := []string{"/root/sub1", "/root/sub2", "/root/sub3"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d paths, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		if i >= len(result) || result[i] != exp {
			t.Errorf("Expected result[%d] to be '%s', got '%s'", i, exp, result[i])
		}
	}
}
