# ============================================
# LogGuardian - CloudWatch Logs Compliance Automation
# ============================================
#
# This Terraform module deploys LogGuardian, an automated compliance solution
# for CloudWatch Logs that ensures encryption and retention policies are applied
# consistently across your AWS environment.
#
# Key Features:
# - Automated encryption with KMS
# - Retention policy enforcement
# - AWS Config integration for compliance monitoring
# - EventBridge scheduled checks
# - CloudWatch dashboard and alarms
#
# Usage:
#   module "logguardian" {
#     source = "./terraform"
#
#     environment      = "prod"
#     lambda_s3_bucket = "my-deployment-bucket"
#     lambda_s3_key    = "logguardian-compliance.zip"
#   }
#
# ============================================

# This file serves as the main orchestration point.
# Actual resources are organized in separate files for clarity:
#
# - versions.tf    : Terraform and provider version constraints
# - locals.tf      : Local variables and data sources (DRY principle)
# - variables.tf   : Input variable definitions with validation
# - iam.tf         : IAM roles and policies (single source of truth)
# - kms.tf         : KMS key for CloudWatch Logs encryption
# - lambda.tf      : Lambda function
# - config.tf      : AWS Config service and compliance rules
# - eventbridge.tf : EventBridge scheduled triggers
# - monitoring.tf  : CloudWatch dashboard and alarms
# - outputs.tf     : Output values for integration
#
# ============================================

# Validate required configurations
# ============================================

resource "null_resource" "validate_config" {
  lifecycle {
    precondition {
      condition     = var.create_kms_key || var.existing_kms_key_arn != null
      error_message = "Either create_kms_key must be true or existing_kms_key_arn must be provided"
    }

    precondition {
      condition     = var.create_config_service || (var.existing_config_bucket != null && var.existing_config_service_role_arn != null)
      error_message = "Either create_config_service must be true or existing Config service resources must be provided"
    }

    precondition {
      condition     = var.create_config_rules || (var.existing_encryption_config_rule != null && var.existing_retention_config_rule != null)
      error_message = "Either create_config_rules must be true or existing Config rule names must be provided"
    }
  }
}