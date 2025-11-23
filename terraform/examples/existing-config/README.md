# Existing AWS Config Example

This example demonstrates deploying LogGuardian in an environment where AWS Config and KMS keys are already managed centrally. This is common in enterprise environments.

## Use Case

Many organizations have:
- Centralized AWS Config setup managed by a security or platform team
- Shared KMS keys for CloudWatch Logs encryption
- Existing S3 buckets for Config snapshots

This example shows how to integrate LogGuardian with existing infrastructure.

## What This Creates

- Lambda function for compliance automation
- AWS Config rules (using existing Config service)
- EventBridge rules for scheduled checks
- CloudWatch dashboard and alarms

## What This Does NOT Create

- KMS key (uses existing)
- AWS Config recorder/delivery channel (uses existing)
- Config S3 bucket (uses existing)

## Prerequisites

1. Existing AWS Config service enabled in the account
2. Existing KMS key for CloudWatch Logs encryption
3. Existing S3 bucket for Config snapshots
4. IAM role for Config service
5. Lambda deployment package in S3

## Usage

1. Update the data sources in `main.tf` to reference your existing resources:

```hcl
data "aws_s3_bucket" "config" {
  bucket = "your-config-bucket-name"
}

data "aws_iam_role" "config" {
  name = "YourConfigRoleName"
}

data "aws_kms_key" "logs" {
  key_id = "alias/your-kms-key-alias"
}
```

2. Deploy:

```bash
terraform init
terraform plan
terraform apply
```

## Important Notes

- Ensure the existing KMS key policy allows the LogGuardian Lambda role to use it
- The existing Config service role needs appropriate permissions
- Config rules created by LogGuardian will appear alongside any existing rules

## Permissions Required

The existing KMS key policy should include:

```json
{
  "Sid": "Allow LogGuardian Lambda",
  "Effect": "Allow",
  "Principal": {
    "AWS": "arn:aws:iam::ACCOUNT_ID:role/logguardian-prod-lambda-role"
  },
  "Action": [
    "kms:DescribeKey",
    "kms:GetKeyPolicy",
    "kms:Decrypt"
  ],
  "Resource": "*"
}
```
