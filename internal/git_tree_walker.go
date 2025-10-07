package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Config represents the application's configuration.
type Config struct {
	DefaultRoots []string
}

// DetermineRoots processes the root directory arguments from the command line or configuration defaults.
func DetermineRoots(args []string, config *Config) ([]string, error) {
	var processedArgs []string
	if len(args) == 0 {
		processedArgs = config.DefaultRoots
	} else {
		for _, arg := range args {
			processedArgs = append(processedArgs, strings.Fields(arg)...)
		}
	}

	var rootPaths []string
	for _, arg := range processedArgs {
		path := os.ExpandEnv(arg)
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute path for %s: %w", path, err)
		}
		rootPaths = append(rootPaths, absPath)
	}

	return rootPaths, nil
}

// FindGitReposRecursive recursively finds Git repositories within a given path.
func FindGitReposRecursive(rootPath string) ([]string, error) {
	var repos []string
	visited := make(map[string]bool)

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if d.IsDir() && d.Name() == ".git" {
			repoPath := filepath.Dir(path)
			if !visited[repoPath] {
				fmt.Printf("Found git repo: %s\n", repoPath)
				repos = append(repos, repoPath)
				visited[repoPath] = true
			}
			return filepath.SkipDir // Prune search
		}

		if _, err := os.Stat(filepath.Join(path, ".ignore")); err == nil {
			fmt.Printf("Skipping %s due to .ignore file\n", path)
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return repos, nil
}

// SortDirectoryEntries sorts the names of subdirectories within the specified directory path.
func SortDirectoryEntries(directoryPath string) ([]string, error) {
	entries, err := os.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	var dirNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		}
	}

	sort.Strings(dirNames)
	return dirNames, nil
}
