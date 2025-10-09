#!/bin/bash
# Release script for git-tree-go
# Usage: ./scripts/release.sh [version]
# Example: ./scripts/release.sh 1.2.3

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
info() {
  echo -e "${BLUE}ℹ${NC} $1"
}

success() {
  echo -e "${GREEN}✓${NC} $1"
}

warning() {
  echo -e "${YELLOW}⚠${NC} $1"
}

error() {
  echo -e "${RED}✗${NC} $1"
  exit 1
}

# Function to validate semantic version
validate_version() {
  local version=$1
  if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "Invalid version format: $version (expected: X.Y.Z)"
  fi
}

# Check if on main branch
check_branch() {
  local current_branch=$(git rev-parse --abbrev-ref HEAD)
  if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
    warning "You are on branch '$current_branch', not main/master"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      error "Aborted"
    fi
  fi
}

# Check for uncommitted changes
check_clean() {
  if [[ -n $(git status -s) ]]; then
    error "Working directory is not clean. Commit or stash changes first."
  fi
  success "Working directory is clean"
}

# Check if tag already exists
check_tag() {
  local version=$1
  if git rev-parse "v$version" >/dev/null 2>&1; then
    error "Tag v$version already exists"
  fi
  success "Tag v$version is available"
}

# Run tests
run_tests() {
  info "Running tests..."
  if make test:spec 2>/dev/null; then
    success "All tests passed"
  else
    error "Tests failed. Fix issues before releasing."
  fi
}

# Update version in files (if needed)
update_version_files() {
  local version=$1
  # Add version updates here if you store version in any files
  # For now, this is a placeholder
  info "Version will be v$version"
}

# Create and push tag
create_tag() {
  local version=$1
  local tag="v$version"

  info "Creating tag $tag..."
  git tag -a "$tag" -m "Release $tag"
  success "Tag $tag created"

  info "Pushing tag to origin..."
  git push origin "$tag"
  success "Tag pushed to origin"

  echo ""
  info "Release workflow has been triggered"
  info "Check progress at: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
}

# Show current version
show_current_version() {
  local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
  info "Latest tag: $latest_tag"
}

# Main script
main() {
  echo "=================================="
  echo "  git-tree-go Release Script"
  echo "=================================="
  echo ""

  # Show current version
  show_current_version
  echo ""

  # Run checks
  check_branch
  check_clean
  check_tag "$version"

  # Run tests
  run_tests

  # Get version from argument or prompt
  local version=$1
  if [[ -z "$version" ]]; then
    read -p "Enter version number (X.Y.Z): " version
  fi

  # Validate version
  validate_version "$version"
  success "Version format is valid: $version"

  # Update version files
  update_version_files "$version"

  # Confirm release
  echo ""
  warning "Ready to release v$version"
  read -p "Proceed with release? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    error "Release cancelled"
  fi

  # Create and push tag
  create_tag "$version"

  echo ""
  success "Release v$version initiated successfully!"
  echo ""
  info "Next steps:"
  echo "  1. Monitor the GitHub Actions workflow"
  echo "  2. Verify the release on GitHub"
  echo "  3. Test the release binaries"
  echo "  4. Announce the release (if applicable)"
  echo ""
}

# Run main function
main "$@"
