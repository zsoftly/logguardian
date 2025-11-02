# LogGuardian 1.4.3 Release Notes

**Release Date:** November 02, 2025

## Overview

LogGuardian is an enterprise-grade AWS CloudWatch log compliance automation tool that ensures log groups meet security and retention requirements.

This release can be deployed as:
- Docker container via GitHub Container Registry (ghcr.io)
- AWS Lambda function for serverless deployments
- Standalone container binary

## 🎯 Highlights

**Version Format Standardization**: This release adopts pure semantic versioning (X.Y.Z) across all systems, removing the "v" prefix to align with industry standards.

## ⚠️ Breaking Changes

### Version Format Migration

**IMPORTANT**: Starting with this release, LogGuardian uses pure semantic versioning without the "v" prefix.

**Before (v1.4.0, v1.4.1, v1.4.2):**
- Git tags: `v1.4.2`
- Docker tags: `1.4.2` (v stripped by tooling)
- AWS SAR: `1.4.2` (required format)
- Conversion needed in multiple places

**After (1.4.3+):**
- Git tags: `1.4.3` (no v prefix)
- Docker tags: `1.4.3` (consistent)
- AWS SAR: `1.4.3` (consistent)
- No conversion needed - single format everywhere

**Migration Impact:**
- Historical tags (v1.4.0, v1.4.1, v1.4.2) remain unchanged
- New tags use format: `1.4.3`, `1.4.4`, etc.
- Docker pull commands: `docker pull ghcr.io/zsoftly/logguardian:1.4.3`

**Why This Change:**
- Aligns with Semantic Versioning specification (semver.org)
- Matches Docker/OCI container standards
- Eliminates AWS SAR conversion requirements
- Simplifies automation and reduces complexity

## 🔧 Changes

### Version Handling Improvements
- Standardized VERSION file format to X.Y.Z (pure semantic versioning)
- Updated GitHub Actions workflow trigger: `[0-9]+.[0-9]+.[0-9]+`
- Simplified Makefile: removed all `sed 's/^v//'` conversions
- Simplified update-version.sh script: single format, no conversions
- Simplified release documentation generator script

### Documentation Improvements
- Enhanced RELEASING.md with automated workflow documentation
- Added critical warnings for auto-generated documentation pull step
- Updated RELEASE_PROCESS.md to remove GitHub CLI dependencies
- Removed all `gh` command references throughout documentation
- Replaced with git commands and GitHub web UI instructions

## 🚀 New Features
- No new features in this release

## 🐛 Bug Fixes
- No bug fixes in this release

## Installation

### Docker (Recommended)
```bash
# Use the new format without v prefix
docker pull ghcr.io/zsoftly/logguardian:1.4.3
docker pull ghcr.io/zsoftly/logguardian:latest
```

### AWS Lambda
Download the Lambda deployment package (logguardian-compliance-1.4.3.zip) from the release artifacts.

### Container Binary
Download the standalone container binary (logguardian-container-1.4.3) from the release artifacts.

## Upgrade Notes

If you have automation or scripts referencing LogGuardian versions:
- Update tag references from `v1.x.x` to `1.x.x` format
- Update Docker pull commands to use version without v prefix
- Update any git tag operations to use new format

## Documentation

- [Release Process](RELEASING.md)
- [Docker Usage](docs/docker-usage.md)
- [Architecture Overview](docs/architecture-overview.md)

