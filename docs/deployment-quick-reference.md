# LogGuardian Multi-Region Deployment - Quick Reference

## 🚀 New Comprehensive Script

We've replaced the simple multi-region script with a comprehensive deployment tool that keeps everything DRY and provides advanced functionality.

### ✅ What's Available Now

| Command | Description | Use Case |
|---------|-------------|----------|
| `deploy-single` | Deploy to one region | Quick testing, single-region customers |
| `deploy-multi` | Deploy to multiple regions | Multi-region deployments |
| `deploy-dev` | Development environment | Cost-optimized dev deployment |
| `deploy-staging` | Staging environment | Pre-production testing |
| `deploy-prod` | Production environment | Full monitoring and compliance |
| `deploy-customer` | Customer infrastructure | Existing KMS keys, Config rules |
| `status` | Check deployment status | Monitor deployments across regions |
| `cleanup` | Remove deployments | Clean up test environments |
| `validate` | Validate template | Pre-deployment validation |

### 🛠️ Script Location and Usage

```bash
# Main script
./scripts/logguardian-deploy.sh

# Legacy wrapper (backward compatibility)  
./scripts/deploy-multi-region.sh  # → redirects to new script
```

## 📚 Quick Examples

### Your Current Sandbox (25 Log Groups)

```bash
# Check current deployment status
./scripts/logguardian-deploy.sh status -e sandbox -r ca-central-1

# Your current deployment is: logguardian-sandbox (not logguardian-sandbox-ca-central-1)
# Dashboard URL: https://ca-central-1.console.aws.amazon.com/cloudwatch/home?region=ca-central-1#dashboards:name=LogGuardian-sandbox
```

### Development Deployments

```bash
# Cost-optimized dev environment
./scripts/logguardian-deploy.sh deploy-dev -r ca-central-1
# → 7-day retention, 3-day S3 expiration, no dashboard, 128MB Lambda

# Single region for quick testing
./scripts/logguardian-deploy.sh deploy-single -e dev -r ca-central-1
```

### Production Multi-Region

```bash
# Canadian regions
./scripts/logguardian-deploy.sh deploy-prod -r ca-central-1,ca-west-1

# North American regions  
./scripts/logguardian-deploy.sh deploy-prod -r ca-central-1,ca-west-1,us-east-1,us-west-2

# Global deployment
./scripts/logguardian-deploy.sh deploy-prod -r ca-central-1,eu-west-1,ap-southeast-1
```

### Customer Infrastructure Integration

```bash
# Customer with existing KMS key and custom tagging
./scripts/logguardian-deploy.sh deploy-customer 
  -e prod 
  -r ca-central-1 
  --existing-kms-key alias/customer-logs-ca 
  --customer-tag-prefix "ACME-LogGuardian" 
  --owner "ACME-Platform-Team"

# Multi-region with different owners per region
./scripts/logguardian-deploy.sh deploy-customer 
  -e prod 
  -r us-east-1 
  --existing-kms-key alias/sox-compliance 
  --retention-days 2555 
  --owner "Compliance-Team" 
  --product-name "SOX-LogGuardian"

./scripts/logguardian-deploy.sh deploy-customer 
  -e prod 
  -r eu-west-1 
  --existing-kms-key alias/gdpr-compliance 
  --retention-days 2190 
  --owner "GDPR-Team" 
  --product-name "GDPR-LogGuardian"
```

### Resource Tagging Examples

```bash
# Enterprise deployment with custom tagging
./scripts/logguardian-deploy.sh deploy-prod 
  -r ca-central-1,ca-west-1 
  --product-name "Enterprise-LogGuardian" 
  --owner "Platform-Engineering-Team" 
  --managed-by "SAM"

# Development with team-specific tagging
./scripts/logguardian-deploy.sh deploy-dev 
  -r ca-central-1 
  --owner "DevOps-Team" 
  --product-name "Dev-LogGuardian"
```

## 🎛️ All Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `-e, --environment` | Environment (dev/staging/prod/sandbox) | sandbox |
| `-r, --regions` | Comma-separated regions | ca-central-1,ca-west-1 |
| `--retention-days` | Log retention in days | 30 |
| `--s3-expiration-days` | S3 data expiration days | 14 |
| `--lambda-memory` | Lambda memory size (MB) | 256 |
| `--lambda-timeout` | Lambda timeout (seconds) | 900 |
| `--create-dashboard` | Create CloudWatch dashboard | true |
| `--enable-staggered` | Enable staggered scheduling | true |
| `--existing-kms-key` | Use existing KMS key alias | (none) |
| `--customer-tag-prefix` | Customer resource tag prefix | (none) |
| `--product-name` | Product name for tagging | LogGuardian |
| `--owner` | Owner/Team name for tagging | ZSoftly |
| `--managed-by` | Management tool for tagging | SAM |
| `--stack-prefix` | Stack name prefix | logguardian |
| `-v, --verbose` | Verbose output | false |

## 🏗️ Environment Defaults

### Development (`deploy-dev`)
- Retention: 7 days
- S3 Expiration: 3 days  
- Lambda Memory: 128 MB
- Dashboard: Disabled
- Cost: ~$5-10/month per region

### Staging (`deploy-staging`)  
- Retention: 30 days
- S3 Expiration: 14 days
- Lambda Memory: 256 MB
- Dashboard: Enabled
- Cost: ~$15-25/month per region

### Production (`deploy-prod`)
- Retention: 365 days
- S3 Expiration: 90 days  
- Lambda Memory: 512 MB
- Dashboard: Enabled
- Cost: ~$50-100/month per region (depends on log volume)

## 🧹 Management Commands

```bash
# Check deployment status
./scripts/logguardian-deploy.sh status -e prod

# Clean up dev environment
./scripts/logguardian-deploy.sh cleanup -e dev

# Clean up specific regions
./scripts/logguardian-deploy.sh cleanup -e sandbox -r ca-central-1

# Validate template before deployment
./scripts/logguardian-deploy.sh validate
```

## 📖 Documentation

- **Multi-Region Guide**: `docs/multi-region-deployment.md`
- **Customer Integration**: `docs/customer-infrastructure-integration.md`
- **AWS Marketplace SAM**: `docs/aws-marketplace-sam.md`

## 🔄 Migration from Old Script

The old `deploy-multi-region.sh` script is now a wrapper that redirects to the new comprehensive script:

```bash
# Old way
./scripts/deploy-multi-region.sh sandbox ca-central-1 ca-west-1

# New way (automatic conversion)
./scripts/logguardian-deploy.sh deploy-multi -e sandbox -r ca-central-1,ca-west-1
```

## 🎯 Best Practices

1. **Start Small**: Use `deploy-single` for initial testing
2. **Environment Progression**: dev → staging → prod
3. **Cost Control**: Use environment-specific commands for appropriate resource sizing
4. **Status Monitoring**: Always run `status` after deployments
5. **Template Validation**: Run `validate` before important deployments
6. **Regional Strategy**: Deploy primary region first, then expand

## 🚨 Important Notes

- **Stack Naming**: New script uses `logguardian-{environment}-{region}` format
- **Existing Deployments**: Your current `logguardian-sandbox` stack uses the old naming pattern
- **S3 Bucket Conflicts**: Each region gets its own S3 bucket with region suffix
- **KMS Keys**: Each deployment creates its own KMS key unless `--existing-kms-key` is specified
- **Config Rules**: Each deployment creates its own Config rules (future enhancement: use existing ones)
