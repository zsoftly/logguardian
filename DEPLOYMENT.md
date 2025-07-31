# LogGuardian Deployment Guide

Comprehensive deployment instructions for LogGuardian CloudFormation templates.

## Prerequisites

- AWS CLI v2 with configured credentials
- Go 1.24+ for Lambda function builds
- AWS Config enabled in target regions
- S3 bucket for deployment artifacts

## Quick Deploy

### Single-File Template (Recommended for Testing)

```bash
# 1. Build and package
make build package

# 2. Upload Lambda code
aws s3 cp dist/logguardian-compliance.zip s3://your-deployment-bucket/

# 3. Deploy simple template
aws cloudformation deploy \
  --template-file templates/00-logguardian-simple.yaml \
  --stack-name logguardian-sandbox \
  --parameter-overrides \
    Environment=sandbox \
    DeploymentBucket=your-deployment-bucket \
    CreateKMSKey=true \
    CreateConfigRules=true \
  --capabilities CAPABILITY_NAMED_IAM
```

## Detailed Deployment Steps

### Step 1: Prepare Deployment Bucket

```bash
# Create S3 bucket for deployment artifacts
aws s3 mb s3://logguardian-deployment-${AWS_ACCOUNT_ID}-${AWS_REGION}

# Set bucket policy for CloudFormation access
aws s3api put-bucket-policy \
  --bucket logguardian-deployment-${AWS_ACCOUNT_ID}-${AWS_REGION} \
  --policy file://bucket-policy.json
```

### Step 2: Build Lambda Function

```bash
# Build the Go Lambda function
make build

# Create deployment package
make package

# Verify package contents
unzip -l dist/logguardian-compliance.zip
```

### Step 3: Validate Templates

```bash
# Validate all CloudFormation templates
make validate-templates

# Or validate manually
cd templates
./90-validate-templates.sh
```

### Step 4: Upload Artifacts

```bash
# Upload Lambda code
aws s3 cp dist/logguardian-compliance.zip s3://${DEPLOYMENT_BUCKET}/

# Upload CloudFormation templates (for modular deployment)
make upload-templates
```

### Step 5: Deploy CloudFormation Stack

Choose your deployment method:

#### Option A: Direct AWS CLI Deployment

```bash
# Simple deployment (single region, basic features)
aws cloudformation deploy \
  --template-file templates/00-logguardian-simple.yaml \
  --stack-name logguardian-sandbox \
  --parameter-overrides Environment=sandbox DeploymentBucket=${DEPLOYMENT_BUCKET} \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ca-central-1

# Full deployment (modular template)
aws cloudformation deploy \
  --template-file templates/01-logguardian-main.yaml \
  --stack-name logguardian-staging \
  --parameter-overrides Environment=staging DeploymentBucket=${DEPLOYMENT_BUCKET} \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ca-central-1
```

#### Option B: Using AWS CLI with Parameter Files

```bash
# Deploy with parameter file
aws cloudformation deploy \
  --template-file templates/01-logguardian-main.yaml \
  --stack-name logguardian-sandbox \
  --parameter-overrides file://templates/parameters/sandbox-parameters.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ${AWS_REGION}
```

#### Option C: Using Makefile Targets

```bash
# Set environment variables
export DEPLOYMENT_BUCKET="your-bucket"
export AWS_REGION="ca-central-1"

# Deploy to sandbox
make deploy-sandbox
```

## Deployment Options

### Simple Deployment

**Use Case**: Development, testing, small production environments

**Features**:
- Single CloudFormation template
- Basic monitoring
- Essential IAM permissions
- Config rules for compliance

**Command**:
```bash
aws cloudformation deploy \
  --template-file templates/00-logguardian-simple.yaml \
  --stack-name logguardian-sandbox \
  --parameter-overrides Environment=sandbox DeploymentBucket=my-bucket \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ca-central-1
```

### Full Deployment (Modular)

**Use Case**: Production environments requiring comprehensive monitoring

**Features**:
- Modular CloudFormation templates
- Comprehensive monitoring dashboard
- Advanced error handling
- Multi-region support
- Detailed alerting

**Command**:
```bash
aws cloudformation deploy \
  --template-file templates/01-logguardian-main.yaml \
  --stack-name logguardian-staging \
  --parameter-overrides Environment=staging DeploymentBucket=my-bucket \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ca-central-1
```

### Multi-Region Deployment

**Use Case**: Enterprise environments with log groups across multiple regions

```bash
# Deploy to multiple regions using AWS CLI
for region in ca-central-1 ca-west-1 eu-central-1; do
  aws cloudformation deploy \
    --template-file templates/01-logguardian-main.yaml \
    --stack-name logguardian-sandbox-$region \
    --parameter-overrides Environment=sandbox DeploymentBucket=my-bucket-$region \
    --capabilities CAPABILITY_NAMED_IAM \
    --region $region
done
```

### Multi-Account Deployment (StackSets)

**Use Case**: Organization-wide deployment across multiple AWS accounts

```bash
# Create StackSet in organization master account
aws cloudformation create-stack-set \
  --stack-set-name logguardian-org \
  --template-body file://templates/01-logguardian-main.yaml \
  --parameters file://templates/parameters/sandbox-parameters.json \
  --capabilities CAPABILITY_NAMED_IAM \
  --operation-preferences RegionConcurrencyType=PARALLEL,MaxConcurrentPercentage=100

# Deploy to organizational units
aws cloudformation create-stack-instances \
  --stack-set-name logguardian-org \
  --deployment-targets OrganizationalUnitIds=ou-root-12345678 \
  --regions ca-central-1 ca-west-1
```

## Configuration Examples

### Development Environment

```bash
aws cloudformation deploy \
  --template-file templates/00-logguardian-simple.yaml \
  --stack-name logguardian-dev \
  --parameter-overrides \
    Environment=dev \
    DeploymentBucket=dev-deployment-bucket \
    DefaultRetentionDays=30 \
    ScheduleExpression="rate(1 hour)" \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ca-central-1
```

### Sandbox Environment

```bash
aws cloudformation deploy \
  --template-file templates/00-logguardian-simple.yaml \
  --stack-name logguardian-sandbox \
  --parameter-overrides \
    Environment=sandbox \
    DeploymentBucket=sandbox-deployment-bucket \
    DefaultRetentionDays=90 \
    ScheduleExpression="rate(12 hours)" \
  --capabilities CAPABILITY_NAMED_IAM \
  --region ca-central-1
```

## Post-Deployment Verification

### Step 1: Verify Stack Deployment

```bash
# Check stack status
aws cloudformation describe-stacks \
  --stack-name logguardian-sandbox \
  --query 'Stacks[0].StackStatus'

# List stack outputs
aws cloudformation describe-stacks \
  --stack-name logguardian-sandbox \
  --query 'Stacks[0].Outputs'
```

### Step 2: Test Lambda Function

```bash
# Get invocation command from stack outputs
INVOCATION_CMD=$(aws cloudformation describe-stacks \
  --stack-name logguardian-sandbox \
  --query 'Stacks[0].Outputs[?OutputKey==`InvocationCommand`].OutputValue' \
  --output text)

# Execute the command
eval $INVOCATION_CMD

# Check response
cat response.json
```

### Step 3: Verify Config Rules

```bash
# Check Config rule status
aws configservice describe-config-rules \
  --config-rule-names cloudwatch-log-group-encrypted cw-loggroup-retention-period-check

# Check compliance status
aws configservice get-compliance-details-by-config-rule \
  --config-rule-name cloudwatch-log-group-encrypted \
  --compliance-types NON_COMPLIANT
```

### Step 4: Monitor CloudWatch Dashboard

```bash
# Get dashboard URL from stack outputs
DASHBOARD_URL=$(aws cloudformation describe-stacks \
  --stack-name logguardian-sandbox \
  --query 'Stacks[0].Outputs[?OutputKey==`CloudWatchDashboardURL`].OutputValue' \
  --output text)

echo "Dashboard URL: $DASHBOARD_URL"
```

## Environment-Specific Configurations

### Parameter Customization

Create custom parameter files for your environment:

```json
{
  "DeploymentBucket": "my-company-logguardian-sandbox",
  "Environment": "sandbox",
  "KMSKeyAlias": "alias/my-company-logs",
  "DefaultRetentionDays": 90,
  "ScheduleExpression": "rate(6 hours)",
  "SupportedRegions": "ca-central-1,ca-west-1,eu-west-1"
}
```

### Regional KMS Keys

Configure region-specific KMS keys:

```bash
# Set region-specific environment variables in Lambda
KMS_KEY_ALIAS_ca_central_1=alias/logs-ca-central-1
KMS_KEY_ALIAS_ca_west_1=alias/logs-ca-west-1
KMS_KEY_ALIAS_eu_west_1=alias/logs-eu-west-1
```

## Troubleshooting

### Common Issues

#### 1. "Template validation failed"

```bash
# Validate specific template
aws cloudformation validate-template \
  --template-body file://templates/logguardian-main.yaml

# Check template syntax
make validate-templates
```

#### 2. "Deployment bucket not found"

```bash
# Check bucket exists
aws s3 ls s3://${DEPLOYMENT_BUCKET}

# Check bucket region
aws s3api get-bucket-location --bucket ${DEPLOYMENT_BUCKET}
```

#### 3. "Lambda code not found"

```bash
# Verify Lambda package exists
aws s3 ls s3://${DEPLOYMENT_BUCKET}/logguardian-compliance.zip

# Re-upload if missing
make package
aws s3 cp dist/logguardian-compliance.zip s3://${DEPLOYMENT_BUCKET}/
```

#### 4. "Config service not enabled"

```bash
# Check Config status
aws configservice describe-configuration-recorders

# Enable Config if needed
aws configservice put-configuration-recorder \
  --configuration-recorder name=default,roleARN=arn:aws:iam::ACCOUNT:role/config-role
```

### Debug Lambda Function

```bash
# Check Lambda logs
aws logs tail /aws/lambda/logguardian-compliance-sandbox --follow

# Check Lambda function configuration
aws lambda get-function --function-name logguardian-compliance-sandbox

# Test Lambda function
aws lambda invoke \
  --function-name logguardian-compliance-sandbox \
  --payload '{"type":"config-rule-evaluation","configRuleName":"cloudwatch-log-group-encrypted","region":"ca-central-1","batchSize":5}' \
  test-output.json
```

### Stack Update Issues

```bash
# Create change set for safe updates
aws cloudformation create-change-set \
  --stack-name logguardian-prod \
  --template-body file://templates/logguardian-main.yaml \
  --change-set-name update-$(date +%Y%m%d) \
  --capabilities CAPABILITY_NAMED_IAM

# Review changes before applying
aws cloudformation describe-change-set \
  --change-set-name update-$(date +%Y%m%d) \
  --stack-name logguardian-prod
```

## Security Considerations

### KMS Key Management

```bash
# Verify KMS key policy allows CloudWatch Logs
aws kms get-key-policy \
  --key-id alias/cloudwatch-logs-compliance \
  --policy-name default

# Test KMS key access
aws kms describe-key --key-id alias/cloudwatch-logs-compliance
```

### IAM Permission Validation

```bash
# Test Lambda execution role permissions
aws sts assume-role \
  --role-arn arn:aws:iam::ACCOUNT:role/LogGuardian-LambdaExecutionRole-prod \
  --role-session-name test-session

# Validate Config permissions
aws configservice describe-config-rules \
  --profile logguardian-test
```

## Monitoring and Maintenance

### Regular Health Checks

```bash
# Check Lambda function health
aws lambda get-function --function-name logguardian-compliance-sandbox

# Check EventBridge rule status
aws events describe-rule --name logguardian-schedule-sandbox

# Check Config rule compliance
aws configservice get-compliance-summary-by-config-rule
```

### Update Procedures

```bash
# Update Lambda code
make build package
aws s3 cp dist/logguardian-compliance.zip s3://${DEPLOYMENT_BUCKET}/
aws cloudformation update-stack \
  --stack-name logguardian-prod \
  --use-previous-template \
  --capabilities CAPABILITY_NAMED_IAM

# Update templates
make validate-templates upload-templates
aws cloudformation update-stack \
  --stack-name logguardian-sandbox \
  --template-url https://${DEPLOYMENT_BUCKET}.s3.amazonaws.com/templates/01-logguardian-main.yaml \
  --capabilities CAPABILITY_NAMED_IAM
```

## Cost Optimization

### Monitor Costs

```bash
# Check Lambda costs
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-02-01 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --group-by Type=DIMENSION,Key=SERVICE

# Optimize Lambda memory
aws lambda put-function-configuration \
  --function-name logguardian-compliance-sandbox \
  --memory-size 256
```

### Resource Cleanup

```bash
# Delete stack when no longer needed
aws cloudformation delete-stack --stack-name logguardian-sandbox

# Clean up S3 artifacts
aws s3 rm s3://${DEPLOYMENT_BUCKET}/ --recursive
aws s3 rb s3://${DEPLOYMENT_BUCKET}
```

## Support

For additional support:

- **Documentation**: [GitHub Repository](https://github.com/zsoftly/logguardian)
- **Issues**: Open an issue on GitHub
- **Community**: Join discussions in GitHub Discussions

---

**Next Steps**: After successful deployment, review the [monitoring guide](../docs/monitoring.md) and [operational procedures](../docs/operations.md).
