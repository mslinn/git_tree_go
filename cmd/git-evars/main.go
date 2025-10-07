package main

import (
	"flag"
	"fmt"
	"git-tree-go/internal"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	zowee := flag.Bool("z", false, "Optimize variable definitions for size.")
	help := flag.Bool("h", false, "Show help message and exit.")
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *zowee {
		fmt.Println("Zowee optimization is not yet implemented in the Go version.")
		os.Exit(1)
	}

	config := &internal.Config{
		DefaultRoots: []string{"HOME/work", "$HOME/sites"}, // Example default roots
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

	for _, repo := range repos {
		varName := envVarName(repo)
		fmt.Printf("export %s=%s\n", varName, repo)
	}
}

func envVarName(path string) string {
	name := filepath.Base(path)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func printHelp() {
	fmt.Println("git-evars - Generate bash environment variables for each git repository.")
	fmt.Println("\nUsage: git-evars [OPTIONS] [ROOTS...]")
	fmt.Println("\nOPTIONS:")
	flag.PrintDefaults()
	fmt.Println("\nROOTS can be directory names or environment variable references (e.g., '$work').")
}
