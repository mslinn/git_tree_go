package main

import (
  "context"
  "fmt"
	"github.com/MakeNowJust/heredoc"
  "os"
  "os/exec"
  "strings"
  "time"

  "git-tree-go/internal"
)

func main() {
  cmd := internal.NewAbstractCommand(os.Args[1:], true)

  // Parse common flags
  remainingArgs := cmd.ParseCommonFlags(showHelp)

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
  fmt.Printf(heredoc.Doc(`
	  git-update - Recursively updates trees of git repositories.

		If no arguments are given, uses default roots (%s) as roots.
		These environment variables point to roots of git repository trees to walk.
		Skips directories containing a .ignore file, and all subdirectories.

		Environment variables that point to the roots of git repository trees must have been exported, for example:

			$ export work=$HOME/work

		Usage: git-update [OPTIONS] [ROOTS...]

		OPTIONS:
			-h, --help           Show this help message and exit.
			-q, --quiet          Suppress normal output, only show errors.
			-s, --serial         Run tasks serially in a single thread.
			-v, --verbose        Increase verbosity. Can be used multiple times (e.g., -v, -vv).

		ROOTS:
		When specifying roots, directory paths can be specified, and environment variables can be used, preceded by a dollar sign.

		Usage examples:

		$ git-update               # Use default environment variables as roots
		$ git-update $work $sites  # Use specific environment variables
		$ git-update $work /path/to/git/tree
	`), strings.Join(config.DefaultRoots, ", "))
}

func processRepo(walker *internal.GitTreeWalker, dir string, threadID int, config *internal.Config) {
  abbrevDir := walker.AbbreviatePath(dir)
  internal.Log(internal.LogNormal, fmt.Sprintf("Updating %s", abbrevDir), internal.ColorGreen)
  internal.Log(internal.LogVerbose, fmt.Sprintf("Thread %d: git -C %s pull", threadID, dir), internal.ColorYellow)

  // Create context with timeout
  ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.GitTimeout)*time.Second)
  defer cancel()

  // Execute git pull
  gitCmd := exec.CommandContext(ctx, "git", "pull")
  gitCmd.Dir = dir

  output, err := gitCmd.CombinedOutput()
  outputStr := string(output)

  if ctx.Err() == context.DeadlineExceeded {
    internal.Log(internal.LogNormal, fmt.Sprintf("[TIMEOUT] Thread %d: git pull timed out in %s", threadID, abbrevDir), internal.ColorRed)
    return
  }

  if err != nil {
    exitCode := -1
    if exitErr, ok := err.(*exec.ExitError); ok {
      exitCode = exitErr.ExitCode()
    }
    internal.Log(internal.LogNormal, fmt.Sprintf("[ERROR] git pull failed in %s (exit code %d):", abbrevDir, exitCode), internal.ColorRed)
    if len(outputStr) > 0 {
      internal.Log(internal.LogNormal, strings.TrimSpace(outputStr), internal.ColorRed)
    }
    return
  }

  // Success
  if internal.GetVerbosity() >= internal.LogVerbose && len(outputStr) > 0 {
    internal.Log(internal.LogNormal, strings.TrimSpace(outputStr), internal.ColorGreen)
  }
}
