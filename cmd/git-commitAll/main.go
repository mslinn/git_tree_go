package main

import (
	"flag"
	"fmt"
	"git-tree-go/internal"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func main() {
	message := flag.String("m", "-", "Commit message.")
	help := flag.Bool("h", false, "Show help message and exit.")
	quiet := flag.Bool("q", false, "Suppress normal output, only show errors.")
	verbose := flag.Bool("v", false, "Increase verbosity.")
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

	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(repoPath string) {
			defer wg.Done()
			commitRepo(repoPath, *message, *quiet, *verbose)
		}(repo)
	}
	wg.Wait()
}

func commitRepo(repoPath, message string, quiet, verbose bool) {
	if !quiet {
		fmt.Printf("Checking %s\n", repoPath)
	}

	r, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Printf("[ERROR] Error opening repository %s: %v\n", repoPath, err)
		return
	}

	w, err := r.Worktree()
	if err != nil {
		log.Printf("[ERROR] Error getting worktree for %s: %v\n", repoPath, err)
		return
	}

	status, err := w.Status()
	if err != nil {
		log.Printf("[ERROR] Error getting status for %s: %v\n", repoPath, err)
		return
	}

	if status.IsClean() {
		if verbose {
			fmt.Printf("No changes to commit in %s\n", repoPath)
		}
		return
	}

	if !quiet {
		fmt.Printf("Committing changes in %s\n", repoPath)
	}

	_, err = w.Add(".")
	if err != nil {
		log.Printf("[ERROR] Error adding files in %s: %v\n", repoPath, err)
		return
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "git-tree-go",
			Email: "git-tree-go@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Printf("[ERROR] Error committing in %s: %v\n", repoPath, err)
		return
	}

	if !quiet {
		fmt.Printf("Pushing changes in %s\n", repoPath)
	}

	err = r.Push(&git.PushOptions{})
	if err != nil {
		log.Printf("[ERROR] Error pushing in %s: %v\n", repoPath, err)
		return
	}

	if !quiet {
		fmt.Printf("Committed and pushed changes in %s\n", repoPath)
	}
}

func printHelp() {
	fmt.Println("git-commitAll - Recursively commits and pushes changes in all git repositories.")
	fmt.Println("\nUsage: git-commitAll [OPTIONS] [ROOTS...]")
	fmt.Println("\nOPTIONS:")
	flag.PrintDefaults()
	fmt.Println("\nROOTS can be directory names or environment variable references (e.g., '$work').")
}
