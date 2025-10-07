package main

import (
	"flag"
	"fmt"
	"git-tree-go/internal"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

func main() {
	help := flag.Bool("h", false, "Show help message and exit.")
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	config := &internal.Config{
		DefaultRoots: []string{"HOME/work", "HOME/sites"}, // Example default roots
	}

	args := flag.Args()
	rootPaths, err := internal.DetermineRoots(args, config)
	if err != nil {
		log.Fatalf("Error determining roots: %v", err)
	}

	var repos []string
	for _, rootPath := range rootPaths {
		r, err := internal.FindGitReposRecursive(rootPath)
		if err != nil {
			log.Printf("Error finding git repos in %s: %v", rootPath, err)
		}
		repos = append(repos, r...)
	}

	for _, repoPath := range repos {
		script, err := replicateRepo(repoPath, rootPaths)
		if err != nil {
			log.Printf("[ERROR] Error replicating repo %s: %v\n", repoPath, err)
			continue
		}
		if script != "" {
			fmt.Println(script)
		}
	}
}

func replicateRepo(repoPath string, rootPaths []string) (string, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("error opening repository: %w", err)
	}

	origin, err := r.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("error getting origin remote: %w", err)
	}

	originURL := origin.Config().URLs[0]

	var relativeDir string
	for _, root := range rootPaths {
		if strings.HasPrefix(repoPath, root) {
			relativeDir, _ = filepath.Rel(root, repoPath)
			break
		}
	}
	if relativeDir == "" {
		relativeDir = filepath.Base(repoPath)
	}

	var script strings.Builder
	script.WriteString(fmt.Sprintf("if [ ! -d \"%s/.git\" ]; then\n", relativeDir))
	script.WriteString(fmt.Sprintf("  mkdir -p '%s'\n", filepath.Dir(relativeDir)))
	script.WriteString(fmt.Sprintf("  pushd '%s' > /dev/null\n", filepath.Dir(relativeDir)))
	script.WriteString(fmt.Sprintf("  git clone '%s' '%s'\n", originURL, filepath.Base(relativeDir)))

	remotes, err := r.Remotes()
	if err != nil {
		return "", fmt.Errorf("error getting remotes: %w", err)
	}

	for _, remote := range remotes {
		if remote.Config().Name != "origin" {
			script.WriteString(fmt.Sprintf("  git remote add %s '%s'\n", remote.Config().Name, remote.Config().URLs[0]))
		}
	}

	script.WriteString("  popd > /dev/null\n")
	script.WriteString("fi")

	return script.String(), nil
}

func printHelp() {
	fmt.Println("git-replicate - Replicates trees of git repositories and writes a bash script to STDOUT.")
	fmt.Println("\nUsage: git-replicate [OPTIONS] [ROOTS...]")
	fmt.Println("\nOPTIONS:")
	flag.PrintDefaults()
	fmt.Println("\nROOTS can be directory names or environment variable references (e.g., '$work').")
}
