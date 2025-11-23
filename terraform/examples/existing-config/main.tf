# LogGuardian - Existing AWS Config Example
#
# This example demonstrates deploying LogGuardian in an environment
# where AWS Config is already enabled. This is common in enterprise
# environments with centralized Config management.

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

# Data sources to reference existing infrastructure
data "aws_s3_bucket" "config" {
  bucket = "my-org-config-bucket"
}

data "aws_iam_role" "config" {
  name = "AWSConfigRole"
}

data "aws_kms_key" "logs" {
  key_id = "alias/cloudwatch-logs-key"
}

module "logguardian" {
  source = "../../"

  # Required variables
  environment      = "prod"
  lambda_s3_bucket = "my-deployment-bucket"
  lambda_s3_key    = "logguardian-compliance.zip"

  # Use existing KMS key (managed by security team)
  create_kms_key       = false
  existing_kms_key_arn = data.aws_kms_key.logs.arn

  # Use existing AWS Config service (centrally managed)
  create_config_service            = false
  existing_config_bucket           = data.aws_s3_bucket.config.id
  existing_config_service_role_arn = data.aws_iam_role.config.arn

  # Create LogGuardian-specific Config rules
  create_config_rules = true

  # EventBridge schedules
  create_eventbridge_rules       = true
  encryption_schedule_expression = "cron(0 3 ? * SUN *)"
  retention_schedule_expression  = "cron(0 4 ? * SUN *)"

  # Monitoring
  create_monitoring_dashboard = true
  enable_cloudwatch_alarms    = true

  # Tags to match existing infrastructure
  additional_tags = {
    ManagedBy   = "Platform-Team"
    ConfigSetup = "centralized"
  }
}

# Outputs
output "lambda_function_name" {
  description = "Name of the LogGuardian Lambda function"
  value       = module.logguardian.lambda_function_name
}

output "encryption_config_rule_name" {
  description = "Name of the encryption Config rule"
  value       = module.logguardian.encryption_config_rule_name
}

output "retention_config_rule_name" {
  description = "Name of the retention Config rule"
  value       = module.logguardian.retention_config_rule_name
}

output "dashboard_url" {
  description = "URL to the CloudWatch dashboard"
  value       = module.logguardian.dashboard_url
}
