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

	if *help || len(flag.Args()) == 0 {
		printHelp()
		os.Exit(0)
	}

	args := flag.Args()
	command := args[len(args)-1]
	rootArgs := args[:len(args)-1]

	config := &internal.Config{
		DefaultRoots: []string{"HOME/work", "$HOME/sites"}, // Example default roots
	}

	rootPaths, err := internal.DetermineRoots(rootArgs, config)
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
			executeCommand(repoPath, command, *quiet, *verbose)
		}(repo)
	}
	wg.Wait()
}

func executeCommand(repoPath, command string, quiet, verbose bool) {
	if !quiet {
		fmt.Printf("Executing in %s: %s\n", repoPath, command)
	}

	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("[ERROR] Command failed in %s: %v\n", repoPath, err)
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
	fmt.Println("git-exec - Executes an arbitrary shell command for each repository.")
	fmt.Println("\nUsage: git-exec [OPTIONS] [ROOTS...] SHELL_COMMAND")
	fmt.Println("\nOPTIONS:")
	flag.PrintDefaults()
	fmt.Println("\nROOTS can be directory names or environment variable references (e.g., '$work').")
}
