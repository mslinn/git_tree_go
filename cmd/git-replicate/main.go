package main

import (
  "fmt"
  "github.com/MakeNowJust/heredoc"
  "os"
  "path/filepath"
  "strings"

  "github.com/mslinn/git_tree_go/internal"
  "github.com/go-git/go-git/v5"
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

  var result []string

  // Process repositories
  walker.FindAndProcessRepos(func(dir, rootArg string) {
    output := replicateOne(dir, rootArg, walker)
    if len(output) > 0 {
      result = append(result, output...)
    }
  })

  // Output results to stdout
  if len(result) > 0 {
    for _, line := range result {
      fmt.Println(line)
    }
  }

  internal.ShutdownLogger()
}

func showHelp() {
  config := internal.NewConfig()
  fmt.Printf(heredoc.Doc(`
    git-replicate - Replicates trees of git repositories and writes a bash script to STDOUT.

    If no directories are given, uses default roots (%s) as roots.
    The script clones the repositories and replicates any remotes.
    Skips directories containing a .ignore file.

    Options:
      -h, --help           Show this help message and exit.
      -q, --quiet          Suppress normal output, only show errors.
      -v, --verbose        Increase verbosity. Can be used multiple times (e.g., -v, -vv).

    Usage: git-replicate [OPTIONS] [ROOTS...]

    ROOTS can be:
      - Environment variable names (e.g., work, sites) - expanded automatically if defined
      - Environment variable references (e.g., '$work', $sites) - with explicit $ prefix
      - Directory paths (e.g., /home/user/projects, .)
    Multiple roots can be specified as separate arguments or in a single quoted string.

    Usage examples:
    $ git-replicate '$work'
    $ git-replicate '$work $sites'
  `), strings.Join(config.DefaultRoots, ", "))
}

func replicateOne(dir, rootArg string, walker *internal.GitTreeWalker) []string {
  output := []string{}

  // Open the repository
  repo, err := git.PlainOpen(dir)
  if err != nil {
    internal.Log(internal.LogDebug, fmt.Sprintf("Error opening repository %s: %v", dir, err), internal.ColorRed)
    return output
  }

  // Get the config
  cfg, err := repo.Config()
  if err != nil {
    internal.Log(internal.LogDebug, fmt.Sprintf("Error getting config for %s: %v", dir, err), internal.ColorRed)
    return output
  }

  // Get origin URL
  originRemote, ok := cfg.Remotes["origin"]
  if !ok || len(originRemote.URLs) == 0 {
    internal.Log(internal.LogDebug, fmt.Sprintf("No origin remote found for %s", dir), internal.ColorYellow)
    return output
  }
  originURL := originRemote.URLs[0]

  // Get the root path
  rootName := strings.Trim(rootArg, "'$")
  rootPath := os.Getenv(rootName)
  if rootPath == "" {
    // If it's not an env var, it might be a direct path
    if paths, ok := walker.RootMap[rootArg]; ok && len(paths) > 0 {
      rootPath = paths[0]
    } else {
      return output
    }
  }

  // Calculate relative directory
  relativeDir := strings.TrimPrefix(dir, rootPath+"/")
  if relativeDir == dir {
    // Not a subdirectory of root, use absolute path
    relativeDir = dir
  }

  // Build the script
  output = append(output, fmt.Sprintf("if [ ! -d \"%s/.git\" ]; then", relativeDir))
  output = append(output, fmt.Sprintf("  mkdir -p '%s'", filepath.Dir(relativeDir)))
  output = append(output, fmt.Sprintf("  pushd '%s' > /dev/null", filepath.Dir(relativeDir)))
  output = append(output, fmt.Sprintf("  git clone '%s' '%s'", originURL, filepath.Base(relativeDir)))

  // Add other remotes
  for remoteName, remote := range cfg.Remotes {
    if remoteName == "origin" || len(remote.URLs) == 0 {
      continue
    }
    output = append(output, fmt.Sprintf("  git remote add %s '%s'", remoteName, remote.URLs[0]))
  }

  output = append(output, "  popd > /dev/null")
  output = append(output, "fi")

  return output
}
