# LogGuardian 1.4.1 Release Notes

**Release Date:** November 02, 2025

## Overview

LogGuardian is an enterprise-grade AWS CloudWatch log compliance automation tool that ensures log groups meet security and retention requirements.

This release can be deployed as:
- Docker container via GitHub Container Registry (ghcr.io)
- AWS Lambda function for serverless deployments
- Standalone container binary

## 🚀 New Features
- No new features in this release

## 🐛 Bug Fixes
- fix: Ensure hermetic tests by clearing and restoring AWS environment variables

## 🔧 Other Changes
- chore: Release v1.4.1

## Installation

### Docker (Recommended)
```bash
docker pull ghcr.io/zsoftly/logguardian:1.4.1
```

### AWS Lambda
Download the Lambda deployment package (logguardian-compliance-1.4.1.zip) from the release artifacts.

### Container Binary
Download the standalone container binary (logguardian-container-1.4.1) from the release artifacts.

