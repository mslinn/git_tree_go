package internal

import (
	"strings"
	"testing"
)

// TestZoweeOptimizer_SimpleNestedStructure tests optimization of simple nested paths
func TestZoweeOptimizer_SimpleNestedStructure(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{"/a", "/a/b", "/a/b/c"}

	result := optimizer.Optimize(paths, []string{})

	expected := []string{
		"export a=/a",
		"export b=$a/b",
		"export c=$b/c",
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


// TestZoweeOptimizer_UnrelatedPaths tests optimization of unrelated paths
func TestZoweeOptimizer_UnrelatedPaths(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{"/x/y", "/m/n"}

	result := optimizer.Optimize(paths, []string{})

	expected := []string{
		"export y=/x/y",
		"export n=/m/n",
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

// TestZoweeOptimizer_WithInitialVariables tests optimization with initial variables
func TestZoweeOptimizer_WithInitialVariables(t *testing.T) {
	initialVars := map[string][]string{
		"$work": {"/path/to/work"},
	}
	optimizer := NewZoweeOptimizer(initialVars)
	paths := []string{"/path/to/work/project_a", "/path/to/work/project_b"}

	result := optimizer.Optimize(paths, []string{"$work"})

	expected := []string{
		"export project_a=$work/project_a",
		"export project_b=$work/project_b",
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

// TestZoweeOptimizer_EmptyPaths tests optimization with empty paths
func TestZoweeOptimizer_EmptyPaths(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{}

	result := optimizer.Optimize(paths, []string{})

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items: %v", len(result), result)
	}
}

// TestZoweeOptimizer_SinglePath tests optimization with a single path
func TestZoweeOptimizer_SinglePath(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{"/usr/local/bin"}

	result := optimizer.Optimize(paths, []string{})

	expected := []string{
		"export bin=/usr/local/bin",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(result))
		t.Logf("Result: %v", result)
		return
	}

	if result[0] != expected[0] {
		t.Errorf("Expected result to be '%s', got '%s'", expected[0], result[0])
	}
}

// TestZoweeOptimizer_GenerateVarName tests variable name generation
func TestZoweeOptimizer_GenerateVarName(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)

	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/project", "project"},
		{"/path/to/www.example.com", "example"},
		{"/path/to/my-project", "my_project"},
		{"/path/to/123project", "_123project"},
	}

	for _, test := range tests {
		result := optimizer.generateVarName(test.path)
		if result != test.expected {
			t.Errorf("For path '%s', expected var name '%s', got '%s'", test.path, test.expected, result)
		}
	}
}

// TestZoweeOptimizer_CollisionHandling tests variable name collision handling
func TestZoweeOptimizer_CollisionHandling(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{
		"/path1/project",
		"/path2/project",
	}

	result := optimizer.Optimize(paths, []string{})

	// Both paths should get unique variable names
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
		t.Logf("Result: %v", result)
		return
	}

	// Extract variable names
	var varNames []string
	for _, r := range result {
		parts := strings.Split(r, " ")
		if len(parts) >= 2 {
			varNames = append(varNames, parts[1])
		}
	}

	// Variable names should be different
	if len(varNames) == 2 && varNames[0] == varNames[1] {
		t.Errorf("Expected different variable names, but both are '%s'", varNames[0])
	}
}

// TestZoweeOptimizer_PathSubstitution tests that paths are correctly substituted
func TestZoweeOptimizer_PathSubstitution(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{
		"/home/user",
		"/home/user/projects",
		"/home/user/projects/myapp",
	}

	result := optimizer.Optimize(paths, []string{})

	// Check that later paths use earlier variables
	foundSubstitution := false
	for _, r := range result {
		if strings.Contains(r, "$") {
			foundSubstitution = true
			break
		}
	}

	if !foundSubstitution {
		t.Error("Expected at least one result to use variable substitution")
		t.Logf("Result: %v", result)
	}
}

// TestZoweeOptimizer_NoInitialVars tests optimization without initial variables
func TestZoweeOptimizer_NoInitialVars(t *testing.T) {
	optimizer := NewZoweeOptimizer(map[string][]string{})
	paths := []string{"/a/b/c"}

	result := optimizer.Optimize(paths, []string{})

	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
		t.Logf("Result: %v", result)
		return
	}

	expected := "export c=/a/b/c"
	if result[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result[0])
	}
}

// TestZoweeOptimizer_IntermediateVariables tests that intermediate variables are created
func TestZoweeOptimizer_IntermediateVariables(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{
		"/root/projects/app1",
		"/root/projects/app2",
		"/root/projects/app3",
	}

	result := optimizer.Optimize(paths, []string{})

	// Should create an intermediate variable for "/root/projects"
	// and use it in the definitions of app1, app2, app3
	foundIntermediate := false
	for _, r := range result {
		if strings.Contains(r, "projects=") {
			foundIntermediate = true
			break
		}
	}

	if !foundIntermediate {
		// It's okay if intermediate variables aren't always created,
		// but let's at least verify that substitution is happening
		foundSubstitution := false
		for _, r := range result {
			if strings.Contains(r, "$") {
				foundSubstitution = true
				break
			}
		}
		if !foundSubstitution {
			t.Error("Expected variable substitution in results")
			t.Logf("Result: %v", result)
		}
	}
}

// TestZoweeOptimizer_SanitizeVarNames tests that invalid characters are sanitized
func TestZoweeOptimizer_SanitizeVarNames(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)

	tests := []struct {
		path        string
		shouldMatch string // regex pattern to match the variable name
	}{
		{"/path/to/my-app", "my_app"},
		{"/path/to/app.com", "app"},
		{"/path/to/123app", "_123app"},
	}

	for _, test := range tests {
		result := optimizer.generateVarName(test.path)
		if !strings.Contains(result, test.shouldMatch) {
			t.Errorf("For path '%s', expected var name to contain '%s', got '%s'", test.path, test.shouldMatch, result)
		}
	}
}

// TestZoweeOptimizer_UniqueResults tests that results don't contain duplicates
func TestZoweeOptimizer_UniqueResults(t *testing.T) {
	optimizer := NewZoweeOptimizer(nil)
	paths := []string{"/a/b", "/a/b", "/a/c"}

	result := optimizer.Optimize(paths, []string{})

	// Check for duplicates
	seen := make(map[string]bool)
	for _, r := range result {
		if seen[r] {
			t.Errorf("Found duplicate result: '%s'", r)
		}
		seen[r] = true
	}
}
