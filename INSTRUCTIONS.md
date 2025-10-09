1) Examine the RSpec unit tests in /mnt/f/work/git/git_tree/spec then
   generate similar tests for the five commands (in the /cmd directory) in this project.

2) The [filepath-securejoin](https://github.com/cyphar/filepath-securejoin) docs say:

    SecureJoin (and SecureJoinVFS) are still provided by this library to support legacy users,
    but new users are strongly suggested to avoid using SecureJoin and instead use the new api
    or switch to `libpathrs`.

5) The git-treeconfig help output should substitute '$HOME' for `/home/user/`.
  For example, the following displays the current dialog:

   ```shell
   $ git-treeconfig
   Welcome to git-tree configuration.
   This utility will help you create a configuration file at: /home/mslinn/.treeconfig.yml
   Press Enter to accept the default value in brackets.

   Git command timeout in seconds? |300| 600
   Default verbosity level (0=quiet, 1=normal, 2=verbose)? |1|
   Default root directories (space-separated)? |sites sitesUbuntu work| dev projects

   Configuration saved to /home/mslinn/.treeconfig.yml
   ```

   Instead, the dialog should display as:

   ```shell
   $ git-treeconfig
   Welcome to git-tree configuration.
   This utility will help you create a configuration file at: $HOME/.treeconfig.yml
   Press Enter to accept the default value in brackets.

   Git command timeout in seconds? |300| 600
   Default verbosity level (0=quiet, 1=normal, 2=verbose)? |1|
   Default root directories (space-separated)? |sites sitesUbuntu work| dev projects

   Configuration saved to $HOME/.treeconfig.yml
   ```

6) Run the unit tests.
