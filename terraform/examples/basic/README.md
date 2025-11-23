# Basic Example

This example demonstrates a minimal LogGuardian deployment with default settings.

## What This Creates

- Lambda function for compliance automation
- KMS key for CloudWatch Logs encryption
- AWS Config rules for encryption and retention compliance
- EventBridge rules for weekly scheduled checks
- CloudWatch dashboard and alarms

## Prerequisites

1. An S3 bucket containing the LogGuardian Lambda deployment package
2. AWS credentials configured

## Usage

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan -var="lambda_s3_bucket=your-bucket-name"

# Apply
terraform apply -var="lambda_s3_bucket=your-bucket-name"
```

## Customization

Edit `main.tf` to customize:
- `environment`: Change from "dev" to "staging" or "prod"
- `lambda_s3_bucket`: Your S3 bucket with the Lambda package
- `owner`: Your team name for resource tagging
