# Release Process

This document describes the release process for `git-tree-go`.


## Prerequisites

1. Ensure you have write access to the repository
2. Make sure all tests pass: `make test`
3. Ensure the `master` branch is up to date
4. Review and update `CHANGELOG.md` if needed


## Creating a Release

### 1. Version Tagging

Create and push a new version tag:

```bash
# For a new version (e.g., v1.2.3)
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### 2. Automated Release Process

Once you push a tag starting with `v`, GitHub Actions will automatically:

1. Run the release workflow (`.github/workflows/release.yml`)
2. Build binaries for all platforms using GoReleaser
3. Create a GitHub release with:
   - All command binaries (`git-commitAll`, `git-evars`, `git-exec`, `git-replicate`, `git-treeconfig`, `git-update`)
   - Archives for each platform (Linux, macOS, Windows)
   - Checksums file
   - Auto-generated changelog


### 3. Supported Platforms

The release process builds for:

- **Operating Systems**: Linux, macOS (Darwin), Windows
- **Architectures**: amd64 (x86_64), arm64

This produces 12 archives (6 platform/arch combinations × 2 formats):

- `.tar.gz` for Linux and macOS
- `.zip` for Windows


### 4. Verify the Release

After the workflow completes:

1. Go to the [Releases page](../../releases)
2. Verify the new release is published
3. Check that all binaries are present
4. Test download and execution on at least one platform


## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version: Incompatible API changes
- **MINOR** version: New functionality (backwards compatible)
- **PATCH** version: Bug fixes (backwards compatible)

Examples:

- `v1.0.0` - Initial stable release
- `v1.1.0` - New feature added
- `v1.1.1` - Bug fix
- `v2.0.0` - Breaking changes


## Changelog

Before creating a release, update the CHANGELOG.md file with:

- New features
- Bug fixes
- Breaking changes
- Deprecations


## Rollback

If a release has issues:

1. Delete the tag locally and remotely:

   ```bash
   git tag -d v1.2.3
   git push origin :refs/tags/v1.2.3
   ```

2. Delete the GitHub release through the web interface

3. Fix the issues and create a new release with a patch version bump


## Manual Release (Emergency)

If GitHub Actions is unavailable, you can create a release manually:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Create a tag
git tag -a v1.2.3 -m "Release v1.2.3"

# Build and release
GITHUB_TOKEN="your_token" goreleaser release --clean
```


## Troubleshooting

### Release workflow fails

1. Check the Actions tab for error details
2. Verify GoReleaser configuration: `goreleaser check`
3. Test locally: `goreleaser release --snapshot --clean`

### Missing binaries

1. Check `.goreleaser.yml` build configuration
2. Verify all commands compile: `make build`
3. Check build logs in GitHub Actions

### Wrong version in binaries

The version is set via ldflags during build. Ensure:

1. The tag follows the `v*` pattern
2. GoReleaser is substituting `{{.Version}}` correctly
