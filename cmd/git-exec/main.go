package main

import (
  "fmt"
  "github.com/MakeNowJust/heredoc"
  "os"
  "os/exec"
  "strings"

  "git-tree-go/internal"
)

func main() {
  cmd := internal.NewAbstractCommand(os.Args[1:], false)

  // Parse common flags
  remainingArgs := cmd.ParseCommonFlags(showHelp)

  if len(remainingArgs) == 0 {
    showHelp()
    os.Exit(1)
  }

  // The last argument is the command to execute, the rest are roots for the walker
  var commandArgs []string
  var shellCommand string

  if len(remainingArgs) > 1 {
    commandArgs = remainingArgs[0 : len(remainingArgs)-1]
    shellCommand = remainingArgs[len(remainingArgs)-1]
  } else {
    // Only command provided, use default roots
    commandArgs = []string{}
    shellCommand = remainingArgs[0]
  }

  rootsToWalk := commandArgs
  if len(commandArgs) == 0 {
    rootsToWalk = cmd.Config.DefaultRoots
  }

  // Create walker
  walker, err := internal.NewGitTreeWalker(rootsToWalk, cmd.Serial)
  if err != nil {
    internal.Log(internal.LogQuiet, fmt.Sprintf("Error: %v", err), internal.ColorRed)
    os.Exit(1)
  }

  // Process repositories
  walker.Process(func(dir string, threadID int, w *internal.GitTreeWalker) {
    executeAndLog(dir, shellCommand)
  })

  internal.ShutdownLogger()
}

func showHelp() {
  config := internal.NewConfig()
  fmt.Printf(heredoc.Doc(`
    git-exec - Executes an arbitrary shell command for each repository.

    If no arguments are given, uses default roots (%s) as roots.
    These environment variables point to roots of git repository trees to walk.
    Skips directories containing a .ignore file, and all subdirectories.

    Environment variables that point to the roots of git repository trees must have been exported, for example:

      $ export work=$HOME/work

    Usage: git-exec [OPTIONS] [ROOTS...] SHELL_COMMAND

    Options:
      -h, --help           Show this help message and exit.
      -q, --quiet          Suppress normal output, only show errors.
      -s, --serial         Run tasks serially in a single thread in the order specified.
      -v, --verbose        Increase verbosity. Can be used multiple times (e.g., -v, -vv).

    ROOTS can be directory names or environment variable references (e.g., '$work').
    Multiple roots can be specified in a single quoted string.

    Usage examples:
    1) For all git repositories under $sites, display their root directories:
      $ git-exec '$sites' pwd

    2) For all git repositories under the current directory and $my_plugins, list the demo/ subdirectory if it exists.
      $ git-exec '. $my_plugins' 'if [ -d demo ]; then realpath demo; fi'

    3) For all subdirectories of the current directory, update Gemfile.lock and install a local copy of the gem:
      $ git-exec . 'bundle update && rake install'
  `), strings.Join(config.DefaultRoots, ", "))
}

func executeAndLog(dir, command string) {
  // Execute the command
  execCmd := exec.Command("sh", "-c", command)
  execCmd.Dir = dir

  output, err := execCmd.CombinedOutput()
  outputStr := strings.TrimSpace(string(output))

  if err != nil {
    // Command failed
    if len(outputStr) > 0 {
      internal.Log(internal.LogQuiet, outputStr, internal.ColorRed)
    } else {
      errorMsg := fmt.Sprintf("Error: Command '%s' failed in %s", command, dir)
      internal.Log(internal.LogQuiet, errorMsg, internal.ColorRed)
    }
  } else {
    // Command succeeded
    if len(outputStr) > 0 {
      internal.LogStdout(outputStr)
    }
  }
}
