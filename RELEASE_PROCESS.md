# LogGuardian Release Process

## Release Overview

Streamlined release process for LogGuardian with automated changelog generation, Docker image publishing, and AWS SAR deployment.

## Release Types

- **Major Release (v1.0.0)**: Breaking changes or major features
- **Minor Release (v0.2.0)**: New features, backwards compatible
- **Patch Release (v0.1.1)**: Bug fixes and minor improvements

## Release Process

### 1. Create Release Branch
```bash
# Ensure you're on the latest main
git checkout main
git pull origin main

# Create release branch
git checkout -b release/v<version>
git push -u origin release/v<version>
```

### 2. Wait for Auto-Generated Documentation
- GitHub Actions automatically detects `release/*` branch creation
- Runs `scripts/01_release_docs_generator.sh` to:
  - Generate CHANGELOG.md from commit history
  - Create RELEASE_NOTES.md with installation instructions
  - Update VERSION file
  - Update version strings in source files
- **Wait for workflow completion before proceeding!**
- Check Actions tab: "Auto-Generate Release Documentation" workflow

### 3. Pull and Review Auto-Generated Changes
```bash
# Pull the auto-generated documentation
git pull origin release/v<version>

# Review generated files
cat CHANGELOG.md | head -50
cat RELEASE_NOTES.md
cat VERSION

# Verify version updates in code
grep "version" cmd/lambda/main.go cmd/container/main.go
```

### 4. Manual Updates (if needed)
```bash
# Edit release notes if needed
vim RELEASE_NOTES.md

# Commit any manual changes
git add .
git commit -m "docs: Finalize release notes for v<version>"
git push origin release/v<version>
```

### 5. Create Release Tag
```bash
# Create and push tag - triggers release workflows
git tag -a v<version> -m "Release v<version>"
git push origin v<version>
```

This triggers:
- Lambda function build and packaging
- Docker image build and push to ghcr.io
- AWS SAR application update
- GitHub Release creation with artifacts

### 6. Verify Release Artifacts
Wait for all workflows to complete, then verify:

```bash
# Check Docker image
docker pull ghcr.io/zsoftly/logguardian:v<version>
docker run --rm ghcr.io/zsoftly/logguardian:v<version> --help

# Check GitHub Release
gh release view v<version>

# Check AWS SAR (if applicable)
aws serverlessrepo get-application \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --region ca-central-1
```

### 7. Merge Back to Main
```bash
# After successful release
git checkout main
git merge --no-ff release/v<version>
git push origin main

# Delete release branch
git push origin --delete release/v<version>
git branch -d release/v<version>
```

## Automated Release Artifacts

### Docker Images (ghcr.io)
Published automatically with tags:
- `ghcr.io/zsoftly/logguardian:latest` (main branch)
- `ghcr.io/zsoftly/logguardian:v<version>` (release tag)
- `ghcr.io/zsoftly/logguardian:<major>` (e.g., v1)
- `ghcr.io/zsoftly/logguardian:<major>.<minor>` (e.g., v1.2)

### Lambda Deployment Package
- `logguardian-compliance.zip` attached to GitHub Release
- Contains compiled Lambda binary

### Container Binary
- `logguardian-container` binary attached to GitHub Release
- Standalone container executable

## Release Checklist

### Pre-Release
- [ ] All tests passing: `make test`
- [ ] Security scan clean: `make security-scan`
- [ ] Linting passed: `golangci-lint run`
- [ ] Documentation updated
- [ ] CLAUDE.md updated with any new patterns

### Release Steps
- [ ] Create release branch: `git checkout -b release/v<version>`
- [ ] Push branch: `git push -u origin release/v<version>`
- [ ] **Wait for auto-generation workflow**
- [ ] Pull changes: `git pull origin release/v<version>`
- [ ] Review CHANGELOG.md and RELEASE_NOTES.md
- [ ] Create tag: `git tag -a v<version> -m "Release v<version>"`
- [ ] Push tag: `git push origin v<version>`
- [ ] **Wait for release workflows to complete**
- [ ] Verify Docker image: `docker pull ghcr.io/zsoftly/logguardian:v<version>`
- [ ] Verify GitHub Release artifacts
- [ ] Merge to main: `git checkout main && git merge --no-ff release/v<version>`
- [ ] Push main: `git push origin main`
- [ ] Delete release branch: `git push origin --delete release/v<version>`

### Post-Release
- [ ] Announce release in discussions/social media
- [ ] Update any dependent projects
- [ ] Monitor for issues

## Emergency Hotfix Process

For critical issues that can't wait:

```bash
# Create hotfix from last release
git checkout v<last-version>
git checkout -b hotfix/v<hotfix-version>

# Apply minimal fix
# ... make changes ...
git commit -m "fix: Critical issue description"

# Create hotfix tag
git tag -a v<hotfix-version> -m "Hotfix v<hotfix-version>"
git push origin v<hotfix-version>

# After release, merge to main
git checkout main
git merge --no-ff hotfix/v<hotfix-version>
git push origin main
```

## Version Numbering

Follow Semantic Versioning (SemVer):
- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- Increment MAJOR for breaking changes
- Increment MINOR for new features (backwards compatible)
- Increment PATCH for bug fixes

## Troubleshooting

### Build Failures
```bash
# Check workflow logs
gh run view --log

# Test locally
make build
make test
make docker-build
```

### Docker Push Issues
```bash
# Verify GitHub Container Registry access
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Build and test locally
docker build -t test-build .
docker run --rm test-build --help
```

### AWS SAR Update Issues
```bash
# Verify SAR application
aws serverlessrepo get-application \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --region ca-central-1

# Test SAM template locally
sam validate
sam build
sam local start-lambda
```

## Rollback Procedure

If issues are discovered post-release:

```bash
# Revert Docker latest tag to previous version
docker pull ghcr.io/zsoftly/logguardian:v<previous-version>
docker tag ghcr.io/zsoftly/logguardian:v<previous-version> ghcr.io/zsoftly/logguardian:latest
docker push ghcr.io/zsoftly/logguardian:latest

# Create GitHub announcement
gh release edit v<version> --notes "⚠️ This release has been superseded. Please use v<previous-version>"

# Fix issues and create new patch release
```

## Release Communication

### Release Notes Template
```markdown
## LogGuardian v<version>

### 🎉 Highlights
- Major feature or improvement
- Performance enhancement
- Security update

### 🐛 Bug Fixes
- Fixed issue #123
- Resolved problem with X

### 📦 Installation
```bash
# Docker
docker pull ghcr.io/zsoftly/logguardian:v<version>

# Lambda
Download from GitHub Release artifacts
```

### 📚 Documentation
See updated docs at [docs/](docs/)

### 🙏 Contributors
Thanks to all contributors!
```

## Automation Scripts Location

- `.github/workflows/release.yml` - Main release workflow (builds Lambda, Container, publishes Docker, creates GitHub Release)
- `.github/workflows/generate-docs.yml` - Auto-generates documentation when release branch is created
- `.github/workflows/ci.yml` - CI pipeline (tests, linting, security, Docker build testing)
- `scripts/01_release_docs_generator.sh` - Generates CHANGELOG.md, RELEASE_NOTES.md, and updates versions