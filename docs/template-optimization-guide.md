# LogGuardian Optimized SAM Template - Single Region Deployment

## Overview

This optimized SAM template makes **every infrastructure component optional**, allowing customers to integrate LogGuardian with their existing AWS infrastructure. The template is designed for single-region deployment to simplify customer understanding and reduce complexity.

## Key Optimizations

### 🎯 **Single Region Focus**
- Simplified deployment model: "Deploy in each AWS region where you want compliance monitoring"
- Faster to market and easier customer understanding
- Most customers only use 1-2 regions anyway
- Clear, focused value proposition

### 🔧 **Truly Optional Infrastructure**

#### 1. **KMS Key Management**
```yaml
CreateKMSKey: true|false
ExistingKMSKeyArn: arn:aws:kms:region:account:key/key-id
```
- **Create new**: Template creates KMS key with proper CloudWatch Logs policy
- **Use existing**: Reference customer's existing KMS key
- Lambda gets the correct key ARN automatically

#### 2. **AWS Config Service**
```yaml
CreateConfigService: true|false
ExistingConfigBucket: bucket-name
ExistingConfigServiceRoleArn: arn:aws:iam::account:role/ConfigRole
```
- **Create new**: Template creates Config recorder, delivery channel, S3 bucket, and IAM role
- **Use existing**: Reference customer's existing Config setup
- Supports customers with organization-wide Config already configured

#### 3. **Config Rules**
```yaml
CreateConfigRules: true|false
ExistingEncryptionConfigRule: customer-encryption-rule
ExistingRetentionConfigRule: customer-retention-rule
```
- **Create new**: Template creates LogGuardian-specific Config rules
- **Use existing**: Reference customer's existing Config rules for encryption and retention
- Lambda automatically uses the correct rule names

#### 4. **EventBridge Scheduling**
```yaml
CreateEventBridgeRules: true|false
```
- **Create scheduled**: Template creates EventBridge rules for automated execution
- **Manual invocation**: Disable scheduling for customers who want manual control
- Perfect for integration with existing automation workflows

#### 5. **CloudWatch Dashboard**
```yaml
CreateMonitoringDashboard: true|false
```
- **Create dashboard**: For customers who want built-in monitoring
- **Skip dashboard**: For customers with existing monitoring (Datadog, Grafana, etc.)

## Deployment Scenarios

### 🚀 **Scenario 1: Greenfield Deployment (New Infrastructure)**
```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=true \
    CreateConfigService=true \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    CreateMonitoringDashboard=true \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

**Creates**: All infrastructure from scratch - ideal for testing or new AWS accounts.

### 🏢 **Scenario 2: Enterprise with Existing Infrastructure**
```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=false \
    ExistingKMSKeyArn=arn:aws:kms:ca-central-1:123456789012:key/enterprise-logs-key \
    CreateConfigService=false \
    ExistingConfigBucket=enterprise-config-bucket \
    ExistingConfigServiceRoleArn=arn:aws:iam::123456789012:role/EnterpriseConfigRole \
    CreateConfigRules=false \
    ExistingEncryptionConfigRule=enterprise-log-encryption-rule \
    ExistingRetentionConfigRule=enterprise-log-retention-rule \
    CreateEventBridgeRules=false \
    CreateMonitoringDashboard=false \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

**Creates**: Only the Lambda function - reuses all existing infrastructure.

### 🔀 **Scenario 3: Hybrid Deployment**
```bash
sam deploy --template-file template.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=true \
    CreateConfigService=false \
    ExistingConfigBucket=corporate-config-bucket \
    ExistingConfigServiceRoleArn=arn:aws:iam::123456789012:role/ConfigRole \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    CreateMonitoringDashboard=false \
  --region ca-central-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

**Creates**: KMS key and Config rules, reuses existing Config service, creates scheduling.

## Smart Defaults and Validation

### ✅ **Built-in Validation**
- Template validates that if `CreateKMSKey=false`, then `ExistingKMSKeyArn` must be provided
- If `CreateConfigService=false`, both `ExistingConfigBucket` and `ExistingConfigServiceRoleArn` are required
- If `CreateConfigRules=false`, both existing Config rule names must be provided

### 🎛️ **Dynamic Environment Variables**
The Lambda function automatically gets the correct environment variables based on what's created vs. reused:

```yaml
Environment:
  Variables:
    KMS_KEY_ARN: !If [ShouldCreateKMSKey, !GetAtt LogGuardianKMSKey.Arn, !Ref ExistingKMSKeyArn]
    ENCRYPTION_CONFIG_RULE: !If [ShouldCreateConfigRules, !Ref EncryptionConfigRule, !Ref ExistingEncryptionConfigRule]
    RETENTION_CONFIG_RULE: !If [ShouldCreateConfigRules, !Ref RetentionConfigRule, !Ref ExistingRetentionConfigRule]
```

### 📊 **Deployment Summary Output**
The template provides a clear summary of what was created vs. reused:

```
DeploymentSummary:
  KMS Key: Created|Using Existing
  Config Service: Created|Using Existing  
  Config Rules: Created|Using Existing
  EventBridge: Created|Disabled
  Dashboard: Created|Disabled
```

## Manual Invocation Support

For customers who disable EventBridge scheduling, the template provides a ready-to-use CLI command:

```bash
aws lambda invoke --function-name logguardian-compliance-prod \
  --payload '{
    "type": "config-rule-evaluation",
    "configRuleName": "customer-encryption-rule",
    "region": "ca-central-1",
    "account": "123456789012",
    "environment": "prod"
  }' response.json --region ca-central-1
```

## Cost Optimization

### 💰 **Pay Only for What You Use**
- **No EventBridge**: $0 scheduling costs for manual invocation scenarios
- **No Dashboard**: $0 CloudWatch dashboard costs for customers with existing monitoring
- **Existing Config**: $0 Config service costs if using existing setup
- **Existing KMS**: $0 additional KMS key costs

### 📈 **Resource Sizing by Environment**
```yaml
# Development
LambdaMemorySize: 128
S3ExpirationDays: 7

# Production  
LambdaMemorySize: 512
S3ExpirationDays: 90
```

## Customer Benefits

### 🎯 **For New Customers**
- ✅ **One-click deployment** with all infrastructure created
- ✅ **Immediate value** with minimal configuration
- ✅ **Best practices** built-in (encryption, lifecycle, monitoring)

### 🏢 **For Enterprise Customers**
- ✅ **Zero infrastructure overlap** - reuse existing investments
- ✅ **Compliance alignment** - use existing Config rules and KMS keys
- ✅ **Integration flexibility** - manual or automated execution
- ✅ **Cost efficiency** - no redundant resources

### 🔄 **For Hybrid Scenarios**
- ✅ **Mix and match** - create some, reuse others
- ✅ **Gradual adoption** - start with manual, add automation later
- ✅ **Customizable tagging** - align with existing tag strategies

## Migration Path

Customers can easily migrate from full to hybrid to existing infrastructure:

1. **Start**: Deploy with all infrastructure created
2. **Migrate**: Redeploy with existing Config service
3. **Optimize**: Redeploy with existing KMS keys and Config rules
4. **Integrate**: Disable scheduling and integrate with existing automation

Each step maintains the same Lambda function while reducing infrastructure footprint.

## Security Best Practices

### 🔒 **Least Privilege IAM**
- Lambda only gets permissions for resources it actually uses
- KMS permissions scope to the specific key (created or existing)
- Config permissions scope to the specific rules (created or existing)

### 🛡️ **Defense in Depth**
- S3 bucket policy restricts Config service access by account
- KMS key policy restricts CloudWatch Logs by ARN pattern
- Lambda execution role follows principle of least privilege

### 📋 **Compliance Ready**
- All resources properly tagged for cost allocation and governance
- Config rules align with common compliance frameworks (SOC, PCI, etc.)
- Audit trail through CloudTrail for all Lambda executions

This optimized template provides maximum flexibility while maintaining security and operational excellence, making LogGuardian truly enterprise-ready for any AWS environment.
