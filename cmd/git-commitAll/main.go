package main

import (
  "context"
  "flag"
  "fmt"
  "github.com/MakeNowJust/heredoc"
  "os"
  "os/exec"
  "strings"
  "time"

  "github.com/mslinn/git_tree_go/internal"
  "github.com/go-git/go-git/v5"
)

var commitMessage string

func main() {
  cmd := internal.NewAbstractCommand(os.Args[1:], true)

  // Add message flag
  remainingArgs := cmd.ParseFlagsWithCallback(showHelp, func(fs *flag.FlagSet) {
    fs.StringVar(&commitMessage, "m", "-", "Use the given string as the commit message")
    fs.StringVar(&commitMessage, "message", "-", "Use the given string as the commit message")
  })

  // Create walker
  walker, err := internal.NewGitTreeWalker(remainingArgs, cmd.Serial)
  if err != nil {
    internal.Log(internal.LogQuiet, fmt.Sprintf("Error: %v", err), internal.ColorRed)
    os.Exit(1)
  }

  // Process repositories
  walker.Process(func(dir string, threadID int, w *internal.GitTreeWalker) {
    processRepo(w, dir, threadID, cmd.Config)
  })

  internal.ShutdownLogger()
}

func showHelp() {
  config := internal.NewConfig()
  fmt.Printf(heredoc.Doc(`git-commitAll - Recursively commits and pushes changes in all git repositories under the specified roots.
    If no directories are given, uses default roots (%s) as roots.
    Skips directories containing a .ignore file, and all subdirectories.
    Repositories in a detached HEAD state are skipped.

    Options:
      -h, --help                Show this help message and exit.
      -m, --message MESSAGE     Use the given string as the commit message.
                                (default: "-")
      -q, --quiet               Suppress normal output, only show errors.
      -s, --serial              Run tasks serially in a single thread in the order specified.
      -v, --verbose             Increase verbosity. Can be used multiple times (e.g., -v, -vv).

    Usage:
      git-commitAll [OPTIONS] [ROOTS...]

    ROOTS can be directory names or environment variable references (e.g., '$work').
    Multiple roots can be specified in a single quoted string.

    Usage examples:
      git-commitAll                                # Commit with default message "-"
      git-commitAll -m "This is a commit message"  # Commit with a custom message
      git-commitAll $work $sites                   # Commit in repositories under specific roots
    `), strings.Join(config.DefaultRoots, ", "))
}

func processRepo(walker *internal.GitTreeWalker, dir string, threadID int, config *internal.Config) {
  shortDir := walker.AbbreviatePath(dir)
  internal.Log(internal.LogVerbose, fmt.Sprintf("Examining %s on thread %d", shortDir, threadID), internal.ColorGreen)

  // Check for .ignore file
  if _, err := os.Stat(dir + "/.ignore"); err == nil {
    internal.Log(internal.LogDebug, fmt.Sprintf("  Skipping %s due to .ignore file", shortDir), internal.ColorGreen)
    return
  }

  // Open the repository
  repo, err := git.PlainOpen(dir)
  if err != nil {
    internal.Log(internal.LogNormal, fmt.Sprintf("Error opening repository %s: %v", shortDir, err), internal.ColorRed)
    return
  }

  // Check if HEAD is detached
  head, err := repo.Head()
  if err != nil {
    internal.Log(internal.LogVerbose, fmt.Sprintf("  Skipping %s because HEAD is detached or invalid", shortDir), internal.ColorYellow)
    return
  }

  if !head.Name().IsBranch() {
    internal.Log(internal.LogVerbose, fmt.Sprintf("  Skipping %s because it is in a detached HEAD state", shortDir), internal.ColorYellow)
    return
  }

  // Create context with timeout
  ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.GitTimeout)*time.Second)
  defer cancel()

  // Check if there are changes
  if !repoHasChanges(ctx, dir) {
    internal.Log(internal.LogDebug, fmt.Sprintf("  No changes to commit in %s", shortDir), internal.ColorGreen)
    return
  }

  // Commit and push changes
  if err := commitChanges(ctx, dir, commitMessage, shortDir); err != nil {
    if ctx.Err() == context.DeadlineExceeded {
      internal.Log(internal.LogNormal, fmt.Sprintf("[TIMEOUT] Thread %d: git operations timed out in %s", threadID, shortDir), internal.ColorRed)
    } else {
      internal.Log(internal.LogNormal, fmt.Sprintf("Error processing %s: %v", shortDir, err), internal.ColorRed)
    }
  }
}

func repoHasChanges(ctx context.Context, dir string) bool {
  cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
  cmd.Dir = dir

  output, err := cmd.Output()
  if err != nil {
    return false
  }

  return len(strings.TrimSpace(string(output))) > 0
}

func repoHasStagedChanges(ctx context.Context, dir string) bool {
  cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only")
  cmd.Dir = dir

  output, err := cmd.Output()
  if err != nil {
    return false
  }

  return len(strings.TrimSpace(string(output))) > 0
}

func commitChanges(ctx context.Context, dir, message, shortDir string) error {
  // Stage all changes
  addCmd := exec.CommandContext(ctx, "git", "add", "--all")
  addCmd.Dir = dir
  if err := addCmd.Run(); err != nil {
    return fmt.Errorf("git add failed: %w", err)
  }

  // Check if there are staged changes
  if !repoHasStagedChanges(ctx, dir) {
    return nil
  }

  // Commit changes
  commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", message, "--quiet", "--no-gpg-sign")
  commitCmd.Dir = dir
  if err := commitCmd.Run(); err != nil {
    return fmt.Errorf("git commit failed: %w", err)
  }

  // Get current branch name
  branchCmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
  branchCmd.Dir = dir
  branchOutput, err := branchCmd.Output()
  if err != nil {
    return fmt.Errorf("failed to get branch name: %w", err)
  }
  currentBranch := strings.TrimSpace(string(branchOutput))

  // Push changes
  pushCmd := exec.CommandContext(ctx, "git", "push", "--set-upstream", "origin", currentBranch)
  pushCmd.Dir = dir
  if err := pushCmd.Run(); err != nil {
    return fmt.Errorf("git push failed: %w", err)
  }

  internal.Log(internal.LogNormal, fmt.Sprintf("Committed and pushed changes in %s", shortDir), internal.ColorGreen)
  return nil
}
