package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git-tree-go/internal"
)

func main() {
	config := internal.NewConfig()
	scanner := bufio.NewScanner(os.Stdin)

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := fmt.Sprintf("%s/.treeconfig.yml", home)

	fmt.Println("Welcome to git-tree configuration.")
	fmt.Printf("This utility will help you create a configuration file at: %s\n", configPath)
	fmt.Println("Press Enter to accept the default value in brackets.")
	fmt.Println()

	// Git timeout
	fmt.Printf("Git command timeout in seconds? |%d| ", config.GitTimeout)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			if timeout, err := strconv.Atoi(input); err == nil {
				config.GitTimeout = timeout
			} else {
				fmt.Fprintf(os.Stderr, "Invalid timeout value, using default\n")
			}
		}
	}

	// Verbosity
	fmt.Printf("Default verbosity level (0=quiet, 1=normal, 2=verbose)? |%d| ", config.Verbosity)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			if verbosity, err := strconv.Atoi(input); err == nil && verbosity >= 0 && verbosity <= 2 {
				config.Verbosity = verbosity
			} else {
				fmt.Fprintf(os.Stderr, "Invalid verbosity value (must be 0-2), using default\n")
			}
		}
	}

	// Default roots
	fmt.Printf("Default root directories (space-separated)? |%s| ", strings.Join(config.DefaultRoots, " "))
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			config.DefaultRoots = strings.Fields(input)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Save configuration
	if err := config.SaveToFile(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("\033[32mConfiguration saved to %s\033[0m\n", configPath)
}
