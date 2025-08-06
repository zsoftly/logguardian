# Multi-Region Deployment Guide for LogGuardian

LogGuardian supports multi-region deployments through the comprehensive deployment script `scripts/logguardian-deploy.sh`. This guide shows you how to deploy across multiple regions efficiently.

## Overview

LogGuardian uses regional AWS services (Config, CloudWatch Logs, KMS), so each region requires its own deployment. The deployment script provides functions to handle common deployment patterns and keep your configurations DRY.

## Deployment Script Features

✅ **Pre-built deployment patterns**: Dev, staging, prod, and customer infrastructure  
✅ **Parameter validation**: Template validation before deployment  
✅ **Multi-region orchestration**: Deploy to multiple regions in sequence  
✅ **Status monitoring**: Check deployment status across all regions  
✅ **Cost optimization**: Environment-specific resource sizing  
✅ **Customer integration**: Support for existing KMS keys and Config rules

## Quick Start

### Common Deployment Commands

```bash
# Deploy to sandbox (development testing)
./scripts/logguardian-deploy.sh deploy-multi -e sandbox -r ca-central-1,ca-west-1

# Deploy to multiple North American regions
./scripts/logguardian-deploy.sh deploy-multi -e prod -r ca-central-1,ca-west-1,us-east-1,us-west-2

# Deploy to global regions
./scripts/logguardian-deploy.sh deploy-multi -e prod -r ca-central-1,eu-west-1,ap-southeast-1

# Deploy single region for testing
./scripts/logguardian-deploy.sh deploy-single -e sandbox -r ca-central-1
```

### Environment-Specific Deployments

```bash
# Development environment (cost-optimized)
./scripts/logguardian-deploy.sh deploy-dev -r ca-central-1

# Staging environment 
./scripts/logguardian-deploy.sh deploy-staging -r ca-central-1,ca-west-1

# Production environment (full monitoring)
./scripts/logguardian-deploy.sh deploy-prod -r ca-central-1,ca-west-1,us-east-1,eu-west-1
```

## Deployment Status and Monitoring

```bash
# Check status across all regions
./scripts/logguardian-deploy.sh status -e prod -r ca-central-1,ca-west-1,us-east-1

# Check sandbox deployment status
./scripts/logguardian-deploy.sh status -e sandbox

# Validate template before deployment
./scripts/logguardian-deploy.sh validate
```

## Customer Infrastructure Integration

### Using Existing KMS Keys

```bash
# Deploy with customer's existing KMS key
./scripts/logguardian-deploy.sh deploy-customer \
  -e prod \
  -r ca-central-1 \
  --existing-kms-key alias/customer-logs-ca-central-1 \
  --customer-tag-prefix "ACME-LogGuardian"

# Multi-region with different KMS keys per region
./scripts/logguardian-deploy.sh deploy-customer \
  -e prod \
  -r ca-central-1 \
  --existing-kms-key alias/customer-logs-canada
  
./scripts/logguardian-deploy.sh deploy-customer \
  -e prod \
  -r ca-west-1 \
  --existing-kms-key alias/customer-logs-canada-west
```

### Custom Retention Policies

```bash
# SOX compliance (7 years)
./scripts/logguardian-deploy.sh deploy-multi \
  -e compliance \
  -r us-east-1 \
  --retention-days 2555 \
  --existing-kms-key alias/customer-sox-logs

# GDPR compliance (6 years)  
./scripts/logguardian-deploy.sh deploy-multi \
  -e compliance \
  -r eu-west-1 \
  --retention-days 2190 \
  --existing-kms-key alias/customer-gdpr-logs \
  --create-dashboard false

# PIPEDA compliance (5 years)
./scripts/logguardian-deploy.sh deploy-multi \
  -e compliance \
  -r ca-central-1 \
  --retention-days 1825
```

## Advanced Deployment Scenarios

### Cost-Optimized Deployment

```bash
# Primary region with full monitoring
./scripts/logguardian-deploy.sh deploy-multi \
  -e prod \
  -r ca-central-1 \
  --create-dashboard true \
  --s3-expiration-days 90 \
  --lambda-memory 512

# Secondary regions with cost optimization
./scripts/logguardian-deploy.sh deploy-multi \
  -e prod \
  -r ca-west-1,us-east-1 \
  --create-dashboard false \
  --s3-expiration-days 30 \
  --lambda-memory 256
```

### Compliance-Driven Multi-Region

```bash
# Script to deploy different compliance requirements per region
#!/bin/bash

# US regions - SOX compliance (7 years)
./scripts/logguardian-deploy.sh deploy-multi \
  -e compliance-us \
  -r us-east-1,us-west-2 \
  --retention-days 2555 \
  --existing-kms-key alias/sox-compliance-logs

# EU regions - GDPR compliance (6 years)
./scripts/logguardian-deploy.sh deploy-multi \
  -e compliance-eu \
  -r eu-west-1,eu-central-1 \
  --retention-days 2190 \
  --existing-kms-key alias/gdpr-compliance-logs

# Canadian regions - PIPEDA compliance (5 years)
./scripts/logguardian-deploy.sh deploy-multi \
  -e compliance-ca \
  -r ca-central-1,ca-west-1 \
  --retention-days 1825 \
  --existing-kms-key alias/pipeda-compliance-logs
```

## Cleanup and Management

```bash
# Clean up sandbox deployments
./scripts/logguardian-deploy.sh cleanup -e sandbox -r ca-central-1,ca-west-1

# Clean up development environment  
./scripts/logguardian-deploy.sh cleanup -e dev

# Clean up specific regions
./scripts/logguardian-deploy.sh cleanup -e staging -r ca-central-1
```

## Script Reference

### All Available Commands

```bash
# Show help and all available options
./scripts/logguardian-deploy.sh help

# Deploy commands
./scripts/logguardian-deploy.sh deploy-single   # Deploy to single region
./scripts/logguardian-deploy.sh deploy-multi    # Deploy to multiple regions
./scripts/logguardian-deploy.sh deploy-dev      # Development environment
./scripts/logguardian-deploy.sh deploy-staging  # Staging environment
./scripts/logguardian-deploy.sh deploy-prod     # Production environment
./scripts/logguardian-deploy.sh deploy-customer # Customer infrastructure

# Management commands
./scripts/logguardian-deploy.sh status          # Show deployment status
./scripts/logguardian-deploy.sh cleanup         # Clean up deployments  
./scripts/logguardian-deploy.sh validate        # Validate template
```

### Common Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `-e, --environment` | Environment (dev/staging/prod/sandbox) | sandbox |
| `-r, --regions` | Comma-separated regions | ca-central-1,ca-west-1 |
| `--retention-days` | Log retention in days | 30 |
| `--s3-expiration-days` | S3 data expiration days | 14 |
| `--lambda-memory` | Lambda memory size (MB) | 256 |
| `--create-dashboard` | Create CloudWatch dashboard | true |
| `--existing-kms-key` | Use existing KMS key alias | (none) |
| `--customer-tag-prefix` | Customer resource tag prefix | (none) |

## Examples for Your Current Setup

### Deploy to Your Sandbox (Current Use Case)

```bash
# Deploy to your sandbox with 25 log groups
./scripts/logguardian-deploy.sh deploy-single -e sandbox -r ca-central-1

# Check deployment status
./scripts/logguardian-deploy.sh status -e sandbox

# Clean up when done testing
./scripts/logguardian-deploy.sh cleanup -e sandbox
```

### Prepare for Production Multi-Region

```bash
# Test staging deployment first
./scripts/logguardian-deploy.sh deploy-staging -r ca-central-1,ca-west-1

# Deploy production when ready
./scripts/logguardian-deploy.sh deploy-prod -r ca-central-1,ca-west-1,us-east-1

# Monitor status
./scripts/logguardian-deploy.sh status -e prod
```
