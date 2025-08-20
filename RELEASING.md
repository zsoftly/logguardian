# Release Process

This document describes the release process for LogGuardian.

## Prerequisites

- AWS CLI configured with `logprod` profile for SAR publishing
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

### 5. Push Release Branch and Create PR

```bash
# Push the release branch
git push origin release/X.Y.Z
```

Create a Pull Request to main using either:
- GitHub UI (github.com)
- GitHub CLI (`gh pr create`)

Include in the PR description:
- List of major changes
- Any breaking changes
- Referenced issues
- Confirmation that tests pass and build succeeds

### 6. Merge PR

After review and approval:
- Merge the PR to main (via GitHub UI or CLI)
- **Note:** Keep the release branch for historical reference (do NOT delete)

### 7. Create and Push Tag from Main

After PR is merged, create the tag from main:

```bash
# Switch to main and pull latest
git checkout main
git pull origin main

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

### 8. Publish to AWS SAR

From the main branch with the merged release:

```bash
# Ensure you're on main with latest changes
git checkout main
git pull origin main

# Publish to SAR
export AWS_PROFILE=logprod
export AWS_DEFAULT_REGION=ca-central-1
make publish
```

### 9. Verify Release

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

# 4. Create PR (via GitHub UI or CLI)
# 5. Wait for review and merge

# 6. After PR is merged, tag from main (MUST have 'v' prefix)
git checkout main && git pull origin main
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0  # This triggers the release pipeline

# 7. Publish to SAR
export AWS_PROFILE=logprod
make publish

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