# LogGuardian v1.4.0 Release Notes

**Release Date:** September 11, 2025

LogGuardian is an enterprise-grade AWS CloudWatch log compliance automation tool that ensures log groups meet security and retention requirements.

This release can be deployed as:
- Docker container via GitHub Container Registry (ghcr.io)
- AWS Lambda function for serverless deployments
- Standalone container binary

## What's New

### Docker Container Support
- Deploy LogGuardian as a Docker container on ECS, Kubernetes, or standalone
- Multi-architecture support (amd64/arm64)
- Published to GitHub Container Registry at `ghcr.io/zsoftly/logguardian`

### Service Adapter Implementation
- Retry logic with exponential backoff for AWS API calls
- Rate limiting to prevent throttling
- Multiple authentication strategies with fallback
- Thread-safe operations

### Developer Improvements
- Pre-commit validation script
- Enhanced CI/CD pipeline with Docker builds
- Improved error handling and CLI help

## Bug Fixes
- Fixed AWS region handling in dry-run mode
- Improved batch size parsing and validation

## Installation

### Docker
```bash
docker pull ghcr.io/zsoftly/logguardian:v1.4.0
```

### Lambda
Download `logguardian-compliance-v1.4.0.zip` from release artifacts.

### Binary
Download `logguardian-container-v1.4.0` from release artifacts.

## Quick Start

```bash
# Docker with dry-run
docker run --rm \
  -e AWS_REGION=ca-central-1 \
  -e CONFIG_RULE_NAME=logguardian-log-retention \
  -e DRY_RUN=true \
  ghcr.io/zsoftly/logguardian:v1.4.0

# Lambda deployment
aws lambda update-function-code \
  --function-name logguardian-compliance \
  --zip-file fileb://logguardian-compliance-v1.4.0.zip
```

## Documentation

- [Docker Usage Guide](docs/docker-usage.md)
- [Architecture Overview](docs/architecture-overview.md)