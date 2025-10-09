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
│   ├── roots.go
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

Make and run:

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

Run all tests:

```shell
$ make test:all
```

Run tests for a specific command:

```shell
$ go test ./cmd/git-commitAll/
$ go test ./cmd/git-exec/
$ go test ./cmd/git-evars/
$ go test ./cmd/git-replicate/
$ go test ./cmd/git-update/
```

Run with verbose output:

```shell
$ go test -v ./cmd/...
```

Skip integration tests (short mode):

```shell
$ go test -short ./cmd/...
```



### Creating Releases

This project uses GoReleaser and GitHub Actions for automated releases. To create a new release:


#### Using the Release Script

The easiest way to create a release:

```shell
$ ./scripts/release.sh 1.2.3
```

This script will:

- Validate the version format
- Check that your working directory is clean
- Run tests
- Create and push a version tag
- Trigger the GitHub Actions release workflow


#### Manual Release

Alternatively, create and push a tag manually:

```shell
$ git tag -a v1.2.3 -m "Release v1.2.3"
$ git push origin v1.2.3
```

#### What Happens Next

Once the tag is pushed, GitHub Actions will automatically:

- Build binaries for all platforms
  (Linux, macOS, Windows on amd64 and arm64)
- Create a GitHub release
- Upload all binaries and checksums
- Generate release notes

For more details, see [RELEASING.md](RELEASING.md).
