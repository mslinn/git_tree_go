package main

import (
	"flag"
	"fmt"
	"git-tree-go/internal"
	"log"
	"os"
	"os/exec"
	"sync"
)

func main() {
	help := flag.Bool("h", false, "Show help message and exit.")
	quiet := flag.Bool("q", false, "Suppress normal output, only show errors.")
	verbose := flag.Bool("v", false, "Increase verbosity.")
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	config := &internal.Config{
		DefaultRoots: []string{"`$HOME/work`", "`$HOME/sites`"}, // Example default roots
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

	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(repoPath string) {
			defer wg.Done()
			updateRepo(repoPath, *quiet, *verbose)
		}(repo)
	}
	wg.Wait()
}

func updateRepo(repoPath string, quiet, verbose bool) {
	if !quiet {
		fmt.Printf("Updating %s\n", repoPath)
	}

	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("[ERROR] git pull failed in %s: %v\n", repoPath, err)
		if len(output) > 0 {
			log.Printf("Output:\n%s", string(output))
		}
		return
	}

	if verbose {
		if len(output) > 0 {
			fmt.Printf("Output for %s:\n%s", repoPath, string(output))
		}
	}
}

func printHelp() {
	fmt.Println("git-update - Recursively updates trees of git repositories.")
	fmt.Println("\nUsage: git-update [OPTIONS] [ROOTS...]")
	fmt.Println("\nOPTIONS:")
	flag.PrintDefaults()
	fmt.Println("\nROOTS:")
	fmt.Println("  When specifying roots, directory paths can be specified, and environment variables can be used, preceded by a dollar sign.")
}
