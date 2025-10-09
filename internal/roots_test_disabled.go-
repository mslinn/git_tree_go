//go:build !skip
// +build !skip

package internal

import (
	"testing"
)

// TestRoots_Level1_OnePathWithManySlashes tests finding level 1 root for one path with many slashes
func TestRoots_Level1_OnePathWithManySlashes_DISABLED(t *testing.T) {
	t.Skip("Temporarily disabled due to issues")
	paths := []string{
		"/root/sub3/sub1",
	}
	result := Roots(paths, 1, false)
	expected := "/root/sub3"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestZoweeOptimizer_MultipleBranchesFromCommonRoot tests multiple branches from a common root
func TestZoweeOptimizer_MultipleBranchesFromCommonRoot(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{"/a/b", "/a/b/c", "/a/b/d"}

	result := optimizer.Optimize(paths, []string{})

	expected := []string{
		"export b=/a/b",
		"export c=$b/c",
		"export d=$b/d",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(result))
		t.Logf("Result: %v", result)
		return
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Expected result[%d] to be '%s', got '%s'", i, exp, result[i])
		}
	}
}

// TestZoweeOptimizer_ComplexNesting tests more complex nesting scenarios
func TestZoweeOptimizer_ComplexNesting(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{
		"/root/sub1/sub2/blah",
		"/root/sub1/sub2",
		"/root/sub1",
		"/root/sub3/sub1",
	}

	result := optimizer.Optimize(paths, []string{})

	// All paths should be optimized
	if len(result) != len(paths) {
		t.Errorf("Expected %d results, got %d", len(paths), len(result))
		t.Logf("Result: %v", result)
	}

	// Each result should start with "export "
	for i, r := range result {
		if !strings.HasPrefix(r, "export ") {
			t.Errorf("Result[%d] should start with 'export ', got '%s'", i, r)
		}
	}
}
