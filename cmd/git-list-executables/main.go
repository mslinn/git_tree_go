package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mslinn/git_tree_go/internal"
)

// Lists executables installed by git-tree-go
func main() {
	descriptions := map[string]string{
		"git-commitAll":   "Commit all changes in the current repository.",
		"git-evars":       "Lists all environment variables used by git.",
		"git-exec":        "Execute a command in each repository of the tree.",
		"git-replicate":   "Replicate a git repository.",
		"git-treeconfig":  "Manage the git-tree configuration.",
		"git-update":      "Update all repositories in the tree.",
		"git-list-executables": "Lists executables installed by git-tree-go.",
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	dir := filepath.Join(cwd, "bin")

	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dir, err)
		os.Exit(1)
	}

	fmt.Printf("Executables installed by git-tree-go v%s in: %s\n\n", internal.Version, dir)
	for _, entry := range entries {
		name := entry.Name()
		description := descriptions[name]
		fmt.Printf("%s: %s\n", name, description)
	}
}