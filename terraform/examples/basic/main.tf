# Basic LogGuardian deployment - creates all new resources

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
  region = var.aws_region
}

module "logguardian" {
  source = "../../"

  # Environment
  environment = var.environment
  aws_region  = var.aws_region

  # KMS - Create new key
  create_kms_key = true
  kms_key_alias  = "alias/logguardian-${var.environment}"

  # Config - Create new service
  create_config_service          = true
  create_encryption_config_rule  = true
  create_retention_config_rule   = true

  # Lambda Configuration
  default_retention_days    = var.default_retention_days
  lambda_log_retention_days = 30
  lambda_memory_size        = 128
  lambda_timeout            = 60
  log_level                 = var.log_level
  dry_run                   = var.dry_run

  # EventBridge - Enable scheduled automation
  create_eventbridge_rules       = true
  encryption_schedule_expression = var.encryption_schedule
  retention_schedule_expression  = var.retention_schedule

  # Monitoring - Enable dashboard
  create_monitoring_dashboard = true

  # S3 Lifecycle
  enable_s3_lifecycle_rules = true
  s3_expiration_days        = 90

  # Tagging
  product_name = "LogGuardian"
  owner        = var.owner
  tags         = var.additional_tags
}
