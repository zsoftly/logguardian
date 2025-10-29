# Existing Config Deployment

Deploy LogGuardian using your existing AWS Config, KMS, and Config rules.

## Use Case

- ✅ You already have AWS Config enabled
- ✅ You have existing KMS keys for CloudWatch Logs
- ✅ You have existing Config rules
- ✅ You want to add LogGuardian without duplicating infrastructure

## What Gets Created

- ✅ Lambda function (new)
- ✅ IAM roles for Lambda (new)
- ✅ EventBridge triggers (new)
- ✅ CloudWatch dashboard (new)
- ❌ No KMS key (uses existing)
- ❌ No Config service (uses existing)
- ❌ No Config rules (uses existing)

## Prerequisites

### 1. Find Your Existing Resources
```bash
# Find KMS key
aws kms list-aliases --query "Aliases[?contains(AliasName,'log')].{Alias:AliasName,KeyId:TargetKeyId}"

# Get full KMS ARN
aws kms describe-key --key-id alias/your-key-alias --query 'KeyMetadata.Arn' --output text

# Find Config bucket
aws configservice describe-delivery-channels --query 'DeliveryChannels[0].s3BucketName' --output text

# Find Config service role
aws iam list-roles --query "Roles[?contains(RoleName,'Config')].Arn"

# Find Config rules
aws configservice describe-config-rules --query 'ConfigRules[?contains(ConfigRuleName,`log`)].ConfigRuleName'
```

### 2. Build Lambda Binary
```bash
cd ../../../
make build
```

## Usage
```bash
# 1. Copy example config
cp terraform.tfvars.example terraform.tfvars

# 2. Update with your existing resource ARNs/names
nano terraform.tfvars

# 3. Initialize
terraform init

# 4. Plan
terraform plan

# 5. Deploy
terraform apply
```

## Cost Savings

By using existing infrastructure:
- **Save ~$7/month** (no new KMS key, Config service, or rules)
- **Only pay for Lambda** (~$0.01/month with weekly schedule)
- **Total cost: ~$0.10/month**

## Validation
```bash
# Verify existing resources are being used
terraform plan | grep "existing"

# Should show:
# - Using existing KMS key
# - Using existing Config service
# - Using existing Config rules
```
