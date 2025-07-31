# LogGuardian CloudFormation Templates

This directory contains comprehensive CloudFormation templates for deploying LogGuardian, an automated CloudWatch log group compliance remediation solution.

## 🔴 **IMPORTANT: Example Values Warning**

**ALL PARAMETER VALUES IN THIS REPOSITORY ARE EXAMPLES ONLY**

Before deploying:
1. **Replace account IDs**: Change `123456789012` to your actual AWS account ID
2. **Update bucket names**: Replace example bucket names with your actual S3 buckets
3. **Verify configurations**: Review all parameter values in `parameters/` directory
4. **Security check**: Ensure no example/placeholder values remain

**Deploying with example values will cause failures and potential security issues.**

## Quick Start

### 1. Simple Single-File Deployment

For basic deployments, use the simplified template:

> ⚠️ **Before deployment**: Update parameter files with your actual AWS account ID and bucket names

```bash
aws cloudformation create-stack \
  --stack-name logguardian-sandbox \
  --template-body file://00-logguardian-simple.yaml \
  --parameters file://parameters/sandbox-parameters.json \
  --capabilities CAPABILITY_NAMED_IAM
```

### 2. Full Modular Deployment

For production environments with nested templates:

> ⚠️ **Before deployment**: Update parameter files with your actual AWS account ID and bucket names

```bash
# Upload templates to S3 (replace with your actual bucket)
aws s3 sync . s3://your-deployment-bucket/templates/

# Deploy main stack  
aws cloudformation create-stack \
  --stack-name logguardian-prod \
  --template-body file://01-logguardian-main.yaml \
  --parameters file://parameters/prod-parameters.json \
  --capabilities CAPABILITY_NAMED_IAM
```

## Template Overview

### Main Templates

- **`00-logguardian-simple.yaml`** - ⭐ **START HERE** - Single-file deployment (all resources inline)
- **`01-logguardian-main.yaml`** - Complete deployment with nested templates

### Component Templates (Nested)

- **`02-iam-roles.yaml`** - IAM roles with least-privilege permissions  
- **`03-lambda-function.yaml`** - Lambda function deployment
- **`04-kms-key.yaml`** - KMS key for log group encryption (optional)
- **`05-config-rules.yaml`** - AWS Config rules for compliance monitoring (optional)
- **`06-eventbridge-rules.yaml`** - EventBridge scheduling and triggers (optional)
- **`07-monitoring.yaml`** - CloudWatch dashboard and alarms (optional)

### Advanced Templates

- **`08-logguardian-stacksets.yaml`** - Multi-account/multi-region deployment

### Utility Scripts

- **`90-validate-templates.sh`** - Template validation script
- **`91-deploy-example.sh`** - Example deployment script

### Parameter Files

- **`parameters/sandbox-parameters.json`** - Sandbox environment settings (full features, short retention)
- **`parameters/prod-parameters.json`** - Production environment settings (optimized, long retention)

> 🔴 **CRITICAL**: All parameter values are **EXAMPLES ONLY**. Account IDs (123456789012), bucket names, and other values must be replaced with your actual values before deployment. **DO NOT USE EXAMPLE VALUES IN PRODUCTION**. See `parameters/README.md` for complete details.

## Architecture

LogGuardian follows this architecture:

1. **AWS Config Rules** evaluate all CloudWatch log groups for compliance
2. **EventBridge Rule** triggers Lambda function on a schedule
3. **Lambda Function** retrieves non-compliant resources and applies fixes:
   - Associates KMS keys for encryption
   - Sets retention policies
4. **CloudWatch Monitoring** tracks compliance metrics and alerts

## Deployment Options

### Option 1: Simple Deployment (Recommended for Testing)

Use `00-logguardian-simple.yaml` for quick deployment:

- All resources in single template
- Fewer dependencies
- Easier to understand and modify
- Perfect for development/testing

### Option 2: Modular Deployment (Recommended for Production)

Use `01-logguardian-main.yaml` with nested templates:

- Separation of concerns
- Reusable components
- Better for large-scale deployments
- Easier maintenance and updates

## Configuration

### Required Parameters

- **`Environment`** - Deployment environment (dev/staging/prod)
- **`DeploymentBucket`** - S3 bucket containing Lambda zip file
- **`LambdaCodeKey`** - S3 key for Lambda deployment package

### Optional Parameters

- **`KMSKeyAlias`** - Custom KMS key alias (default: logguardian)
- **`DefaultRetentionDays`** - Default retention period for log groups in days (varies by template: 30 days for main template, 365 days for simple template)

### Environment-Specific Settings

Each environment has different defaults:
- **Dev**: Smaller memory, hourly checks, 30-day retention
- **Staging**: Medium settings, twice daily checks, 90-day retention  
- **Prod**: Higher memory, daily checks, 365-day retention

## Security Features

### Least-Privilege IAM Permissions

Lambda execution role only has permissions for:
- Reading Config rule compliance details
- Modifying CloudWatch log groups (encryption & retention)
- Using designated KMS key
- Publishing custom metrics

### KMS Encryption

- Customer-managed KMS key for log group encryption
- Automatic key rotation enabled
- Proper key policies for CloudWatch Logs service

### Network Security

- No VPC configuration required (uses AWS service endpoints)
- All communication over HTTPS
- No internet access needed

## Monitoring & Observability

### CloudWatch Dashboard

Provides visibility into:
- Lambda function performance (duration, errors, invocations)  
- Compliance metrics (processed, remediated, error counts)
- Recent error logs for troubleshooting

### CloudWatch Alarms

Monitors:
- Lambda function errors
- High execution duration  
- Function throttling
- Low compliance rates

### Custom Metrics

LogGuardian publishes these metrics to CloudWatch:
- `LogGuardian/LogGroupsProcessed` - Total log groups evaluated
- `LogGuardian/LogGroupsRemediated` - Successfully fixed log groups
- `LogGuardian/RemediationErrors` - Failed remediation attempts
- `LogGuardian/ComplianceRate` - Overall compliance percentage

## Troubleshooting

### Common Issues

1. **Lambda timeout errors**: Increase `LambdaTimeout` parameter
2. **KMS permission errors**: Verify KMS key policy allows CloudWatch Logs
3. **Config rule not triggering**: Ensure Config service is enabled
4. **No resources processed**: Check EventBridge rule is enabled

### Validation

Use the validation script to check templates:

```bash
./90-validate-templates.sh
```

### Logs

Check Lambda function logs at:
```
/aws/lambda/logguardian-{environment}
```

## Cleanup

To remove all resources:

```bash
aws cloudformation delete-stack --stack-name logguardian-dev
```

Note: KMS key will be scheduled for deletion with 7-day waiting period.

## Examples

### Deploy to Development

```bash
export DEPLOYMENT_BUCKET=my-logguardian-bucket

# Build and package Lambda
make package

# Upload to S3
aws s3 cp dist/logguardian-compliance.zip s3://$DEPLOYMENT_BUCKET/

# Deploy simple template
aws cloudformation create-stack \
  --stack-name logguardian-sandbox \
  --template-body file://00-logguardian-simple.yaml \
  --parameters ParameterKey=DeploymentBucket,ParameterValue=$DEPLOYMENT_BUCKET \
               ParameterKey=Environment,ParameterValue=sandbox \
  --capabilities CAPABILITY_NAMED_IAM
```

### Deploy to Production with Nested Templates

```bash
# Upload all templates
aws s3 sync templates/ s3://$DEPLOYMENT_BUCKET/templates/

# Deploy main template
aws cloudformation create-stack \
  --stack-name logguardian-prod \
  --template-body file://01-logguardian-main.yaml \
  --parameters file://parameters/prod-parameters.json \
  --capabilities CAPABILITY_NAMED_IAM
```

## Contributing

When modifying templates:

1. Update parameter files to match template changes
2. Run validation script before committing
3. Test in development environment first
4. Update this README if adding new features

## Support

For issues and questions:
1. Check CloudWatch logs for Lambda errors
2. Verify AWS Config is properly configured
3. Ensure required IAM permissions are in place
4. Review EventBridge rule configuration
