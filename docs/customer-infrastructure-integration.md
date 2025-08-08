# Customer Infrastructure Integration Examples

This document provides examples and template modifications for customers who want to integrate LogGuardian with their existing AWS infrastructure.

## Common Customer Requirements

1. **Use existing KMS keys** instead of creating new ones
2. **Leverage existing Config rules** instead of creating duplicates  
3. **Custom retention policies** per region/compliance requirement
4. **Integration with existing monitoring** and alerting systems
5. **Custom IAM roles** and policies

## Template Modifications for Customer Infrastructure

### Enhanced Parameters Section

Add these parameters to support customer infrastructure integration:

```yaml
# Add to Parameters section in template.yaml

ExistingKMSKeyArn:
  Type: String
  Default: ""
  Description: "ARN of existing KMS key to use (leave empty to create new key)"

CreateKMSKey:
  Type: String
  Default: "true"
  Description: "Create new KMS key or use existing one specified in ExistingKMSKeyArn"
  AllowedValues: ["true", "false"]

ExistingEncryptionConfigRule:
  Type: String
  Default: ""  
  Description: "Name of existing Config rule for encryption compliance (leave empty to create new)"

ExistingRetentionConfigRule:
  Type: String
  Default: ""
  Description: "Name of existing Config rule for retention compliance (leave empty to create new)"

CreateConfigRules:
  Type: String
  Default: "true"
  Description: "Create new Config rules or use existing ones"
  AllowedValues: ["true", "false"]

ExistingConfigServiceRoleArn:
  Type: String
  Default: ""
  Description: "ARN of existing Config service role (leave empty to create new)"

CustomerTagPrefix:
  Type: String
  Default: "LogGuardian"
  Description: "Prefix for resource tags to match customer tagging strategy"
```

### Enhanced Conditions Section

```yaml
# Add to Conditions section in template.yaml

ShouldCreateKMSKey: !Equals [!Ref CreateKMSKey, "true"]
ShouldCreateConfigRules: !Equals [!Ref CreateConfigRules, "true"]
HasExistingKMSKey: !Not [!Equals [!Ref ExistingKMSKeyArn, ""]]
HasExistingConfigRole: !Not [!Equals [!Ref ExistingConfigServiceRoleArn, ""]]
HasExistingEncryptionRule: !Not [!Equals [!Ref ExistingEncryptionConfigRule, ""]]
HasExistingRetentionRule: !Not [!Equals [!Ref ExistingRetentionConfigRule, ""]]
```

### Conditional Resource Creation

```yaml
# Modified KMS Key Resource
LogGuardianKMSKey:
  Type: AWS::KMS::Key
  Condition: ShouldCreateKMSKey  # Only create if requested
  Properties:
    Description: !Sub "${CustomerTagPrefix} KMS key for ${Environment} environment"
    # ... rest of properties

# Modified Config Service Role  
ConfigServiceRole:
  Type: AWS::IAM::Role
  Condition: !Not [!Condition HasExistingConfigRole]  # Only create if no existing role
  Properties:
    # ... properties

# Modified Config Rules with conditional creation
EncryptionConfigRule:
  Type: AWS::Config::ConfigRule
  Condition: !And 
    - !Condition ShouldCreateConfigRules
    - !Not [!Condition HasExistingEncryptionRule]
  Properties:
    ConfigRuleName: !Sub "${CustomerTagPrefix}-encryption-${Environment}"
    # ... rest of properties
```

### Dynamic Lambda Environment Variables

```yaml
# In Lambda Function Properties
Environment:
  Variables:
    KMS_KEY_ARN: !If 
      - HasExistingKMSKey
      - !Ref ExistingKMSKeyArn  
      - !GetAtt LogGuardianKMSKey.Arn
    ENCRYPTION_CONFIG_RULE: !If
      - HasExistingEncryptionRule
      - !Ref ExistingEncryptionConfigRule
      - !Ref EncryptionConfigRule
    RETENTION_CONFIG_RULE: !If  
      - HasExistingRetentionRule
      - !Ref ExistingRetentionConfigRule
      - !Ref RetentionConfigRule
    CUSTOMER_TAG_PREFIX: !Ref CustomerTagPrefix
```

## Customer Deployment Examples

### Example 1: Large Enterprise with Existing KMS Infrastructure

```bash
# Customer has existing KMS keys per region and wants to use them
./scripts/logguardian-deploy.sh deploy \
  -e prod \
  -r ca-central-1 \
  --existing-kms-key alias/enterprise-logs-ca \
  --customer-tag-prefix "ACME-LogGuardian" \
  --owner "ACME-Platform-Team" \
  --retention-days 2555
  
# This creates resources tagged with:
# Product=ACME-LogGuardian, Owner=ACME-Platform-Team, Environment=prod, ManagedBy=SAM
```

### Example 2: Multi-Region Enterprise with Different Tagging per Region

```bash
# US regions - SOX compliance with specific owner
./scripts/logguardian-deploy.sh deploy \
  -e compliance \
  -r us-east-1 \
  --existing-kms-key alias/sox-compliance-logs \
  --customer-tag-prefix "MegaCorp-SOX-LogGuardian" \
  --owner "MegaCorp-Compliance-Team" \
  --retention-days 2555

# EU regions - GDPR compliance with different owner
./scripts/logguardian-deploy.sh deploy \
  -e compliance \
  -r eu-west-1 \
  --existing-kms-key alias/gdpr-compliance-logs \
  --customer-tag-prefix "MegaCorp-GDPR-LogGuardian" \
  --owner "MegaCorp-EU-Legal-Team" \
  --retention-days 2190
```

### Example 2: Existing Config Service Setup

```bash
# Customer already has Config service and rules configured
sam deploy --template-file template-enhanced.yaml \
  --stack-name logguardian-existing-config \
  --parameter-overrides \
    Environment=prod \
    CreateConfigRules=false \
    ExistingEncryptionConfigRule="customer-log-encryption-rule" \
    ExistingRetentionConfigRule="customer-log-retention-rule" \
    ExistingConfigServiceRoleArn="arn:aws:iam::123456789012:role/ConfigServiceRole" \
    CreateKMSKey=false \
    ExistingKMSKeyArn="arn:aws:kms:ca-central-1:123456789012:key/existing-key-id" \
  --region ca-central-1 \
  --capabilities CAPABILITY_IAM

# Customer then configures EventBridge to trigger Lambda with existing rule names:
aws events put-rule \
  --name logguardian-existing-encryption-trigger \
  --event-pattern '{
    "source": ["aws.config"],
    "detail-type": ["Config Rules Compliance Change"],
    "detail": {
      "configRuleName": ["customer-log-encryption-rule"]
    }
  }'

aws events put-targets \
  --rule logguardian-existing-encryption-trigger \
  --targets "Id"="1","Arn"="arn:aws:lambda:ca-central-1:123456789012:function:logguardian-compliance-prod"
```

### Example 3: Multi-Region with Different Customer Requirements

```bash
# Different customer configurations per region
declare -A region_configs=(
  # US East: Existing KMS, new Config rules, SOX compliance
  ["us-east-1"]="CreateKMSKey=false ExistingKMSKeyArn=arn:aws:kms:us-east-1:123456789012:key/sox-key DefaultRetentionDays=2555"
  
  # EU West: New KMS, existing Config rules, GDPR compliance  
  ["eu-west-1"]="CreateConfigRules=false ExistingEncryptionConfigRule=gdpr-encryption-rule DefaultRetentionDays=2190"
  
  # Canada: All new resources, PIPEDA compliance
  ["ca-central-1"]="DefaultRetentionDays=1825 CustomerTagPrefix=ACME-Canada-LogGuardian"
)

for region in "${!region_configs[@]}"; do
  echo "Deploying to $region with config: ${region_configs[$region]}"
  
  sam deploy --template-file template-enhanced.yaml \
    --stack-name logguardian-custom-$region \
    --parameter-overrides \
      Environment=prod \
      CreateMonitoringDashboard=true \
      ${region_configs[$region]} \
    --region $region \
    --capabilities CAPABILITY_IAM
done
```

### Example 4: Integration with Existing Monitoring

```bash
# Customer wants LogGuardian without CloudWatch dashboard (using existing Grafana/Datadog)
sam deploy --template-file template.yaml \
  --stack-name logguardian-external-monitoring \
  --parameter-overrides \
    Environment=prod \
    CreateMonitoringDashboard=false \
    CreateKMSKey=false \
    ExistingKMSKeyArn=arn:aws:kms:ca-central-1:123456789012:key/monitoring-key \
    CustomerTagPrefix="DatadogMonitored" \
  --region ca-central-1 \
  --capabilities CAPABILITY_IAM

# Customer then configures external monitoring to watch Lambda metrics:
# - AWS/Lambda Duration, Errors, Invocations  
# - LogGuardian custom metrics (LogGroupsProcessed, RemediationErrors)
```

## EventBridge Integration Patterns

### Pattern 1: Customer Config Rule Triggers

```json
{
  "Rules": [
    {
      "Name": "customer-encryption-compliance-trigger",
      "EventPattern": {
        "source": ["aws.config"],
        "detail-type": ["Config Rules Compliance Change"],  
        "detail": {
          "configRuleName": ["customer-existing-encryption-rule"],
          "newEvaluationResult": {
            "complianceType": ["NON_COMPLIANT"]
          }
        }
      },
      "Targets": [
        {
          "Id": "1",
          "Arn": "arn:aws:lambda:region:account:function:logguardian-compliance-prod",
          "Input": "{\"type\":\"config-rule-evaluation\",\"configRuleName\":\"customer-existing-encryption-rule\",\"region\":\"ca-central-1\",\"batchSize\":25}"
        }
      ]
    }
  ]
}
```

### Pattern 2: Cross-Account Config Rules

```bash
# For customers with Config in separate security account
sam deploy --template-file template-cross-account.yaml \
  --stack-name logguardian-cross-account \
  --parameter-overrides \
    Environment=prod \
    CreateConfigRules=false \
    SecurityAccountId=999999999999 \
    CrossAccountConfigRole=arn:aws:iam::999999999999:role/CrossAccountConfigAccess \
  --region ca-central-1 \
  --capabilities CAPABILITY_IAM
```

## Advanced Customer Integration

### Custom Lambda Input Processing

For customers with complex requirements, the Lambda can be modified to handle custom inputs:

```go
// In internal/handler/handler.go
type CustomerConfigEvent struct {
    Type                string `json:"type"`
    ConfigRuleName      string `json:"configRuleName,omitempty"`
    CustomerRuleName    string `json:"customerRuleName,omitempty"`    // Customer's existing rule
    CustomerKMSKeyArn   string `json:"customerKmsKeyArn,omitempty"`   // Customer's KMS key
    CustomerTagPrefix   string `json:"customerTagPrefix,omitempty"`   // Customer tagging
    Region              string `json:"region,omitempty"`
    BatchSize           int    `json:"batchSize,omitempty"`
}

// Handle customer-specific configuration
func (h *Handler) handleCustomerConfig(ctx context.Context, event CustomerConfigEvent) (*Response, error) {
    // Use customer's KMS key if provided
    kmsKeyArn := event.CustomerKMSKeyArn
    if kmsKeyArn == "" {
        kmsKeyArn = os.Getenv("KMS_KEY_ARN") // Fall back to default
    }
    
    // Use customer's Config rule name if provided  
    configRule := event.CustomerRuleName
    if configRule == "" {
        configRule = event.ConfigRuleName // Fall back to standard rule
    }
    
    // Process with customer-specific settings
    return h.processComplianceWithCustomerSettings(ctx, configRule, kmsKeyArn, event.CustomerTagPrefix)
}
```

### IAM Permissions for Customer Integration

```yaml
# Additional IAM permissions needed for customer infrastructure integration
CustomerIntegrationPolicy:
  Type: AWS::IAM::Policy
  Properties:
    PolicyName: LogGuardianCustomerIntegration
    PolicyDocument:
      Version: '2012-10-17'
      Statement:
        - Effect: Allow
          Action:
            - kms:DescribeKey
            - kms:GetKeyPolicy
          Resource: 
            - !If [HasExistingKMSKey, !Ref ExistingKMSKeyArn, !Ref AWS::NoValue]
        - Effect: Allow  
          Action:
            - config:DescribeConfigRules
            - config:GetComplianceDetailsByConfigRule
          Resource:
            - !Sub "arn:aws:config:${AWS::Region}:${AWS::AccountId}:config-rule/${ExistingEncryptionConfigRule}"
            - !Sub "arn:aws:config:${AWS::Region}:${AWS::AccountId}:config-rule/${ExistingRetentionConfigRule}"
          Condition:
            StringEquals:
              "config:ConfigRuleName": 
                - !Ref ExistingEncryptionConfigRule
                - !Ref ExistingRetentionConfigRule
```

## Testing Customer Configurations

### Validation Script

```bash
#!/bin/bash
# validate-customer-config.sh

STACK_NAME=$1
REGION=$2

if [ -z "$STACK_NAME" ] || [ -z "$REGION" ]; then
  echo "Usage: $0 <stack-name> <region>"
  exit 1
fi

echo "Validating LogGuardian customer configuration..."
echo "Stack: $STACK_NAME"  
echo "Region: $REGION"
echo ""

# Get stack outputs
outputs=$(aws cloudformation describe-stacks \
  --stack-name $STACK_NAME \
  --region $REGION \
  --query 'Stacks[0].Outputs')

# Get Lambda function name
function_arn=$(echo $outputs | jq -r '.[] | select(.OutputKey=="LogGuardianFunctionArn") | .OutputValue')
function_name=$(echo $function_arn | cut -d':' -f7)

echo "Lambda Function: $function_name"

# Test Lambda with customer configuration
aws lambda invoke \
  --function-name $function_name \
  --region $REGION \
  --payload '{
    "type": "validation",
    "region": "'$REGION'",
    "customerValidation": true
  }' \
  validation-response.json

echo "Validation response:"
cat validation-response.json | jq '.'

# Check if KMS key exists and is accessible
kms_key_arn=$(echo $outputs | jq -r '.[] | select(.OutputKey=="KMSKeyArn") | .OutputValue')
if [ "$kms_key_arn" != "null" ]; then
  echo ""
  echo "Validating KMS key access..."
  aws kms describe-key --key-id $kms_key_arn --region $REGION > /dev/null
  echo "✅ KMS key accessible"
fi

echo ""
echo "Customer configuration validation complete!"
```

These examples provide comprehensive guidance for customers who want to integrate LogGuardian with their existing AWS infrastructure while maintaining their security and compliance requirements.
