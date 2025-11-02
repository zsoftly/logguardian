# Release Process

This document describes the release process for LogGuardian.

## Prerequisites

- AWS CLI configured with access to both dev and prod accounts
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

**⚠️ CRITICAL: After pushing, automated documentation generation will run!**

### 6. Wait for Automated Documentation Generation

GitHub Actions will automatically:
- Detect the `release/*` branch
- Run `scripts/01_release_docs_generator.sh`
- Generate/update CHANGELOG.md and RELEASE_NOTES.md
- Commit and push these changes back to your release branch

**You MUST wait for this workflow to complete before proceeding!**

Check workflow status:
- Navigate to GitHub repository in browser
- Click "Actions" tab
- Look for "Auto-Generate Release Documentation" workflow
- Wait for it to complete (green checkmark)

### 7. Pull and Review Auto-Generated Documentation

**CRITICAL:** Pull the auto-generated changes:

```bash
# Pull the automated commits
git pull origin release/X.Y.Z

# Review the generated documentation
cat CHANGELOG.md | head -50
cat RELEASE_NOTES.md

# Verify VERSION was updated correctly
cat VERSION
```

### 8. Manual Edits (Optional)

If you need to make manual edits to the release notes:

```bash
# Edit release notes
vim RELEASE_NOTES.md

# Commit and push your changes
git add RELEASE_NOTES.md
git commit -m "docs: Finalize release notes for X.Y.Z"
git push origin release/X.Y.Z
```

### 9. Create and Push Tag from Release Branch

**NOW** you can create the tag. While still on the release branch:

```bash
# Still on release/X.Y.Z branch
# Create and push tag (semantic versioning format X.Y.Z)
git tag -a X.Y.Z -m "Release version X.Y.Z

- List major changes
- Include any breaking changes
- Reference issues fixed"

git push origin X.Y.Z
```

**IMPORTANT:**
- Tags use pure semantic versioning format (e.g., `1.2.0`, no "v" prefix)
- The tag push automatically triggers GitHub Actions to create a release with artifacts
- Tag is created from the release branch, NOT from main

### 10. Publish to AWS SAR

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

### 11. Create PR to Main

After the release is complete (tag pushed, SAR published), create a Pull Request to main:

**Using GitHub Web UI:**
1. Navigate to your repository on github.com
2. Click "Pull requests" tab
3. Click "New pull request"
4. Set base: `main`, compare: `release/X.Y.Z`
5. Click "Create pull request"
6. Title: `Release X.Y.Z`
7. Body:
```
## Summary
- Release X.Y.Z completed
- Tag already created and pushed
- SAR application already published
- GitHub Actions release pipeline completed

## Changes
- List of major changes
- Any breaking changes
- Referenced issues

## Release Artifacts
- GitHub Release: Created via automated pipeline
- SAR Application: Published to production
```
8. Click "Create pull request"

### 12. Merge PR

After review and approval:
- Merge the PR to main (via GitHub UI or CLI)
- **Note:** Keep the release branch for historical reference (do NOT delete)
- This merge brings the version changes to main for consistency

### 13. Verify Release

- Check GitHub Actions for successful release pipeline
- Verify GitHub release was created with artifacts
- Confirm SAR application is updated in AWS Console
- Verify release branch remains for historical reference

## Naming Conventions

### Branches
- Release branches: `release/X.Y.Z`
- Example: `release/1.2.0`

### Tags
- Tags: `X.Y.Z` (pure semantic versioning)
- Example: `1.2.0`
- **Tags matching pattern `[0-9]+.[0-9]+.[0-9]+` trigger the release pipeline**

### Version Numbering

We follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

## Quick Release Command Summary

For a standard release (example with version 1.2.0):

```bash
# 1. Create release branch from main
git checkout main && git pull origin main
git checkout -b release/1.2.0

# 2. Update version and test
./scripts/update-version.sh 1.2.0
make clean build test package

# 3. Commit and push release branch
git add -A
git commit -m "chore: Release version 1.2.0"
git push origin release/1.2.0

# 4. ⚠️ CRITICAL: Wait for automated documentation generation
# Check GitHub Actions tab in browser:
# https://github.com/your-org/logguardian/actions

# 5. Pull auto-generated CHANGELOG.md and RELEASE_NOTES.md
git pull origin release/1.2.0

# 6. Review generated documentation
cat CHANGELOG.md | head -50
cat RELEASE_NOTES.md

# 7. (Optional) Make manual edits to release notes
vim RELEASE_NOTES.md
git add RELEASE_NOTES.md
git commit -m "docs: Finalize release notes"
git push origin release/1.2.0

# 8. Create and push tag from release branch (semantic versioning)
# Still on release/1.2.0 branch
git tag -a 1.2.0 -m "Release version 1.2.0"
git push origin 1.2.0  # This triggers the release pipeline

# 9. Publish to SAR from release branch
# First authenticate to dev account and test
make publish
# Verify in dev account, then authenticate to prod account
make publish

# 10. Create PR to main (after release is complete)
# Use GitHub UI: github.com -> Pull requests -> New pull request
# base: main, compare: release/1.2.0

# 11. Wait for review and merge to main (via GitHub UI)

# Note: Release branches are kept for historical reference - do NOT delete
```

## Rollback Process

If a release needs to be rolled back:

1. Delete the tag locally and remotely:
```bash
git tag -d X.Y.Z
git push origin :refs/tags/X.Y.Z
```

2. Fix the issues
3. Create a new patch version
4. Follow the normal release process

## Notes

- The VERSION file is the single source of truth for versioning
- VERSION file format: **X.Y.Z** (pure semantic versioning, no "v" prefix)
  - Git tags use same format: `1.4.2` (not `v1.4.2`)
  - Consistent across all systems: Git, Docker, AWS SAR, Makefiles
  - No conversion or transformation needed
- Never use default/fallback versions - VERSION file is required
- All version references in Makefiles read directly from VERSION file
- Documentation uses generic "latest" references to avoid updates
- Release branches (`release/*`) are kept permanently for audit trail
- Each release has a corresponding branch showing exactly what was released

## Version Format Migration

**Historical Note:** Prior to v1.4.2, tags used "v" prefix (v1.4.0, v1.4.1). Starting from 1.4.2, we adopted pure semantic versioning (X.Y.Z) to align with Docker/OCI and AWS SAR standards.