variable "environment" {
  description = "Environment name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "sam_created_kms" {
  description = "Did SAM create the KMS key?"
  type        = bool
  default     = true
}

variable "sam_created_config" {
  description = "Did SAM create Config service?"
  type        = bool
  default     = false
}

variable "sam_created_encryption_rule" {
  description = "Did SAM create encryption Config rule?"
  type        = bool
  default     = true
}

variable "sam_created_retention_rule" {
  description = "Did SAM create retention Config rule?"
  type        = bool
  default     = true
}

variable "sam_created_eventbridge" {
  description = "Did SAM create EventBridge rules?"
  type        = bool
  default     = true
}

variable "sam_created_dashboard" {
  description = "Did SAM create CloudWatch dashboard?"
  type        = bool
  default     = true
}

variable "existing_kms_key_arn" {
  description = "Existing KMS key ARN (if SAM didn't create one)"
  type        = string
  default     = null
}

variable "default_retention_days" {
  description = "Default retention days"
  type        = number
  default     = 90
}

variable "owner" {
  description = "Owner/Team"
  type        = string
  default     = "DevOps"
}
