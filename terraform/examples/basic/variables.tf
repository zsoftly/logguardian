variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ca-central-1"
}

variable "default_retention_days" {
  description = "Default retention period for log groups"
  type        = number
  default     = 30
}

variable "log_level" {
  description = "Lambda log level"
  type        = string
  default     = "INFO"
}

variable "dry_run" {
  description = "Enable dry-run mode"
  type        = bool
  default     = false
}

variable "encryption_schedule" {
  description = "Encryption check schedule"
  type        = string
  default     = "cron(0 3 ? * SUN *)"
}

variable "retention_schedule" {
  description = "Retention check schedule"
  type        = string
  default     = "cron(0 4 ? * SUN *)"
}

variable "owner" {
  description = "Owner/Team"
  type        = string
  default     = "DevOps"
}

variable "additional_tags" {
  description = "Additional resource tags"
  type        = map(string)
  default     = {}
}
