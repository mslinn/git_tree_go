package main

import (
  "fmt"
  "github.com/MakeNowJust/heredoc"
  "os"
  "path/filepath"
  "strings"

  "github.com/mslinn/git_tree_go/internal"
  flag "github.com/spf13/pflag"
)

func main() {
  cmd := internal.NewAbstractCommand(os.Args[1:], true)

  // Add zowee flag
  var zowee bool
  remainingArgs := cmd.ParseFlagsWithCallback(showHelp, func(fs *flag.FlagSet) {
    fs.BoolVarP(&zowee, "zowee", "z", false, "Optimize variable definitions for size")
  })

  // Create walker
  walker, err := internal.NewGitTreeWalker(remainingArgs, cmd.Serial)
  if err != nil {
    internal.Log(internal.LogQuiet, fmt.Sprintf("Error: %v", err), internal.ColorRed)
    os.Exit(1)
  }

  var result []string

  if zowee {
    // Use zowee optimizer
    var allPaths []string
    walker.FindAndProcessRepos(func(dir, rootArg string) {
      allPaths = append(allPaths, dir)
    })

    optimizer := internal.NewZoweeOptimizer(walker.RootMap)
    result = optimizer.Optimize(allPaths, walker.DisplayRoots)
  } else {
    // Simple mode
    walker.FindAndProcessRepos(func(dir, rootArg string) {
      varDef := makeEnvVarWithSubstitution(dir, rootArg, walker)
      if varDef != "" {
        result = append(result, varDef)
      }
    })
  }

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
    git-evars v%s - Generate bash environment variables for each git repository found under specified directory trees.

    Examines trees of git repositories and writes a bash script to STDOUT.
    If no directories are given, uses default roots (%s) as roots.
    These environment variables point to roots of git repository trees to walk.
    Skips directories containing a .ignore file, and all subdirectories.

    Does not redefine existing environment variables; messages are written to STDERR to indicate environment variables that are not redefined.

    Environment variables that point to the roots of git repository trees must have been exported, for example:

      $ export work=$HOME/work

    Usage: git-evars [OPTIONS] [ROOTS...]

    Options:
      -h, --help           Show this help message and exit.
      -q, --quiet          Suppress normal output, only show errors.
      -v, --verbose        Increase verbosity. Can be used multiple times (e.g., -v, -vv).
      -z, --zowee          Optimize variable definitions for size.

    ROOTS can be:
      - Environment variable names (e.g., work, sites) - expanded automatically if defined
      - Environment variable references (e.g., '$work', $sites) - with explicit $ prefix
      - Directory paths (e.g., /home/user/projects, .)
    Multiple roots can be specified as separate arguments or in a single quoted string.

    Usage examples:
    $ git-evars                 # Use default environment variables as roots
    $ git-evars '$work $sites'  # Use specific environment variables
  `), internal.Version, strings.Join(config.DefaultRoots, ", "))
}

func envVarName(path string) string {
  name := filepath.Base(path)
  if name == "" || name == "." || name == "/" {
    return ""
  }

  // Replace hyphens and spaces with underscores
  name = strings.ReplaceAll(name, "-", "_")
  name = strings.ReplaceAll(name, " ", "_")

  // Handle special case: www.something.com -> something
  if strings.HasPrefix(name, "www.") {
    parts := strings.Split(name, ".")
    if len(parts) > 1 {
      name = parts[1]
    }
  }

  // Remove extension if present
  if strings.Contains(name, ".") {
    parts := strings.Split(name, ".")
    name = parts[0]
  }

  return name
}

func makeEnvVarWithSubstitution(dir, rootArg string, walker *internal.GitTreeWalker) string {
  // Get the root name (without $ and quotes)
  rootName := strings.Trim(rootArg, "'$")

  // Get the root path
  rootPath := os.Getenv(rootName)
  if rootPath == "" {
    // If it's not an env var, it might be a direct path
    if paths, ok := walker.RootMap[rootArg]; ok && len(paths) > 0 {
      rootPath = paths[0]
    } else {
      return ""
    }
  }

  // Check if dir starts with root path
  if !strings.HasPrefix(dir, rootPath) {
    // Fallback to absolute path
    varName := envVarName(dir)
    if varName == "" {
      return ""
    }
    return fmt.Sprintf("export %s=%s", varName, dir)
  }

  // Create relative path
  relativeDir := strings.TrimPrefix(dir, rootPath+"/")
  if relativeDir == "" {
    relativeDir = filepath.Base(dir)
  }

  varName := envVarName(relativeDir)
  if varName == "" {
    return ""
  }

  return fmt.Sprintf("export %s=$%s/%s", varName, rootName, relativeDir)
}
