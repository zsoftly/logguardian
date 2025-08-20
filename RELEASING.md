# Release Process

This document describes the release process for LogGuardian.

## Prerequisites

- AWS CLI configured with access to both dev and prod accounts
- GitHub CLI (`gh`) installed for creating releases
- Go 1.24+ installed
- SAM CLI installed
- Push access to the repository
- Main branch must be up to date

## Release Workflow

**Important:** We use a release branch workflow. Never push directly to main. All releases go through a PR.

## Release Steps

### 1. Create Release Branch

Start from an up-to-date main branch:

```bash
git checkout main
git pull origin main
git checkout -b release/X.Y.Z  # Example: release/1.2.0 (no 'v' prefix)
```

**Note:** Branch naming convention is `release/X.Y.Z` without 'v' prefix

### 2. Update Version

Use the helper script to update the version:

```bash
./scripts/update-version.sh X.Y.Z
```

Or manually:
```bash
echo "X.Y.Z" > VERSION
sed -i 's/SemanticVersion: .*/SemanticVersion: X.Y.Z/' template.yaml
```

### 3. Build and Test

```bash
make clean
make build
make test
make package
```

### 4. Commit Changes

```bash
git add -A
git commit -m "chore: Release version X.Y.Z

Brief description of changes"
```

### 5. Push Release Branch

```bash
# Push the release branch
git push origin release/X.Y.Z
```

### 6. Create and Push Tag from Release Branch

While still on the release branch, create and push the tag:

```bash
# Still on release/X.Y.Z branch
# Create and push tag (MUST start with 'v' to trigger release pipeline)
git tag -a vX.Y.Z -m "Release version X.Y.Z

- List major changes
- Include any breaking changes
- Reference issues fixed"

git push origin vX.Y.Z
```

**IMPORTANT:** 
- Tags MUST start with 'v' (e.g., `v1.2.0`) to trigger the release pipeline
- The tag push automatically triggers GitHub Actions to create a release with artifacts
- Tag is created from the release branch, NOT from main

### 7. Publish to AWS SAR

From the release branch, first deploy to dev for testing, then to prod:

```bash
# Still on release/X.Y.Z branch

# First authenticate to dev account and deploy for testing
# Set your AWS credentials for the dev account
make publish

# Test the application in dev account
# Verify functionality and configuration

# If all tests pass, authenticate to prod account where public SAR is located
# Set your AWS credentials for the prod account
make publish
```

### 8. Create PR to Main

After the release is complete (tag pushed, SAR published), create a Pull Request to main:

```bash
# Create PR using GitHub CLI
gh pr create --title "Release vX.Y.Z" --body "## Summary
- Release vX.Y.Z completed
- Tag already created and pushed
- SAR application already published
- GitHub Actions release pipeline completed

## Changes
- List of major changes
- Any breaking changes
- Referenced issues

## Release Artifacts
- GitHub Release: Created via automated pipeline
- SAR Application: Published to production"
```

Or use GitHub UI (github.com) to create the PR.

### 9. Merge PR

After review and approval:
- Merge the PR to main (via GitHub UI or CLI)
- **Note:** Keep the release branch for historical reference (do NOT delete)
- This merge brings the version changes to main for consistency

### 10. Verify Release

- Check GitHub Actions for successful release pipeline
- Verify GitHub release was created with artifacts
- Confirm SAR application is updated in AWS Console
- Verify release branch remains for historical reference

## Naming Conventions

### Branches
- Release branches: `release/X.Y.Z` (NO 'v' prefix)
- Example: `release/1.2.0`

### Tags
- Tags: `vX.Y.Z` (MUST have 'v' prefix)
- Example: `v1.2.0`
- **Only tags starting with 'v' trigger the release pipeline**

### Version Numbering

We follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

## Quick Release Command Summary

For a standard release (example with version 1.2.0):

```bash
# 1. Create release branch from main (no 'v' prefix)
git checkout main && git pull origin main
git checkout -b release/1.2.0

# 2. Update version and test
./scripts/update-version.sh 1.2.0
make clean build test package

# 3. Commit and push release branch
git add -A
git commit -m "chore: Release version 1.2.0"
git push origin release/1.2.0

# 4. Create and push tag from release branch (MUST have 'v' prefix)
# Still on release/1.2.0 branch
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0  # This triggers the release pipeline

# 5. Publish to SAR from release branch
# First authenticate to dev account and test
make publish
# Verify in dev account, then authenticate to prod account
make publish

# 6. Create PR to main (after release is complete)
gh pr create --title "Release v1.2.0" --body "Release completed"

# 7. Wait for review and merge to main

# Note: Release branches are kept for historical reference - do NOT delete
```

## Rollback Process

If a release needs to be rolled back:

1. Delete the tag locally and remotely:
```bash
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
```

2. Fix the issues
3. Create a new patch version
4. Follow the normal release process

## Notes

- The VERSION file is the single source of truth for versioning
- Never use default/fallback versions - VERSION file is required
- All version references in Makefiles read from VERSION file
- Documentation uses generic "latest" references to avoid updates
- Release branches (`release/*`) are kept permanently for audit trail
- Each release has a corresponding branch showing exactly what was released