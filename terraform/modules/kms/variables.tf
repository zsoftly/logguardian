variable "environment" {
  description = "Environment name (dev, staging, prod, test)"
  type        = string

  validation {
    condition     = can(regex("^(dev|staging|prod|test)$", var.environment))
    error_message = "Environment must be one of: dev, staging, prod, test."
  }
}

variable "create_kms_key" {
  description = "Whether to create a new KMS key or use an existing one"
  type        = bool
  default     = true
}

variable "existing_kms_key_arn" {
  description = "ARN of existing KMS key (required if create_kms_key=false)"
  type        = string
  default     = null

  validation {
    condition     = var.existing_kms_key_arn == null || can(regex("^arn:aws:kms:[a-z0-9-]+:[0-9]{12}:key/[a-f0-9-]+$", var.existing_kms_key_arn))
    error_message = "KMS key ARN must be valid format."
  }
}

variable "kms_key_alias" {
  description = "Alias for the KMS key (must start with 'alias/')"
  type        = string
  default     = "alias/logguardian-cloudwatch-logs"

  validation {
    condition     = can(regex("^alias/[a-zA-Z0-9:/_-]+$", var.kms_key_alias))
    error_message = "KMS key alias must start with 'alias/' and contain only alphanumeric characters, '/', '_', or '-'."
  }
}

variable "enable_key_rotation" {
  description = "Enable automatic key rotation (recommended)"
  type        = bool
  default     = true
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
