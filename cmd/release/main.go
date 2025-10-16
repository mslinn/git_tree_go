// Release tool for git-tree-go project maintainers
//
// This is a development tool for creating new releases of git-tree-go.
// It is NOT installed with 'go install' and should only be used by project maintainers.
//
// Build with: go build -o release ./cmd/release
// Usage: ./release [OPTIONS] [VERSION]
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
)

type Options struct {
	skipTests bool
	debug     bool
}

func main() {
	opts := Options{}
	var showHelp bool
	flag.CommandLine.SortFlags = true
	flag.BoolVarP(&opts.debug, "debug", "d", false, "Run GoReleaser in debug mode in the CI workflow")
	flag.BoolVarP(&showHelp, "help", "h", false, "Display this help message")
	flag.BoolVarP(&opts.skipTests, "skip-tests", "s", false, "Skip running integration tests")
	flag.Usage = usage
	flag.Parse()

	if showHelp {
		usage()
	}

	fmt.Println("==================================")
	fmt.Println("  git-tree-go Release Script")
	fmt.Println("==================================")
	fmt.Println()

	// Get project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		errorExit(fmt.Sprintf("Failed to find project root: %v", err))
	}

	// Show current version
	showCurrentVersion(projectRoot)
	fmt.Println()

	// Get version from argument or prompt
	version := ""
	if flag.NArg() > 0 {
		version = flag.Arg(0)
	} else {
		nextVersion := getNextVersion(projectRoot)
		version = promptVersion(nextVersion)
	}

	// Validate version
	if err := validateVersion(version); err != nil {
		errorExit(err.Error())
	}
	success(fmt.Sprintf("Version format is valid: %s", version))

	// Run checks
	checkBranch(projectRoot)
	checkClean(projectRoot)
	checkTag(projectRoot, version)

	// Run tests
	if !opts.skipTests {
		runTests(projectRoot)
	} else {
		warning("Skipping tests.")
	}

	// Update version files
	updateVersionFiles(projectRoot, version)

	// Confirmation
	fmt.Println()
	if !confirmDefault("Proceed with release v"+version+"?", true) {
		errorExit("Release cancelled")
	}

	// Create and push tag
	createTag(projectRoot, version, opts.debug)

	fmt.Println()
	success(fmt.Sprintf("Release v%s initiated successfully!", version))
	fmt.Println()
	info("Next steps:")
	fmt.Println("  1. Monitor the GitHub Actions workflow")
	fmt.Println("  2. Verify the release on GitHub")
	fmt.Println("  3. Test the release binaries")
	fmt.Println("  4. Announce the release (if applicable)")
	fmt.Println()

	// Display release URL
	repoURL, err := getRepoURL(projectRoot)
	if err == nil && repoURL != "" {
		info(fmt.Sprintf("Check progress at: https://github.com/%s/actions", repoURL))
	}
}

func usage() {
	nextVersion := getNextVersion(".")
	fmt.Fprintf(os.Stderr, "Release a new version to GitHub\n\n")
	fmt.Fprintf(os.Stderr, "Usage: release [OPTIONS] [VERSION]\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nVERSION: The version to release (e.g., %s)\n", nextVersion)
	os.Exit(0)
}

func info(msg string) {
	fmt.Printf("%sℹ%s  %s\n", colorBlue, colorReset, msg)
}

func success(msg string) {
	fmt.Printf("%s✓%s  %s\n", colorGreen, colorReset, msg)
}

func warning(msg string) {
	fmt.Printf("%s⚠%s  %s\n", colorYellow, colorReset, msg)
}

func errorMsg(msg string) {
	fmt.Printf("%s✗%s  %s\n", colorRed, colorReset, msg)
}

func errorExit(msg string) {
	errorMsg(msg)
	os.Exit(1)
}

func runCommand(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func runCommandVerbose(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getProjectRoot finds the git root directory
func getProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Try to get git root using git -C
	gitRoot, err := runCommand(cwd, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return cwd, nil // Fallback to current directory
	}
	return gitRoot, nil
}

func getNextVersion(dir string) string {
	// Get version from git tags and increment
	output, err := runCommand(dir, "git", "describe", "--tags", "--abbrev=0")
	incrementedVersion := "0.1.0"
	if err == nil {
		latestTag := strings.TrimPrefix(output, "v")
		parts := strings.Split(latestTag, ".")
		if len(parts) == 3 {
			// Increment patch version
			var major, minor, patch int
			fmt.Sscanf(latestTag, "%d.%d.%d", &major, &minor, &patch)
			incrementedVersion = fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
		}
	}

	return incrementedVersion
}

func validateVersion(version string) error {
	matched, _ := regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+$`, version)
	if !matched {
		return fmt.Errorf("invalid version format: %s (expected: X.Y.Z)", version)
	}
	return nil
}

func checkBranch(dir string) {
	branch, err := runCommand(dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		errorExit("Failed to get current branch")
	}

	if branch != "main" && branch != "master" {
		warning(fmt.Sprintf("You are on branch '%s', not main/master", branch))
		if !confirmDefault("Continue anyway?", false) {
			errorExit("Aborted")
		}
	}
	success(fmt.Sprintf("On branch: %s", branch))
}

func checkClean(dir string) {
	output, _ := runCommand(dir, "git", "status", "-s")
	if output != "" {
		warning("Working directory is not clean. Committing...")
		fmt.Println()

		info("Adding all changes...")
		if err := runCommandVerbose(dir, "git", "add", "-A"); err != nil {
			errorExit("Failed to add changes")
		}

		info("Committing changes...")
		if err := runCommandVerbose(dir, "git", "commit", "-m", "Committing leftovers"); err != nil {
			errorExit("Failed to commit changes")
		}

		info("Pushing changes to remote...")
		if err := runCommandVerbose(dir, "git", "push", "origin"); err != nil {
			errorExit("Failed to push changes")
		}

		success("Changes committed and pushed")
	} else {
		success("Working directory is clean")
	}
}

func checkTag(dir, version string) {
	tag := fmt.Sprintf("v%s", version)
	_, err := runCommand(dir, "git", "rev-parse", tag)
	if err == nil {
		errorExit(fmt.Sprintf("Tag %s already exists", tag))
	}
	success(fmt.Sprintf("Tag %s is available", tag))
}

func runTests(dir string) {
	info("Running tests...")

	// Try make test first
	if err := runCommandVerbose(dir, "make", "test:spec"); err != nil {
		errorExit("Tests failed. Fix issues before releasing.")
	}
	success("All tests passed")
}

func updateVersionFiles(dir, version string) {
	// Update internal/version.go
	versionFilePath := filepath.Join(dir, "internal", "version.go")
	content, err := os.ReadFile(versionFilePath)
	if err != nil {
		errorExit(fmt.Sprintf("Failed to read version file: %v", err))
	}

	// Replace the version string
	versionPattern := regexp.MustCompile(`const Version = ".*"`)
	newContent := versionPattern.ReplaceAllString(string(content), fmt.Sprintf(`const Version = "%s"`, version))

	if err := os.WriteFile(versionFilePath, []byte(newContent), 0644); err != nil {
		errorExit(fmt.Sprintf("Failed to write version file: %v", err))
	}

	success(fmt.Sprintf("Updated internal/version.go to %s", version))
}

func createTag(dir, version string, debug bool) {
	tag := fmt.Sprintf("v%s", version)
	tagMessage := fmt.Sprintf("Release %s", tag)

	if debug {
		tagMessage += "\n\n[debug]"
		warning("GoReleaser will run in debug mode in the CI workflow.")
	}

	info(fmt.Sprintf("Creating tag %s...", tag))
	if err := runCommandVerbose(dir, "git", "tag", "-a", tag, "-m", tagMessage); err != nil {
		errorExit("Failed to create tag")
	}
	success(fmt.Sprintf("Tag %s created", tag))

	info("Pushing tag to origin...")
	if err := runCommandVerbose(dir, "git", "push", "origin", tag); err != nil {
		errorExit("Failed to push tag")
	}
	success("Tag pushed to origin")
}

func showCurrentVersion(dir string) {
	latestTag, err := runCommand(dir, "git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		latestTag = "none"
	}
	info(fmt.Sprintf("The most recent version tag is %s.", latestTag))
}

func promptVersion(defaultVersion string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("What version number should this release have (accept the default with Enter) [%s] ", defaultVersion)
	version, _ := reader.ReadString('\n')
	version = strings.TrimSpace(version)
	if version == "" {
		version = defaultVersion
	}
	return version
}

func confirmDefault(prompt string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)

	suffix := "(y/N)"
	if defaultYes {
		suffix = "(Y/n)"
	}

	fmt.Printf("%s %s ", prompt, suffix)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	// If empty response, use default
	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}

func getRepoURL(dir string) (string, error) {
	repoURL, err := runCommand(dir, "git", "config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}

	// Extract repo path from git URL
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimPrefix(repoURL, "git@github.com:")
	repoURL = strings.TrimPrefix(repoURL, "https://github.com/")

	return repoURL, nil
}
