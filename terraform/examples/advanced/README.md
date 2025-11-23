# Advanced Example

This example demonstrates a full-featured LogGuardian deployment with custom configurations suitable for production environments.

## What This Creates

- Lambda function with increased memory and timeout
- KMS key with custom alias
- AWS Config rules for compliance monitoring
- Daily EventBridge scheduled checks (instead of weekly)
- CloudWatch dashboard and alarms
- SNS topic with email subscription for alerts
- Multi-region support configuration

## Features Demonstrated

- **Custom Schedules**: Daily compliance checks instead of weekly
- **SNS Notifications**: Email alerts for compliance violations
- **Multi-Region**: Configure supported regions for log group management
- **Extended Retention**: 90-day retention policy enforcement
- **Custom Tags**: Cost allocation and organizational tags
- **Additional Environment Variables**: Feature flags and custom settings

## Prerequisites

1. An S3 bucket containing the LogGuardian Lambda deployment package
2. AWS credentials configured
3. Valid email address for SNS subscription confirmation

## Usage

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply (you'll be prompted for variables or can use a tfvars file)
terraform apply
```

## Post-Deployment

1. Check your email and confirm the SNS subscription
2. Access the CloudWatch dashboard via the output URL
3. Verify the Lambda function is deployed correctly

## Customization

Key areas to customize for your environment:

- `environment`: Set to your environment name
- `lambda_s3_bucket`: Your deployment artifact bucket
- `supported_regions`: Regions where you have CloudWatch Log Groups
- `alarm_sns_topic_arn`: Your existing SNS topic (or use the one created)
- `additional_tags`: Your organization's tagging requirements
