# LogGuardian Deployment Examples

## AWS CLI CloudFormation Deployments

### Get SAR Template
```bash
# Get the application template from SAR (latest version)
aws serverlessrepo create-cloud-formation-template \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --region ca-central-1
```

### New Infrastructure (Greenfield)
```bash
aws cloudformation deploy \
  --template-file logguardian-template.yaml \
  --stack-name logguardian \
  --parameter-overrides \
    Environment=prod \
    ProductName=MyCompany-LogGuardian \
    Owner=DevOps-Team \
    KMSKeyAlias=alias/logguardian-logs \
    DefaultRetentionDays=365 \
    CreateKMSKey=true \
    CreateConfigService=true \
    CreateConfigRules=true \
    CreateEventBridgeRules=true \
    EncryptionScheduleExpression="cron(0 2 ? * * *)" \
    RetentionScheduleExpression="cron(0 3 ? * * *)" \
  --capabilities CAPABILITY_NAMED_IAM
```

### Existing Infrastructure (Enterprise)
```bash
aws cloudformation deploy \
  --template-file logguardian-template.yaml \
  --stack-name logguardian \
  --parameter-overrides \
    Environment=prod \
    CreateKMSKey=false \
    ExistingKMSKeyArn=arn:aws:kms:ca-central-1:123456789012:key/abcd1234 \
    KMSKeyAlias=alias/enterprise-logs \
    CreateConfigService=false \
    ExistingConfigBucket=enterprise-config-bucket \
    ExistingConfigServiceRoleArn=arn:aws:iam::123456789012:role/ConfigRole \
    CreateConfigRules=false \
    ExistingEncryptionConfigRule=enterprise-encryption-rule \
    ExistingRetentionConfigRule=enterprise-retention-rule \
    DefaultRetentionDays=365 \
    LambdaMemorySize=512 \
  --capabilities CAPABILITY_NAMED_IAM
```

### Manual Invocation Only
```bash
aws cloudformation deploy \
  --template-file logguardian-template.yaml \
  --stack-name logguardian \
  --parameter-overrides \
    Environment=audit \
    CreateEventBridgeRules=false \
    DefaultRetentionDays=90 \
    LambdaMemorySize=512 \
  --capabilities CAPABILITY_NAMED_IAM
```

## Terraform Integration

### Using SAR Application
```hcl
resource "aws_cloudformation_stack" "logguardian" {
  name = "logguardian"
  
  template_url = "https://s3.amazonaws.com/aws-sam-cli-managed-default-samclisourcebucket-xxx/logguardian-template.yaml"
  
  parameters = {
    Environment                    = "prod"
    ProductName                   = "TerraformCorp-LogGuardian"
    Owner                        = "Platform-Engineering"
    ManagedBy                    = "Terraform"
    CreateKMSKey                 = "false"
    ExistingKMSKeyArn           = aws_kms_key.logs.arn
    KMSKeyAlias                 = aws_kms_alias.logs.name
    CreateConfigService         = "false"
    ExistingConfigBucket        = aws_s3_bucket.config.bucket
    ExistingConfigServiceRoleArn = aws_iam_role.config.arn
    DefaultRetentionDays        = "365"
    LambdaMemorySize           = "512"
    CreateEventBridgeRules     = "true"
    EncryptionScheduleExpression = "cron(0 2 ? * SUN *)"
    RetentionScheduleExpression  = "cron(0 3 ? * SUN *)"
  }
  
  capabilities = ["CAPABILITY_NAMED_IAM"]
  
  tags = {
    Environment = "prod"
    ManagedBy   = "Terraform"
  }
}
```

### With Existing Resources
```hcl
# Existing KMS key for logs
resource "aws_kms_key" "logs" {
  description             = "Enterprise logs encryption key"
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_alias" "logs" {
  name          = "alias/enterprise-logs"
  target_key_id = aws_kms_key.logs.key_id
}

# Existing Config bucket
resource "aws_s3_bucket" "config" {
  bucket = "enterprise-config-${data.aws_caller_identity.current.account_id}"
}

# LogGuardian deployment
resource "aws_cloudformation_stack" "logguardian" {
  name = "logguardian"
  
  template_url = "https://s3.amazonaws.com/aws-sam-cli-managed-default-samclisourcebucket-xxx/logguardian-template.yaml"
  
  parameters = {
    Environment                    = "prod"
    CreateKMSKey                  = "false"
    ExistingKMSKeyArn            = aws_kms_key.logs.arn
    CreateConfigService          = "false"
    ExistingConfigBucket         = aws_s3_bucket.config.bucket
    DefaultRetentionDays         = "365"
  }
  
  capabilities = ["CAPABILITY_NAMED_IAM"]
  
  depends_on = [
    aws_kms_key.logs,
    aws_s3_bucket.config
  ]
}

data "aws_caller_identity" "current" {}
```
