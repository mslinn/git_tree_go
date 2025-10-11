// scripts/verify_version.go: Verifies that the version in internal/version.go matches
// the latest version in CHANGELOG.md and ensures release notes exist.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Read version from internal/version.go
	versionFile, err := os.ReadFile("internal/version.go")
	if err != nil {
		log.Fatalf("Failed to read internal/version.go: %v", err)
	}
	versionRe := regexp.MustCompile(`Version\s*=\s*"([^"]+)"`)
	matches := versionRe.FindStringSubmatch(string(versionFile))
	if len(matches) < 2 {
		log.Fatal("Version not found in internal/version.go")
	}
	goVersion := matches[1]

	// Read latest version from CHANGELOG.md
	changelogFile, err := os.Open("CHANGELOG.md")
	if err != nil {
		log.Fatalf("Failed to read CHANGELOG.md: %v", err)
	}
	defer changelogFile.Close()

	changelogRe := regexp.MustCompile(`##\s*\[?([0-9]+\.[0-9]+\.[0-9]+)\]?`)
	scanner := bufio.NewScanner(changelogFile)
	var changelogVersion string
	var hasNotes bool
	for scanner.Scan() {
		if matches := changelogRe.FindStringSubmatch(scanner.Text()); len(matches) > 1 {
			changelogVersion = matches[1]
			break
		}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "##") {
			hasNotes = true
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading CHANGELOG.md: %v", err)
	}

	if changelogVersion == "" {
		log.Fatal("No version found in CHANGELOG.md")
	}
	if goVersion != changelogVersion {
		log.Fatalf("Version mismatch: internal/version.go (%s) != CHANGELOG.md (%s)", goVersion, changelogVersion)
	}
	if !hasNotes {
		log.Fatal("No release notes found for version %s in CHANGELOG.md", changelogVersion)
	}

	fmt.Printf("Version %s verified successfully\n", goVersion)
}