package internal

import (
  "fmt"
  "path/filepath"
  "regexp"
  "sort"
  "strings"
)

// ZoweeOptimizer optimizes environment variable definitions for git-evars.
type ZoweeOptimizer struct {
  definedVars      map[string]string
  intermediateVars map[string]string
}

// NewZoweeOptimizer creates a new ZoweeOptimizer.
func NewZoweeOptimizer(initialVars map[string][]string) *ZoweeOptimizer {
  zo := &ZoweeOptimizer{
    definedVars:      make(map[string]string),
    intermediateVars: make(map[string]string),
  }

  // Initialize with the initial variables
  for varRef, paths := range initialVars {
    varName := strings.Trim(varRef, "'$")
    if len(paths) > 0 {
      zo.definedVars[varName] = paths[0]
    }
  }

  return zo
}

// Optimize optimizes a list of paths to generate environment variable definitions.
func (zo *ZoweeOptimizer) Optimize(paths []string, initialRoots []string) []string {
  output := []string{}

  // Find common prefixes and define intermediate variables
  zo.defineIntermediateVars(paths)

  for _, path := range paths {
    varName := zo.generateVarName(path)
    if varName == "" {
      continue
    }

    // Skip defining a var for a root that was passed in
    if contains(initialRoots, "$"+varName) && zo.definedVars[varName] == path {
      continue
    }

    bestSubstitution := zo.findBestSubstitution(path)

    var value string
    if bestSubstitution != nil {
      relativePath := strings.TrimPrefix(path, bestSubstitution["path"]+"/")
      value = fmt.Sprintf("$%s/%s", bestSubstitution["var"], relativePath)
    } else {
      value = path
    }

    output = append(output, fmt.Sprintf("export %s=%s", varName, value))
    zo.definedVars[varName] = path
  }

  // Combine intermediate variables and output
  result := []string{}

  // Sort intermediate vars by their path for consistent output
  var intermediatePaths []string
  for path := range zo.intermediateVars {
    intermediatePaths = append(intermediatePaths, path)
  }
  sort.Strings(intermediatePaths)

  for _, path := range intermediatePaths {
    result = append(result, zo.intermediateVars[path])
  }

  result = append(result, output...)
  return uniqueStrings(result)
}

// generateVarName generates a valid environment variable name from a path.
func (zo *ZoweeOptimizer) generateVarName(path string) string {
  basename := filepath.Base(path)
  if basename == "" {
    return ""
  }

  parts := strings.Split(basename, ".")
  var name string
  if parts[0] == "www" && len(parts) > 1 {
    name = parts[1]
  } else {
    name = parts[0]
  }
  name = strings.ReplaceAll(name, "-", "_")

  // Check for collision
  if existingPath, exists := zo.definedVars[name]; exists && existingPath != path {
    // Collision. Try to disambiguate.
    parentName := filepath.Base(filepath.Dir(path))
    name = parentName + "_" + name
  }

  // Sanitize the name
  reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
  name = reg.ReplaceAllString(name, "_")

  // Prepend underscore if it starts with a digit
  if matched, _ := regexp.MatchString(`^[0-9]`, name); matched {
    name = "_" + name
  }

  return name
}

// defineIntermediateVars defines intermediate variables based on common prefixes.
func (zo *ZoweeOptimizer) defineIntermediateVars(paths []string) {
  prefixes := make(map[string]int)

  for _, path := range paths {
    parts := strings.Split(path, "/")
    for i := 1; i < len(parts); i++ {
      prefix := strings.Join(parts[:i], "/")
      prefixes[prefix]++
    }
  }

  // Sort by length to define shorter prefixes first
  var sortedPrefixes []string
  for prefix := range prefixes {
    sortedPrefixes = append(sortedPrefixes, prefix)
  }
  sort.Slice(sortedPrefixes, func(i, j int) bool {
    return len(sortedPrefixes[i]) < len(sortedPrefixes[j])
  })

  for _, prefix := range sortedPrefixes {
    // An intermediate variable is useful if it's a prefix to at least 2 paths
    isUseful := prefixes[prefix] > 1 &&
      !mapContainsValue(zo.definedVars, prefix) &&
      !hasParentInPaths(prefix, paths) &&
      !hasConflictingDefined(prefix, zo.definedVars)

    isNotAnInputPath := !contains(paths, prefix)

    if !isUseful || !isNotAnInputPath {
      continue
    }

    varName := zo.generateVarName(prefix)
    if varName == "" {
      continue
    }

    bestSubstitution := zo.findBestSubstitution(prefix)
    var value string
    if bestSubstitution != nil {
      relativePath := strings.TrimPrefix(prefix, bestSubstitution["path"]+"/")
      value = fmt.Sprintf("$%s/%s", bestSubstitution["var"], relativePath)
    } else {
      value = prefix
    }

    if _, exists := zo.definedVars[varName]; !exists {
      zo.definedVars[varName] = prefix
      zo.intermediateVars[prefix] = fmt.Sprintf("export %s=%s", varName, value)
    }
  }
}

// findBestSubstitution finds the best substitution for a given path.
func (zo *ZoweeOptimizer) findBestSubstitution(path string) map[string]string {
  var bestSubstitution map[string]string
  longestMatch := 0

  for subVar, subPath := range zo.definedVars {
    if strings.HasPrefix(path, subPath+"/") && len(subPath) > longestMatch {
      bestSubstitution = map[string]string{
        "var":  subVar,
        "path": subPath,
      }
      longestMatch = len(subPath)
    }
  }

  return bestSubstitution
}

// Helper functions

func contains(slice []string, item string) bool {
  for _, s := range slice {
    if s == item {
      return true
    }
  }
  return false
}

func mapContainsValue(m map[string]string, value string) bool {
  for _, v := range m {
    if v == value {
      return true
    }
  }
  return false
}

func hasParentInPaths(prefix string, paths []string) bool {
  dir := filepath.Dir(prefix)
  for _, p := range paths {
    if dir == p {
      return true
    }
  }
  return false
}

func hasConflictingDefined(prefix string, defined map[string]string) bool {
  for _, v := range defined {
    if strings.HasPrefix(prefix, v) || strings.HasPrefix(v, prefix) {
      return true
    }
  }
  return false
}

func uniqueStrings(slice []string) []string {
  seen := make(map[string]bool)
  result := []string{}
  for _, s := range slice {
    if !seen[s] {
      seen[s] = true
      result = append(result, s)
    }
  }
  return result
}
