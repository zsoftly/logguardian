variable "environment" {
  description = "Environment name (dev, staging, prod, test)"
  type        = string

  validation {
    condition     = can(regex("^(dev|staging|prod|test)$", var.environment))
    error_message = "Environment must be one of: dev, staging, prod, test."
  }
}

variable "kms_key_arn" {
  description = "ARN of KMS key for CloudWatch Logs encryption"
  type        = string

  validation {
    condition     = can(regex("^arn:aws:kms:[a-z0-9-]+:[0-9]{12}:key/[a-f0-9-]+$", var.kms_key_arn))
    error_message = "KMS key ARN must be valid format."
  }
}

variable "product_name" {
  description = "Product name for resource tagging"
  type        = string
  default     = "LogGuardian"
}

variable "owner" {
  description = "Owner/Team responsible for the resources"
  type        = string
  default     = "DevOps"
}

variable "managed_by" {
  description = "How these resources are managed"
  type        = string
  default     = "Terraform"
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}
