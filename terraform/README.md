# LogGuardian Terraform Module

Terraform module for deploying LogGuardian - an automated compliance solution for CloudWatch Logs that ensures encryption and retention policies are applied consistently across your AWS environment.

## Features

- **Automated Encryption**: Enforce KMS encryption on CloudWatch Log Groups
- **Retention Policy Enforcement**: Ensure all log groups have appropriate retention policies
- **AWS Config Integration**: Continuous compliance monitoring with AWS Config rules
- **EventBridge Scheduling**: Automated scheduled compliance checks
- **CloudWatch Monitoring**: Dashboard and alarms for operational visibility
- **SNS Notifications**: Alert on compliance violations for manual review

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.5.0 |
| aws | ~> 5.0 |

## Usage

### Basic Usage

```hcl
module "logguardian" {
  source = "./terraform"

  environment      = "prod"
  lambda_s3_bucket = "my-deployment-bucket"
  lambda_s3_key    = "logguardian-compliance.zip"
}
```

### With Existing KMS Key

```hcl
module "logguardian" {
  source = "./terraform"

  environment      = "prod"
  lambda_s3_bucket = "my-deployment-bucket"

  # Use existing KMS key
  create_kms_key       = false
  existing_kms_key_arn = "arn:aws:kms:ca-central-1:123456789012:key/12345678-1234-1234-1234-123456789012"
}
```

### With Existing AWS Config Service

```hcl
module "logguardian" {
  source = "./terraform"

  environment      = "prod"
  lambda_s3_bucket = "my-deployment-bucket"

  # Use existing Config service
  create_config_service            = false
  existing_config_bucket           = "my-config-bucket"
  existing_config_service_role_arn = "arn:aws:iam::123456789012:role/ConfigServiceRole"
}
```

### With SNS Notifications

```hcl
module "logguardian" {
  source = "./terraform"

  environment      = "prod"
  lambda_s3_bucket = "my-deployment-bucket"

  # Enable notifications
  alarm_sns_topic_arn = "arn:aws:sns:ca-central-1:123456789012:alerts"
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Deployment environment (dev, staging, prod, sandbox) | `string` | n/a | yes |
| lambda_s3_bucket | S3 bucket containing the Lambda deployment package | `string` | n/a | yes |
| lambda_s3_key | S3 key for the Lambda deployment package | `string` | `"logguardian-compliance.zip"` | no |
| product_name | Product name for resource tagging and naming | `string` | `"LogGuardian"` | no |
| owner | Owner/Team responsible for the resources | `string` | `"Platform-Engineering"` | no |
| lambda_memory_size | Lambda memory allocation in MB | `number` | `128` | no |
| lambda_timeout | Lambda timeout in seconds | `number` | `300` | no |
| lambda_log_level | Lambda logging level (ERROR, WARN, INFO, DEBUG) | `string` | `"INFO"` | no |
| lambda_log_retention_days | Retention period for Lambda's own logs | `number` | `7` | no |
| default_retention_days | Default retention period for managed log groups | `number` | `30` | no |
| batch_size | Number of log groups to process in parallel | `number` | `10` | no |
| create_kms_key | Create a new KMS key for encryption | `bool` | `true` | no |
| existing_kms_key_arn | ARN of existing KMS key (if create_kms_key = false) | `string` | `null` | no |
| kms_key_alias | Alias for the KMS key | `string` | `null` | no |
| kms_deletion_window_days | KMS key deletion window in days | `number` | `30` | no |
| create_config_service | Create AWS Config service resources | `bool` | `false` | no |
| existing_config_bucket | Existing S3 bucket for Config snapshots | `string` | `null` | no |
| existing_config_service_role_arn | Existing IAM role ARN for Config service | `string` | `null` | no |
| config_bucket_expiration_days | S3 lifecycle expiration for Config snapshots | `number` | `90` | no |
| create_config_rules | Create AWS Config rules for compliance | `bool` | `true` | no |
| existing_encryption_config_rule | Name of existing encryption Config rule | `string` | `null` | no |
| existing_retention_config_rule | Name of existing retention Config rule | `string` | `null` | no |
| create_eventbridge_rules | Create EventBridge scheduled rules | `bool` | `true` | no |
| encryption_schedule_expression | Schedule for encryption compliance checks | `string` | `"cron(0 2 ? * SUN *)"` | no |
| retention_schedule_expression | Schedule for retention compliance checks | `string` | `"cron(0 3 ? * SUN *)"` | no |
| create_monitoring_dashboard | Create CloudWatch dashboard | `bool` | `true` | no |
| enable_cloudwatch_alarms | Enable CloudWatch alarms | `bool` | `true` | no |
| alarm_sns_topic_arn | SNS topic ARN for alarms and notifications | `string` | `null` | no |
| lambda_code_signing_config_arn | Lambda code signing config ARN (security best practice) | `string` | `null` | no |
| supported_regions | List of AWS regions to manage | `list(string)` | `[]` | no |
| additional_lambda_env_vars | Additional Lambda environment variables | `map(string)` | `{}` | no |
| additional_tags | Additional tags for all resources | `map(string)` | `{}` | no |
| config_recorder_resource_types | Resource types to record in AWS Config | `list(string)` | `["AWS::Logs::LogGroup"]` | no |

## Outputs

| Name | Description |
|------|-------------|
| lambda_function_name | Name of the LogGuardian Lambda function |
| lambda_function_arn | ARN of the LogGuardian Lambda function |
| lambda_role_arn | ARN of the Lambda execution role |
| lambda_log_group_name | CloudWatch Log Group for Lambda logs |
| kms_key_id | ID of the KMS key for encryption |
| kms_key_alias | Alias of the KMS key |
| config_bucket_name | S3 bucket for AWS Config snapshots |
| config_role_arn | IAM role ARN for AWS Config |
| encryption_config_rule_name | Name of the encryption Config rule |
| retention_config_rule_name | Name of the retention Config rule |
| encryption_schedule_rule_name | Name of the encryption EventBridge rule |
| retention_schedule_rule_name | Name of the retention EventBridge rule |
| dashboard_name | Name of the CloudWatch dashboard |
| dashboard_url | URL to the CloudWatch dashboard |
| manual_invocation_command | AWS CLI command to manually invoke LogGuardian |
| test_invocation_payload | Sample payload for testing |
| deployment_summary | Summary of the deployment configuration |

## Architecture

```text
┌─────────────────────────────────────────────────────────────────┐
│                        EventBridge                               │
│  ┌──────────────────┐  ┌──────────────────┐                     │
│  │ Encryption Check │  │ Retention Check  │                     │
│  │ (Weekly Sunday)  │  │ (Weekly Sunday)  │                     │
│  └────────┬─────────┘  └────────┬─────────┘                     │
└───────────┼─────────────────────┼───────────────────────────────┘
            │                     │
            ▼                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Lambda Function                               │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  LogGuardian Compliance Automation                         │ │
│  │  - Process Config rule evaluations                         │ │
│  │  - Apply KMS encryption to log groups                      │ │
│  │  - Set retention policies on log groups                    │ │
│  └────────────────────────────────────────────────────────────┘ │
└──────────────────────────┬──────────────────────────────────────┘
                           │
            ┌──────────────┼──────────────┐
            ▼              ▼              ▼
     ┌──────────┐   ┌──────────┐   ┌──────────┐
     │   KMS    │   │CloudWatch│   │  Config  │
     │   Key    │   │   Logs   │   │  Rules   │
     └──────────┘   └──────────┘   └──────────┘
```

## Deployment

### Using the Deploy Script

```bash
cd terraform/scripts
./deploy.sh prod my-lambda-bucket
```

### Manual Deployment

```bash
cd terraform

# Initialize
terraform init

# Plan
terraform plan -var="environment=prod" -var="lambda_s3_bucket=my-bucket"

# Apply
terraform apply -var="environment=prod" -var="lambda_s3_bucket=my-bucket"
```

## Cleanup

```bash
cd terraform/scripts
./destroy.sh
```

## Security Considerations

- **KMS Key Rotation**: Enabled by default (365-day rotation)
- **Least Privilege IAM**: Lambda role has minimal required permissions
- **S3 Bucket Security**: Config bucket has versioning, encryption, and public access blocked
- **Manual Remediation**: Config remediation requires manual approval via SNS

## License

Copyright (c) zsoftly. All rights reserved.
