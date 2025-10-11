# Change Log

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

