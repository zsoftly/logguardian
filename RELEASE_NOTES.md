# LogGuardian 1.4.2 Release Notes

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
- fix: Update VERSION file format to include 'v' prefix for consistency with git tags and AWS SAR requirements
- fix: Remove v prefix from VERSION file for SAR compatibility

## 🔧 Other Changes
- chore: Bump version to v1.4.2 and update documentation for deployment guide
- docs: Auto-generate release documentation for 1.4.1

## Installation

### Docker (Recommended)
```bash
docker pull ghcr.io/zsoftly/logguardian:1.4.2
```

### AWS Lambda
Download the Lambda deployment package (logguardian-compliance-1.4.2.zip) from the release artifacts.

### Container Binary
Download the standalone container binary (logguardian-container-1.4.2) from the release artifacts.

