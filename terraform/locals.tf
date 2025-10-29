locals {
  # Determine Config rule names (created or existing)
  encryption_config_rule = var.create_encryption_config_rule ? module.config.encryption_config_rule_name : var.existing_encryption_config_rule
  retention_config_rule  = var.create_retention_config_rule ? module.config.retention_config_rule_name : var.existing_retention_config_rule

  # Common tags applied to all resources
  common_tags = merge(
    var.tags,
    {
      Project     = var.product_name
      Environment = var.environment
      ManagedBy   = var.managed_by
      Terraform   = "true"
    }
  )
}
