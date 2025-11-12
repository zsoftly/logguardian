variable "environment" {
  description = "Environment name (dev, staging, prod, test)"
  type        = string

  validation {
    condition     = can(regex("^(dev|staging|prod|test)$", var.environment))
    error_message = "Environment must be one of: dev, staging, prod, test."
  }
}

variable "create_config_bucket" {
  description = "Whether to create S3 bucket for AWS Config"
  type        = bool
  default     = false
}

variable "enable_lifecycle_rules" {
  description = "Enable S3 lifecycle rules for cost optimization"
  type        = bool
  default     = true
}

variable "s3_expiration_days" {
  description = "Days after which Config data is permanently deleted"
  type        = number
  default     = 90

  validation {
    condition     = var.s3_expiration_days >= 1 && var.s3_expiration_days <= 3653
    error_message = "Expiration days must be between 1 and 3653 (10 years)."
  }
}

variable "product_name" {
  description = "Product name for resource tagging"
  type        = string
  default     = "LogGuardian"
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}
