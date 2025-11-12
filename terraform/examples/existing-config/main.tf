# LogGuardian deployment using existing AWS infrastructure

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

  # KMS - Use existing key
  create_kms_key       = false
  existing_kms_key_arn = var.existing_kms_key_arn
  kms_key_alias        = var.kms_key_alias

  # Config - Use existing service
  create_config_service    = false
  existing_config_bucket   = var.existing_config_bucket
  existing_config_role_arn = var.existing_config_role_arn

  # Config Rules - Use existing rules
  create_encryption_config_rule   = false
  existing_encryption_config_rule = var.existing_encryption_config_rule
  create_retention_config_rule    = false
  existing_retention_config_rule  = var.existing_retention_config_rule

  # Lambda Configuration
  default_retention_days    = var.default_retention_days
  lambda_log_retention_days = 30
  lambda_memory_size        = 128
  lambda_timeout            = 60
  log_level                 = var.log_level

  # EventBridge - Enable scheduled automation
  create_eventbridge_rules       = true
  encryption_schedule_expression = var.encryption_schedule
  retention_schedule_expression  = var.retention_schedule

  # Monitoring - Enable dashboard
  create_monitoring_dashboard = true

  # Tagging
  product_name = "LogGuardian"
  owner        = var.owner
  tags         = var.additional_tags
}
