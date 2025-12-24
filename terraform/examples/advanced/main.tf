# LogGuardian - Advanced Example
#
# This example demonstrates a full-featured deployment of LogGuardian
# with custom schedules, SNS notifications, and additional configuration.

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ca-central-1"
}

# SNS Topic for alerts (encrypted at rest - security best practice)
resource "aws_sns_topic" "alerts" {
  name              = "logguardian-alerts"
  kms_master_key_id = "alias/aws/sns" # AWS-managed key
}

resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = "platform-team@example.com"
}

module "logguardian" {
  source = "../../"

  # Required variables
  environment      = "prod"
  lambda_s3_bucket = "my-deployment-bucket"
  lambda_s3_key    = "logguardian-compliance.zip"

  # Naming and ownership
  product_name = "LogGuardian"
  owner        = "Platform-Engineering"

  # Lambda configuration
  lambda_memory_size        = 256
  lambda_timeout            = 600
  lambda_log_level          = "INFO"
  lambda_log_retention_days = 30

  # Compliance settings
  default_retention_days = 90
  batch_size             = 20

  # KMS configuration
  create_kms_key           = true
  kms_key_alias            = "logguardian-prod-logs"
  kms_deletion_window_days = 30

  # Config rules
  create_config_rules = true

  # EventBridge schedules - daily checks at different times
  create_eventbridge_rules       = true
  encryption_schedule_expression = "cron(0 1 * * ? *)" # Daily at 1 AM UTC
  retention_schedule_expression  = "cron(0 2 * * ? *)" # Daily at 2 AM UTC

  # Monitoring and alerting
  create_monitoring_dashboard = true
  enable_cloudwatch_alarms    = true
  alarm_sns_topic_arn         = aws_sns_topic.alerts.arn

  # Multi-region support (optional)
  supported_regions = [
    "ca-central-1",
    "us-east-1",
    "us-west-2"
  ]

  # Additional Lambda environment variables
  additional_lambda_env_vars = {
    FEATURE_FLAG_STRICT_MODE = "true"
    MAX_RETRY_ATTEMPTS       = "3"
  }

  # Additional tags for cost allocation and organization
  additional_tags = {
    CostCenter  = "platform-infrastructure"
    Project     = "compliance-automation"
    Compliance  = "required"
    BackupPolicy = "none"
  }
}

# Outputs
output "lambda_function_arn" {
  description = "ARN of the LogGuardian Lambda function"
  value       = module.logguardian.lambda_function_arn
}

output "kms_key_id" {
  description = "KMS key ID for CloudWatch Logs encryption"
  value       = module.logguardian.kms_key_id
}

output "dashboard_url" {
  description = "URL to the CloudWatch dashboard"
  value       = module.logguardian.dashboard_url
}

output "deployment_summary" {
  description = "Summary of the deployment"
  value       = module.logguardian.deployment_summary
}

output "sns_topic_arn" {
  description = "SNS topic for alerts"
  value       = aws_sns_topic.alerts.arn
}
