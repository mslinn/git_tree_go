package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func main() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" || !strings.HasPrefix(githubToken, "ghp_") {
		log.Fatal("GITHUB_TOKEN must be a valid GitHub Personal Access Token (starts with 'ghp_')")
	}

	sshKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		log.Fatalf("SSH key not found at %s", sshKeyPath)
	}
	info, err := os.Stat(sshKeyPath)
	if err != nil {
		log.Fatalf("Failed to stat SSH key %s: %v", sshKeyPath, err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		log.Fatalf("SSH key %s has insecure permissions (%o). Run: chmod 600 %s", sshKeyPath, info.Mode().Perm(), sshKeyPath)
	}
	sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
	if err != nil {
		log.Fatalf("Failed to load SSH key %s: %v", sshKeyPath, err)
	}

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
	tag := "v" + goVersion
	fmt.Printf("Extracted version: %s\n", goVersion)

	changelogFile, err := os.Open("CHANGELOG.md")
	if err != nil {
		log.Fatalf("Failed to read CHANGELOG.md: %v", err)
	}
	defer changelogFile.Close()

	changelogRe := regexp.MustCompile(`##\s*([0-9]+\.[0-9]+\.[0-9]+)\s*/\s*\d{4}-\d{2}-\d{2}`)
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
	if err = scanner.Err(); err != nil {
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

	updateChangelog(goVersion)

	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalf("Failed to open repo: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatalf("Failed to get worktree: %v", err)
	}
	status, err := worktree.Status()
	if err != nil {
		log.Fatalf("Failed to get status: %v", err)
	}
	if !status.IsClean() {
		if err := worktree.AddWithOptions(&git.AddOptions{All: true}); err != nil {
			log.Fatalf("Failed to stage changes: %v", err)
		}
		commitMsg := fmt.Sprintf("Release v%s", goVersion)
		_, err = worktree.Commit(commitMsg, &git.CommitOptions{})
		if err != nil {
			log.Fatalf("Failed to commit: %v", err)
		}
		fmt.Printf("Committed changes: %s\n", commitMsg)
	} else {
		fmt.Println("No changes to commit")
	}

	err = repo.Push(&git.PushOptions{
		RemoteURL: "git@github.com:mslinn/git_tree_go.git",
		Auth:      sshAuth,
		RefSpecs:  []config.RefSpec{config.RefSpec("refs/heads/master:refs/heads/master")},
	})
	if err != nil {
		log.Fatalf("Failed to push to master: %v. Run: ssh -T git@github.com", err)
	}
	fmt.Println("Pushed changes to master")

	_, err = repo.Tag(tag)
	if err == nil {
		fmt.Printf("Tag %s already exists, skipping creation\n", tag)
	} else {
		h, err := repo.Head()
		if err != nil {
			log.Fatalf("Failed to get head: %v", err)
		}
		_, err = repo.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
			Tagger: &object.Signature{
				Name:  "mslinn",
				Email: "mslinn@users.noreply.github.com",
				When:  time.Now(),
			},
			Message: fmt.Sprintf("Release v%s", goVersion),
		})
		if err != nil {
			log.Fatalf("Failed to create tag: %v", err)
		}
		fmt.Printf("Created tag %s\n", tag)
	}

	// Check if tag exists remotely
	tagExistsRemotely := false
	remoteTags, err := repo.Tags()
	if err != nil {
		log.Fatalf("Failed to list tags: %v", err)
	}
	err = remoteTags.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == tag {
			tagExistsRemotely = true
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to iterate tags: %v", err)
	}
	if tagExistsRemotely {
		fmt.Printf("Tag %s already exists on remote, skipping push\n", tag)
	} else {
		err = repo.Push(&git.PushOptions{
			RemoteURL: "git@github.com:mslinn/git_tree_go.git",
			Auth:      sshAuth,
			RefSpecs:  []config.RefSpec{config.RefSpec("refs/tags/" + tag + ":refs/tags/" + tag)},
		})
		if err != nil {
			log.Fatalf("Failed to push tag: %v", err)
		}
		fmt.Printf("Pushed tag %s\n", tag)
	}

	for _, cmd := range []string{"clean", "install"} {
		c := exec.Command("make", cmd)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			log.Fatalf("Failed to run make %s: %v", cmd, err)
		}
	}

	tarball := fmt.Sprintf("git_tree_go_%s.tar.gz", goVersion)
	c := exec.Command("tar", "-czf", tarball, "--exclude=.git", ".")
	if err := c.Run(); err != nil {
		log.Fatalf("Failed to create tarball: %v", err)
	}

	checksum, err := exec.Command("sha256sum", tarball).Output()
	if err != nil {
		log.Fatalf("Failed to create checksum: %v", err)
	}
	checksumFile := fmt.Sprintf("git_tree_go_%s.sha256", goVersion)
	if err := os.WriteFile(checksumFile, checksum, 0644); err != nil {
		log.Fatalf("Failed to write checksum: %v", err)
	}

	c = exec.Command("gh", "release", "create", tag, tarball, checksumFile, "--notes-file", "CHANGELOG.md")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		log.Fatalf("Failed to create GitHub release: %v", err)
	}

	fmt.Println("Release completed successfully")
}

func updateChangelog(version string) {
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
	if err = scanner.Err(); err != nil {
		log.Fatalf("Error reading CHANGELOG.md: %v", err)
	}

	versionRe := regexp.MustCompile(`##\s*` + regexp.QuoteMeta(version) + `\s*/\s*\d{4}-\d{2}-\d{2}`)
	for _, line := range lines {
		if versionRe.MatchString(line) {
			fmt.Printf("Version %s already exists in CHANGELOG.md, skipping update\n", version)
			return
		}
	}

	insertIndex := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "# Change Log") {
			insertIndex = i + 1
			break
		}
	}

	newHeader := fmt.Sprintf("## %s / %s", version, time.Now().Format("2006-01-02"))
	newContent := []string{
		"",
		newHeader,
		"",
		"- ",
		"",
	}
	if insertIndex > 0 {
		newContent = append(lines[:insertIndex], append(newContent, lines[insertIndex:]...)...)
	} else {
		newContent = append(newContent, lines...)
	}

	output, err := os.Create("CHANGELOG.md")
	if err != nil {
		log.Fatalf("Failed to write CHANGELOG.md: %v", err)
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for _, line := range newContent {
		fmt.Fprintln(writer, line)
	}
	if err = writer.Flush(); err != nil {
		log.Fatalf("Failed to flush CHANGELOG.md: %v", err)
	}

	fmt.Printf("Updated CHANGELOG.md with version %s\n", version)
}
