# `git-tree-go` ![Latest Release](https://img.shields.io/github/v/release/mslinn/git_tree_go)

This Go package installs commands that walk through one or more git
directory trees (breadth-first) and act on each repository.
Directories containing a file called `.ignore` are ignored, as well as all subdirectories.
Multiple goroutines are normally used to dramatically boost performance,
but serial processing is available for deterministic results.


## Command Summary

- The `git-commitAll` command commits and pushes all changes to each repository
  in the tree. Repositories that are in a detached `HEAD` state are skipped.

- The `git-evars` command writes a script that defines environment variables
  pointing to each git repository.

- The `git-exec` command executes an arbitrary shell command for each repository.

- The `git-list-executables` command lists all the executables created by this package.

- The `git-replicate` command writes a script that clones the repositories in the tree,
  and adds any defined remotes.

  - Any git repos that have already been cloned into the target directory tree
    are skipped. This means you can rerun `git-replicate` as many times as you
    want, without ill effects.

  - All remotes in each repository are replicated.

- The `git-update` command updates each repository in the trees.


## Installation

### Prerequisites

[Install the Go language](https://go.dev/doc/install) if you need to.

You need Go 1.24 or later installed on your system.

```shell
$ go version
go version go1.24.2 linux/amd64
```

### Install Release

To install the command-line programs from releases on the GitHub repository
without manually cloning, use `go install`.

The following provides the most recently released version:

```shell
$ go install github.com/mslinn/git_tree_go/cmd/...@latest
go: downloading github.com/mslinn/git_tree_go v0.1.9
go: downloading github.com/MakeNowJust/heredoc v1.0.0
go: downloading github.com/ProtonMail/go-crypto v1.3.0
go: downloading dario.cat/mergo v1.0.2
go: downloading github.com/sergi/go-diff v1.4.0
go: downloading github.com/pjbgf/sha1cd v0.5.0
go: downloading golang.org/x/sys v0.37.0
go: downloading github.com/kevinburke/ssh_config v1.4.0
go: downloading github.com/skeema/knownhosts v1.3.2
go: downloading golang.org/x/crypto v0.43.0
go: downloading golang.org/x/net v0.46.0
go: downloading github.com/cyphar/filepath-securejoin v0.5.0
go: downloading github.com/klauspost/cpuid/v2 v2.3.0
```

The following provides version 0.1.9:

```shell
$ go install github.com/mslinn/git_tree_go/cmd/...@v0.1.9
go: downloading github.com/mslinn/git_tree_go v0.1.9
```


### Building from Source

Download, build and install to `$HOME/go/bin/` like this:

```shell
$ git clone https://github.com/mslinn/git_tree_go.git
$ cd git_tree_go
$ make install
```


## Configuration

The `git-tree-go` commands can be configured to suit your preferences.
Configuration settings are resolved in the following order of precedence,
where items higher in the list override those lower down:

1. **Environment Variables**
2. **User Configuration File** (`~/.treeconfig.yml`)
3. **Default values** built into the program.

This allows for flexible customization of the program's behavior.


### Interactive Setup: `git-treeconfig`

The easiest way to get started is to use the `git-treeconfig` command.
This interactive tool will ask you a few questions
and create a configuration file for you at `~/.treeconfig.yml`.

This is the help message produced by `git-treeconfig -h`:

```text
git-treeconfig - Configure git-tree settings
This utility creates a configuration file at $HOME/.treeconfig.yml
Press Enter to accept the default value in brackets.

Usage: git-treeconfig [OPTIONS]

OPTIONS:
  -h   Show this help message
```

Lets use `git-treeconfig` now.

```shell
$ git-treeconfig
Welcome to git-tree configuration.
This utility will help you create a configuration file at: /home/user/.treeconfig.yml
Press Enter to accept the default value in brackets.

Git command timeout in seconds? |300| 600
Default verbosity level (0=quiet, 1=normal, 2=verbose)? |1|
Default root directories (space-separated)? |sites sitesUbuntu work| dev projects

Configuration saved to /home/user/.treeconfig.yml
```

### Configuration File

The `git-treeconfig` command generates a YAML file (`~/.treeconfig.yml`) that you can also edit manually.
Here is an example:

```yaml
---
git_timeout: 600
verbosity: 1
default_roots:
- dev
- projects
```

### Environment Variables

For temporary overrides or use in CI/CD environments, you can use environment variables.
They must be prefixed with `GIT_TREE_` and be in uppercase.

- `export GIT_TREE_GIT_TIMEOUT=900`
- `export GIT_TREE_VERBOSITY=2`
- `export GIT_TREE_DEFAULT_ROOTS="dev projects personal"` (space-separated string)


## Use Cases

### Dependent Package Maintenance

A directory tree holds Jekyll plugins, packaged as 25 gems.
They depend on one another, and must be built in a particular order.
Sometimes an operation must be performed on all of the plugins,
and then all the gems must be rebuilt.

Most operations do not require that the projects be processed in
any particular order, however the build process must be invoked on the
primary dependencies first.
It is quite tedious to do this 25 times, over and over.

This use case is fulfilled by the `git-exec` command provided by the `git-tree-go` package.
See below for further details.


### Replicating Trees of Git Repositories

Whenever setting up an operating system for a new development computer,
one of the tedious tasks that must be performed is to replicate
one or more directory trees of Git repositories.

It is a bad idea to attempt to copy an entire Git repository between computers,
because the `.git` directories within them can be quite large.
So large, in fact, that it might take much more time to copy than re-cloning.

The reason is that copying the entire Git repository actually means copying the same information twice:
first the `.git` hidden directory, complete with all the history for the project,
and then again for the files in the currently checked out branch.
Git repos store the entire development history of the project in their `.git` directories,
so as they accumulate history they eventually become much larger than the
code that is checked out at any given time.

This use case is fulfilled by the `git-replicate` and `git-evars`
commands provided by this package, working together.

```text
    ┌───────────────┐     ┌───────────────────────────┐
    │  Old Computer │     │        New Computer       │
    │ git-replicate │ ═══►│   git-evars>$work/.evars  │
    └───────────────┘     └───────────────────────────┘
```

`git-replicate` generates a script that recreates the same git directory trees
in the new computer that are found in the old computer.
Ignored subtrees of Git repositories are not processed.

The script generated by `git-replicate` is then copied to the new computer and run;
this recreates the Git repositories that are not ignored.
`git-evars` is then used to generate the contents of the script that generates
environment variable pointing to every Git repository in the new computer.
The diagram calls this script `$work/.evars`.

## Usage

### Single- And Multi-Processing

All of these commands are default to multi-processing mode using
[goroutines](https://dev.to/gophers/what-are-goroutines-and-how-are-they-scheduled-2nj3).
You may notice that your computer's fan gets louder when you run these commands on large numbers of Git repositories.

For builds and other sequential tasks, however, multiprocessing is inappropriate.
Instead, it is necessary to build components in the proper order.
Doing all the work as a single process is a straightforward way of ensuring proper task ordering.

Use the `-s/--serial` option when the order that Git projects are processed matters.
All of the commands support this option.
Execution will take much longer than without the option,
because performing most tasks take longer to perform in sequence than performing them via multiprocessing.

### `git-commitAll`

This is the help message produced by `git-commitAll -h`:

```text
git-commitAll - Recursively commits and pushes changes in all git repositories under the specified roots.
If no directories are given, uses default roots (sites, sitesUbuntu, work) as roots.
Skips directories containing a .ignore file, and all subdirectories.
Repositories in a detached HEAD state are skipped.

Options:
  -h, --help              Show this help message and exit.
  -m, --message MESSAGE   Use the given string as the commit message.
                          (default: "-")
  -q, --quiet             Suppress normal output, only show errors.
  -s, --serial            Run tasks serially in a single thread in the order specified.
  -v, --verbose           Increase verbosity. Can be used multiple times (e.g., -v, -vv).

Usage:
  git-commitAll [OPTIONS] [ROOTS...]

Usage examples:
  git-commitAll                      # Commit default repositories with the default message ("-")
  git-commitAll -m "Commit message"  # Commit default repositories with the same message
  git-commitAll $work $sites         # Commit in repositories under specific roots with the default message
```

```shell
$ git-commitAll
Processing $sites $sitesUbuntu $work

All work is complete.
```


### `git-evars`

This is the help message produced by `git-evars -h`:

```text
git-evars - Generate environment variable definitions for git repositories in directory trees.

Examines trees of git repositories and generates a bash script that defines
environment variables pointing to each git repository.
If no directories are given, default roots are used (sites, sitesUbuntu, work) as roots.
These environment variables point to roots of git repository trees to walk.
Skips directories containing a .ignore file, and all subdirectories.

Does not redefine existing environment variables; messages are written to STDERR to indicate environment
variables that are not redefined.

Environment variables that point to the roots of git repository trees must have been exported, for example:

  $ export work=$HOME/work

Usage: git-evars [OPTIONS] [ROOTS...]

Options:
  -h, --help           Show this help message and exit.
  -q, --quiet          Suppress normal output, only show errors.
  -z, --zowee          Optimize variable definitions for size.
  -v, --verbose        Increase verbosity. Can be used multiple times (e.g., -v, -vv).

ROOTS can be directory names or environment variable references enclosed within single quotes (e.g., '$work').
The environment variable reference must be contained within single quotes to prevent expansion by the shell.
Multiple roots can be specified in a single quoted string.

Usage examples:
$ git-evars                 # Use default environment variables as roots
$ git-evars '$work $sites'  # Use specific environment variables
```

The following appends to any script in the `$work` directory called `.evars`.
The script defines environment variables that point to each git repository pointed to by `$work`:

```shell
$ git-evars '$work' >> $work/.evars
$ source $work/.evars
$ cd $snakesHaveMoreFun
...
$ cd $camels_get_it_done
```


#### Generated Script from `git-evars`

Following is a sample of environment variable definitions.
The `-z`/`--zowee` option generates intermediate environment variable definitions,
making them much easier to work with.

```shell
$ git-evars -z '$sites'
export mnt=/mnt
export c=$mnt/c
export _6of26=$sites/6of26
export computers=$sites/computers.mslinn.com
export ebooks=$sites/ebooks
...
```


### `git-exec`

This is the help message produced by `git-exec -h`:

```text
git-exec - Executes an arbitrary shell command for each repository.

If no arguments are given, uses default roots (sites, sitesUbuntu, work) as roots.
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
```

#### Example 1

For all subdirectories of current directory,
update `Gemfile.lock` and install a local copy of the gem:

```shell
$ git-exec '$jekyll_plugins' 'bundle && bundle update && rake install'
```

#### Example 2

List the projects under the directory pointed to by `$my_plugins`
that have a `demo/` subdirectory:

```shell
$ git-exec '$my_plugins' 'if [ -d demo ]; then realpath demo; fi'
```


### `git-replicate`

This is the help message produced by `git-replicate -h`:

```text
git-replicate - Writes a bash script to STDOUT for replicating trees of Git repositories.

If no directories are given, uses default roots (sites, sitesUbuntu, work) as roots.
The script clones the repositories and replicates any remotes.
Skips directories containing a .ignore file.

Options:
  -h, --help           Show this help message and exit.
  -q, --quiet          Suppress normal output, only show errors.
  -v, --verbose        Increase verbosity. Can be used multiple times (e.g., -v, -vv).

Usage: git-replicate [OPTIONS] [ROOTS...]

ROOTS can be directory names or environment variable references (e.g., '$work').
Multiple roots can be specified in a single quoted string.

Usage examples:
$ git-replicate '$work'
$ git-replicate '$work $sites'

When `git-replicate` completes, edit the generated script to suit, then
copy it to the target machine and run it.
```

### `git-list-executables`

This is the help message produced by `git-list-executables -h`:

```text
git-list-executables - Lists executables installed by git-tree-go.

Usage: git-list-executables [OPTIONS]

OPTIONS:
  -h, --help           Show this help message and exit.
```

Example:

```shell
$ git-list-executables
Executables installed by git-tree-go in: /mnt/f/work/git/git_tree_go/bin

git-commitAll: Commit all changes in the current repository.
git-evars: Lists all environment variables used by git.
git-exec: Execute a command in each repository of the tree.
git-list-executables: Lists executables installed by git-tree-go.
git-replicate: Replicate a git repository.
git-treeconfig: Manage the git-tree configuration.
git-update: Update all repositories in the tree.
```


### `git-update`

This is the help message produced by `git-update -h`:

```text
git-update - Recursively updates trees of git repositories by running git pull.

If no arguments are given, uses default roots (sites, sitesUbuntu, work) as roots.
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
```


## Development

See [DEVELOPMENT.md](DEVELOPMENT.md).


## License

The package is available as open source under the terms of the
[MIT License](https://opensource.org/licenses/MIT).


## Additional Information

More information is available on
[Mike Slinn's website](https://www.mslinn.com/git/1100-git-tree.html)
