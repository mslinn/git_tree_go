# `git-tree-go`

This Go package installs commands that walk one or more git directory trees and
act on each repository.
Directories containing a file called `.ignore` are ignored.
Ignoring a directory means all subdirectories are also ignored.
Multiple goroutines are used to dramatically boost performance.

## Commands

- The `git-commitAll` command commits and pushes all changes to each repository
  in the tree. Repositories in a detached `HEAD` state are skipped.

- The `git-evars` command writes a script that defines environment variables
  pointing to each git repository.

- The `git-exec` command executes an arbitrary bash command for each repository.

- The `git-replicate` command writes a script that clones the repos in the tree,
  and adds any defined remotes.

  - Any git repos that have already been cloned into the target directory tree
    are skipped. This means you can rerun `git-replicate` as many times as you
    want, without ill effects.

  - All remotes in each repo are replicated.

- The `git-update` command updates each repository in the trees.


## Installation

### Prerequisites

You need Go 1.24 or later installed on your system.

```shell
$ go version
go version go1.24.2 linux/amd64
```


### Building from Source

Download, build and install to `$GOPATH/bin` like this:

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

One of my directory trees holds Jekyll plugins, packaged as 25 gems.
They depend on one another, and must be built in a particular order.
Sometimes an operation must be performed on all of the plugins, and then rebuild them all.

Most operations do not require that the projects be processed in any particular order, however
the build process must be invoked on the dependencies first.
It is quite tedious to do this 25 times, over and over.

This use case is now fulfilled by the `git-exec` command provided by the `git-tree-go` package.
See below for further details.


### Replicating Trees of Git Repositories

Whenever I set up an operating system for a new development computer,
one of the tedious tasks that must be performed is to replicate
the directory trees of Git repositories.

It is a bad idea to attempt to copy an entire Git repository between computers,
because the `.git` directories within them can be quite large.
So large, in fact, that it might take much more time to copy than re-cloning.

The reason is that copying the entire Git repository actually means copying the same information twice:
first the `.git` hidden directory, complete with all the history for the project,
and then again for the files in the currently checked out branch.
Git repos store the entire development history of the project in their `.git` directories,
so as they accumulate history they eventually become much larger than the
code that is checked out at any given time.

This use case is fulfilled by the `git-replicate` and `git-evars` commands provided by this package.


## Usage

### Single- And Multi-Threading

All of these commands are inherently multi-threaded using Go's goroutines.
They consume up to 75% of the CPU cores that your system can provide.
You may notice that your computer's fan gets louder when you run these commands on large numbers of Git repositories.

For builds and other sequential tasks, however, parallelism is inappropriate.
Instead, it is necessary to build components in the proper order.
Doing all the work on a single thread is a straightforward way of ensuring proper task ordering.

Use the `-s/--serial` option when the order that Git projects are processed matters.
All of the commands support this option.
Execution will take much longer than without the option,
because performing most tasks take longer to perform in sequence than performing them in parallel.

### `git-commitAll`

```text
git-commitAll - Recursively commits and pushes changes in all git repositories under the specified roots.
If no directories are given, uses default roots (sites, sitesUbuntu, work) as roots.
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
  git-commitAll [OPTIONS] [DIRECTORY...]

Usage examples:
  git-commitAll                                # Commit with default message "-"
  git-commitAll -m "This is a commit message"  # Commit with a custom message
  git-commitAll $work $sites                   # Commit in repositories under specific roots
```

```shell
$ git-commitAll
Processing $sites $sitesUbuntu $work

All work is complete.
```


### `git-evars`

The `git-evars` command writes a script that defines environment variables pointing to each git repository.
This command should be run on the target computer.

Only one parameter is required:
an environment variable reference, pointing to the top-level directory to replicate.
The environment variable reference must be contained within single quotes to prevent expansion by the shell.

The following appends to any script in the `$work` directory called `.evars`.
The script defines environment variables that point to each git repo pointed to by `$work`:

```shell
$ git-evars '$work' >> $work/.evars
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

### `git-exec` Usage

The `git-exec` command can be run on any computer.

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

### `git-replicate` Usage

This command generates a shell script to replicate a tree of git repositories.
ROOTS can be directory names or environment variable references (e.g., '$work').
Multiple roots can be specified in a single quoted string.

```shell
$ git-replicate '$work' > work.sh                # Replicate repos under $work
$ git-replicate '$work $sites' > replicate.sh    # Replicate repos under $work and $sites
```

When `git-replicate` completes,
edit the generated script to suit, then
copy it to the target machine and run it.

### `git-update`

The `git-update` command updates each repository in the tree by running `git pull`.


## Development

### Makefile

[`Makefile`](https://en.wikipedia.org/wiki/Make_(software))
automates common development tasks for this Go project.
The main commands are:

```shell
make         # Short form for 'make all'.
make all     # The default command. It formats, vets, and builds all the Go commands.
make build   # Compile all commands and places the binaries in the bin/ directory.
make clean   # Delete the bin/ directory to clean up build files.
make fmt     # Format all Go code in the project.
make help    # Display a help message with all available commands.
make install # Install all the commands to your GOPATH/bin, making them executable from your terminal.
make test    # Run all the tests in the project.
make tidy    # Tidy the go.mod and go.sum files.
make vet     # Analyze the code for potential issues.
```


### Project Structure

```text
git_tree_go/
├── cmd/                    # Command implementations
│   ├── git-commitAll/
│   ├── git-evars/
│   ├── git-exec/
│   ├── git-replicate/
│   ├── git-treeconfig/
│   └── git-update/
├── internal/               # Internal packages
│   ├── abstract_command.go
│   ├── config.go
│   ├── util.go
│   ├── git_tree_walker.go
│   ├── log.go
│   ├── task.go
│   ├── thread_pool.go
│   └── zowee_optimizer.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```


### Building

Build just one command:

```shell
make git-commitAll
make git-evars
make git-exec
make git-replicate
make git-treeconfig
make git-update
```

Example:

```shell
$ make
Formatting code...
internal/util.go
internal/task.go
Vetting code...
Building all commands...
  Building git-commitAll...
  Building git-evars...
  Building git-exec...
  Building git-replicate...
  Building git-treeconfig...
  Building git-update...
Build complete!
```

### Running

Commands such as `git-exec` can be run several ways.
The most direct is to use `go run` and point to the source of the file to compile and run.
Unlike `go build`, `go run` does not leave a permanent executable file in your project directory.

```shell
$ go run ./cmd/git-exec $work pwd
```

Alternatively, build everything to `bin/` first:

```shell
$ make build
$ ./bin/git-exec $work pwd
```

Alternatively, build just the command to `bin/` first:

```shell
$ make git-exec
$ ./bin/git-exec $work pwd
```


### Testing

```shell
$ make test
Running tests...
?       git-tree-go/cmd/git-commitAll   [no test files]
?       git-tree-go/cmd/git-evars       [no test files]
?       git-tree-go/cmd/git-exec        [no test files]
?       git-tree-go/cmd/git-replicate   [no test files]
?       git-tree-go/cmd/git-treeconfig  [no test files]
?       git-tree-go/cmd/git-update      [no test files]
=== RUN   TestAbstractCommand_Initialization
--- PASS: TestAbstractCommand_Initialization (0.00s)
=== RUN   TestAbstractCommand_ArgumentHandling
--- PASS: TestAbstractCommand_ArgumentHandling (0.00s)
=== RUN   TestAbstractCommand_ParseCommonFlags_Quiet
--- PASS: TestAbstractCommand_ParseCommonFlags_Quiet (0.00s)
=== RUN   TestAbstractCommand_ParseCommonFlags_Serial
--- PASS: TestAbstractCommand_ParseCommonFlags_Serial (0.00s)
=== RUN   TestAbstractCommand_ParseCommonFlags_Verbose
FAIL    git-tree-go/internal    0.003s
FAIL
make: *** [Makefile:49: test] Error 1
```

Or just look at failures:

```shell
$ make test | grep FAIL
--- FAIL: TestRoots_Level1_OnePathWithManySlashes (0.00s)
--- FAIL: TestRoots_DeeperLevel (0.00s)
--- FAIL: TestZoweeOptimizer_MultipleBranchesFromCommonRoot (0.00s)
--- FAIL: TestZoweeOptimizer_ComplexNesting (0.00s)
FAIL
FAIL    git-tree-go/internal    0.677s
FAIL
```



## License

The package is available as open source under the terms of the
[MIT License](https://opensource.org/licenses/MIT).


## Additional Information

More information is available on
[Mike Slinn's website](https://www.mslinn.com/git/1100-git-tree.html)
