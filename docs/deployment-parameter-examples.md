# LogGuardian Deployment Parameter Examples

This file contains parameter examples for different deployment scenarios using the optimized SAM template.

## Greenfield Deployment (All New Infrastructure)

**Use Case**: New AWS account or testing environment
**Creates**: All infrastructure from scratch

```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=true \
    KMSKeyAlias=alias/logguardian-logs \
    CreateConfigService=true \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    ScheduleExpression="cron(0 2 ? * SAT *)" \
    EnableStaggeredScheduling=true \
    CreateMonitoringDashboard=true \
    DefaultRetentionDays=365 \
    LambdaMemorySize=256 \
    S3ExpirationDays=90 \
    ProductName=LogGuardian \
    Owner=Platform-Team \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Enterprise with Existing Infrastructure

**Use Case**: Large enterprise with existing Config service and KMS keys
**Creates**: Only Lambda function, reuses everything else

```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=false \
    ExistingKMSKeyArn=arn:aws:kms:ca-central-1:123456789012:key/12345678-1234-1234-1234-123456789012 \
    KMSKeyAlias=alias/enterprise-logs \
    CreateConfigService=false \
    ExistingConfigBucket=enterprise-config-ca-central-1 \
    ExistingConfigServiceRoleArn=arn:aws:iam::123456789012:role/EnterpriseConfigServiceRole \
    CreateConfigRules=false \
    ExistingEncryptionConfigRule=enterprise-log-encryption-compliance \
    ExistingRetentionConfigRule=enterprise-log-retention-compliance \
    CreateEventBridgeRules=false \
    CreateMonitoringDashboard=false \
    DefaultRetentionDays=2555 \
    LambdaMemorySize=512 \
    Owner=ACME-Security-Team \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Hybrid Deployment (Mix of New and Existing)

**Use Case**: Organization with existing Config but wants new KMS key and rules
**Creates**: KMS key, Config rules, and scheduling; reuses Config service

```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=true \
    KMSKeyAlias=alias/logguardian-compliance-logs \
    CreateConfigService=false \
    ExistingConfigBucket=corporate-config-bucket \
    ExistingConfigServiceRoleArn=arn:aws:iam::123456789012:role/CorporateConfigRole \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    ScheduleExpression="rate(12 hours)" \
    EnableStaggeredScheduling=true \
    CreateMonitoringDashboard=false \
    DefaultRetentionDays=1825 \
    LambdaMemorySize=256 \
    Owner=GlobalBank-CloudOps \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Manual Invocation Only

**Use Case**: Integration with existing automation workflows
**Creates**: Lambda and required infrastructure, but no scheduling

```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=true \
    CreateConfigService=true \
    CreateConfigRules=true \
    CreateEventBridgeRules=false \
    CreateMonitoringDashboard=false \
    DefaultRetentionDays=90 \
    LambdaMemorySize=256 \
    ProductName=LogGuardian-Manual \
    Owner=Automation-Team \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Development Environment

**Use Case**: Development and testing with minimal retention and cost
**Creates**: All infrastructure with development-optimized settings

```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-dev \
  --parameter-overrides \
    Environment=dev \
    CreateKMSKey=true \
    CreateConfigService=true \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    ScheduleExpression="rate(1 hour)" \
    EnableStaggeredScheduling=false \
    CreateMonitoringDashboard=true \
    DefaultRetentionDays=7 \
    LambdaMemorySize=128 \
    S3ExpirationDays=3 \
    EnableS3LifecycleRules=true \
    ProductName=LogGuardian-Dev \
    Owner=DevOps-Team \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Multi-Account Organizational Deployment

**Use Case**: Using existing organizational Config rules deployed via StackSets
**Creates**: Lambda function, reuses org-level Config infrastructure

```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=false \
    ExistingKMSKeyArn=arn:aws:kms:ca-central-1:123456789012:key/org-compliance-key \
    KMSKeyAlias=alias/org-compliance-logs \
    CreateConfigService=false \
    ExistingConfigBucket=organization-config-123456789012-ca-central-1 \
    ExistingConfigServiceRoleArn=arn:aws:iam::123456789012:role/OrganizationConfigRole \
    CreateConfigRules=false \
    ExistingEncryptionConfigRule=org-cloudwatch-log-group-encrypted \
    ExistingRetentionConfigRule=org-cw-loggroup-retention-period-check \
    CreateEventBridgeRules=true \
    ScheduleExpression="cron(0 3 ? * SUN *)" \
    CreateMonitoringDashboard=true \
    DefaultRetentionDays=365 \
    LambdaMemorySize=256 \
    Owner=Enterprise-Compliance \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Compliance-Specific Deployments

### SOX Compliance (7 years retention)
```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-sox \
  --parameter-overrides \
    Environment=sox-compliance \
    CreateKMSKey=false \
    ExistingKMSKeyArn=arn:aws:kms:us-east-1:123456789012:key/sox-compliance-key \
    CreateConfigService=true \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    DefaultRetentionDays=2555 \
    S3ExpirationDays=2555 \
    Owner=SOX-Compliance-Team \
  --region us-east-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

### GDPR Compliance (6 years retention)
```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-gdpr \
  --parameter-overrides \
    Environment=gdpr-compliance \
    CreateKMSKey=false \
    ExistingKMSKeyArn=arn:aws:kms:eu-west-1:123456789012:key/gdpr-compliance-key \
    CreateConfigService=true \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    DefaultRetentionDays=2190 \
    S3ExpirationDays=2190 \
    Owner=GDPR-Compliance-Team \
  --region eu-west-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Validation Commands

After deployment, validate the configuration:

```bash
# Check stack status
aws cloudformation describe-stacks \
  --stack-name logguardian-prod \
  --query 'Stacks[0].StackStatus'

# Get deployment summary
aws cloudformation describe-stacks \
  --stack-name logguardian-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`DeploymentSummary`].OutputValue' \
  --output text

# Test Lambda function
aws lambda invoke \
  --function-name logguardian-compliance-prod \
  --payload '{
    "type": "config-rule-evaluation",
    "configRuleName": "logguardian-encryption-prod",
    "region": "ca-central-1",
    "account": "123456789012",
    "environment": "prod",
    "batchSize": 5
  }' \
  test-response.json

# View response
cat test-response.json | jq '.'
```

## Parameter Reference

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `Environment` | Yes | `prod` | Deployment environment |
| `CreateKMSKey` | Yes | `true` | Create new KMS key |
| `ExistingKMSKeyArn` | Conditional | `""` | Required if CreateKMSKey=false |
| `KMSKeyAlias` | Yes | `alias/logguardian-logs` | KMS key alias |
| `CreateConfigService` | Yes | `true` | Create Config service resources |
| `ExistingConfigBucket` | Conditional | `""` | Required if CreateConfigService=false |
| `ExistingConfigServiceRoleArn` | Conditional | `""` | Required if CreateConfigService=false |
| `CreateConfigRules` | Yes | `true` | Create new Config rules |
| `ExistingEncryptionConfigRule` | Conditional | `""` | Required if CreateConfigRules=false |
| `ExistingRetentionConfigRule` | Conditional | `""` | Required if CreateConfigRules=false |
| `CreateEventBridgeRules` | Yes | `true` | Create EventBridge scheduling |
| `ScheduleExpression` | No | `cron(0 2 ? * SAT *)` | Schedule for automated runs |
| `CreateMonitoringDashboard` | Yes | `false` | Create CloudWatch dashboard |
| `DefaultRetentionDays` | Yes | `30` | Default log retention period |
| `LambdaMemorySize` | Yes | `256` | Lambda memory allocation |
| `Owner` | Yes | `ZSoftly` | Resource owner tag |

## Common Validation Errors

### Missing Required Parameters
```
Error: ExistingKMSKeyArn is required when CreateKMSKey=false
```
**Solution**: Provide the ARN of an existing KMS key

### Invalid KMS Key ARN
```
Error: Access denied when describing KMS key
```
**Solution**: Ensure the deployment role has access to the KMS key

### Config Service Not Enabled
```
Error: Configuration recorder not found
```
**Solution**: Enable Config service or set CreateConfigService=true

### Config Rule Not Found
```
Error: Config rule 'rule-name' does not exist
```
**Solution**: Verify the Config rule name or set CreateConfigRules=true

These examples cover the most common deployment scenarios and provide a starting point for customizing LogGuardian deployments to fit specific customer requirements.
