# LogGuardian v1.4.1 Release Notes

**Release Date:** November 2, 2025

LogGuardian is an enterprise-grade AWS CloudWatch log compliance automation tool that ensures log groups meet security and retention requirements.

This release can be deployed as:
- Docker container via GitHub Container Registry (ghcr.io)
- AWS Lambda function for serverless deployments
- Standalone container binary

## Overview

This patch release updates AWS SDK dependencies to their latest versions and improves test reliability across different development environments.

## 🐛 Bug Fixes

- **Hermetic Tests**: Fixed test isolation by ensuring AWS environment variables (AWS_REGION, AWS_DEFAULT_REGION) are properly cleared and restored during test execution, preventing environment-dependent test failures

## 🔧 Dependency Updates

Updated AWS SDK packages to latest stable versions:
- aws-lambda-go: v1.47.0 → v1.50.0
- aws-sdk-go-v2: v1.38.0 → v1.39.5
- aws-sdk-go-v2/config: v1.26.1 → v1.31.16
- aws-sdk-go-v2/service/cloudwatch: v1.48.0 → v1.52.0
- aws-sdk-go-v2/service/cloudwatchlogs: v1.29.0 → v1.58.6
- aws-sdk-go-v2/service/kms: v1.27.0 → v1.47.0
- aws-sdk-go-v2/service/configservice: v1.54.0 → v1.59.1
- aws-sdk-go-v2/service/sts: v1.26.5 → v1.39.0

These updates provide improved performance, bug fixes, and security enhancements from the AWS SDK.

## Installation

### Docker (Recommended)
```bash
docker pull ghcr.io/zsoftly/logguardian:v1.4.1
docker pull ghcr.io/zsoftly/logguardian:latest
```

### AWS Lambda
Download `logguardian-compliance-v1.4.1.zip` from the [release artifacts](https://github.com/zsoftly/logguardian/releases/tag/v1.4.1).

### Container Binary
Download `logguardian-container-v1.4.1` from the [release artifacts](https://github.com/zsoftly/logguardian/releases/tag/v1.4.1).

## Quick Start

```bash
# Docker with dry-run
docker run --rm \
  -e AWS_REGION=ca-central-1 \
  -e CONFIG_RULE_NAME=cw-lg-retention-min \
  -e DRY_RUN=true \
  ghcr.io/zsoftly/logguardian:v1.4.1

# Lambda deployment
aws lambda update-function-code \
  --function-name logguardian-compliance \
  --zip-file fileb://logguardian-compliance-v1.4.1.zip
```

## Upgrade Notes

This release is fully backward compatible with v1.4.0. No configuration changes are required.

## Documentation

- [Docker Usage Guide](docs/docker-usage.md)
- [Architecture Overview](docs/architecture-overview.md)
- [CLAUDE.md - Project Instructions](CLAUDE.md)

## Full Changelog

See [CHANGELOG.md](CHANGELOG.md) for complete release history.
