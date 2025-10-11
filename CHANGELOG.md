## [0.1.13] - 2025-10-11

### Added
- 

### Changed
- 

### Deprecated
- 

### Removed
- 

### Fixed
- 

### Security
- 

# Change Log

## 0.1.13 / 2025-10-10

- Fixed path condensation in `git-exec` command. Output now properly abbreviates paths using environment variable names
  (e.g., `$sites/project` instead of `/mnt/d/sites/project`).
- Added version number to help output for all commands (displays as "v0.1.13" in command help text).

## 0.1.12 / 2025-10-10

- Enabled all previously skipped tests in `cmd/git-evars/main_test.go` and `cmd/git-replicate/main_test.go`.
  Fixed [#1](https://github.com/mslinn/git_tree_go/issues/1) and [#2](https://github.com/mslinn/git_tree_go/issues/2).
- Enhanced logger thread safety by preventing double-close of logger channel and adding `ResetLogger()` for test isolation.


## 0.1.11 / 2025-10-10

- Now condenses paths output: When environment variables are used as roots (e.g., `'$work'`),
  paths are now displayed in condensed form using the variable name. For example,
  `/mnt/f/work/CanPolitique` is shown as `$work/CanPolitique`. This makes output more concise
  and readable. Fixed [#4](https://github.com/mslinn/git_tree_go/issues/4).


## 0.1.10 / 2025-10-10

- Fixed [#3 Default roots are ignored](https://github.com/mslinn/git_tree_go/issues/3)


## 0.1.6 / 2025-10-10

- Added `git-list-executables` command to list all executables created by this package.


## 0.1.0 / 2025-10-05

- Initial release

