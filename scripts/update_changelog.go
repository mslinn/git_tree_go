// scripts/update_changelog.go: Inserts a new version header and release notes template
// into CHANGELOG.md based on the version in internal/version.go.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
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
	version := matches[1]

	// Read existing CHANGELOG.md
	changelogFile, err := os.Open("CHANGELOG.md")
	if err != nil {
		log.Fatalf("Failed to read CHANGELOG.md: %v", err)
	}
	defer changelogFile.Close()

	var lines []string
	scanner := bufio.NewScanner(changelogFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading CHANGELOG.md: %v", err)
	}

	// Check if version already exists
	versionRe = regexp.MustCompile(`##\s*\[?` + regexp.QuoteMeta(version) + `\]?`)
	for _, line := range lines {
		if versionRe.MatchString(line) {
			log.Fatalf("Version %s already exists in CHANGELOG.md", version)
		}
	}

	// Create new changelog content
	newHeader := fmt.Sprintf("## [%s] - %s", version, time.Now().Format("2006-01-02"))
	newContent := []string{
		newHeader,
		"",
		"### Added",
		"- ",
		"",
		"### Changed",
		"- ",
		"",
		"### Deprecated",
		"- ",
		"",
		"### Removed",
		"- ",
		"",
		"### Fixed",
		"- ",
		"",
		"### Security",
		"- ",
		"",
	}
	newContent = append(newContent, lines...)

	// Write updated CHANGELOG.md
	output, err := os.Create("CHANGELOG.md")
	if err != nil {
		log.Fatalf("Failed to write CHANGELOG.md: %v", err)
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for _, line := range newContent {
		fmt.Fprintln(writer, line)
	}
	if err := writer.Flush(); err != nil {
		log.Fatalf("Failed to flush CHANGELOG.md: %v", err)
	}

	fmt.Printf("Updated CHANGELOG.md with version %s\n", version)
}