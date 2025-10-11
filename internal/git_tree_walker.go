package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var ignoredDirectories = []string{".", "..", ".venv"}

// GitTreeWalker is used to walk a directory tree and find git repositories.
type GitTreeWalker struct {
	Config       *Config
	DisplayRoots []string
	RootMap      map[string][]string
	Serial       bool
}

// NewGitTreeWalker creates a new GitTreeWalker.
func NewGitTreeWalker(args []string, serial bool) (*GitTreeWalker, error) {
	walker := &GitTreeWalker{
		Config:       NewConfig(),
		DisplayRoots: []string{},
		RootMap:      make(map[string][]string),
		Serial:       serial,
	}

	err := walker.determineRoots(args)
	if err != nil {
		return nil, err
	}

	return walker, nil
}

// AbbreviatePath abbreviates a directory path by replacing expanded root prefixes with their original display representations.
// If a root was specified as an environment variable (e.g., $work), this will condense any path under that root to use
// the variable name (e.g., /mnt/f/work/foo -> $work/foo).
func (w *GitTreeWalker) AbbreviatePath(dir string) string {
	// Try to abbreviate using the root map
	// This will match the longest root prefix and replace it with the display representation
	longestMatch := ""
	longestDisplayRoot := ""

	for displayRoot, expandedPaths := range w.RootMap {
		for _, expandedPath := range expandedPaths {
			// Check for prefix match with separator, or exact match
			if (strings.HasPrefix(dir, expandedPath+string(filepath.Separator)) || dir == expandedPath) && len(expandedPath) > len(longestMatch) {
				longestMatch = expandedPath
				longestDisplayRoot = displayRoot
			}
		}
	}

	if longestMatch != "" {
		return strings.Replace(dir, longestMatch, longestDisplayRoot, 1)
	}

	return dir
}

// Process processes the git repositories using the provided function.
func (w *GitTreeWalker) Process(processFunc func(dir string, threadID int, walker *GitTreeWalker)) {
	Log(LogVerbose, fmt.Sprintf("Processing %s", strings.Join(w.DisplayRoots, " ")), ColorGreen)

	if w.Serial {
		w.processSerially(processFunc)
	} else {
		w.processMultithreaded(processFunc)
	}
}

func (w *GitTreeWalker) processSerially(processFunc func(dir string, threadID int, walker *GitTreeWalker)) {
	Log(LogVerbose, "Running in serial mode.", ColorYellow)
	w.FindAndProcessRepos(func(dir, rootArg string) {
		processFunc(dir, 0, w)
	})
}

func (w *GitTreeWalker) processMultithreaded(processFunc func(dir string, threadID int, walker *GitTreeWalker)) {
	pool := NewThreadPoolManager(0.75)
	if pool == nil {
		Log(LogQuiet, "Failed to create thread pool", ColorRed)
		return
	}

	pool.Start(func(task interface{}, workerID int) {
		if dir, ok := task.(string); ok {
			processFunc(dir, workerID, w)
		}
	})

	// Find all repositories and add them to the work queue
	w.FindAndProcessRepos(func(dir, rootArg string) {
		pool.AddTask(dir)
	})

	pool.WaitForCompletion()
}

// FindAndProcessRepos finds git repos and yields them to the callback.
func (w *GitTreeWalker) FindAndProcessRepos(callback func(dir, rootArg string)) {
	visited := make(map[string]bool)

	// Sort root map keys for consistent ordering
	var rootArgs []string
	for rootArg := range w.RootMap {
		rootArgs = append(rootArgs, rootArg)
	}
	sort.Strings(rootArgs)

	for _, rootArg := range rootArgs {
		paths := w.RootMap[rootArg]
		sort.Strings(paths)
		for _, rootPath := range paths {
			w.findGitReposRecursive(rootPath, visited, func(dir string) {
				callback(dir, rootArg)
			})
		}
	}
}

func (w *GitTreeWalker) determineRoots(args []string) error {
	processedArgs := args
	if len(args) == 0 {
		processedArgs = w.Config.DefaultRoots
	}

	// Split any args that contain spaces
	var expandedArgs []string
	for _, arg := range processedArgs {
		expandedArgs = append(expandedArgs, strings.Fields(arg)...)
	}

	w.DisplayRoots = expandedArgs

	for _, arg := range expandedArgs {
		if err := w.processRootArg(arg); err != nil {
			return err
		}
	}

	return nil
}

// processRootArg processes a root argument (environment variable or path).
// This function handles three types of inputs:
// 1. Explicit environment variable references: $VAR or '$VAR' - must be defined
// 2. Implicit environment variable names: VAR (alphanumeric/underscore only) - expanded if defined
// 3. Literal paths: any other string - used as-is
//
// This allows users to specify roots in their config file without the $ prefix
// (e.g., "sites" instead of "$sites"), making the configuration more readable
// while still supporting explicit $ prefixes when needed.
func (w *GitTreeWalker) processRootArg(arg string) error {
	path := arg
	displayName := arg

	// Match $VAR or '$VAR' patterns (explicit environment variable reference)
	envVarPattern := regexp.MustCompile(`^(?:'\$[a-zA-Z_]\w*'|\$[a-zA-Z_]\w*)$`)
	if match := envVarPattern.FindString(arg); match != "" {
		varName := strings.Trim(match, "'$")
		envValue := os.Getenv(varName)
		if envValue == "" {
			return fmt.Errorf("environment variable '%s' is undefined", arg)
		}
		path = envValue
		// Keep the display name with $ prefix
		displayName = "$" + varName
	} else {
		// Check if arg looks like an environment variable name (alphanumeric/underscore only)
		// If so, try to expand it as an environment variable (implicit reference)
		varNamePattern := regexp.MustCompile(`^[a-zA-Z_]\w*$`)
		if varNamePattern.MatchString(arg) {
			envValue := os.Getenv(arg)
			if envValue != "" {
				// Environment variable exists, use its value
				path = envValue
				// Add $ prefix to display name since it's an environment variable
				displayName = "$" + arg
			}
			// If environment variable doesn't exist, treat arg as a literal path
		}
	}

	if path != "" {
		// Split by spaces and expand each path
		paths := strings.Fields(path)
		var absPaths []string
		for _, p := range paths {
			absPath, err := filepath.Abs(p)
			if err != nil {
				return fmt.Errorf("could not get absolute path for %s: %w", p, err)
			}
			absPaths = append(absPaths, absPath)
		}
		w.RootMap[displayName] = absPaths
	}

	return nil
}

func (w *GitTreeWalker) findGitReposRecursive(rootPath string, visited map[string]bool, callback func(dir string)) {
	// Check if the directory exists
	info, err := os.Stat(rootPath)
	if err != nil || !info.IsDir() {
		return
	}

	// Check for .ignore file
	if _, err := os.Stat(filepath.Join(rootPath, ".ignore")); err == nil {
		Log(LogDebug, fmt.Sprintf("  Skipping %s due to .ignore file", rootPath), ColorGreen)
		return
	}

	Log(LogDebug, fmt.Sprintf("Scanning %s", rootPath), ColorGreen)

	// Check if this is a git repository
	gitDirOrFile := filepath.Join(rootPath, ".git")
	if info, err := os.Stat(gitDirOrFile); err == nil {
		if info.IsDir() {
			Log(LogDebug, fmt.Sprintf("  Found %s", gitDirOrFile), ColorGreen)
			if !visited[rootPath] {
				visited[rootPath] = true
				callback(rootPath)
			}
			return // Prune search
		} else {
			Log(LogNormal, fmt.Sprintf("  %s is a file, not a directory; skipping", gitDirOrFile), ColorGreen)
			return
		}
	} else {
		Log(LogDebug, fmt.Sprintf("  %s is not a git directory", rootPath), ColorGreen)
	}

	// Recurse into subdirectories
	entries := sortDirectoryEntries(rootPath)
	for _, entry := range entries {
		if isIgnoredDirectory(entry) {
			continue
		}
		w.findGitReposRecursive(filepath.Join(rootPath, entry), visited, callback)
	}
}

func sortDirectoryEntries(directoryPath string) []string {
	entries, err := os.ReadDir(directoryPath)
	if err != nil {
		Log(LogNormal, fmt.Sprintf("Error scanning %s: %s", directoryPath, err.Error()), ColorRed)
		return []string{}
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	sort.Strings(dirs)
	return dirs
}

func isIgnoredDirectory(name string) bool {
	for _, ignored := range ignoredDirectories {
		if name == ignored {
			return true
		}
	}
	return false
}
