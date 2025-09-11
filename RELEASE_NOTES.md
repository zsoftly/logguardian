# LogGuardian v1.4.0 Release Notes

**Release Date:** September 11, 2025

## Overview

LogGuardian is an enterprise-grade AWS CloudWatch log compliance automation tool that ensures log groups meet security and retention requirements.

This release can be deployed as:
- Docker container via GitHub Container Registry (ghcr.io)
- AWS Lambda function for serverless deployments
- Standalone container binary

## 🚀 New Features
- feat: Add Docker container support and update documentation
- feat: Add usage information to command line arguments for LogGuardian container
- feat: Enhance CI workflow with Docker build and test steps; add pre-commit script for local checks
- feat: Implement ServiceAdapter with retry logic and rate limiting; add tests for service adapter functionality
- Add dry-run compliance service and related tests

## 🐛 Bug Fixes
- fix: Update AWS region for dry-run container execution
- fix: Improve batch size parsing and handle errors gracefully

## 🔧 Other Changes
- docs: Update containerization design document to reflect service adapter implementation details and authentication strategies
- docs: Update containerization design document with implementation status and decision dates
- docs: Update upgrade guide with detailed CLI change set review steps and parameter preservation instructions
- docs: Enhance CLI update instructions with change set review process
- docs: Update Terraform deployment examples and add warning for production validation

## Installation

### Docker (Recommended)
```bash
docker pull ghcr.io/zsoftly/logguardian:v1.4.0
```

### AWS Lambda
Download the Lambda deployment package (logguardian-compliance-v1.4.0.zip) from the release artifacts.

### Container Binary
Download the standalone container binary (logguardian-container-v1.4.0) from the release artifacts.

