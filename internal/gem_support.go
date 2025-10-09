package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// CommonPrefix returns the longest path prefix that is a prefix of all paths in the array.
// If the array is empty, return ”.
// If only the leading slash matches and allowRootMatch is true, return '/', else return ”.
func CommonPrefix(paths []string, allowRootMatch bool) string {
	if len(paths) == 0 {
		return ""
	}

	// Check for relative paths
	for _, path := range paths {
		if !strings.HasPrefix(path, "/") {
			fmt.Fprintf(os.Stderr, "Error: common_prefix received relative path: %s\n", path)
			os.Exit(1)
		}
	}

	if len(paths) == 1 {
		dir := filepath.Dir(paths[0])
		if dir == "/" && !allowRootMatch {
			return ""
		}
		return dir
	}

	sorted := make([]string, len(paths))
	copy(sorted, paths)
	sort.Strings(sorted)

	first := strings.Split(sorted[0], "/")
	last := strings.Split(sorted[len(sorted)-1], "/")

	i := 0
	for i < len(first) && i < len(last) && first[i] == last[i] {
		i++
	}

	result := strings.Join(first[:i], "/")
	if result == "" && allowRootMatch {
		return "/"
	}
	return result
}

// Roots returns the common root directory for the given paths up to the specified level.
// Level is 1-indexed (minimum # of leading directory names in result).
func Roots(paths []string, level int, allowRootMatch bool) string {
	if level <= 0 {
		fmt.Fprintf(os.Stderr, "Error: level must be positive, but it is %d.\n", level)
		os.Exit(1)
	}

	if len(paths) == 0 {
		if allowRootMatch {
			return "/"
		}
		return ""
	}

	if len(paths) == 1 {
		root := filepath.Dir(paths[0])
		if root == "/" {
			if allowRootMatch {
				return "/"
			}
			return ""
		}
		return root
	}

	for {
		paths = TrimToLevel(paths, level)
		if len(paths) == 1 {
			return paths[0]
		}

		level--
		if level == 0 {
			break
		}
	}

	if allowRootMatch {
		return "/"
	}
	return ""
}

// TrimToLevel trims paths to the specified level (1-indexed).
func TrimToLevel(paths []string, level int) []string {
	result := make([]string, 0, len(paths))
	seen := make(map[string]bool)

	for _, path := range paths {
		parts := strings.Split(path, "/")
		// Filter out empty strings
		var elements []string
		for _, part := range parts {
			if part != "" {
				elements = append(elements, part)
			}
		}

		if len(elements) >= level {
			elements = elements[:level]
		}
		trimmed := "/" + strings.Join(elements, "/")

		if !seen[trimmed] {
			result = append(result, trimmed)
			seen[trimmed] = true
		}
	}

	sort.Strings(result)
	return result
}

// DerefSymlink returns the real path of a symlink.
func DerefSymlink(symlink string) (string, error) {
	realPath, err := filepath.EvalSymlinks(symlink)
	if err != nil {
		return "", err
	}
	return realPath, nil
}

// EnsureEndsWith ensures that a string ends with the specified suffix.
func EnsureEndsWith(str, suffix string) string {
	return strings.TrimSuffix(str, suffix) + suffix
}

// ExpandEnv expands environment variables in a string.
// Supports $VAR, ${VAR}, and %VAR% formats.
func ExpandEnv(str string) string {
	// Match $VAR, ${VAR}, and %VAR%
	re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)|\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}|%([a-zA-Z_][a-zA-Z0-9_]*)%`)

	result := re.ReplaceAllStringFunc(str, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			varName = strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		} else if strings.HasPrefix(match, "$") {
			varName = strings.TrimPrefix(match, "$")
		} else if strings.HasPrefix(match, "%") {
			varName = strings.Trim(match, "%")
		}

		return os.Getenv(varName)
	})

	return result
}
